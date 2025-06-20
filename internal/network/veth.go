package network

import (
	"fmt"
	"os"

	"github.com/vishvananda/netlink"
)

type Veth struct {
	HostName   string
	PeerName   string
	BridgeName string
}

func NewVeth(hostName, peerName, bridgeName string) *Veth {
	return &Veth{
		HostName:   hostName,
		PeerName:   peerName,
		BridgeName: bridgeName,
	}
}

func (v *Veth) CreateAndAttach(pid int) error {
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: v.HostName,
		},
		PeerName: v.PeerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return fmt.Errorf("failed to add veth network:%w", err)
	}

	br, err := netlink.LinkByName(v.BridgeName)
	if err != nil {
		return fmt.Errorf("failed to get bridge network:%w", err)
	}

	hostLink, _ := netlink.LinkByName(v.HostName)
	if err := netlink.LinkSetMaster(hostLink, br); err != nil {
		return fmt.Errorf("failed to attach veth to bridge:%w", err)
	}

	if err := netlink.LinkSetUp(hostLink); err != nil {
		return fmt.Errorf("failed to bring up the host veth:%w", err)
	}

	peerLink, _ := netlink.LinkByName(v.PeerName)
	if err := netlink.LinkSetUp(peerLink); err != nil {
		return fmt.Errorf("failed to bring up the peer veth:%w", err)
	}

	err = v.movePeerToNamespace(pid)
	if err != nil {
		return err
	}

	return nil
}

func (v *Veth) movePeerToNamespace(pid int) error {
	netNsPath := fmt.Sprintf("/proc/%d/ns/net", pid)
	f, err := os.Open(netNsPath)
	if err != nil {
		return fmt.Errorf("failed to open namespace file:%w", err)
	}
	defer f.Close()

	peer, err := netlink.LinkByName(v.PeerName)
	if err != nil {
		return fmt.Errorf("failed to find peer veth network:%w", err)
	}

	if err := netlink.LinkSetNsFd(peer, int(f.Fd())); err != nil {
		return fmt.Errorf("failed to move peer to namespace:%w", err)
	}

	return nil
}

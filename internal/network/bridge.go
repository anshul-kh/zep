package network

import (
	"fmt"
	"os/exec"

	"github.com/anshul-kh/zep-core/internal/helper"
	"github.com/vishvananda/netlink"
)

type Bridge struct {
	GateWay string
	CIDR    string
	Name    string
}

const (
	BRIDGE_GATEWAY = "184.23.0.1"
	BRIDGE_CIDR    = "24"
	BRIDGE_NAME    = "zep0"
)

func NewBridge() *Bridge {
	return &Bridge{
		GateWay: BRIDGE_GATEWAY,
		CIDR:    BRIDGE_CIDR,
		Name:    BRIDGE_NAME,
	}
}

func (b *Bridge) SetUpBridge() error {

	extBr, err := netlink.LinkByName(b.Name)
	if err == nil {
		netlink.LinkSetUp(extBr)
		return nil
	}

	la := netlink.NewLinkAttrs()
	la.Name = b.Name
	br := &netlink.Bridge{
		LinkAttrs: la,
	}

	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("failed to add bridge network:%w", err)
	}

	addr, err := netlink.ParseAddr(fmt.Sprintf("%s/%s", b.GateWay, b.CIDR))
	if err != nil {
		return fmt.Errorf("failed to parse address for bridge network:%w", err)
	}

	if err := netlink.AddrAdd(br, addr); err != nil {
		return fmt.Errorf("failed to add bridge address:%w", err)
	}

	if err := netlink.LinkSetUp(br); err != nil {
		return fmt.Errorf("failed to bring up bridge network:%w", err)
	}

	if err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run(); err != nil {
		return fmt.Errorf("failed to enable ip_forward:%w", err)
	}

	netIf, err := helper.GetDefaultInterface()
	if err != nil {
		return err
	}

	err = b.setUpNAT(addr.IPNet.String(), netIf)
	if err != nil {
		return err
	}

	return nil
}

func (b *Bridge) StopBridge() error {
	br, err := netlink.LinkByName(b.Name)
	if err != nil {
		return fmt.Errorf("failed to find bridge network:%w", err)
	}

	err = netlink.LinkSetDown(br)
	if err != nil {
		return fmt.Errorf("failed to stop bridge network:%w", err)
	}

	return nil
}

func (b *Bridge) setUpNAT(bridgeSubnet, externalIf string) error {

	cmds := [][]string{
		// NAT -> internet
		{"-t", "nat", "-A", "POSTROUTING", "-s", bridgeSubnet, "-o", externalIf, "-j", "MASQUERADE"},
		// bridge -> external interface
		{"-A", "FORWARD", "-i", b.Name, "-o", externalIf, "-j", "ACCEPT"},
		// return back
		{"-A", "FORWARD", "-i", externalIf, "-o", b.Name, "-m", "state", "--state", "RELATED,ESTABLISHED", "-j", "ACCEPT"},
	}

	for _, args := range cmds {
		cmd := exec.Command("iptables", args...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("iptables %v failed:%w", args, err)
		}
	}

	return nil
}

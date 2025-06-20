package network

import (
	"fmt"
	"os/exec"

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

	_, err := netlink.LinkByName(b.Name)
	if err == nil {
		fmt.Printf("bridge network is already running. :%v", err)
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

	// TODO: setup NAT

	return nil
}

// func to setup nat

package helper

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/vishvananda/netlink"
)

func GetAbsPath(relPath string) (string, error) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute file path:%w", err)
	}
	return absPath, nil
}

func GetDefaultInterface() (string, error) {
	dst := &net.IPNet{
		IP:   net.ParseIP("8.8.8.8"),
		Mask: net.CIDRMask(32, 32),
	}

	routes, err := netlink.RouteGet(dst.IP)
	if err != nil {
		return "", fmt.Errorf("failed to get route:%w", err)
	}

	if len(routes) <= 0 {
		return "", fmt.Errorf("no route found:%w", err)
	}

	link, err := netlink.LinkByIndex(routes[0].LinkIndex)
	if err != nil {
		return "", fmt.Errorf("failed to get link:%w", err)
	}

	return link.Attrs().Name, nil
}

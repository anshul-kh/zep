package network

import (
	"fmt"
	"net"
)

/**
* * ip-allocator manages ip related function for the processes
* * it maps , assigns , and frees ip for each process
 */

type IPAllocator struct {
	network  *net.IPNet
	current  net.IP
	freedIPs []string
	IPStore  map[int]string
}

func NewIPAllcator(cidr string) (*IPAllocator, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("failed to create ip allocator:%w", err)
	}

	ip = ip.Mask(ipnet.Mask)
	ip[3] += 2

	return &IPAllocator{
		network:  ipnet,
		current:  ip,
		freedIPs: []string{},
		IPStore:  make(map[int]string, 10),
	}, nil
}

func (a *IPAllocator) NextIP() (string, error) {

	freedIP, err := a.GetFreedIP()
	if err == nil {
		return freedIP, err
	}

	ip := make(net.IP, len(a.current))
	copy(ip, a.current)

	for i := len(a.current); i >= 0; i-- {
		a.current[i]++

		if a.current[i] != 0 {
			break
		}
	}

	if !a.network.Contains(ip) {
		return "", fmt.Errorf("ip range exhausted")
	}

	return ip.String(), nil
}

func (a *IPAllocator) FreeIP(pid int) error {
	ip, ok := a.IPStore[pid]
	if !ok {
		return fmt.Errorf("ip does not exist")
	}

	parsedIp := net.ParseIP(ip)

	if !a.network.Contains(parsedIp) {
		return fmt.Errorf("ip does not lies in current network")
	}
	a.freedIPs = append(a.freedIPs, ip)

	delete(a.IPStore, pid)

	return nil
}

func (a *IPAllocator) GetFreedIP() (string, error) {
	if len(a.freedIPs) <= 0 {
		return "", fmt.Errorf("no freed up IP available")
	}

	ip := a.freedIPs[0]

	a.freedIPs = a.freedIPs[1:]

	return ip, nil
}

func (a *IPAllocator) AppendIPStore(pid int, ip string) {
	a.IPStore[pid] = ip
}

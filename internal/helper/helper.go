package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
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

func DecodeJSONFromResponse(resp *http.Response, target interface{}) error {

	if resp == nil {
		return fmt.Errorf("response is null")
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Error : %s", string(body))
	}

	decoder := json.NewDecoder(resp.Body)
	fmt.Print(resp.Body)
	err := decoder.Decode(&target)
	if err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}

	return nil
}

func JSONToBuffer(v interface{}) (bytes.Buffer, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return bytes.Buffer{}, err
	}

	return *bytes.NewBuffer(data), nil
}

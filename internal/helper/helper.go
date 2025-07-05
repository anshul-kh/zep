package helper

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/vishvananda/netlink"
)

// * type for response from master process
type MasterResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// * type for stats data from response stats command from zep (cli)
type ProcessStats struct {
	Name string      `json:"name"`
	CPU  float64     `json:"cpu"`
	Mem  MemoryStats `json:"mem"`
}

// * type for usage of memory data
type MemoryStats struct {
	Data   float64 `json:"data"`
	HWM    float64 `json:"hwm"`
	Locked float64 `json:"locked"`
	RSS    float64 `json:"rss"`
	Stack  float64 `json:"stack"`
	Swap   float64 `json:"swap"`
	VMS    float64 `json:"vms"`
}

// * type for stats response after running stats command from zep (cli)
type StatsResponse struct {
	Success bool           `json:"success"`
	Msg     string         `json:"message"`
	Data    []ProcessStats `json:"data,omitempty"`
}

// * type for list command reponse
type ProcListResponse struct {
	Data map[int]string `json:"data"`
}

// helper to get absolute path
func GetAbsPath(relPath string) (string, error) {
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute file path:%w", err)
	}
	return absPath, nil
}

// helper to get defualt interface (eg: wlan or eth)
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

// coverts json to buffer
func JSONToBuffer(v interface{}) (bytes.Buffer, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return bytes.Buffer{}, err
	}

	return *bytes.NewBuffer(data), nil
}

/** =============================================
* respnse handlers for response of master process
* ===============================================*/

func DecodeJSONFromResponse(resp *http.Response, target interface{}) error {

	if resp == nil {
		return fmt.Errorf("response is null")
	}

	defer resp.Body.Close()

	// if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	// 	body, _ := io.ReadAll(resp.Body)
	// 	return fmt.Errorf("Error : %s", string(body))
	// }

	decoder := json.NewDecoder(resp.Body)
	err := decoder.Decode(&target)
	if err != nil {
		return fmt.Errorf("failed to decode json: %w", err)
	}

	return nil
}

func HandleFinalResponse(resp *http.Response) {
	var response MasterResponse
	err := DecodeJSONFromResponse(resp, &response)
	if err != nil {
		log.Printf("error from daemon:%v", err.Error())
		return
	}

	fmt.Print(response.Msg)
}

func HandleFinalDataResponse(resp *http.Response) {
	var response MasterResponse
	err := DecodeJSONFromResponse(resp, &response)
	if err != nil {
		log.Printf("error from daemon:%v", err.Error())
		return
	}

	fmt.Print(response.Data)
}

func HandleStatsResponse(resp *http.Response) {
	var response StatsResponse
	err := DecodeJSONFromResponse(resp, &response)
	if err != nil {
		log.Printf("error from daemon:%v", err.Error())
		return
	}

	fmt.Printf("%-12s %-8s %-10s %-10s %-10s\n",
		"NAME", "CPU%", "RSS(MB)", "VMS(MB)", "SWAP(MB)")
	fmt.Println(strings.Repeat("-", 54))

	for _, proc := range response.Data {
		fmt.Printf("%-12s %-8.2f %-10.2f %-10.2f %-10.2f\n",
			proc.Name,
			proc.CPU*100,
			proc.Mem.RSS/(1024*1024),
			proc.Mem.VMS/(1024*1024),
			proc.Mem.Swap/(1024*1024),
		)
	}
}

func StreamResponseBody(resp *http.Response) {
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	if strings.TrimSpace(contentType) == "application/json" {
		HandleFinalResponse(resp)
		return
	}

	reader := bufio.NewReader(resp.Body)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("Stream closed by server.")
			} else {
				fmt.Println("Error reading:", err)
			}
			break
		}

		fmt.Print(line)
	}
}

func HandleProcListResponse(resp *http.Response) {
	var response MasterResponse

	err := DecodeJSONFromResponse(resp, &response)
	if err != nil {
		log.Printf("error from daemon: %v", err.Error())
		return
	}

	rawMap, ok := response.Data.(map[string]interface{})
	if !ok {
		fmt.Println("Invalid data format from daemon")
		return
	}

	procMap := make(map[int]string)
	for k, v := range rawMap {
		id, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		name, ok := v.(string)
		if !ok {
			continue
		}
		procMap[id] = name
	}

	fmt.Printf("%-8s %-20s\n", "ZID", "NAME")
	fmt.Println(strings.Repeat("-", 28))

	ids := make([]int, 0, len(procMap))
	for id := range procMap {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, id := range ids {
		name := procMap[id]
		fmt.Printf("%-8d %-20s\n", id, name)
	}
}

// * PrettyPrintJSON takes any object and prints it nicely formatted
func PrettyPrintJSON(data interface{}) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Println("Failed to pretty-print JSON:", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

/*
* * PrintAnyJSONFromResponse reads json from an HTTP response body
* * and prints it as JSON for debugging purposes.
 */
func PrintAnyJSONFromResponse(resp *http.Response) {
	if resp == nil {
		fmt.Println("Response is nil")
		return
	}

	defer resp.Body.Close()

	// read the body into raw bytes
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed reading response body: %v\n", err)
		return
	}

	if len(bodyBytes) == 0 {
		return
	}

	var parsed interface{}
	err = json.Unmarshal(bodyBytes, &parsed)
	if err != nil {
		log.Printf("failed decoding JSON: %v\nbody:\n%s\n", err, string(bodyBytes))
		return
	}

	PrettyPrintJSON(parsed)
}

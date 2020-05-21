package qemu

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/digitalocean/go-qemu/qmp"
)

type qomListRequest struct {
	Execute   string                  `json:"execute"`
	Arguments qomListRequestArguments `json:"arguments"`
}

type qomListRequestArguments struct {
	Path string `json:"path"`
}

type qomListResponse struct {
	Return []qomListReturn `json:"return"`
}

type qomListReturn struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func qmpQomList(qmpMonitor *qmp.SocketMonitor, path string) ([]qomListReturn, error) {
	request, _ := json.Marshal(qomListRequest{
		Execute: "qom-list",
		Arguments: qomListRequestArguments{
			Path: path,
		},
	})
	result, err := qmpMonitor.Run(request)
	if err != nil {
		return nil, err
	}
	var response qomListResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return nil, err
	}
	return response.Return, nil
}

type qomGetRequest struct {
	Execute   string                 `json:"execute"`
	Arguments qomGetRequestArguments `json:"arguments"`
}

type qomGetRequestArguments struct {
	Path     string `json:"path"`
	Property string `json:"property"`
}

type qomGetResponse struct {
	Return string `json:"return"`
}

func qmpQomGet(qmpMonitor *qmp.SocketMonitor, path string, property string) (string, error) {
	request, _ := json.Marshal(qomGetRequest{
		Execute: "qom-get",
		Arguments: qomGetRequestArguments{
			Path:     path,
			Property: property,
		},
	})
	result, err := qmpMonitor.Run(request)
	if err != nil {
		return "", err
	}
	var response qomGetResponse
	if err := json.Unmarshal(result, &response); err != nil {
		return "", err
	}
	return response.Return, nil
}

type netDevice struct {
	Path       string
	Name       string
	Type       string
	MacAddress string
}

func getNetDevices(qmpMonitor *qmp.SocketMonitor) ([]netDevice, error) {
	devices := []netDevice{}
	for _, parentPath := range []string{"/machine/peripheral", "/machine/peripheral-anon"} {
		listResponse, err := qmpQomList(qmpMonitor, parentPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get qmp qom list %v: %w", parentPath, err)
		}
		for _, p := range listResponse {
			if strings.HasPrefix(p.Type, "child<") {
				path := fmt.Sprintf("%s/%s", parentPath, p.Name)
				r, err := qmpQomList(qmpMonitor, path)
				if err != nil {
					return nil, fmt.Errorf("failed to get qmp qom list %v: %w", path, err)
				}
				isNetdev := false
				for _, d := range r {
					if d.Name == "netdev" {
						isNetdev = true
						break
					}
				}
				if isNetdev {
					device := netDevice{
						Path: path,
					}
					for _, d := range r {
						if d.Name != "type" && d.Name != "netdev" && d.Name != "mac" {
							continue
						}
						value, err := qmpQomGet(qmpMonitor, path, d.Name)
						if err != nil {
							return nil, fmt.Errorf("failed to get qmp qom property %v %v: %w", path, d.Name, err)
						}
						switch d.Name {
						case "type":
							device.Type = value
						case "netdev":
							device.Name = value
						case "mac":
							device.MacAddress = value
						}
					}
					devices = append(devices, device)
				}
			}
		}
	}
	return devices, nil
}

package registry

import (
	"bytes"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type OSInfo struct {
	Name    string
	Arch    string
	Version string
}

func GetOSMetadata() map[string]interface{} {
	var osInfo OSInfo

	switch runtime.GOOS {
	case "windows":
		osInfo = GetInfoForWindows()
	case "darwin":
		osInfo = GetInfo("-srm")
	case "linux":
		osInfo = GetInfo("-srio")
	case "freebsd":
		osInfo = GetInfo("-sri")
	case "openbsd":
		osInfo = GetInfo("-srm")
	case "netbsd":
		osInfo = GetInfo("-srm")
	default:
		osInfo = OSInfo{
			Name: runtime.GOOS,
			Arch: runtime.GOARCH,
		}
	}

	return map[string]interface{}{
		"type": osInfo.Name,
		"details": map[string]interface{}{
			"arch":    osInfo.Arch,
			"version": osInfo.Version,
		},
	}
}

func GetInfo(flags string) OSInfo {
	out, err := _uname(flags)
	tries := 0
	for strings.Contains(out, "broken pipe") && tries < 3 {
		out, err = _uname(flags)
		time.Sleep(500 * time.Millisecond)
		tries++
	}
	if strings.Contains(out, "broken pipe") || err != nil {
		out = ""
	}

	if err != nil {
		log.Printf("[ERROR] failed to get the OS info: %s", err)
	}
	core := _retrieveCore(out)
	return OSInfo{
		Name:    runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: core,
	}
}

func _uname(flags string) (string, error) {
	cmd := exec.Command("uname", flags)
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	return out.String(), err
}

func _retrieveCore(osStr string) string {
	osStr = strings.Replace(osStr, "\n", "", -1)
	osStr = strings.Replace(osStr, "\r\n", "", -1)
	osInfo := strings.Split(osStr, " ")

	var core string
	if len(osInfo) > 1 {
		core = osInfo[1]
	}
	return core
}

func GetInfoForWindows() OSInfo {
	cmd := exec.Command("cmd", "ver")
	cmd.Stdin = strings.NewReader("some input")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		log.Printf("[ERROR] failed to get the OS info: %s", err)
		return OSInfo{
			Name: runtime.GOOS,
			Arch: runtime.GOARCH,
		}
	}

	osStr := strings.Replace(out.String(), "\n", "", -1)
	osStr = strings.Replace(osStr, "\r\n", "", -1)
	tmp1 := strings.Index(osStr, "[Version")
	tmp2 := strings.Index(osStr, "]")
	var ver string
	if tmp1 == -1 || tmp2 == -1 {
		ver = ""
	} else {
		ver = osStr[tmp1+9 : tmp2]
	}

	osInfo := OSInfo{
		Name:    runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: ver,
	}
	return osInfo
}

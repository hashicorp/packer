package metadata

import (
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

// CommandExecutor is an interface for executing commands.
type CommandExecutor interface {
	Exec(name string, arg ...string) ([]byte, error)
}

// DefaultExecutor is the default implementation of CommandExecutor.
type DefaultExecutor struct{}

// Exec executes a command and returns the combined output.
func (d DefaultExecutor) Exec(name string, arg ...string) ([]byte, error) {
	cmd := exec.Command(name, arg...)
	return cmd.CombinedOutput()
}

var executor CommandExecutor = DefaultExecutor{}

func GetOSMetadata() map[string]interface{} {
	var osInfo OSInfo

	switch runtime.GOOS {
	case "windows":
		osInfo = GetInfoForWindows(executor)
	case "darwin":
		osInfo = GetInfo(executor, "-srm")
	case "linux":
		osInfo = GetInfo(executor, "-srio")
	case "freebsd":
		osInfo = GetInfo(executor, "-sri")
	case "openbsd":
		osInfo = GetInfo(executor, "-srm")
	case "netbsd":
		osInfo = GetInfo(executor, "-srm")
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

func GetInfo(exec CommandExecutor, flags string) OSInfo {
	out, err := uname(exec, flags)
	tries := 0
	for strings.Contains(out, "broken pipe") && tries < 3 {
		out, err = uname(exec, flags)
		time.Sleep(500 * time.Millisecond)
		tries++
	}
	if strings.Contains(out, "broken pipe") || err != nil {
		out = ""
	}

	if err != nil {
		log.Printf("[ERROR] failed to get the OS info: %s", err)
	}
	core := retrieveCore(out)
	return OSInfo{
		Name:    runtime.GOOS,
		Arch:    runtime.GOARCH,
		Version: core,
	}
}

func uname(exec CommandExecutor, flags string) (string, error) {
	output, err := exec.Exec("uname", flags)
	return string(output), err
}

func retrieveCore(osStr string) string {
	osStr = strings.Replace(osStr, "\n", "", -1)
	osStr = strings.Replace(osStr, "\r\n", "", -1)
	osInfo := strings.Split(osStr, " ")

	var core string
	if len(osInfo) > 1 {
		core = osInfo[1]
	}
	return core
}

func GetInfoForWindows(exec CommandExecutor) OSInfo {
	out, err := exec.Exec("cmd", "ver")
	if err != nil {
		log.Printf("[ERROR] failed to get the OS info: %s", err)
		return OSInfo{
			Name: runtime.GOOS,
			Arch: runtime.GOARCH,
		}
	}

	osStr := strings.Replace(string(out), "\n", "", -1)
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

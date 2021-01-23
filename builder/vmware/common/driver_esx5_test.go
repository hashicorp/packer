package common

import (
	"fmt"
	"net"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

func TestESX5Driver_implDriver(t *testing.T) {
	var _ Driver = new(ESX5Driver)
}

func TestESX5Driver_UpdateVMX(t *testing.T) {
	var driver ESX5Driver
	data := make(map[string]string)
	driver.UpdateVMX("0.0.0.0", "", 5900, data)
	if _, ok := data["remotedisplay.vnc.ip"]; ok {
		// Do not add the remotedisplay.vnc.ip on ESXi
		t.Fatal("invalid VMX data key: remotedisplay.vnc.ip")
	}
	if enabled := data["remotedisplay.vnc.enabled"]; enabled != "TRUE" {
		t.Errorf("bad VMX data for key remotedisplay.vnc.enabled: %v", enabled)
	}
	if port := data["remotedisplay.vnc.port"]; port != fmt.Sprint(port) {
		t.Errorf("bad VMX data for key remotedisplay.vnc.port: %v", port)
	}
}

func TestESX5Driver_implOutputDir(t *testing.T) {
	var _ OutputDir = new(ESX5Driver)
}

func TestESX5Driver_implVNCAddressFinder(t *testing.T) {
	var _ VNCAddressFinder = new(ESX5Driver)
}

func TestESX5Driver_implRemoteDriver(t *testing.T) {
	var _ RemoteDriver = new(ESX5Driver)
}

func TestESX5Driver_HostIP(t *testing.T) {
	expected_host := "127.0.0.1"

	//create mock SSH server
	listen, _ := net.Listen("tcp", fmt.Sprintf("%s:0", expected_host))
	port := listen.Addr().(*net.TCPAddr).Port
	defer listen.Close()

	driver := ESX5Driver{Host: "localhost", Port: port}
	state := new(multistep.BasicStateBag)

	if host, _ := driver.HostIP(state); host != expected_host {
		t.Error(fmt.Sprintf("Expected string, %s but got %s", expected_host, host))
	}
}

func TestESX5Driver_CommHost(t *testing.T) {
	const expected_host = "127.0.0.1"

	conf := make(map[string]interface{})
	conf["communicator"] = "winrm"
	conf["winrm_username"] = "username"
	conf["winrm_password"] = "password"
	conf["winrm_host"] = expected_host

	var commConfig communicator.Config
	err := config.Decode(&commConfig, nil, conf)
	if err != nil {
		t.Fatalf("error decoding config: %v", err)
	}
	state := new(multistep.BasicStateBag)
	sshConfig := SSHConfig{Comm: commConfig}
	state.Put("sshConfig", &sshConfig)
	driver := ESX5Driver{CommConfig: *(&sshConfig.Comm)}

	host, err := driver.CommHost(state)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if host != expected_host {
		t.Errorf("bad host name: %s", host)
	}
	address, ok := state.GetOk("vm_address")
	if !ok {
		t.Error("state not updated with vm_address")
	}
	if address.(string) != expected_host {
		t.Errorf("bad vm_address: %s", address.(string))
	}
}

func TestESX5Driver_VerifyOvfTool(t *testing.T) {
	driver := ESX5Driver{}
	// should always skip validation if export is skipped, so this should always
	// pass even when ovftool is not installed.
	err := driver.VerifyOvfTool(true, false)
	if err != nil {
		t.Fatalf("shouldn't fail ever because should always skip check")
	}
}

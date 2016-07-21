package iso

import (
	"fmt"
	"net"
	"testing"

	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
)

func TestESX5Driver_implDriver(t *testing.T) {
	var _ vmwcommon.Driver = new(ESX5Driver)
}

func TestESX5Driver_implOutputDir(t *testing.T) {
	var _ vmwcommon.OutputDir = new(ESX5Driver)
}

func TestESX5Driver_implVNCAddressFinder(t *testing.T) {
	var _ vmwcommon.VNCAddressFinder = new(ESX5Driver)
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

	driver := ESX5Driver{Host: "localhost", Port: uint(port)}

	if host, _ := driver.HostIP(); host != expected_host {
		t.Error(fmt.Sprintf("Expected string, %s but got %s", expected_host, host))
	}
}

func TestESX5Driver_UpdateVMX(t *testing.T) {
	var driver ESX5Driver
	data := make(map[string]string)
	driver.UpdateVMX("0.0.0.0", 5900, data)
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

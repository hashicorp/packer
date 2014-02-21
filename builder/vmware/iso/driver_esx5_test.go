package iso

import (
	"fmt"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
	"net"
	"testing"
)

func TestESX5Driver_implDriver(t *testing.T) {
	var _ vmwcommon.Driver = new(ESX5Driver)
}

func TestESX5Driver_implOutputDir(t *testing.T) {
	var _ vmwcommon.OutputDir = new(ESX5Driver)
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

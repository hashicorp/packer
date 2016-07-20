package iso

import (
	"fmt"
	"net"
	"testing"

	"github.com/mitchellh/multistep"
	vmwcommon "github.com/mitchellh/packer/builder/vmware/common"
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

func TestESX5Driver_CommHost(t *testing.T) {
	const expected_host = "127.0.0.1"

	config := testConfig()
	config["communicator"] = "winrm"
	config["winrm_username"] = "username"
	config["winrm_password"] = "password"
	config["winrm_host"] = expected_host

	var b Builder
	warns, err := b.Prepare(config)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if host := b.config.CommConfig.Host(); host != expected_host {
		t.Fatalf("setup failed, bad host name: %s", host)
	}

	state := new(multistep.BasicStateBag)
	state.Put("config", &b.config)

	var driver ESX5Driver
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

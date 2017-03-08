package vm

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"ssh_username":     "foo",
		"shutdown_command": "foo",
		"remote_host":      "esx",
		"source_vm":        "foovm",

		packer.BuildNameConfigKey: "foo",
	}
}

func testConfigErr(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}
}

func testConfigOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

func TestConfig_Defaults(t *testing.T) {
	c := testConfig(t)
	config, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.OutputDir != "output-foo" {
		t.Errorf("bad output dir: %s", config.OutputDir)
	}

	if config.VMName != "packer-foo" {
		t.Errorf("bad vm name: %s", config.VMName)
	}
}

func TestNewConfig_Cpu(t *testing.T) {
	c := testConfig(t)

	delete(c, "cpu")
	config, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.Cpu != 0 {
		t.Fatalf("bad number of Cpus: %d", config.Cpu)
	}

	c["cpu"] = 4
	config, warns, errs = NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.Cpu != 4 {
		t.Fatalf("bad number of Cpus: %d", config.Cpu)
	}
}

func TestNewConfig_MemSize(t *testing.T) {
	c := testConfig(t)

	delete(c, "mem_size")
	config, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.MemSize != 0 {
		t.Fatalf("bad size: %d", config.MemSize)
	}

	c["mem_size"] = 4096
	config, warns, errs = NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.MemSize != 4096 {
		t.Fatalf("bad size: %d", config.MemSize)
	}
}

func TestNewConfig_DiskThick(t *testing.T) {
	c := testConfig(t)

	delete(c, "disk_thick")
	config, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.DiskThick != false {
		t.Fatalf("bad disk_thick: %t", config.DiskThick)
	}

	c["disk_thick"] = true
	config, warns, errs = NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.DiskThick != true {
		t.Fatalf("bad disk_thick: %t", config.DiskThick)
	}
}

func TestNewConfig_DiskSize(t *testing.T) {
	c := testConfig(t)

	delete(c, "disk_size")
	config, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.DiskSize != 0 {
		t.Fatalf("bad size: %d", config.DiskSize)
	}

	c["disk_size"] = 60000
	config, warns, errs = NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.DiskSize != 60000 {
		t.Fatalf("bad size: %d", config.DiskSize)
	}
}

func TestNewConfig_NetworkAdapter(t *testing.T) {
	c := testConfig(t)

	delete(c, "network_adapter")
	config, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	c["network_adapter"] = "vmxnet3"
	config, warns, errs = NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.NetworkAdapter != "vmxnet3" {
		t.Fatalf("bad network adapter: %s", config.NetworkAdapter)
	}
}

func TestNewConfig_NetworkName(t *testing.T) {
	c := testConfig(t)

	delete(c, "network_name")
	config, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	c["network_name"] = "Test Network"
	config, warns, errs = NewConfig(c)
	testConfigOk(t, warns, errs)

	if config.NetworkName != "Test Network" {
		t.Fatalf("bad network name: %s", config.NetworkName)
	}
}

func TestNewConfig_InvalidKey(t *testing.T) {
	c := testConfig(t)

	// Add a random key
	c["i_should_not_be_valid"] = true
	_, warns, errs := NewConfig(c)
	testConfigErr(t, warns, errs)
}

func TestNewConfig_OutputDir(t *testing.T) {
	c := testConfig(t)

	// Test with existing dir
	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	c["output_directory"] = dir
	_, warns, errs := NewConfig(c)
	testConfigOk(t, warns, errs)

	// Test with a good one
	c["output_directory"] = "i-hope-i-dont-exist"
	_, warns, errs = NewConfig(c)
	testConfigOk(t, warns, errs)
}

func TestNewConfig_CommConfig(t *testing.T) {
	// Test Winrm
	{
		c := testConfig(t)
		c["communicator"] = "winrm"
		c["winrm_username"] = "username"
		c["winrm_password"] = "password"
		c["winrm_host"] = "1.2.3.4"

		config, warns, errs := NewConfig(c)
		testConfigOk(t, warns, errs)

		if config.CommConfig.WinRMUser != "username" {
			t.Errorf("bad winrm_username: %s", config.CommConfig.WinRMUser)
		}
		if config.CommConfig.WinRMPassword != "password" {
			t.Errorf("bad winrm_password: %s", config.CommConfig.WinRMPassword)
		}
		if host := config.CommConfig.Host(); host != "1.2.3.4" {
			t.Errorf("bad host: %s", host)
		}
	}

	// Test SSH
	{
		c := testConfig(t)
		c["communicator"] = "ssh"
		c["ssh_username"] = "username"
		c["ssh_password"] = "password"
		c["ssh_host"] = "1.2.3.4"

		config, warns, errs := NewConfig(c)
		testConfigOk(t, warns, errs)

		if config.CommConfig.SSHUsername != "username" {
			t.Errorf("bad ssh_username: %s", config.CommConfig.SSHUsername)
		}
		if config.CommConfig.SSHPassword != "password" {
			t.Errorf("bad ssh_password: %s", config.CommConfig.SSHPassword)
		}
		if host := config.CommConfig.Host(); host != "1.2.3.4" {
			t.Errorf("bad host: %s", host)
		}
	}
}

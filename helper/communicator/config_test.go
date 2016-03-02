package communicator

import (
	"testing"

	"github.com/mitchellh/packer/template/interpolate"
)

func testConfig() *Config {
	return &Config{
		SSHUsername: "root",
	}
}

func TestConfigType(t *testing.T) {
	c := testConfig()
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.Type != "ssh" {
		t.Fatalf("bad: %#v", c)
	}
}

func TestConfig_none(t *testing.T) {
	c := &Config{Type: "none"}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}
}

func TestConfig_badtype(t *testing.T) {
	c := &Config{Type: "foo"}
	if err := c.Prepare(testContext(t)); len(err) != 1 {
		t.Fatalf("bad: %#v", err)
	}
}

func TestConfig_winrm_noport(t *testing.T) {
	c := &Config{
		Type:      "winrm",
		WinRMUser: "admin",
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5985 {
		t.Fatalf("WinRMPort doesn't match default port 5985 when SSL is not enabled and no port is specified.")
	}

}

func TestConfig_winrm_noport_ssl(t *testing.T) {
	c := &Config{
		Type:        "winrm",
		WinRMUser:   "admin",
		WinRMUseSSL: true,
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5986 {
		t.Fatalf("WinRMPort doesn't match default port 5986 when SSL is enabled and no port is specified.")
	}

}

func TestConfig_winrm_port(t *testing.T) {
	c := &Config{
		Type:      "winrm",
		WinRMUser: "admin",
		WinRMPort: 5509,
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5509 {
		t.Fatalf("WinRMPort doesn't match custom port 5509 when SSL is not enabled.")
	}

}

func TestConfig_winrm_port_ssl(t *testing.T) {
	c := &Config{
		Type:        "winrm",
		WinRMUser:   "admin",
		WinRMPort:   5510,
		WinRMUseSSL: true,
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}

	if c.WinRMPort != 5510 {
		t.Fatalf("WinRMPort doesn't match custom port 5510 when SSL is enabled.")
	}

}

func TestConfig_winrm(t *testing.T) {
	c := &Config{
		Type:      "winrm",
		WinRMUser: "admin",
	}
	if err := c.Prepare(testContext(t)); len(err) > 0 {
		t.Fatalf("bad: %#v", err)
	}
}

func testContext(t *testing.T) *interpolate.Context {
	return nil
}

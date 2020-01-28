package clone

import (
	"testing"
	"time"
)

func TestCloneConfig_MinimalConfig(t *testing.T) {
	c := new(Config)
	warns, errs := c.Prepare(minimalConfig())
	testConfigOk(t, warns, errs)
}

func TestCloneConfig_MandatoryParameters(t *testing.T) {
	params := []string{"vcenter_server", "username", "password", "template", "vm_name", "host"}
	for _, param := range params {
		raw := minimalConfig()
		raw[param] = ""
		c := new(Config)
		warns, err := c.Prepare(raw)
		testConfigErr(t, param, warns, err)
	}
}

func TestCloneConfig_Timeout(t *testing.T) {
	raw := minimalConfig()
	raw["shutdown_timeout"] = "3m"
	conf := new(Config)
	warns, err := conf.Prepare(raw)
	testConfigOk(t, warns, err)
	if conf.ShutdownConfig.Timeout != 3*time.Minute {
		t.Fatalf("shutdown_timeout sould be equal 3 minutes, got %v", conf.ShutdownConfig.Timeout)
	}
}

func TestCloneConfig_RAMReservation(t *testing.T) {
	raw := minimalConfig()
	raw["RAM_reservation"] = 1000
	raw["RAM_reserve_all"] = true
	c := new(Config)
	warns, err := c.Prepare(raw)
	testConfigErr(t, "RAM_reservation", warns, err)
}

func minimalConfig() map[string]interface{} {
	return map[string]interface{}{
		"vcenter_server": "vcenter.domain.local",
		"username":       "root",
		"password":       "vmware",
		"template":       "ubuntu",
		"vm_name":        "vm1",
		"host":           "esxi1.domain.local",
		"ssh_username":   "root",
		"ssh_password":   "secret",
	}
}

func testConfigOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Errorf("Should be no warnings: %#v", warns)
	}
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func testConfigErr(t *testing.T, context string, warns []string, err error) {
	if len(warns) > 0 {
		t.Errorf("Should be no warnings: %#v", warns)
	}
	if err == nil {
		t.Error("An error is not raised for", context)
	}
}

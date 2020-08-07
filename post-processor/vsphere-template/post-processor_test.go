package vsphere_template

import (
	"testing"
)

func getTestConfig() Config {
	return Config{
		Username: "me",
		Password: "notpassword",
		Host:     "myhost",
	}
}

func TestConfigure_Good(t *testing.T) {
	var p PostProcessor

	config := getTestConfig()

	err := p.Configure(config)
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}

func TestConfigure_ReRegisterVM(t *testing.T) {
	var p PostProcessor

	config := getTestConfig()

	err := p.Configure(config)
	if err != nil {
		t.Errorf("Error: %s", err)
	}

	if p.config.ReregisterVM.False() {
		t.Errorf("This should default to unset, not false.")
	}
}

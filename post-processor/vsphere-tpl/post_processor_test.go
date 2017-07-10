package vsphere_tpl

import (
	"testing"
)

func TestConfigureURL(t *testing.T) {
	var p PostProcessor
	p.config.Username = "me"
	p.config.Password = "notpassword"
	p.config.Host = "myhost"
	p.config.Datacenter = "mydc"
	p.config.VMName = "my vm"
	p.config.Insecure = true

	if err := p.configureURL(); err != nil {
		t.Errorf("Error: %s", err)
	}
}

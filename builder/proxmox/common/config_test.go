package proxmox

import (
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func mandatoryConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"proxmox_url":  "https://my-proxmox.my-domain:8006/api2/json",
		"username":     "apiuser@pve",
		"password":     "supersecret",
		"node":         "my-proxmox",
		"ssh_username": "root",
	}
}

func TestRequiredParameters(t *testing.T) {
	var c Config
	_, _, err := c.Prepare(&c, make(map[string]interface{}))
	if err == nil {
		t.Fatal("Expected empty configuration to fail")
	}
	errs, ok := err.(*packersdk.MultiError)
	if !ok {
		t.Fatal("Expected errors to be packersdk.MultiError")
	}

	required := []string{"username", "password", "proxmox_url", "node", "ssh_username"}
	for _, param := range required {
		found := false
		for _, err := range errs.Errors {
			if strings.Contains(err.Error(), param) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected error about missing parameter %q", required)
		}
	}
}

func TestAgentSetToFalse(t *testing.T) {
	cfg := mandatoryConfig(t)
	cfg["qemu_agent"] = false

	var c Config
	_, _, err := c.Prepare(&c, cfg)
	if err != nil {
		t.Fatal(err)
	}

	if c.Agent != false {
		t.Errorf("Expected Agent to be false, got %t", c.Agent)
	}
}

func TestPacketQueueSupportForNetworkAdapters(t *testing.T) {
	drivertests := []struct {
		expectedToFail bool
		model          string
	}{
		{expectedToFail: false, model: "virtio"},
		{expectedToFail: true, model: "e1000"},
		{expectedToFail: true, model: "e1000-82540em"},
		{expectedToFail: true, model: "e1000-82544gc"},
		{expectedToFail: true, model: "e1000-82545em"},
		{expectedToFail: true, model: "i82551"},
		{expectedToFail: true, model: "i82557b"},
		{expectedToFail: true, model: "i82559er"},
		{expectedToFail: true, model: "ne2k_isa"},
		{expectedToFail: true, model: "ne2k_pci"},
		{expectedToFail: true, model: "pcnet"},
		{expectedToFail: true, model: "rtl8139"},
		{expectedToFail: true, model: "vmxnet3"},
	}

	for _, tt := range drivertests {
		device := make(map[string]interface{})
		device["bridge"] = "vmbr0"
		device["model"] = tt.model
		device["packet_queues"] = 2

		devices := make([]map[string]interface{}, 0)
		devices = append(devices, device)

		cfg := mandatoryConfig(t)
		cfg["network_adapters"] = devices

		var c Config
		_, _, err := c.Prepare(&c, cfg)

		if tt.expectedToFail == true && err == nil {
			t.Error("expected config preparation to fail, but no error occured")
		}

		if tt.expectedToFail == false && err != nil {
			t.Errorf("expected config preparation to succeed, but %s", err.Error())
		}
	}
}

package proxmox

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template"
)

func TestRequiredParameters(t *testing.T) {
	_, _, err := NewConfig(make(map[string]interface{}))
	if err == nil {
		t.Fatal("Expected empty configuration to fail")
	}
	errs, ok := err.(*packer.MultiError)
	if !ok {
		t.Fatal("Expected errors to be packer.MultiError")
	}

	required := []string{"username", "password", "proxmox_url", "iso_file", "node", "ssh_username"}
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

func TestBasicExampleFromDocsIsValid(t *testing.T) {
	const config = `{
  "builders": [
    {
      "type": "proxmox",
      "proxmox_url": "https://my-proxmox.my-domain:8006/api2/json",
      "insecure_skip_tls_verify": true,
      "username": "apiuser@pve",
      "password": "supersecret",

      "node": "my-proxmox",
      "network_adapters": [
        {
          "bridge": "vmbr0"
        }
      ],
      "disks": [
        {
          "type": "scsi",
          "disk_size": "5G",
          "storage_pool": "local-lvm",
          "storage_pool_type": "lvm"
        }
      ],

      "iso_file": "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso",
      "http_directory":"config",
      "boot_wait": "10s",
      "boot_command": [
        "<up><tab> ip=dhcp inst.cmdline inst.ks=http://{{.HTTPIP}}:{{.HTTPPort}}/ks.cfg<enter>"
      ],

      "ssh_username": "root",
      "ssh_timeout": "15m",
      "ssh_password": "packer",

      "unmount_iso": true,
      "template_name": "fedora-29",
      "template_description": "Fedora 29-1.2, generated on {{ isotime \"2006-01-02T15:04:05Z\" }}"
    }
  ]
}`
	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatal(err)
	}

	b := &Builder{}
	warn, err := b.Prepare(tpl.Builders["proxmox"].Config)
	if err != nil {
		t.Fatal(err, warn)
	}

	// The example config does not set a number of optional fields. Validate that:
	// Memory 0 is too small, using default: 512
	// Number of cores 0 is too small, using default: 1
	// Number of sockets 0 is too small, using default: 1
	// OS not set, using default 'other'
	// NIC 0 model not set, using default 'e1000'
	// Disk 0 cache mode not set, using default 'none'
	// Agent not set, default is true

	if b.config.Memory != 512 {
		t.Errorf("Expected Memory to be 512, got %d", b.config.Memory)
	}
	if b.config.Cores != 1 {
		t.Errorf("Expected Cores to be 1, got %d", b.config.Cores)
	}
	if b.config.Sockets != 1 {
		t.Errorf("Expected Sockets to be 1, got %d", b.config.Sockets)
	}
	if b.config.OS != "other" {
		t.Errorf("Expected OS to be 'other', got %s", b.config.OS)
	}
	if b.config.NICs[0].Model != "e1000" {
		t.Errorf("Expected NIC model to be 'e1000', got %s", b.config.NICs[0].Model)
	}
	if b.config.Disks[0].CacheMode != "none" {
		t.Errorf("Expected disk cache mode to be 'none', got %s", b.config.Disks[0].CacheMode)
	}
	if b.config.Agent != true {
		t.Errorf("Expected Agent to be true, got %t", b.config.Agent)
	}
}

func TestAgentSetToFalse(t *testing.T) {
	// only the mandatory attributes are specified
	const config = `{
		"builders": [
			{
				"type": "proxmox",
				"proxmox_url": "https://my-proxmox.my-domain:8006/api2/json",
				"username": "apiuser@pve",
				"password": "supersecret",
				"iso_file": "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso",
				"ssh_username": "root",
				"node": "my-proxmox",
				"qemu_agent": false
			}
		]
	}`

	tpl, err := template.Parse(strings.NewReader(config))
	if err != nil {
		t.Fatal(err)
	}

	b := &Builder{}
	warn, err := b.Prepare(tpl.Builders["proxmox"].Config)
	if err != nil {
		t.Fatal(err, warn)
	}

	if b.config.Agent != false {
		t.Errorf("Expected Agent to be false, got %t", b.config.Agent)
	}
}

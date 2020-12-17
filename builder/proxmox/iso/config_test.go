package proxmoxiso

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/template"
)

func TestBasicExampleFromDocsIsValid(t *testing.T) {
	const config = `{
  "builders": [
    {
      "type": "proxmox-iso",
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
	_, _, err = b.Prepare(tpl.Builders["proxmox-iso"].Config)
	if err != nil {
		t.Fatal(err)
	}

	// The example config does not set a number of optional fields. Validate that:
	// Memory 0 is too small, using default: 512
	// Number of cores 0 is too small, using default: 1
	// Number of sockets 0 is too small, using default: 1
	// CPU type not set, using default 'kvm64'
	// OS not set, using default 'other'
	// NIC 0 model not set, using default 'e1000'
	// Disk 0 cache mode not set, using default 'none'
	// Agent not set, default is true
	// SCSI controller not set, using default 'lsi'
	// Firewall toggle not set, using default: 0
	// Disable KVM not set, using default: 0

	if b.config.Memory != 512 {
		t.Errorf("Expected Memory to be 512, got %d", b.config.Memory)
	}
	if b.config.Cores != 1 {
		t.Errorf("Expected Cores to be 1, got %d", b.config.Cores)
	}
	if b.config.Sockets != 1 {
		t.Errorf("Expected Sockets to be 1, got %d", b.config.Sockets)
	}
	if b.config.CPUType != "kvm64" {
		t.Errorf("Expected CPU type to be 'kvm64', got %s", b.config.CPUType)
	}
	if b.config.OS != "other" {
		t.Errorf("Expected OS to be 'other', got %s", b.config.OS)
	}
	if b.config.NICs[0].Model != "e1000" {
		t.Errorf("Expected NIC model to be 'e1000', got %s", b.config.NICs[0].Model)
	}
	if b.config.NICs[0].Firewall != false {
		t.Errorf("Expected NIC firewall to be false, got %t", b.config.NICs[0].Firewall)
	}
	if b.config.Disks[0].CacheMode != "none" {
		t.Errorf("Expected disk cache mode to be 'none', got %s", b.config.Disks[0].CacheMode)
	}
	if b.config.Agent != true {
		t.Errorf("Expected Agent to be true, got %t", b.config.Agent)
	}
	if b.config.DisableKVM != false {
		t.Errorf("Expected Disable KVM toggle to be false, got %t", b.config.DisableKVM)
	}
	if b.config.SCSIController != "lsi" {
		t.Errorf("Expected SCSI controller to be 'lsi', got %s", b.config.SCSIController)
	}
	if b.config.CloudInit != false {
		t.Errorf("Expected CloudInit to be false, got %t", b.config.CloudInit)
	}
}

func TestAgentSetToFalse(t *testing.T) {
	cfg := mandatoryConfig(t)
	cfg["qemu_agent"] = false

	var c Config
	_, warn, err := c.Prepare(cfg)
	if err != nil {
		t.Fatal(err, warn)
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
		_, _, err := c.Prepare(cfg)

		if tt.expectedToFail == true && err == nil {
			t.Error("expected config preparation to fail, but no error occured")
		}

		if tt.expectedToFail == false && err != nil {
			t.Errorf("expected config preparation to succeed, but %s", err.Error())
		}
	}
}

func TestHardDiskControllerIOThreadSupport(t *testing.T) {
	drivertests := []struct {
		expectedToFail bool
		controller     string
		disk_type      string
	}{
		// io thread is only supported by virtio-scsi-single controller
		// and only for virtio and scsi disks
		{expectedToFail: false, controller: "virtio-scsi-single", disk_type: "scsi"},
		{expectedToFail: false, controller: "virtio-scsi-single", disk_type: "virtio"},
		{expectedToFail: true, controller: "virtio-scsi-single", disk_type: "sata"},
		{expectedToFail: true, controller: "lsi", disk_type: "scsi"},
		{expectedToFail: true, controller: "lsi53c810", disk_type: "virtio"},
	}

	for _, tt := range drivertests {
		nic := make(map[string]interface{})
		nic["bridge"] = "vmbr0"

		nics := make([]map[string]interface{}, 0)
		nics = append(nics, nic)

		disk := make(map[string]interface{})
		disk["type"] = tt.disk_type
		disk["io_thread"] = true
		disk["storage_pool"] = "local-lvm"
		disk["storage_pool_type"] = "lvm"

		disks := make([]map[string]interface{}, 0)
		disks = append(disks, disk)

		cfg := mandatoryConfig(t)
		cfg["network_adapters"] = nics
		cfg["disks"] = disks
		cfg["scsi_controller"] = tt.controller

		var c Config
		_, _, err := c.Prepare(cfg)

		if tt.expectedToFail == true && err == nil {
			t.Error("expected config preparation to fail, but no error occured")
		}

		if tt.expectedToFail == false && err != nil {
			t.Errorf("expected config preparation to succeed, but %s", err.Error())
		}
	}
}

func mandatoryConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"proxmox_url":  "https://my-proxmox.my-domain:8006/api2/json",
		"username":     "apiuser@pve",
		"password":     "supersecret",
		"node":         "my-proxmox",
		"ssh_username": "root",
		"iso_file":     "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso",
	}
}

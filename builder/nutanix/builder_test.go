package nutanix

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"type":             "nutanix",
		"nutanix_endpoint": "https://127.0.0.1",
		"nutanix_port":     9440,
		"nutanix_insecure": true,
		"nutanix_username": "user",
		"nutanix_password": "password",
		"new_image_name":   "test_image",
		"shutdown_command": "C:\\Windows\\system32\\sysprep\\sysprep.exe /generalize /oobe ",
		"shutdown_timeout": "15m",
	}
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilder_Prepare_BadType(t *testing.T) {
	b := &Builder{}
	config := testConfig()
	// should be int
	config["nutanix_port"] = "port_number"

	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatalf("prepare should fail")
	}
}

func TestBuilder_Prepare_Windows(t *testing.T) {
	b := &Builder{}
	config := testConfig()
	config["communicator"] = "winrm"
	config["winrm_username"] = "Administrator"
	config["winrm_password"] = "p@ssword1"

	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("prepare should not fail. errors: %v", err)
	}
}

func TestBuilder_Prepare_Linux(t *testing.T) {
	b := &Builder{}
	config := testConfig()
	config["communicator"] = "ssh"
	config["ssh_username"] = "root"
	config["ssh_password"] = "p@ssword1"

	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("prepare should not fail. errors: %v", err)
	}
}

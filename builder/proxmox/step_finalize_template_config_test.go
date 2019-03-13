package proxmox

import (
	"context"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type finalizerMock struct {
	getConfig func() (map[string]interface{}, error)
	setConfig func(map[string]interface{}) (string, error)
}

func (m finalizerMock) GetVmConfig(*proxmox.VmRef) (map[string]interface{}, error) {
	return m.getConfig()
}
func (m finalizerMock) SetVmConfig(vmref *proxmox.VmRef, c map[string]interface{}) (string, error) {
	return m.setConfig(c)
}

func TestTemplateFinalize(t *testing.T) {
	finalizer := finalizerMock{
		getConfig: func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"name":        "dummy",
				"description": "Packer ephemeral build VM",
				"ide2":        "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso,media=cdrom",
			}, nil
		},
		setConfig: func(c map[string]interface{}) (string, error) {
			if c["name"] != "my-template" {
				t.Errorf("Expected name to be my-template, got %q", c["name"])
			}
			if c["description"] != "foo" {
				t.Errorf("Expected description to be foo, got %q", c["description"])
			}
			if c["ide2"] != "none,media=cdrom" {
				t.Errorf("Expected ide2 to be none,media=cdrom, got %q", c["ide2"])
			}

			return "", nil
		},
	}

	state := new(multistep.BasicStateBag)
	state.Put("ui", packer.TestUi(t))
	state.Put("config", &Config{
		TemplateName:        "my-template",
		TemplateDescription: "foo",
		UnmountISO:          true,
	})
	state.Put("vmRef", proxmox.NewVmRef(1))
	state.Put("proxmoxClient", finalizer)

	step := stepFinalizeTemplateConfig{}
	action := step.Run(context.TODO(), state)
	if action != multistep.ActionContinue {
		t.Error("Expected action to be Continue, got Halt")
	}
}

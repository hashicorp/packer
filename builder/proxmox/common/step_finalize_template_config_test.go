package proxmox

import (
	"context"
	"fmt"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type finalizerMock struct {
	getConfig func() (map[string]interface{}, error)
	setConfig func(map[string]interface{}) (string, error)
}

func (m finalizerMock) GetVmConfig(*proxmox.VmRef) (map[string]interface{}, error) {
	return m.getConfig()
}
func (m finalizerMock) SetVmConfig(vmref *proxmox.VmRef, c map[string]interface{}) (interface{}, error) {
	return m.setConfig(c)
}

var _ templateFinalizer = finalizerMock{}

func TestTemplateFinalize(t *testing.T) {
	cs := []struct {
		name                string
		builderConfig       *Config
		initialVMConfig     map[string]interface{}
		getConfigErr        error
		expectCallSetConfig bool
		expectedVMConfig    map[string]interface{}
		setConfigErr        error
		expectedAction      multistep.StepAction
	}{
		{
			name:          "empty config changes only description",
			builderConfig: &Config{},
			initialVMConfig: map[string]interface{}{
				"name":        "dummy",
				"description": "Packer ephemeral build VM",
				"ide2":        "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso,media=cdrom",
			},
			expectCallSetConfig: true,
			expectedVMConfig: map[string]interface{}{
				"name":        nil,
				"description": "",
				"ide2":        nil,
			},
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "all options",
			builderConfig: &Config{
				TemplateName:        "my-template",
				TemplateDescription: "some-description",
			},
			initialVMConfig: map[string]interface{}{
				"name":        "dummy",
				"description": "Packer ephemeral build VM",
			},
			expectCallSetConfig: true,
			expectedVMConfig: map[string]interface{}{
				"name":        "my-template",
				"description": "some-description",
			},
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "all options with cloud-init",
			builderConfig: &Config{
				TemplateName:        "my-template",
				TemplateDescription: "some-description",
				CloudInit:           true,
			},
			initialVMConfig: map[string]interface{}{
				"name":        "dummy",
				"description": "Packer ephemeral build VM",
				"bootdisk":    "virtio0",
				"virtio0":     "ceph01:base-223-disk-0,cache=unsafe,media=disk,size=32G",
			},
			expectCallSetConfig: true,
			expectedVMConfig: map[string]interface{}{
				"name":        "my-template",
				"description": "some-description",
				"ide3":        "ceph01:cloudinit",
			},
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "no available controller for cloud-init drive",
			builderConfig: &Config{
				TemplateName:        "my-template",
				TemplateDescription: "some-description",
				CloudInit:           true,
			},
			initialVMConfig: map[string]interface{}{
				"name":        "dummy",
				"description": "Packer ephemeral build VM",
				"ide0":        "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso,media=cdrom",
				"ide1":        "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso,media=cdrom",
				"ide2":        "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso,media=cdrom",
				"ide3":        "local:iso/Fedora-Server-dvd-x86_64-29-1.2.iso,media=cdrom",
				"bootdisk":    "virtio0",
				"virtio0":     "ceph01:base-223-disk-0,cache=unsafe,media=disk,size=32G",
			},
			expectCallSetConfig: false,
			expectedAction:      multistep.ActionHalt,
		},
		{
			name: "GetVmConfig error should return halt",
			builderConfig: &Config{
				TemplateName:        "my-template",
				TemplateDescription: "some-description",
				CloudInit:           true,
			},
			getConfigErr:        fmt.Errorf("some error"),
			expectCallSetConfig: false,
			expectedAction:      multistep.ActionHalt,
		},
		{
			name: "SetVmConfig error should return halt",
			builderConfig: &Config{
				TemplateName:        "my-template",
				TemplateDescription: "some-description",
			},
			initialVMConfig: map[string]interface{}{
				"name":        "dummy",
				"description": "Packer ephemeral build VM",
			},
			expectCallSetConfig: true,
			setConfigErr:        fmt.Errorf("some error"),
			expectedAction:      multistep.ActionHalt,
		},
	}

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			finalizer := finalizerMock{
				getConfig: func() (map[string]interface{}, error) {
					return c.initialVMConfig, c.getConfigErr
				},
				setConfig: func(cfg map[string]interface{}) (string, error) {
					if !c.expectCallSetConfig {
						t.Error("Did not expect SetVmConfig to be called")
					}
					for key, val := range c.expectedVMConfig {
						if cfg[key] != val {
							t.Errorf("Expected %q to be %q, got %q", key, val, cfg[key])
						}
					}

					return "", c.setConfigErr
				},
			}

			state := new(multistep.BasicStateBag)
			state.Put("ui", packersdk.TestUi(t))
			state.Put("config", c.builderConfig)
			state.Put("vmRef", proxmox.NewVmRef(1))
			state.Put("proxmoxClient", finalizer)

			step := stepFinalizeTemplateConfig{}
			action := step.Run(context.TODO(), state)
			if action != c.expectedAction {
				t.Errorf("Expected action to be %v, got %v", c.expectedAction, action)
			}
		})
	}
}

package proxmox

import (
	"context"
	"fmt"
	"testing"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type converterMock struct {
	shutdownVm     func(*proxmox.VmRef) (string, error)
	createTemplate func(*proxmox.VmRef) error
}

func (m converterMock) ShutdownVm(r *proxmox.VmRef) (string, error) {
	return m.shutdownVm(r)
}
func (m converterMock) CreateTemplate(r *proxmox.VmRef) error {
	return m.createTemplate(r)
}

var _ templateConverter = converterMock{}

func TestConvertToTemplate(t *testing.T) {
	cs := []struct {
		name                     string
		shutdownErr              error
		expectCallCreateTemplate bool
		createTemplateErr        error
		expectedAction           multistep.StepAction
		expectTemplateIdSet      bool
	}{
		{
			name:                     "no errors returns continue and sets template id",
			expectCallCreateTemplate: true,
			expectedAction:           multistep.ActionContinue,
			expectTemplateIdSet:      true,
		},
		{
			name:                     "when shutdown fails, don't try to create template and halt",
			shutdownErr:              fmt.Errorf("failed to stop vm"),
			expectCallCreateTemplate: false,
			expectedAction:           multistep.ActionHalt,
			expectTemplateIdSet:      false,
		},
		{
			name:                     "when create template fails, halt",
			expectCallCreateTemplate: true,
			createTemplateErr:        fmt.Errorf("failed to stop vm"),
			expectedAction:           multistep.ActionHalt,
			expectTemplateIdSet:      false,
		},
	}

	const vmid = 123

	for _, c := range cs {
		t.Run(c.name, func(t *testing.T) {
			converter := converterMock{
				shutdownVm: func(r *proxmox.VmRef) (string, error) {
					if r.VmId() != vmid {
						t.Errorf("ShutdownVm called with unexpected id, expected %d, got %d", vmid, r.VmId())
					}
					return "", c.shutdownErr
				},
				createTemplate: func(r *proxmox.VmRef) error {
					if r.VmId() != vmid {
						t.Errorf("CreateTemplate called with unexpected id, expected %d, got %d", vmid, r.VmId())
					}
					if !c.expectCallCreateTemplate {
						t.Error("Did not expect CreateTemplate to be called")
					}

					return c.createTemplateErr
				},
			}

			state := new(multistep.BasicStateBag)
			state.Put("ui", packer.TestUi(t))
			state.Put("vmRef", proxmox.NewVmRef(vmid))
			state.Put("proxmoxClient", converter)

			step := stepConvertToTemplate{}
			action := step.Run(context.TODO(), state)
			if action != c.expectedAction {
				t.Errorf("Expected action to be %v, got %v", c.expectedAction, action)
			}

			id, wasSet := state.GetOk("template_id")

			if c.expectTemplateIdSet != wasSet {
				t.Errorf("Expected template_id state present=%v was present=%v", c.expectTemplateIdSet, wasSet)
			}

			if c.expectTemplateIdSet && id != vmid {
				t.Errorf("Expected template_id state to be set to %d, got %v", vmid, id)
			}
		})
	}
}

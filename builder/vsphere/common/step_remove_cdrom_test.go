package common

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

func TestStepRemoveCDRom_Run(t *testing.T) {
	tc := []struct {
		name           string
		step           *StepRemoveCDRom
		expectedAction multistep.StepAction
		vmMock         *driver.VirtualMachineMock
		expectedVmMock *driver.VirtualMachineMock
		fail           bool
		errMessage     string
	}{
		{
			name: "Eject CD-ROM drives",
			step: &StepRemoveCDRom{
				Config: &RemoveCDRomConfig{},
			},
			expectedAction: multistep.ActionContinue,
			vmMock:         new(driver.VirtualMachineMock),
			expectedVmMock: &driver.VirtualMachineMock{
				EjectCdromsCalled: true,
			},
			fail: false,
		},
		{
			name: "Failed to eject CD-ROM drives",
			step: &StepRemoveCDRom{
				Config: &RemoveCDRomConfig{},
			},
			expectedAction: multistep.ActionHalt,
			vmMock: &driver.VirtualMachineMock{
				EjectCdromsErr: fmt.Errorf("failed to eject cd-rom drives"),
			},
			expectedVmMock: &driver.VirtualMachineMock{
				EjectCdromsCalled: true,
			},
			fail:       true,
			errMessage: "failed to eject cd-rom drives",
		},
		{
			name: "Eject and delete CD-ROM drives",
			step: &StepRemoveCDRom{
				Config: &RemoveCDRomConfig{
					RemoveCdrom: true,
				},
			},
			expectedAction: multistep.ActionContinue,
			vmMock:         new(driver.VirtualMachineMock),
			expectedVmMock: &driver.VirtualMachineMock{
				EjectCdromsCalled:  true,
				RemoveCdromsCalled: true,
			},
			fail: false,
		},
		{
			name: "Fail to delete CD-ROM drives",
			step: &StepRemoveCDRom{
				Config: &RemoveCDRomConfig{
					RemoveCdrom: true,
				},
			},
			expectedAction: multistep.ActionHalt,
			vmMock: &driver.VirtualMachineMock{
				RemoveCdromsErr: fmt.Errorf("failed to delete cd-rom devices"),
			},
			expectedVmMock: &driver.VirtualMachineMock{
				EjectCdromsCalled:  true,
				RemoveCdromsCalled: true,
			},
			fail:       true,
			errMessage: "failed to delete cd-rom devices",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			state := basicStateBag(nil)
			state.Put("vm", c.vmMock)

			if action := c.step.Run(context.TODO(), state); action != c.expectedAction {
				t.Fatalf("unexpected action %v", action)
			}
			err, ok := state.Get("error").(error)
			if ok {
				if err.Error() != c.errMessage {
					t.Fatalf("unexpected error %s", err.Error())
				}
			} else {
				if c.fail {
					t.Fatalf("expected to fail but it didn't")
				}
			}

			if diff := cmp.Diff(c.vmMock, c.expectedVmMock,
				cmpopts.IgnoreInterfaces(struct{ error }{})); diff != "" {
				t.Fatalf("unexpected VirtualMachine calls: %s", diff)
			}
		})
	}
}

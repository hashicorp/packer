package clone

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/packer/builder/vsphere/driver"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestCDRomConfig_Prepare(t *testing.T) {
	// Data validation
	tc := []struct {
		name           string
		config         *CDRomConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name:           "Should not fail for empty config",
			config:         new(CDRomConfig),
			fail:           false,
			expectedErrMsg: "",
		},
		{
			name:           "Valid cdroom type ide",
			config:         &CDRomConfig{CdromType: "ide"},
			fail:           false,
			expectedErrMsg: "",
		},
		{
			name:           "Valid cdroom type sata",
			config:         &CDRomConfig{CdromType: "ide"},
			fail:           false,
			expectedErrMsg: "",
		},
		{
			name:           "Invalid cdroom type",
			config:         &CDRomConfig{CdromType: "invalid"},
			fail:           true,
			expectedErrMsg: "'cdrom_type' must be 'ide' or 'sata'",
		},
	}

	for _, c := range tc {
		errs := c.config.Prepare()
		if c.fail {
			if len(errs) == 0 {
				t.Fatalf("Config preprare should fail")
			}
			if errs[0].Error() != c.expectedErrMsg {
				t.Fatalf("Expected error message: %s but was '%s'", c.expectedErrMsg, errs[0].Error())
			}
		} else {
			if len(errs) != 0 {
				t.Fatalf("Config preprare should not fail")
			}
		}
	}
}

func TestStepAddCDRom_Run(t *testing.T) {
	tc := []struct {
		name           string
		state          *multistep.BasicStateBag
		step           *StepAddCDRom
		vmMock         *driver.VirtualMachineMock
		expectedAction multistep.StepAction
		expectedVmMock *driver.VirtualMachineMock
		fail           bool
		errMessage     string
	}{
		{
			name:  "CDRom SATA type with all cd paths set",
			state: cdPathStateBag(),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
				},
			},
			vmMock:         new(driver.VirtualMachineMock),
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				FindSATAControllerCalled: true,
				AddCdromCalled:           true,
				AddCdromCalledTimes:      1,
				AddCdromTypes:            []string{"sata"},
				AddCdromPaths:            []string{"cd/path"},
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Add SATA Controller",
			state: basicStateBag(),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
				},
			},
			vmMock: &driver.VirtualMachineMock{
				FindSATAControllerErr: driver.ErrNoSataController,
			},
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				FindSATAControllerCalled: true,
				FindSATAControllerErr:    driver.ErrNoSataController,
				AddSATAControllerCalled:  true,
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Fail to add SATA Controller",
			state: basicStateBag(),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
				},
			},
			vmMock: &driver.VirtualMachineMock{
				FindSATAControllerErr: driver.ErrNoSataController,
				AddSATAControllerErr:  fmt.Errorf("AddSATAController error"),
			},
			expectedAction: multistep.ActionHalt,
			expectedVmMock: &driver.VirtualMachineMock{
				FindSATAControllerCalled: true,
				AddSATAControllerCalled:  true,
			},
			fail:       true,
			errMessage: fmt.Sprintf("error adding SATA controller: %v", fmt.Errorf("AddSATAController error")),
		},
		{
			name:  "IDE CDRom Type and Iso Path set",
			state: basicStateBag(),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "ide",
				},
			},
			vmMock:         new(driver.VirtualMachineMock),
			expectedAction: multistep.ActionContinue,
			expectedVmMock: new(driver.VirtualMachineMock),
			fail:           false,
			errMessage:     "",
		},
		{
			name:  "Fail to add cdrom from state cd_path",
			state: cdPathStateBag(),
			step: &StepAddCDRom{
				Config: new(CDRomConfig),
			},
			vmMock: &driver.VirtualMachineMock{
				AddCdromErr: fmt.Errorf("AddCdrom error"),
			},
			expectedAction: multistep.ActionHalt,
			expectedVmMock: &driver.VirtualMachineMock{
				AddCdromCalled:      true,
				AddCdromCalledTimes: 1,
				AddCdromTypes:       []string{""},
				AddCdromPaths:       []string{"cd/path"},
			},
			fail:       true,
			errMessage: fmt.Sprintf("error mounting a CD 'cd/path': %v", fmt.Errorf("AddCdrom error")),
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			c.state.Put("vm", c.vmMock)
			if action := c.step.Run(context.TODO(), c.state); action != c.expectedAction {
				t.Fatalf("unexpected action %v", action)
			}
			err, ok := c.state.Get("error").(error)
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

func cdPathStateBag() *multistep.BasicStateBag {
	state := basicStateBag()
	state.Put("cd_path", "cd/path")
	return state
}

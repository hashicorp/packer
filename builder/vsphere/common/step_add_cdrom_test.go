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
			name:  "CDRom SATA type with all iso paths set",
			state: cdAndIsoRemotePathStateBag(),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "sata",
					ISOPaths:  []string{"iso/path"},
				},
			},
			vmMock:         new(driver.VirtualMachineMock),
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				FindSATAControllerCalled: true,
				AddCdromCalled:           true,
				AddCdromCalledTimes:      3,
				AddCdromTypes:            []string{"sata", "sata", "sata"},
				AddCdromPaths:            []string{"remote/path", "iso/path", "cd/path"},
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Add SATA Controller",
			state: basicStateBag(nil),
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
			state: basicStateBag(nil),
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
			state: basicStateBag(nil),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					CdromType: "ide",
					ISOPaths:  []string{"iso/path"},
				},
			},
			vmMock:         new(driver.VirtualMachineMock),
			expectedAction: multistep.ActionContinue,
			expectedVmMock: &driver.VirtualMachineMock{
				AddCdromCalled:      true,
				AddCdromCalledTimes: 1,
				AddCdromTypes:       []string{"ide"},
				AddCdromPaths:       []string{"iso/path"},
			},
			fail:       false,
			errMessage: "",
		},
		{
			name:  "Fail to add cdrom from ISOPaths",
			state: basicStateBag(nil),
			step: &StepAddCDRom{
				Config: &CDRomConfig{
					ISOPaths: []string{"iso/path"},
				},
			},
			vmMock: &driver.VirtualMachineMock{
				AddCdromErr: fmt.Errorf("AddCdrom error"),
			},
			expectedAction: multistep.ActionHalt,
			expectedVmMock: &driver.VirtualMachineMock{
				AddCdromCalled:      true,
				AddCdromCalledTimes: 1,
				AddCdromTypes:       []string{""},
				AddCdromPaths:       []string{"iso/path"},
			},
			fail:       true,
			errMessage: fmt.Sprintf("error mounting an image 'iso/path': %v", fmt.Errorf("AddCdrom error")),
		},
		{
			name:  "Fail to add cdrom from state iso_remote_path",
			state: isoRemotePathStateBag(),
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
				AddCdromPaths:       []string{"remote/path"},
			},
			fail:       true,
			errMessage: fmt.Sprintf("error mounting an image 'remote/path': %v", fmt.Errorf("AddCdrom error")),
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

func cdAndIsoRemotePathStateBag() *multistep.BasicStateBag {
	state := basicStateBag(nil)
	state.Put("iso_remote_path", "remote/path")
	state.Put("cd_path", "cd/path")
	return state
}

func isoRemotePathStateBag() *multistep.BasicStateBag {
	state := basicStateBag(nil)
	state.Put("iso_remote_path", "remote/path")
	return state
}

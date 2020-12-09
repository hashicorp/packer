package common

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

func TestHardwareConfig_Prepare(t *testing.T) {
	tc := []struct {
		name           string
		config         *HardwareConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name:   "Validate empty config",
			config: &HardwareConfig{},
			fail:   false,
		},
		{
			name: "Validate RAMReservation RAMReserveAll cannot be used together",
			config: &HardwareConfig{
				RAMReservation: 2,
				RAMReserveAll:  true,
			},
			fail:           true,
			expectedErrMsg: "'RAM_reservation' and 'RAM_reserve_all' cannot be used together",
		},
		{
			name: "Invalid firmware",
			config: &HardwareConfig{
				Firmware: "invalid",
			},
			fail:           true,
			expectedErrMsg: "'firmware' must be '', 'bios', 'efi' or 'efi-secure'",
		},
		{
			name: "Validate 'bios' firmware",
			config: &HardwareConfig{
				Firmware: "bios",
			},
			fail: false,
		},
		{
			name: "Validate 'efi' firmware",
			config: &HardwareConfig{
				Firmware: "efi",
			},
			fail: false,
		},
		{
			name: "Validate 'efi-secure' firmware",
			config: &HardwareConfig{
				Firmware: "efi-secure",
			},
			fail: false,
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
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
		})
	}
}

func TestStepConfigureHardware_Run(t *testing.T) {
	tc := []struct {
		name            string
		step            *StepConfigureHardware
		action          multistep.StepAction
		configureError  error
		configureCalled bool
		hardwareConfig  *driver.HardwareConfig
	}{
		{
			name:            "Configure hardware",
			step:            basicStepConfigureHardware(),
			action:          multistep.ActionContinue,
			configureError:  nil,
			configureCalled: true,
			hardwareConfig:  driverHardwareConfigFromConfig(basicStepConfigureHardware().Config),
		},
		{
			name:            "Don't configure hardware when config is empty",
			step:            &StepConfigureHardware{Config: &HardwareConfig{}},
			action:          multistep.ActionContinue,
			configureError:  nil,
			configureCalled: false,
		},
		{
			name:            "Halt when configure return error",
			step:            basicStepConfigureHardware(),
			action:          multistep.ActionHalt,
			configureError:  errors.New("failed to configure"),
			configureCalled: true,
			hardwareConfig:  driverHardwareConfigFromConfig(basicStepConfigureHardware().Config),
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			state := basicStateBag(nil)
			vmMock := new(driver.VirtualMachineMock)
			vmMock.ConfigureError = c.configureError
			state.Put("vm", vmMock)

			action := c.step.Run(context.TODO(), state)
			if action != c.action {
				t.Fatalf("expected action '%v' but actual action was '%v'", c.action, action)
			}
			if vmMock.ConfigureCalled != c.configureCalled {
				t.Fatalf("expecting vm.Configure called to %t but was %t", c.configureCalled, vmMock.ConfigureCalled)
			}
			if diff := cmp.Diff(vmMock.ConfigureHardwareConfig, c.hardwareConfig); diff != "" {
				t.Fatalf("wrong driver.HardwareConfig: %s", diff)
			}

			err, ok := state.GetOk("error")
			containsError := c.configureError != nil
			if containsError != ok {
				t.Fatalf("Contain error - expecting %t but was %t", containsError, ok)
			}
			if containsError {
				if !strings.Contains(err.(error).Error(), c.configureError.Error()) {
					t.Fatalf("Destroy should fail with error message '%s' but failed with '%s'", c.configureError.Error(), err.(error).Error())
				}
			}
		})
	}
}

func basicStepConfigureHardware() *StepConfigureHardware {
	return &StepConfigureHardware{
		Config: &HardwareConfig{
			CPUs:           1,
			CpuCores:       1,
			CPUReservation: 1,
			CPULimit:       4000,
			RAM:            1024,
			RAMReserveAll:  true,
			Firmware:       "efi-secure",
			ForceBIOSSetup: true,
		},
	}
}

func driverHardwareConfigFromConfig(config *HardwareConfig) *driver.HardwareConfig {
	return &driver.HardwareConfig{
		CPUs:                config.CPUs,
		CpuCores:            config.CpuCores,
		CPUReservation:      config.CPUReservation,
		CPULimit:            config.CPULimit,
		RAM:                 config.RAM,
		RAMReservation:      config.RAMReservation,
		RAMReserveAll:       config.RAMReserveAll,
		NestedHV:            config.NestedHV,
		CpuHotAddEnabled:    config.CpuHotAddEnabled,
		MemoryHotAddEnabled: config.MemoryHotAddEnabled,
		VideoRAM:            config.VideoRAM,
		VGPUProfile:         config.VGPUProfile,
		Firmware:            config.Firmware,
		ForceBIOSSetup:      config.ForceBIOSSetup,
	}
}

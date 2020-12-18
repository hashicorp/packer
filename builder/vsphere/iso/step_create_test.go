package iso

import (
	"context"
	"errors"
	"io/ioutil"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

func TestCreateConfig_Prepare(t *testing.T) {
	// Empty config - check defaults
	config := &CreateConfig{
		// Storage is required
		StorageConfig: common.StorageConfig{
			Storage: []common.DiskConfig{
				{
					DiskSize: 32768,
				},
			},
		},
	}
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("Config preprare should not fail: %s", errs[0])
	}
	if config.GuestOSType != "otherGuest" {
		t.Fatalf("GuestOSType should default to 'otherGuest'")
	}
	if len(config.StorageConfig.DiskControllerType) != 1 {
		t.Fatalf("DiskControllerType should have at least one element as default")
	}

	// Data validation
	tc := []struct {
		name           string
		config         *CreateConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name: "Storage validate disk_size",
			config: &CreateConfig{
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize:            0,
							DiskThinProvisioned: true,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "storage[0].'disk_size' is required",
		},
		{
			name: "Storage validate disk_controller_index",
			config: &CreateConfig{
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize:            32768,
							DiskControllerIndex: 3,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "storage[0].'disk_controller_index' references an unknown disk controller",
		},
		{
			name: "USBController validate 'usb' and 'xhci' can be set together",
			config: &CreateConfig{
				USBController: []string{"usb", "xhci"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail: false,
		},
		{
			name: "USBController validate '1' and '0' can be set together",
			config: &CreateConfig{
				USBController: []string{"1", "0"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail: false,
		},
		{
			name: "USBController validate 'true' and 'false' can be set together",
			config: &CreateConfig{
				USBController: []string{"true", "false"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail: false,
		},
		{
			name: "USBController validate 'true' and 'usb' cannot be set together",
			config: &CreateConfig{
				USBController: []string{"true", "usb"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "there can only be one usb controller and one xhci controller",
		},
		{
			name: "USBController validate '1' and 'usb' cannot be set together",
			config: &CreateConfig{
				USBController: []string{"1", "usb"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "there can only be one usb controller and one xhci controller",
		},
		{
			name: "USBController validate 'xhci' cannot be set more that once",
			config: &CreateConfig{
				USBController: []string{"xhci", "xhci"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "there can only be one usb controller and one xhci controller",
		},
		{
			name: "USBController validate unknown value cannot be set",
			config: &CreateConfig{
				USBController: []string{"unknown"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "usb_controller[0] references an unknown usb controller",
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
				t.Fatalf("Config preprare should not fail: %s", errs[0])
			}
		}
	}
}

func TestStepCreateVM_Run(t *testing.T) {
	state := basicStateBag()
	driverMock := driver.NewDriverMock()
	state.Put("driver", driverMock)
	step := basicStepCreateVM()
	step.Force = true
	vmPath := path.Join(step.Location.Folder, step.Location.VMName)

	if action := step.Run(context.TODO(), state); action == multistep.ActionHalt {
		t.Fatalf("Should not halt.")
	}

	// Pre clean VM
	if !driverMock.PreCleanVMCalled {
		t.Fatalf("driver.PreCleanVM should be called.")
	}
	if driverMock.PreCleanForce != step.Force {
		t.Fatalf("Force PreCleanVM should be %t but was %t.", step.Force, driverMock.PreCleanForce)
	}
	if driverMock.PreCleanVMPath != vmPath {
		t.Fatalf("VM path expected to be %s but was %s", vmPath, driverMock.PreCleanVMPath)
	}

	if !driverMock.CreateVMCalled {
		t.Fatalf("driver.CreateVM should be called.")
	}
	if diff := cmp.Diff(driverMock.CreateConfig, driverCreateConfig(step.Config, step.Location)); diff != "" {
		t.Fatalf("wrong driver.CreateConfig: %s", diff)
	}
	vm, ok := state.GetOk("vm")
	if !ok {
		t.Fatal("state must contain the VM")
	}
	if vm != driverMock.VM {
		t.Fatalf("state doesn't contain the created VM.")
	}
}

func TestStepCreateVM_RunHalt(t *testing.T) {
	state := basicStateBag()
	step := basicStepCreateVM()

	// PreCleanVM fails
	driverMock := driver.NewDriverMock()
	driverMock.PreCleanShouldFail = true
	state.Put("driver", driverMock)
	if action := step.Run(context.TODO(), state); action != multistep.ActionHalt {
		t.Fatalf("Step should halt.")
	}
	if !driverMock.PreCleanVMCalled {
		t.Fatalf("driver.PreCleanVM should be called")
	}

	// CreateVM fails
	driverMock = driver.NewDriverMock()
	driverMock.CreateVMShouldFail = true
	state.Put("driver", driverMock)
	if action := step.Run(context.TODO(), state); action != multistep.ActionHalt {
		t.Fatalf("Step should halt.")
	}
	if !driverMock.PreCleanVMCalled {
		t.Fatalf("driver.PreCleanVM should be called")
	}
	if !driverMock.CreateVMCalled {
		t.Fatalf("driver.PreCleanVM should be called")
	}
	if _, ok := state.GetOk("vm"); ok {
		t.Fatal("state should not contain a VM")
	}
}

func TestStepCreateVM_Cleanup(t *testing.T) {
	state := basicStateBag()
	step := basicStepCreateVM()
	vm := new(driver.VirtualMachineMock)
	state.Put("vm", vm)

	// Clean up when state is cancelled
	state.Put(multistep.StateCancelled, true)
	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("vm.Destroy should be called")
	}
	vm.DestroyCalled = false
	state.Remove(multistep.StateCancelled)

	// Clean up when state is halted
	state.Put(multistep.StateHalted, true)
	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("vm.Destroy should be called")
	}
	vm.DestroyCalled = false
	state.Remove(multistep.StateHalted)

	// Clean up when state is destroy_vm is set
	state.Put("destroy_vm", true)
	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("vm.Destroy should be called")
	}
	vm.DestroyCalled = false
	state.Remove("destroy_vm")

	// Don't clean up if state is not set with previous values
	step.Cleanup(state)
	if vm.DestroyCalled {
		t.Fatalf("vm.Destroy should not be called")
	}

	// Destroy fail
	errorBuffer := &strings.Builder{}
	ui := &packersdk.BasicUi{
		Reader:      strings.NewReader(""),
		Writer:      ioutil.Discard,
		ErrorWriter: errorBuffer,
	}
	state.Put("ui", ui)
	state.Put(multistep.StateCancelled, true)
	vm.DestroyError = errors.New("destroy failed")

	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("vm.Destroy should be called")
	}
	if !strings.Contains(errorBuffer.String(), vm.DestroyError.Error()) {
		t.Fatalf("Destroy should fail with error message '%s' but failed with '%s'", vm.DestroyError.Error(), errorBuffer.String())
	}
	vm.DestroyCalled = false
	state.Remove(multistep.StateCancelled)

	// Should not destroy if VM is not set
	state.Remove("vm")
	state.Put(multistep.StateCancelled, true)
	step.Cleanup(state)
	if vm.DestroyCalled {
		t.Fatalf("vm.Destroy should not be called")
	}
}

func basicStepCreateVM() *StepCreateVM {
	step := &StepCreateVM{
		Config:   createConfig(),
		Location: basicLocationConfig(),
	}
	return step
}

func basicLocationConfig() *common.LocationConfig {
	return &common.LocationConfig{
		VMName:       "test-vm",
		Folder:       "test-folder",
		Cluster:      "test-cluster",
		Host:         "test-host",
		ResourcePool: "test-resource-pool",
		Datastore:    "test-datastore",
	}
}

func createConfig() *CreateConfig {
	return &CreateConfig{
		Version:     1,
		GuestOSType: "ubuntu64Guest",
		StorageConfig: common.StorageConfig{
			DiskControllerType: []string{"pvscsi"},
			Storage: []common.DiskConfig{
				{
					DiskSize:            32768,
					DiskThinProvisioned: true,
				},
			},
		},
		NICs: []NIC{
			{
				Network:     "VM Network",
				NetworkCard: "vmxnet3",
			},
		},
	}
}

func driverCreateConfig(config *CreateConfig, location *common.LocationConfig) *driver.CreateConfig {
	var networkCards []driver.NIC
	for _, nic := range config.NICs {
		networkCards = append(networkCards, driver.NIC{
			Network:     nic.Network,
			NetworkCard: nic.NetworkCard,
			MacAddress:  nic.MacAddress,
			Passthrough: nic.Passthrough,
		})
	}

	var disks []driver.Disk
	for _, disk := range config.StorageConfig.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	return &driver.CreateConfig{
		StorageConfig: driver.StorageConfig{
			DiskControllerType: config.StorageConfig.DiskControllerType,
			Storage:            disks,
		},
		Annotation:    config.Notes,
		Name:          location.VMName,
		Folder:        location.Folder,
		Cluster:       location.Cluster,
		Host:          location.Host,
		ResourcePool:  location.ResourcePool,
		Datastore:     location.Datastore,
		GuestOS:       config.GuestOSType,
		NICs:          networkCards,
		USBController: config.USBController,
		Version:       config.Version,
	}
}

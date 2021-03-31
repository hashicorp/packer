package clone

import (
	"bytes"
	"context"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/vsphere/common"
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

func TestCreateConfig_Prepare(t *testing.T) {
	tc := []struct {
		name           string
		config         *CloneConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name: "Valid config",
			config: &CloneConfig{
				Template: "template name",
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 0,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "storage[0].'disk_size' is required",
		},
		{
			name: "Storage validate disk_size",
			config: &CloneConfig{
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
			config: &CloneConfig{
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
			name: "Validate template is set",
			config: &CloneConfig{
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "'template' is required",
		},
		{
			name: "Validate LinkedClone and DiskSize set at the same time",
			config: &CloneConfig{
				Template:    "template name",
				LinkedClone: true,
				DiskSize:    32768,
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "'linked_clone' and 'disk_size' cannot be used together",
		},
		{
			name: "Validate MacAddress and Network not set at the same time",
			config: &CloneConfig{
				Template:   "template name",
				MacAddress: "some mac address",
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "'network' is required when 'mac_address' is specified",
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
					t.Fatalf("Config preprare should not fail: %s", errs[0])
				}
			}
		})
	}
}

func TestStepCreateVM_Run(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	driverMock := driver.NewDriverMock()
	state.Put("driver", driverMock)
	step := basicStepCloneVM()
	step.Force = true
	vmPath := path.Join(step.Location.Folder, step.Location.VMName)
	vmMock := new(driver.VirtualMachineMock)
	driverMock.VM = vmMock

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

	if !driverMock.FindVMCalled {
		t.Fatalf("driver.FindVM should be called.")
	}
	if !vmMock.CloneCalled {
		t.Fatalf("vm.Clone should be called.")
	}

	if diff := cmp.Diff(vmMock.CloneConfig, driverCreateConfig(step.Config, step.Location)); diff != "" {
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

//
//func TestStepCreateVM_RunHalt(t *testing.T) {
//	state := basicStateBag()
//	step := basicStepCreateVM()
//
//	// PreCleanVM fails
//	driverMock := driver.NewDriverMock()
//	driverMock.PreCleanShouldFail = true
//	state.Put("driver", driverMock)
//	if action := step.Run(context.TODO(), state); action != multistep.ActionHalt {
//		t.Fatalf("Step should halt.")
//	}
//	if !driverMock.PreCleanVMCalled {
//		t.Fatalf("driver.PreCleanVM should be called")
//	}
//
//	// CreateVM fails
//	driverMock = driver.NewDriverMock()
//	driverMock.CreateVMShouldFail = true
//	state.Put("driver", driverMock)
//	if action := step.Run(context.TODO(), state); action != multistep.ActionHalt {
//		t.Fatalf("Step should halt.")
//	}
//	if !driverMock.PreCleanVMCalled {
//		t.Fatalf("driver.PreCleanVM should be called")
//	}
//	if !driverMock.CreateVMCalled {
//		t.Fatalf("driver.PreCleanVM should be called")
//	}
//	if _, ok := state.GetOk("vm"); ok {
//		t.Fatal("state should not contain a VM")
//	}
//}
//
//func TestStepCreateVM_Cleanup(t *testing.T) {
//	state := basicStateBag()
//	step := basicStepCreateVM()
//	vm := new(driver.VirtualMachineMock)
//	state.Put("vm", vm)
//
//	// Clean up when state is cancelled
//	state.Put(multistep.StateCancelled, true)
//	step.Cleanup(state)
//	if !vm.DestroyCalled {
//		t.Fatalf("vm.Destroy should be called")
//	}
//	vm.DestroyCalled = false
//	state.Remove(multistep.StateCancelled)
//
//	// Clean up when state is halted
//	state.Put(multistep.StateHalted, true)
//	step.Cleanup(state)
//	if !vm.DestroyCalled {
//		t.Fatalf("vm.Destroy should be called")
//	}
//	vm.DestroyCalled = false
//	state.Remove(multistep.StateHalted)
//
//	// Clean up when state is destroy_vm is set
//	state.Put("destroy_vm", true)
//	step.Cleanup(state)
//	if !vm.DestroyCalled {
//		t.Fatalf("vm.Destroy should be called")
//	}
//	vm.DestroyCalled = false
//	state.Remove("destroy_vm")
//
//	// Don't clean up if state is not set with previous values
//	step.Cleanup(state)
//	if vm.DestroyCalled {
//		t.Fatalf("vm.Destroy should not be called")
//	}
//
//	// Destroy fail
//	errorBuffer := &strings.Builder{}
//	ui := &packersdk.BasicUi{
//		Reader:      strings.NewReader(""),
//		Writer:      ioutil.Discard,
//		ErrorWriter: errorBuffer,
//	}
//	state.Put("ui", ui)
//	state.Put(multistep.StateCancelled, true)
//	vm.DestroyError = errors.New("destroy failed")
//
//	step.Cleanup(state)
//	if !vm.DestroyCalled {
//		t.Fatalf("vm.Destroy should be called")
//	}
//	if !strings.Contains(errorBuffer.String(), vm.DestroyError.Error()) {
//		t.Fatalf("Destroy should fail with error message '%s' but failed with '%s'", vm.DestroyError.Error(), errorBuffer.String())
//	}
//	vm.DestroyCalled = false
//	state.Remove(multistep.StateCancelled)
//
//	// Should not destroy if VM is not set
//	state.Remove("vm")
//	state.Put(multistep.StateCancelled, true)
//	step.Cleanup(state)
//	if vm.DestroyCalled {
//		t.Fatalf("vm.Destroy should not be called")
//	}
//}

func basicStepCloneVM() *StepCloneVM {
	step := &StepCloneVM{
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

func createConfig() *CloneConfig {
	return &CloneConfig{
		Template: "template name",
		StorageConfig: common.StorageConfig{
			DiskControllerType: []string{"pvscsi"},
			Storage: []common.DiskConfig{
				{
					DiskSize:            32768,
					DiskThinProvisioned: true,
				},
			},
		},
	}
}

func driverCreateConfig(config *CloneConfig, location *common.LocationConfig) *driver.CloneConfig {
	var disks []driver.Disk
	for _, disk := range config.StorageConfig.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	return &driver.CloneConfig{
		StorageConfig: driver.StorageConfig{
			DiskControllerType: config.StorageConfig.DiskControllerType,
			Storage:            disks,
		},
		Annotation:      config.Notes,
		Name:            location.VMName,
		Folder:          location.Folder,
		Cluster:         location.Cluster,
		Host:            location.Host,
		ResourcePool:    location.ResourcePool,
		Datastore:       location.Datastore,
		LinkedClone:     config.LinkedClone,
		Network:         config.Network,
		MacAddress:      config.MacAddress,
		VAppProperties:  config.VAppConfig.Properties,
		PrimaryDiskSize: config.DiskSize,
	}
}

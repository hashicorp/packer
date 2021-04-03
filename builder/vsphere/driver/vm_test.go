package driver

import (
	"context"
	"testing"

	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
)

// ReconfigureFail changes the behavior of simulator.VirtualMachine
type ReconfigureFail struct {
	*simulator.VirtualMachine
}

// Override simulator.VirtualMachine.ReconfigVMTask to inject faults
func (vm *ReconfigureFail) ReconfigVMTask(req *types.ReconfigVM_Task) soap.HasFault {
	task := simulator.CreateTask(req.This, "reconfigure", func(*simulator.Task) (types.AnyType, types.BaseMethodFault) {
		return nil, &types.TaskInProgress{}
	})

	return &methods.ReconfigVM_TaskBody{
		Res: &types.ReconfigVM_TaskResponse{
			Returnval: task.Run(),
		},
	}
}

func TestVirtualMachineDriver_Configure(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	vm, machine := sim.ChooseSimulatorPreCreatedVM()

	// Happy test
	hardwareConfig := &HardwareConfig{
		CPUs:           1,
		CpuCores:       1,
		CPUReservation: 2500,
		CPULimit:       1,
		RAM:            1024,
		RAMReserveAll:  true,
		VideoRAM:       512,
		VGPUProfile:    "grid_m10-8q",
		Firmware:       "efi-secure",
		ForceBIOSSetup: true,
	}
	if err = vm.Configure(hardwareConfig); err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	//Fail test
	//Wrap the existing vm object with the mocked reconfigure task which will return a fault
	simulator.Map.Put(&ReconfigureFail{machine})
	if err = vm.Configure(&HardwareConfig{}); err == nil {
		t.Fatalf("Configure should fail")
	}
}

func TestVirtualMachineDriver_CreateVMWithMultipleDisks(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	_, datastore := sim.ChooseSimulatorPreCreatedDatastore()

	config := &CreateConfig{
		Name:      "mock name",
		Host:      "DC0_H0",
		Datastore: datastore.Name,
		NICs: []NIC{
			{
				Network:     "VM Network",
				NetworkCard: "vmxnet3",
			},
		},
		StorageConfig: StorageConfig{
			DiskControllerType: []string{"pvscsi"},
			Storage: []Disk{
				{
					DiskSize:            3072,
					DiskThinProvisioned: true,
					ControllerIndex:     0,
				},
				{
					DiskSize:            20480,
					DiskThinProvisioned: true,
					ControllerIndex:     0,
				},
			},
		},
	}

	vm, err := sim.driver.CreateVM(config)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}

	devices, err := vm.Devices()
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}

	var disks []*types.VirtualDisk
	for _, device := range devices {
		switch d := device.(type) {
		case *types.VirtualDisk:
			disks = append(disks, d)
		}
	}

	if len(disks) != 2 {
		t.Fatalf("unexpected number of devices")
	}
}

func TestVirtualMachineDriver_CloneWithPrimaryDiskResize(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	_, datastore := sim.ChooseSimulatorPreCreatedDatastore()
	vm, _ := sim.ChooseSimulatorPreCreatedVM()

	config := &CloneConfig{
		Name:            "mock name",
		Host:            "DC0_H0",
		Datastore:       datastore.Name,
		PrimaryDiskSize: 204800,
		StorageConfig: StorageConfig{
			DiskControllerType: []string{"pvscsi"},
			Storage: []Disk{
				{
					DiskSize:            3072,
					DiskThinProvisioned: true,
					ControllerIndex:     0,
				},
				{
					DiskSize:            20480,
					DiskThinProvisioned: true,
					ControllerIndex:     0,
				},
			},
		},
	}

	clonedVM, err := vm.Clone(context.TODO(), config)
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}

	devices, err := clonedVM.Devices()
	if err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}

	var disks []*types.VirtualDisk
	for _, device := range devices {
		switch d := device.(type) {
		case *types.VirtualDisk:
			disks = append(disks, d)
		}
	}

	if len(disks) != 3 {
		t.Fatalf("unexpected number of devices")
	}

	if disks[0].CapacityInKB != config.PrimaryDiskSize*1024 {
		t.Fatalf("unexpected disk size for primary disk: %d", disks[0].CapacityInKB)
	}
	if disks[1].CapacityInKB != config.StorageConfig.Storage[0].DiskSize*1024 {
		t.Fatalf("unexpected disk size for primary disk: %d", disks[1].CapacityInKB)
	}
	if disks[2].CapacityInKB != config.StorageConfig.Storage[1].DiskSize*1024 {
		t.Fatalf("unexpected disk size for primary disk: %d", disks[2].CapacityInKB)
	}
}

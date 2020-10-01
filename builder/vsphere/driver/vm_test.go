package driver

import (
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

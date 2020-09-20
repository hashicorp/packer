package driver

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/types"
)

func TestVirtualMachineDriver_FindAndAddSATAController(t *testing.T) {
	model := simulator.VPX()
	model.Machine = 1
	defer model.Remove()

	s, err := NewSimulatorServer(model)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer s.Close()

	driverSim, err := NewSimulatorDriver(s)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	vm := ChooseSimulatorPreCreatedVM(driverSim)

	_, err = vm.FindSATAController()
	if err != nil && !strings.Contains(err.Error(), "no available SATA controller") {
		t.Fatalf("unexpected error: %s", err.Error())
	}
	if err == nil {
		t.Fatalf("vm should not have sata controller")
	}

	if err := vm.AddSATAController(); err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	sc, err := vm.FindSATAController()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if sc == nil {
		t.Fatalf("SATA controller wasn't added properly")
	}
}

func TestVirtualMachineDriver_CreateAndRemoveCdrom(t *testing.T) {
	model := simulator.VPX()
	model.Machine = 1
	defer model.Remove()

	s, err := NewSimulatorServer(model)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer s.Close()

	driverSim, err := NewSimulatorDriver(s)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	//Simulator shortcut to choose any pre created VM.
	vm := ChooseSimulatorPreCreatedVM(driverSim)

	// Add SATA Controller
	if err := vm.AddSATAController(); err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	// Verify if controller was created
	sc, err := vm.FindSATAController()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if sc == nil {
		t.Fatalf("SATA controller wasn't added properly")
	}

	// Create CDROM
	controller := sc.GetVirtualController()
	cdrom, err := vm.CreateCdrom(controller)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if cdrom == nil {
		t.Fatalf("CDrom wasn't created properly")
	}

	// Verify if CDROM was created
	devices, err := vm.Devices()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	if len(cdroms) != 1 {
		t.Fatalf("unexpected numbers of cdrom: %d", len(cdroms))
	}

	// Remove CDROM
	err = vm.RemoveCdroms()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	// Verify if CDROM was removed
	devices, err = vm.Devices()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	cdroms = devices.SelectByType((*types.VirtualCdrom)(nil))
	if len(cdroms) != 0 {
		t.Fatalf("unexpected numbers of cdrom: %d", len(cdroms))
	}
}

func TestVirtualMachineDriver_EjectCdrom(t *testing.T) {
	model := simulator.VPX()
	model.Machine = 1
	defer model.Remove()

	s, err := NewSimulatorServer(model)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer s.Close()

	driverSim, err := NewSimulatorDriver(s)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	//Simulator shortcut to choose any pre created VM.
	vm := ChooseSimulatorPreCreatedVM(driverSim)

	// Add SATA Controller
	if err := vm.AddSATAController(); err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}

	// Verify if controller was created
	sc, err := vm.FindSATAController()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if sc == nil {
		t.Fatalf("SATA controller wasn't added properly")
	}

	// Create CDROM
	controller := sc.GetVirtualController()
	cdrom, err := vm.CreateCdrom(controller)
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	if cdrom == nil {
		t.Fatalf("CDrom wasn't created properly")
	}

	// Verify if CDROM was created
	devices, err := vm.Devices()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	if len(cdroms) != 1 {
		t.Fatalf("unexpected numbers of cdrom: %d", len(cdroms))
	}

	// Remove CDROM
	err = vm.EjectCdroms()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	// Verify if CDROM was removed
	devices, err = vm.Devices()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	cdroms = devices.SelectByType((*types.VirtualCdrom)(nil))
	if len(cdroms) != 1 {
		t.Fatalf("unexpected numbers of cdrom: %d", len(cdroms))
	}
	cd, ok := cdroms[0].(*types.VirtualCdrom)
	if !ok {
		t.Fatalf("Wrong cdrom type")
	}
	if diff := cmp.Diff(cd.Backing, &types.VirtualCdromRemotePassthroughBackingInfo{}); diff != "" {
		t.Fatalf("Wrong cdrom backing info: %s", diff)
	}
	if diff := cmp.Diff(cd.Connectable, &types.VirtualDeviceConnectInfo{}); diff != "" {
		t.Fatalf("Wrong cdrom connect info: %s", diff)
	}
}

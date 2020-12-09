package common

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

type parseBootDeviceIdentifierTest struct {
	generation         uint
	deviceIdentifier   string
	controllerType     string
	controllerNumber   uint
	controllerLocation uint
	failInParse        bool // true if ParseBootDeviceIdentifier should return an error
	haltStep           bool // true if Step.Run should return Halt action
	shouldCallSet      bool // true if driver.SetFirstBootDevice should have been called
	setDvdProps        bool // true to set DvdDeviceProperties state
}

var parseIdentifierTests = [...]parseBootDeviceIdentifierTest{
	{1, "IDE", "IDE", 0, 0, false, false, true, false},
	{1, "idE", "IDE", 0, 0, false, false, true, false},
	{1, "CD", "CD", 0, 0, false, false, false, false},
	{1, "CD", "CD", 0, 0, false, false, true, true},
	{1, "cD", "CD", 0, 0, false, false, false, false},
	{1, "DVD", "CD", 0, 0, false, false, false, false},
	{1, "DVD", "CD", 0, 0, false, false, true, true},
	{1, "Dvd", "CD", 0, 0, false, false, false, false},
	{1, "FLOPPY", "FLOPPY", 0, 0, false, false, true, false},
	{1, "FloppY", "FLOPPY", 0, 0, false, false, true, false},
	{1, "NET", "NET", 0, 0, false, false, true, false},
	{1, "net", "NET", 0, 0, false, false, true, false},
	{1, "", "", 0, 0, true, false, false, false},
	{1, "bad", "", 0, 0, true, true, false, false},
	{1, "IDE:0:0", "", 0, 0, true, true, true, false},
	{1, "SCSI:0:0", "", 0, 0, true, true, true, false},
	{2, "IDE", "", 0, 0, true, true, true, false},
	{2, "idE", "", 0, 0, true, true, true, false},
	{2, "CD", "CD", 0, 0, false, false, false, false},
	{2, "CD", "CD", 0, 0, false, false, true, true},
	{2, "cD", "CD", 0, 0, false, false, false, false},
	{2, "DVD", "CD", 0, 0, false, false, false, false},
	{2, "DVD", "CD", 0, 0, false, false, true, true},
	{2, "Dvd", "CD", 0, 0, false, false, false, false},
	{2, "FLOPPY", "", 0, 0, true, true, true, false},
	{2, "FloppY", "", 0, 0, true, true, true, false},
	{2, "NET", "NET", 0, 0, false, false, true, false},
	{2, "net", "NET", 0, 0, false, false, true, false},
	{2, "", "", 0, 0, true, false, false, false},
	{2, "bad", "", 0, 0, true, true, false, false},
	{2, "IDE:0:0", "IDE", 0, 0, false, false, true, false},
	{2, "SCSI:0:0", "SCSI", 0, 0, false, false, true, false},
	{2, "Ide:0:0", "IDE", 0, 0, false, false, true, false},
	{2, "sCsI:0:0", "SCSI", 0, 0, false, false, true, false},
	{2, "IDEscsi:0:0", "", 0, 0, true, true, false, false},
	{2, "SCSIide:0:0", "", 0, 0, true, true, false, false},
	{2, "IDE:0", "", 0, 0, true, true, false, false},
	{2, "SCSI:0", "", 0, 0, true, true, false, false},
	{2, "IDE:0:a", "", 0, 0, true, true, false, false},
	{2, "SCSI:0:a", "", 0, 0, true, true, false, false},
	{2, "IDE:0:653", "", 0, 0, true, true, false, false},
	{2, "SCSI:-10:0", "", 0, 0, true, true, false, false},
}

func TestStepSetFirstBootDevice_impl(t *testing.T) {
	var _ multistep.Step = new(StepSetFirstBootDevice)
}

func TestStepSetFirstBootDevice_ParseIdentifier(t *testing.T) {

	for _, identifierTest := range parseIdentifierTests {

		controllerType, controllerNumber, controllerLocation, err := ParseBootDeviceIdentifier(
			identifierTest.deviceIdentifier,
			identifierTest.generation)

		if (err != nil) != identifierTest.failInParse {

			t.Fatalf("Test %q (gen %v): failInParse: %v but err: %v", identifierTest.deviceIdentifier,
				identifierTest.generation, identifierTest.failInParse, err)

		}

		switch {

		case controllerType != identifierTest.controllerType:
			t.Fatalf("Test %q (gen %v): controllerType: %q != %q", identifierTest.deviceIdentifier, identifierTest.generation,
				identifierTest.controllerType, controllerType)

		case controllerNumber != identifierTest.controllerNumber:
			t.Fatalf("Test %q (gen %v): controllerNumber: %v != %v", identifierTest.deviceIdentifier, identifierTest.generation,
				identifierTest.controllerNumber, controllerNumber)

		case controllerLocation != identifierTest.controllerLocation:
			t.Fatalf("Test %q (gen %v): controllerLocation: %v != %v", identifierTest.deviceIdentifier, identifierTest.generation,
				identifierTest.controllerLocation, controllerLocation)

		}
	}
}

func TestStepSetFirstBootDevice(t *testing.T) {

	step := new(StepSetFirstBootDevice)

	for _, identifierTest := range parseIdentifierTests {

		state := testState(t)
		driver := state.Get("driver").(*DriverMock)

		// requires the vmName state value
		vmName := "foo"
		state.Put("vmName", vmName)

		// pretend that we mounted a DVD somewhere (CD:0:0)
		if identifierTest.setDvdProps {
			var dvdControllerProperties DvdControllerProperties
			dvdControllerProperties.ControllerNumber = 0
			dvdControllerProperties.ControllerLocation = 0
			dvdControllerProperties.Existing = false
			state.Put("os.dvd.properties", dvdControllerProperties)
		}

		step.Generation = identifierTest.generation
		step.FirstBootDevice = identifierTest.deviceIdentifier

		action := step.Run(context.Background(), state)
		if (action != multistep.ActionContinue) != identifierTest.haltStep {
			t.Fatalf("Test %q (gen %v): Bad action: %v", identifierTest.deviceIdentifier, identifierTest.generation, action)
		}

		if identifierTest.haltStep {

			if _, ok := state.GetOk("error"); !ok {
				t.Fatalf("Test %q (gen %v): Should have error", identifierTest.deviceIdentifier, identifierTest.generation)
			}

			// don't perform the remaining checks..
			continue

		} else {

			if _, ok := state.GetOk("error"); ok {
				t.Fatalf("Test %q (gen %v): Should NOT have error", identifierTest.deviceIdentifier, identifierTest.generation)
			}

		}

		if driver.SetFirstBootDevice_Called != identifierTest.shouldCallSet {
			if identifierTest.shouldCallSet {
				t.Fatalf("Test %q (gen %v): Should have called SetFirstBootDevice", identifierTest.deviceIdentifier, identifierTest.generation)
			}

			t.Fatalf("Test %q (gen %v): Should NOT have called SetFirstBootDevice", identifierTest.deviceIdentifier, identifierTest.generation)
		}

		if (driver.SetFirstBootDevice_Called) &&
			((driver.SetFirstBootDevice_VmName != vmName) ||
				(driver.SetFirstBootDevice_ControllerType != identifierTest.controllerType) ||
				(driver.SetFirstBootDevice_ControllerNumber != identifierTest.controllerNumber) ||
				(driver.SetFirstBootDevice_ControllerLocation != identifierTest.controllerLocation) ||
				(driver.SetFirstBootDevice_Generation != identifierTest.generation)) {

			t.Fatalf("Test %q (gen %v): Called SetFirstBootDevice with unexpected arguments.", identifierTest.deviceIdentifier, identifierTest.generation)

		}
	}
}

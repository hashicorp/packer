package common

import (
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepSetFirstBootDevice_impl(t *testing.T) {
	var _ multistep.Step = new(StepSetFirstBootDevice)
}

func TestStepSetFirstBootDevice(t *testing.T) {
//	t.Fatal("Fail IT!")
}

type parseBootDeviceIdentifierTest struct {
	generation         uint
	deviceIdentifier   string
	controllerType     string
	controllerNumber   uint
	controllerLocation uint
	shouldError        bool
}

func TestStepSetFirstBootDevice_ParseIdentifier(t *testing.T) {

	identifierTests := [...]parseBootDeviceIdentifierTest{
		{1, "IDE", "IDE", 0, 0, false},
		{1, "idE", "IDE", 0, 0, false},
		{1, "CD", "CD", 0, 0, false},
		{1, "cD", "CD", 0, 0, false},
		{1, "DVD", "CD", 0, 0, false},
		{1, "Dvd", "CD", 0, 0, false},
		{1, "FLOPPY", "FLOPPY", 0, 0, false},
		{1, "FloppY", "FLOPPY", 0, 0, false},
		{1, "NET", "NET", 0, 0, false},
		{1, "net", "NET", 0, 0, false},
		{1, "", "", 0, 0, true},
		{1, "bad", "", 0, 0, true},
		{1, "IDE:0:0", "", 0, 0, true},
		{1, "SCSI:0:0", "", 0, 0, true},
	}

	for _, identifierTest := range identifierTests {

		controllerType, controllerNumber, controllerLocation, err := ParseBootDeviceIdentifier(
			identifierTest.deviceIdentifier,
			identifierTest.generation)

		if (err != nil) != identifierTest.shouldError {

			t.Fatalf("Test %q (gen %v): shouldError: %v but err: %v", identifierTest.deviceIdentifier, 
				identifierTest.generation, identifierTest.shouldError, err)
			
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
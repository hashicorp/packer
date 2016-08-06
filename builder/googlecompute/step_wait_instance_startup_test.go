package googlecompute

import (
	"github.com/mitchellh/multistep"
	"testing"
)

func TestStepWaitInstanceStartup(t *testing.T) {
	state := testState(t)
	step := new(StepWaitInstanceStartup)
	config := state.Get("config").(*Config)
	driver := state.Get("driver").(*DriverMock)
	
	testZone := "test-zone"
	testInstanceName := "test-instance-name"

	config.Zone = testZone
	state.Put("instance_name", testInstanceName)
	// The done log triggers step completion.
	driver.GetSerialPortOutputResult = StartupScriptDoneLog
	
	// Run the step.
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("StepWaitInstanceStartup did not return a Continue action: %#v", action)
	}
	
	// Check that GetSerialPortOutput was called properly.
	if driver.GetSerialPortOutputZone != testZone {
		t.Fatalf(
			"GetSerialPortOutput wrong zone. Expected: %s, Actual: %s", driver.GetSerialPortOutputZone,
			testZone)
	}
	if driver.GetSerialPortOutputName != testInstanceName {
		t.Fatalf(
			"GetSerialPortOutput wrong instance name. Expected: %s, Actual: %s", driver.GetSerialPortOutputName,
			testInstanceName)
	}
}
package common

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

type bootOrderTest struct {
	bootOrder []string
}

var bootOrderTests = [...]bootOrderTest{
	{[]string{"SCSI:0:0"}},
}

func TestStepSetBootOrder(t *testing.T) {
	step := new(StepSetBootOrder)

	for _, d := range bootOrderTests {
		state := testState(t)
		driver := state.Get("driver").(*DriverMock)
		vmName := "test"

		state.Put("vmName", vmName)
		step.BootOrder = d.bootOrder

		action := step.Run(context.Background(), state)

		if multistep.ActionContinue != action {
			t.Fatalf("Should have returned action %v but got %v", multistep.ActionContinue, action)
		}

		if vmName != driver.SetBootOrder_VmName {
			t.Fatalf("Should have set VmName to %v but got %v", vmName, driver.SetBootOrder_VmName)
		}

		if !driver.SetBootOrder_Called {
			t.Fatalf("Should have called SetBootOrder")
		}

		if !reflect.DeepEqual(d.bootOrder, driver.SetBootOrder_BootOrder) {
			t.Fatalf("Should have set BootOrder to %v but got %v", d.bootOrder, driver.SetBootOrder_BootOrder)
		}
	}
}

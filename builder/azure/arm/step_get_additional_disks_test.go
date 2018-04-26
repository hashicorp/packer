package arm

import (
	"context"
	"fmt"
	"testing"

	"github.com/Azure/azure-sdk-for-go/arm/compute"

	"github.com/hashicorp/packer/builder/azure/common/constants"

	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepGetAdditionalDiskShouldFailIfGetFails(t *testing.T) {
	var testSubject = &StepGetDataDisk{
		query: func(string, string) (compute.VirtualMachine, error) {
			return createVirtualMachineWithDataDisksFromUri("test.vhd"), fmt.Errorf("!! Unit Test FAIL !!")
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepGetAdditionalDisks()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepGetAdditionalDiskShouldPassIfGetPasses(t *testing.T) {
	var testSubject = &StepGetDataDisk{
		query: func(string, string) (compute.VirtualMachine, error) {
			return createVirtualMachineWithDataDisksFromUri("test.vhd"), nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepGetAdditionalDisks()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func TestStepGetAdditionalDiskShouldTakeValidateArgumentsFromStateBag(t *testing.T) {
	var actualResourceGroupName string
	var actualComputeName string

	var testSubject = &StepGetDataDisk{
		query: func(resourceGroupName string, computeName string) (compute.VirtualMachine, error) {
			actualResourceGroupName = resourceGroupName
			actualComputeName = computeName

			return createVirtualMachineWithDataDisksFromUri("test.vhd"), nil
		},
		say:   func(message string) {},
		error: func(e error) {},
	}

	stateBag := createTestStateBagStepGetAdditionalDisks()
	var result = testSubject.Run(context.Background(), stateBag)

	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	var expectedComputeName = stateBag.Get(constants.ArmComputeName).(string)
	var expectedResourceGroupName = stateBag.Get(constants.ArmResourceGroupName).(string)

	if actualComputeName != expectedComputeName {
		t.Fatal("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	if actualResourceGroupName != expectedResourceGroupName {
		t.Fatal("Expected the step to source 'constants.ArmResourceGroupName' from the state bag, but it did not.")
	}

	expectedAdditionalDiskVhds, ok := stateBag.GetOk(constants.ArmAdditionalDiskVhds)
	if !ok {
		t.Fatalf("Expected the state bag to have a value for '%s', but it did not.", constants.ArmAdditionalDiskVhds)
	}

	expectedAdditionalDiskVhd := expectedAdditionalDiskVhds.([]string)
	if expectedAdditionalDiskVhd[0] != "test.vhd" {
		t.Fatalf("Expected the value of stateBag[%s] to be 'test.vhd', but got '%s'.", constants.ArmAdditionalDiskVhds, expectedAdditionalDiskVhd[0])
	}
}

func createTestStateBagStepGetAdditionalDisks() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmComputeName, "Unit Test: ComputeName")
	stateBag.Put(constants.ArmResourceGroupName, "Unit Test: ResourceGroupName")

	return stateBag
}

func createVirtualMachineWithDataDisksFromUri(vhdUri string) compute.VirtualMachine {
	vm := compute.VirtualMachine{
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			StorageProfile: &compute.StorageProfile{
				OsDisk: &compute.OSDisk{
					Vhd: &compute.VirtualHardDisk{
						URI: &vhdUri,
					},
				},
				DataDisks: &[]compute.DataDisk{
					{
						Vhd: &compute.VirtualHardDisk{
							URI: &vhdUri,
						},
					},
				},
			},
		},
	}

	return vm
}

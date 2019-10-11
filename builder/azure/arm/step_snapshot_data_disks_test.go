package arm

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepSnapshotDataDisksShouldFailIfSnapshotFails(t *testing.T) {
	var testSubject = &StepSnapshotDataDisks{
		create: func(context.Context, string, string, string, map[string]*string, string) error {
			return fmt.Errorf("!! Unit Test FAIL !!")
		},
		say:    func(message string) {},
		error:  func(e error) {},
		enable: func() bool { return true },
	}

	stateBag := createTestStateBagStepSnapshotDataDisks()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionHalt {
		t.Fatalf("Expected the step to return 'ActionHalt', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == false {
		t.Fatalf("Expected the step to set stateBag['%s'], but it was not.", constants.Error)
	}
}

func TestStepSnapshotDataDisksShouldNotExecute(t *testing.T) {
	var testSubject = &StepSnapshotDataDisks{
		create: func(context.Context, string, string, string, map[string]*string, string) error {
			return fmt.Errorf("!! Unit Test FAIL !!")
		},
		say:    func(message string) {},
		error:  func(e error) {},
		enable: func() bool { return false },
	}

	var result = testSubject.Run(context.Background(), nil)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}
}

func TestStepSnapshotDataDisksShouldPassIfSnapshotPasses(t *testing.T) {
	var testSubject = &StepSnapshotDataDisks{
		create: func(context.Context, string, string, string, map[string]*string, string) error {
			return nil
		},
		say:    func(message string) {},
		error:  func(e error) {},
		enable: func() bool { return true },
	}

	stateBag := createTestStateBagStepSnapshotDataDisks()

	var result = testSubject.Run(context.Background(), stateBag)
	if result != multistep.ActionContinue {
		t.Fatalf("Expected the step to return 'ActionContinue', but got '%d'.", result)
	}

	if _, ok := stateBag.GetOk(constants.Error); ok == true {
		t.Fatalf("Expected the step to not set stateBag['%s'], but it was.", constants.Error)
	}
}

func createTestStateBagStepSnapshotDataDisks() multistep.StateBag {
	stateBag := new(multistep.BasicStateBag)

	stateBag.Put(constants.ArmManagedImageResourceGroupName, "Unit Test: ResourceGroupName")
	stateBag.Put(constants.ArmLocation, "Unit Test: Location")

	value := "Unit Test: Tags"
	tags := map[string]*string{
		"tag01": &value,
	}
	stateBag.Put(constants.ArmTags, tags)

	stateBag.Put(constants.ArmAdditionalDiskVhds, []string{"subscriptions/123-456-789/resourceGroups/existingresourcegroup/providers/Microsoft.Compute/disks/osdisk"})
	stateBag.Put(constants.ArmManagedImageDataDiskSnapshotPrefix, "Unit Test: ManagedImageDataDiskSnapshotPrefix")

	return stateBag
}

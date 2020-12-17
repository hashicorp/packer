package common

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/assert"
)

func TestStepCreateDisks_impl(t *testing.T) {
	var _ multistep.Step = new(StepCreateDisks)
}

func strPtr(s string) *string {
	return &s
}
func NewTestCreateDiskStep() *StepCreateDisks {
	return &StepCreateDisks{
		OutputDir:          strPtr("output_dir"),
		CreateMainDisk:     true,
		DiskName:           "disk_name",
		MainDiskSize:       uint(1024),
		AdditionalDiskSize: []uint{},
		DiskAdapterType:    "fake_adapter",
		DiskTypeId:         "1",
	}
}

func TestStepCreateDisks_MainOnly(t *testing.T) {
	state := testState(t)
	step := NewTestCreateDiskStep()

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	driver := state.Get("driver").(*DriverMock)
	if !driver.CreateDiskCalled {
		t.Fatalf("Should have called create disk.")
	}

	diskFullPaths, ok := state.Get("disk_full_paths").([]string)
	if !ok {
		t.Fatalf("Should be able to load disk_full_paths from state")
	}

	assert.Equal(t, diskFullPaths, []string{filepath.Join("output_dir", "disk_name.vmdk")})

	// Cleanup
	step.Cleanup(state)
}

func TestStepCreateDisks_MainAndExtra(t *testing.T) {
	state := testState(t)
	step := NewTestCreateDiskStep()
	step.AdditionalDiskSize = []uint{1024, 2048, 4096}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	driver := state.Get("driver").(*DriverMock)
	if !driver.CreateDiskCalled {
		t.Fatalf("Should have called create disk.")
	}

	diskFullPaths, ok := state.Get("disk_full_paths").([]string)
	if !ok {
		t.Fatalf("Should be able to load disk_full_paths from state")
	}

	assert.Equal(t, diskFullPaths,
		[]string{
			filepath.Join("output_dir", "disk_name.vmdk"),
			filepath.Join("output_dir", "disk_name-1.vmdk"),
			filepath.Join("output_dir", "disk_name-2.vmdk"),
			filepath.Join("output_dir", "disk_name-3.vmdk"),
		})
	// Cleanup
	step.Cleanup(state)
}

func TestStepCreateDisks_ExtraOnly(t *testing.T) {
	state := testState(t)
	step := NewTestCreateDiskStep()
	step.CreateMainDisk = false
	step.AdditionalDiskSize = []uint{1024, 2048, 4096}

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	driver := state.Get("driver").(*DriverMock)
	if !driver.CreateDiskCalled {
		t.Fatalf("Should have called create disk.")
	}

	diskFullPaths, ok := state.Get("disk_full_paths").([]string)
	if !ok {
		t.Fatalf("Should be able to load disk_full_paths from state")
	}

	assert.Equal(t, diskFullPaths,
		[]string{
			filepath.Join("output_dir", "disk_name-1.vmdk"),
			filepath.Join("output_dir", "disk_name-2.vmdk"),
			filepath.Join("output_dir", "disk_name-3.vmdk"),
		})

	// Cleanup
	step.Cleanup(state)
}

func TestStepCreateDisks_Nothing(t *testing.T) {
	state := testState(t)
	step := NewTestCreateDiskStep()
	step.CreateMainDisk = false

	if action := step.Run(context.Background(), state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	driver := state.Get("driver").(*DriverMock)
	if driver.CreateDiskCalled {
		t.Fatalf("Should not have called create disk.")
	}

	// Cleanup
	step.Cleanup(state)
}

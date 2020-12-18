package qemu

import (
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/stretchr/testify/assert"
)

func copyTestState(t *testing.T, d *DriverMock) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", packersdk.TestUi(t))
	state.Put("driver", d)
	state.Put("iso_path", "example_source.qcow2")

	return state
}

func Test_StepCopySkip(t *testing.T) {
	testcases := []stepCopyDisk{
		stepCopyDisk{
			DiskImage:      false,
			UseBackingFile: false,
		},
		stepCopyDisk{
			DiskImage:      true,
			UseBackingFile: true,
		},
		stepCopyDisk{
			DiskImage:      false,
			UseBackingFile: true,
		},
	}

	for _, tc := range testcases {
		d := new(DriverMock)
		state := copyTestState(t, d)
		action := tc.Run(context.TODO(), state)
		if action != multistep.ActionContinue {
			t.Fatalf("Should have gotten an ActionContinue")
		}

		if d.CopyCalled || d.QemuImgCalled {
			t.Fatalf("Should have skipped step since DiskImage and UseBackingFile are not set")
		}
	}
}

func Test_StepCopyCalled(t *testing.T) {
	step := stepCopyDisk{
		DiskImage: true,
		Format:    "qcow2",
		VMName:    "output.qcow2",
	}

	d := new(DriverMock)
	state := copyTestState(t, d)
	action := step.Run(context.TODO(), state)
	if action != multistep.ActionContinue {
		t.Fatalf("Should have gotten an ActionContinue")
	}

	if !d.CopyCalled {
		t.Fatalf("Should have copied since all extensions are qcow2")
	}
	if d.QemuImgCalled {
		t.Fatalf("Should not have called qemu-img when formats match")
	}
}

func Test_StepQemuImgCalled(t *testing.T) {
	step := stepCopyDisk{
		DiskImage: true,
		Format:    "raw",
		VMName:    "output.qcow2",
	}

	d := new(DriverMock)
	state := copyTestState(t, d)
	action := step.Run(context.TODO(), state)
	if action != multistep.ActionContinue {
		t.Fatalf("Should have gotten an ActionContinue")
	}
	if d.CopyCalled {
		t.Fatalf("Should not have copied since extensions don't match")
	}
	if !d.QemuImgCalled {
		t.Fatalf("Should have called qemu-img since extensions don't match")
	}
}

func Test_StepQemuImgCalledWithExtraArgs(t *testing.T) {
	step := &stepCopyDisk{
		DiskImage: true,
		Format:    "raw",
		VMName:    "output.qcow2",
		QemuImgArgs: QemuImgArgs{
			Convert: []string{"-o", "preallocation=full"},
		},
	}

	d := new(DriverMock)
	state := copyTestState(t, d)
	action := step.Run(context.TODO(), state)
	if action != multistep.ActionContinue {
		t.Fatalf("Should have gotten an ActionContinue")
	}
	if d.CopyCalled {
		t.Fatalf("Should not have copied since extensions don't match")
	}
	if !d.QemuImgCalled {
		t.Fatalf("Should have called qemu-img since extensions don't match")
	}
	assert.Equal(
		t,
		d.QemuImgCalls,
		[]string{"convert", "-o", "preallocation=full", "-O", "raw",
			"example_source.qcow2", "output.qcow2"},
		"should have added user extra args")
}

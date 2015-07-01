package common

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/multistep"
)

func TestStepCleanVMX_impl(t *testing.T) {
	var _ multistep.Step = new(StepCleanVMX)
}

func TestStepCleanVMX(t *testing.T) {
	state := testState(t)
	step := new(StepCleanVMX)

	vmxPath := testVMXFile(t)
	defer os.Remove(vmxPath)
	state.Put("vmx_path", vmxPath)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}
}

func TestStepCleanVMX_floppyPath(t *testing.T) {
	state := testState(t)
	step := new(StepCleanVMX)

	vmxPath := testVMXFile(t)
	defer os.Remove(vmxPath)
	if err := ioutil.WriteFile(vmxPath, []byte(testVMXFloppyPath), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	state.Put("vmx_path", vmxPath)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the resulting data
	vmxContents, err := ioutil.ReadFile(vmxPath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	vmxData := ParseVMX(string(vmxContents))

	cases := []struct {
		Key   string
		Value string
	}{
		{"floppy0.present", "FALSE"},
		{"floppy0.fileType", ""},
		{"floppy0.fileName", ""},
	}

	for _, tc := range cases {
		if tc.Value == "" {
			if _, ok := vmxData[tc.Key]; ok {
				t.Fatalf("should not have key: %s", tc.Key)
			}
		} else {
			if vmxData[tc.Key] != tc.Value {
				t.Fatalf("bad: %s %#v", tc.Key, vmxData[tc.Key])
			}
		}
	}
}

func TestStepCleanVMX_isoPath(t *testing.T) {
	state := testState(t)
	step := new(StepCleanVMX)

	vmxPath := testVMXFile(t)
	defer os.Remove(vmxPath)
	if err := ioutil.WriteFile(vmxPath, []byte(testVMXISOPath), 0644); err != nil {
		t.Fatalf("err: %s", err)
	}

	state.Put("vmx_path", vmxPath)

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Test the resulting data
	vmxContents, err := ioutil.ReadFile(vmxPath)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	vmxData := ParseVMX(string(vmxContents))

	cases := []struct {
		Key   string
		Value string
	}{
		{"ide0:0.fileName", "auto detect"},
		{"ide0:0.deviceType", "cdrom-raw"},
		{"ide0:1.fileName", "bar"},
		{"foo", "bar"},
	}

	for _, tc := range cases {
		if tc.Value == "" {
			if _, ok := vmxData[tc.Key]; ok {
				t.Fatalf("should not have key: %s", tc.Key)
			}
		} else {
			if vmxData[tc.Key] != tc.Value {
				t.Fatalf("bad: %s %#v", tc.Key, vmxData[tc.Key])
			}
		}
	}
}

const testVMXFloppyPath = `
floppy0.present = "TRUE"
floppy0.fileType = "file"
`

const testVMXISOPath = `
ide0:0.deviceType = "cdrom-image"
ide0:0.fileName = "foo"
ide0:1.fileName = "bar"
foo = "bar"
`

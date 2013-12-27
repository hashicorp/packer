package common

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/multistep"
)

func testVMXFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()

	return tf.Name()
}

func TestStepConfigureVMX_impl(t *testing.T) {
	var _ multistep.Step = new(StepConfigureVMX)
}

func TestStepConfigureVMX(t *testing.T) {
	state := testState(t)
	step := new(StepConfigureVMX)
	step.CustomData = map[string]string{
		"foo": "bar",
	}

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
		// Stuff we set
		{"msg.autoanswer", "true"},
		{"uuid.action", "create"},

		// Custom data
		{"foo", "bar"},

		// Stuff that should NOT exist
		{"floppy0.present", ""},
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

func TestStepConfigureVMX_floppyPath(t *testing.T) {
	state := testState(t)
	step := new(StepConfigureVMX)

	vmxPath := testVMXFile(t)
	defer os.Remove(vmxPath)

	state.Put("floppy_path", "foo")
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
		{"floppy0.present", "TRUE"},
		{"floppy0.filetype", "file"},
		{"floppy0.filename", "foo"},
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

func TestStepConfigureVMX_generatedAddresses(t *testing.T) {
	state := testState(t)
	step := new(StepConfigureVMX)

	vmxPath := testVMXFile(t)
	defer os.Remove(vmxPath)

	err := WriteVMX(vmxPath, map[string]string{
		"foo": "bar",
		"ethernet0.generatedAddress":       "foo",
		"ethernet1.generatedAddress":       "foo",
		"ethernet1.generatedAddressOffset": "foo",
	})
	if err != nil {
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
		{"foo", "bar"},
		{"ethernet0.generatedaddress", ""},
		{"ethernet1.generatedaddress", ""},
		{"ethernet1.generatedaddressoffset", ""},
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

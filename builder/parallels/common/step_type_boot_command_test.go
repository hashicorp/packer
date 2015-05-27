package common

import (
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"strings"
	"testing"
)

func TestStepTypeBootCommand(t *testing.T) {
	state := testState(t)

	var bootcommand = []string{
		"1234567890-=<enter><wait>",
		"!@#$%^&*()_+<enter>",
		"qwertyuiop[]<enter>",
		"QWERTYUIOP{}<enter>",
		"asdfghjkl;'`<enter>",
		`ASDFGHJKL:"~<enter>`,
		"\\zxcvbnm,./<enter>",
		"|ZXCVBNM<>?<enter>",
		" <enter>",
	}

	step := StepTypeBootCommand{
		BootCommand:    bootcommand,
		HostInterfaces: []string{},
		VMName:         "myVM",
		Ctx:            *testConfigTemplate(t),
	}

	comm := new(packer.MockCommunicator)
	state.Put("communicator", comm)

	driver := state.Get("driver").(*DriverMock)
	driver.VersionResult = "foo"
	state.Put("http_port", uint(0))

	// Test the run
	if action := step.Run(state); action != multistep.ActionContinue {
		t.Fatalf("bad action: %#v", action)
	}
	if _, ok := state.GetOk("error"); ok {
		t.Fatal("should NOT have error")
	}

	// Verify
	var expected = [][]string{
		[]string{"02", "82", "03", "83", "04", "84", "05", "85", "06", "86", "07", "87", "08", "88", "09", "89", "0a", "8a", "0b", "8b", "0c", "8c", "0d", "8d", "1c", "9c"},
		[]string{},
		[]string{"2a", "02", "82", "aa", "2a", "03", "83", "aa", "2a", "04", "84", "aa", "2a", "05", "85", "aa", "2a", "06", "86", "aa", "2a", "07", "87", "aa", "2a", "08", "88", "aa", "2a", "09", "89", "aa", "2a", "0a", "8a", "aa", "2a", "0b", "8b", "aa", "2a", "0c", "8c", "aa", "2a", "0d", "8d", "aa", "1c", "9c"},
		[]string{"10", "90", "11", "91", "12", "92", "13", "93", "14", "94", "15", "95", "16", "96", "17", "97", "18", "98", "19", "99", "1a", "9a", "1b", "9b", "1c", "9c"},
		[]string{"2a", "10", "90", "aa", "2a", "11", "91", "aa", "2a", "12", "92", "aa", "2a", "13", "93", "aa", "2a", "14", "94", "aa", "2a", "15", "95", "aa", "2a", "16", "96", "aa", "2a", "17", "97", "aa", "2a", "18", "98", "aa", "2a", "19", "99", "aa", "2a", "1a", "9a", "aa", "2a", "1b", "9b", "aa", "1c", "9c"},
		[]string{"1e", "9e", "1f", "9f", "20", "a0", "21", "a1", "22", "a2", "23", "a3", "24", "a4", "25", "a5", "26", "a6", "27", "a7", "28", "a8", "29", "a9", "1c", "9c"},
		[]string{"2a", "1e", "9e", "aa", "2a", "1f", "9f", "aa", "2a", "20", "a0", "aa", "2a", "21", "a1", "aa", "2a", "22", "a2", "aa", "2a", "23", "a3", "aa", "2a", "24", "a4", "aa", "2a", "25", "a5", "aa", "2a", "26", "a6", "aa", "2a", "27", "a7", "aa", "2a", "28", "a8", "aa", "2a", "29", "a9", "aa", "1c", "9c"},
		[]string{"2b", "ab", "2c", "ac", "2d", "ad", "2e", "ae", "2f", "af", "30", "b0", "31", "b1", "32", "b2", "33", "b3", "34", "b4", "35", "b5", "1c", "9c"},
		[]string{"2a", "2b", "ab", "aa", "2a", "2c", "ac", "aa", "2a", "2d", "ad", "aa", "2a", "2e", "ae", "aa", "2a", "2f", "af", "aa", "2a", "30", "b0", "aa", "2a", "31", "b1", "aa", "2a", "32", "b2", "aa", "2a", "33", "b3", "aa", "2a", "34", "b4", "aa", "2a", "35", "b5", "aa", "1c", "9c"},
		[]string{"39", "b9", "1c", "9c"},
	}
	fail := false

	for i := range driver.SendKeyScanCodesCalls {
		t.Logf("prltype %s\n", strings.Join(driver.SendKeyScanCodesCalls[i], " "))
	}

	for i := range expected {
		for j := range expected[i] {
			if driver.SendKeyScanCodesCalls[i][j] != expected[i][j] {
				fail = true
			}
		}
	}
	if fail {
		t.Fatalf("Sent bad scancodes: %#v\n Expected: %#v", driver.SendKeyScanCodesCalls, expected)
	}
}

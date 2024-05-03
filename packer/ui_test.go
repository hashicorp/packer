// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package packer

import (
	"bytes"
	"strings"
	"testing"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// This reads the output from the bytes.Buffer in our test object
// and then resets the buffer.
func readWriter(ui *packersdk.BasicUi) (result string) {
	buffer := ui.Writer.(*bytes.Buffer)
	result = buffer.String()
	buffer.Reset()
	return
}

// Reset the input Reader than add some input to it.
func writeReader(ui *packersdk.BasicUi, input string) {
	buffer := ui.Reader.(*bytes.Buffer)
	buffer.WriteString(input)
}

func readErrorWriter(ui *packersdk.BasicUi) (result string) {
	buffer := ui.ErrorWriter.(*bytes.Buffer)
	result = buffer.String()
	buffer.Reset()
	return
}

func testUi() *packersdk.BasicUi {
	return &packersdk.BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: new(bytes.Buffer),
		TTY:         new(testTTY),
	}
}

type testTTY struct {
	say string
}

func (tty *testTTY) Close() error { return nil }
func (tty *testTTY) ReadString() (string, error) {
	return tty.say, nil
}

func TestColoredUi(t *testing.T) {
	bufferUi := testUi()
	ui := &ColoredUi{UiColorYellow, UiColorRed, bufferUi, &UiProgressBar{}}

	if !ui.supportsColors() {
		t.Skip("skipping for ui without color support")
	}

	ui.Say("foo")
	result := readWriter(bufferUi)
	if result != "\033[1;33mfoo\033[0m\n" {
		t.Fatalf("invalid output: %s", result)
	}

	ui.Message("foo")
	result = readWriter(bufferUi)
	if result != "\033[0;33mfoo\033[0m\n" {
		t.Fatalf("invalid output: %s", result)
	}

	ui.Error("foo")
	result = readWriter(bufferUi)
	if result != "" {
		t.Fatalf("invalid output: %s", result)
	}

	result = readErrorWriter(bufferUi)
	if result != "\033[1;31mfoo\033[0m\n" {
		t.Fatalf("invalid output: %s", result)
	}
}

func TestColoredUi_noColorEnv(t *testing.T) {
	bufferUi := testUi()
	ui := &ColoredUi{UiColorYellow, UiColorRed, bufferUi, &UiProgressBar{}}

	// Set the env var to get rid of the color
	t.Setenv("PACKER_NO_COLOR", "1")

	ui.Say("foo")
	result := readWriter(bufferUi)
	if result != "foo\n" {
		t.Fatalf("invalid output: %s", result)
	}

	ui.Message("foo")
	result = readWriter(bufferUi)
	if result != "foo\n" {
		t.Fatalf("invalid output: %s", result)
	}

	ui.Error("foo")
	result = readErrorWriter(bufferUi)
	if result != "foo\n" {
		t.Fatalf("invalid output: %s", result)
	}
}

func TestTargetedUI(t *testing.T) {
	bufferUi := testUi()
	targetedUi := &TargetedUI{
		Target: "foo",
		Ui:     bufferUi,
	}

	var actual, expected string
	targetedUi.Say("foo")
	actual = readWriter(bufferUi)
	expected = "==> foo: foo\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	targetedUi.Message("foo")
	actual = readWriter(bufferUi)
	expected = "    foo: foo\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	targetedUi.Error("bar")
	actual = readErrorWriter(bufferUi)
	expected = "==> foo: bar\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	targetedUi.Say("foo\nbar")
	actual = readWriter(bufferUi)
	expected = "==> foo: foo\n==> foo: bar\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestColoredUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &ColoredUi{}
	if _, ok := raw.(packersdk.Ui); !ok {
		t.Fatalf("ColoredUi must implement Ui")
	}
}

func TestTargetedUI_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &TargetedUI{}
	if _, ok := raw.(packersdk.Ui); !ok {
		t.Fatalf("TargetedUI must implement Ui")
	}
}

func TestBasicUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &packersdk.BasicUi{}
	if _, ok := raw.(packersdk.Ui); !ok {
		t.Fatalf("BasicUi must implement Ui")
	}
}

func TestBasicUi_Error(t *testing.T) {
	bufferUi := testUi()

	var actual, expected string
	bufferUi.Error("foo")
	actual = readErrorWriter(bufferUi)
	expected = "foo\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	bufferUi.ErrorWriter = nil
	bufferUi.Error("5")
	actual = readWriter(bufferUi)
	expected = "5\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestBasicUi_Say(t *testing.T) {
	bufferUi := testUi()

	var actual, expected string

	bufferUi.Say("foo")
	actual = readWriter(bufferUi)
	expected = "foo\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	bufferUi.Say("5")
	actual = readWriter(bufferUi)
	expected = "5\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestBasicUi_Ask(t *testing.T) {

	var actual, expected string
	var err error

	var testCases = []struct {
		Prompt, Input, Answer string
	}{
		{"[c]ontinue or [a]bort", "c\n", "c"},
		{"[c]ontinue or [a]bort", "c", "c"},
		// Empty input shouldn't give an error
		{"Name", "Joe Bloggs\n", "Joe Bloggs"},
		{"Name", "Joe Bloggs", "Joe Bloggs"},
		{"Name", "\n", ""},
	}

	for _, testCase := range testCases {
		// Because of the internal bufio we can't easily reset the input, so create a new one each time
		bufferUi := testUi()
		bufferUi.TTY = &testTTY{testCase.Input}

		actual, err = bufferUi.Ask(testCase.Prompt)
		if err != nil {
			t.Fatal(err)
		}

		if actual != testCase.Answer {
			t.Fatalf("bad answer: %#v", actual)
		}

		actual = readWriter(bufferUi)
		expected = testCase.Prompt + " "
		if actual != expected {
			t.Fatalf("bad prompt: %#v", actual)
		}
	}

}

func TestMachineReadableUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &MachineReadableUi{}
	if _, ok := raw.(packersdk.Ui); !ok {
		t.Fatalf("MachineReadableUi must implement Ui")
	}
}

func TestMachineReadableUi(t *testing.T) {
	var data, expected string

	buf := new(bytes.Buffer)
	ui := &MachineReadableUi{Writer: buf}

	// No target
	ui.Machine("foo", "bar", "baz")
	data = strings.SplitN(buf.String(), ",", 2)[1]
	expected = ",foo,bar,baz\n"
	if data != expected {
		t.Fatalf("bad: %s", data)
	}

	// Target
	buf.Reset()
	ui.Machine("mitchellh,foo", "bar", "baz")
	data = strings.SplitN(buf.String(), ",", 2)[1]
	expected = "mitchellh,foo,bar,baz\n"
	if data != expected {
		t.Fatalf("bad: %s", data)
	}

	// Commas
	buf.Reset()
	ui.Machine("foo", "foo,bar")
	data = strings.SplitN(buf.String(), ",", 2)[1]
	expected = ",foo,foo%!(PACKER_COMMA)bar\n"
	if data != expected {
		t.Fatalf("bad: %s", data)
	}

	// New lines
	buf.Reset()
	ui.Machine("foo", "foo\n")
	data = strings.SplitN(buf.String(), ",", 2)[1]
	expected = ",foo,foo\\n\n"
	if data != expected {
		t.Fatalf("bad: %#v", data)
	}
}

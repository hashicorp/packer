package packer

import (
	"bytes"
	"strings"
	"testing"
)

func testUi() *BasicUi {
	return &BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestColoredUi(t *testing.T) {
	bufferUi := testUi()
	ui := &ColoredUi{UiColorYellow, UiColorRed, bufferUi}

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
	if result != "\033[1;31mfoo\033[0m\n" {
		t.Fatalf("invalid output: %s", result)
	}
}

func TestTargettedUi(t *testing.T) {
	bufferUi := testUi()
	targettedUi := &TargettedUi{
		Target: "foo",
		Ui:     bufferUi,
	}

	var actual, expected string
	targettedUi.Say("foo")
	actual = readWriter(bufferUi)
	expected = "==> foo: foo\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	targettedUi.Message("foo")
	actual = readWriter(bufferUi)
	expected = "    foo: foo\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	targettedUi.Error("bar")
	actual = readWriter(bufferUi)
	expected = "==> foo: bar\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

	targettedUi.Say("foo\nbar")
	actual = readWriter(bufferUi)
	expected = "==> foo: foo\n==> foo: bar\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestColoredUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &ColoredUi{}
	if _, ok := raw.(Ui); !ok {
		t.Fatalf("ColoredUi must implement Ui")
	}
}

func TestTargettedUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &TargettedUi{}
	if _, ok := raw.(Ui); !ok {
		t.Fatalf("TargettedUi must implement Ui")
	}
}

func TestBasicUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &BasicUi{}
	if _, ok := raw.(Ui); !ok {
		t.Fatalf("BasicUi must implement Ui")
	}
}

func TestBasicUi_Error(t *testing.T) {
	bufferUi := testUi()

	var actual, expected string
	bufferUi.Error("foo")
	actual = readWriter(bufferUi)
	expected = "foo\n"
	if actual != expected {
		t.Fatalf("bad: %#v", actual)
	}

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

func TestMachineReadableUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &MachineReadableUi{}
	if _, ok := raw.(Ui); !ok {
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

// This reads the output from the bytes.Buffer in our test object
// and then resets the buffer.
func readWriter(ui *BasicUi) (result string) {
	buffer := ui.Writer.(*bytes.Buffer)
	result = buffer.String()
	buffer.Reset()
	return
}

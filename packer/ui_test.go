package packer

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"testing"
)

func testUi() *ReaderWriterUi {
	return &ReaderWriterUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}
}

func TestColoredUi(t *testing.T) {
	bufferUi := testUi()
	ui := &ColoredUi{UiColorYellow, UiColorRed, bufferUi}

	ui.Say("foo")
	result := readWriter(bufferUi)
	if result != "\033[1;33;40mfoo\033[0m\n" {
		t.Fatalf("invalid output: %s", result)
	}

	ui.Message("foo")
	result = readWriter(bufferUi)
	if result != "\033[0;33;40mfoo\033[0m\n" {
		t.Fatalf("invalid output: %s", result)
	}

	ui.Error("foo")
	result = readWriter(bufferUi)
	if result != "\033[1;31;40mfoo\033[0m\n" {
		t.Fatalf("invalid output: %s", result)
	}
}

func TestPrefixedUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	bufferUi := testUi()
	prefixUi := &PrefixedUi{"mitchell", "bar", bufferUi}

	prefixUi.Say("foo")
	assert.Equal(readWriter(bufferUi), "mitchell: foo\n", "should have prefix")

	prefixUi.Message("foo")
	assert.Equal(readWriter(bufferUi), "bar: foo\n", "should have prefix")

	prefixUi.Error("bar")
	assert.Equal(readWriter(bufferUi), "mitchell: bar\n", "should have prefix")
}

func TestColoredUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &ColoredUi{}
	if _, ok := raw.(Ui); !ok {
		t.Fatalf("ColoredUi must implement Ui")
	}
}

func TestPrefixedUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &PrefixedUi{}
	if _, ok := raw.(Ui); !ok {
		t.Fatalf("PrefixedUi must implement Ui")
	}
}

func TestReaderWriterUi_ImplUi(t *testing.T) {
	var raw interface{}
	raw = &ReaderWriterUi{}
	if _, ok := raw.(Ui); !ok {
		t.Fatalf("ReaderWriterUi must implement Ui")
	}
}

func TestReaderWriterUi_Error(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	bufferUi := testUi()

	bufferUi.Error("foo")
	assert.Equal(readWriter(bufferUi), "foo\n", "basic output")

	bufferUi.Error("5")
	assert.Equal(readWriter(bufferUi), "5\n", "formatting")
}

func TestReaderWriterUi_Say(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	bufferUi := testUi()

	bufferUi.Say("foo")
	assert.Equal(readWriter(bufferUi), "foo\n", "basic output")

	bufferUi.Say("5")
	assert.Equal(readWriter(bufferUi), "5\n", "formatting")
}

// This reads the output from the bytes.Buffer in our test object
// and then resets the buffer.
func readWriter(ui *ReaderWriterUi) (result string) {
	buffer := ui.Writer.(*bytes.Buffer)
	result = buffer.String()
	buffer.Reset()
	return
}

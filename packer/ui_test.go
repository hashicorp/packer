package packer

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"testing"
)

func testUi() *ReaderWriterUi {
	return &ReaderWriterUi{
		new(bytes.Buffer),
		new(bytes.Buffer),
	}
}

func TestPrefixedUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	bufferUi := testUi()
	prefixUi := &PrefixedUi{"mitchell", bufferUi}

	prefixUi.Say("foo")
	assert.Equal(readWriter(bufferUi), "mitchell: foo\n", "should have prefix")

	prefixUi.Error("bar")
	assert.Equal(readWriter(bufferUi), "mitchell: bar\n", "should have prefix")
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

	bufferUi.Error("%d", 5)
	assert.Equal(readWriter(bufferUi), "5\n", "formatting")
}

func TestReaderWriterUi_Say(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	bufferUi := testUi()

	bufferUi.Say("foo")
	assert.Equal(readWriter(bufferUi), "foo\n", "basic output")

	bufferUi.Say("%d", 5)
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

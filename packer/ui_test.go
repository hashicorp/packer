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

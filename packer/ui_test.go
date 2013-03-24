package packer

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"testing"
)

// Our test Ui that just writes to bytes.Buffers.
var bufferUi = &ReaderWriterUi{new(bytes.Buffer), new(bytes.Buffer)}

func TestReaderWriterUi_Say(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	bufferUi.Say("foo")
	assert.Equal(readWriter(bufferUi), "foo", "basic output")

	bufferUi.Say("%d", 5)
	assert.Equal(readWriter(bufferUi), "5", "formatting")
}

// This reads the output from the bytes.Buffer in our test object
// and then resets the buffer.
func readWriter(ui *ReaderWriterUi) (result string) {
	buffer := ui.Writer.(*bytes.Buffer)
	result = buffer.String()
	buffer.Reset()
	return
}

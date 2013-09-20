package shell

import (
	"bytes"
	"io"
	"testing"
)

func TestUnixReader_impl(t *testing.T) {
	var raw interface{}
	raw = new(UnixReader)
	if _, ok := raw.(io.Reader); !ok {
		t.Fatal("should be reader")
	}
}

func TestUnixReader(t *testing.T) {
	input := "one\r\ntwo\nthree\r\n"
	expected := "one\ntwo\nthree\n"

	r := &UnixReader{
		Reader: bytes.NewReader([]byte(input)),
	}

	result := new(bytes.Buffer)
	if _, err := io.Copy(result, r); err != nil {
		t.Fatalf("err: %s", err)
	}

	if result.String() != expected {
		t.Fatalf("bad: %#v", result.String())
	}
}

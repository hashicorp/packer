// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

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
	input := "one\r\ntwo\n\r\nthree\r\n"
	expected := "one\ntwo\n\nthree\n"

	unixReaderTest(t, input, expected)
}

func TestUnixReader_unixOnly(t *testing.T) {
	input := "\none\n\ntwo\nthree\n\n"
	expected := "\none\n\ntwo\nthree\n\n"

	unixReaderTest(t, input, expected)
}

func TestUnixReader_readsLastLine(t *testing.T) {
	input := "one\ntwo"
	expected := "one\ntwo\n"

	unixReaderTest(t, input, expected)
}

func unixReaderTest(t *testing.T, input string, expected string) {
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

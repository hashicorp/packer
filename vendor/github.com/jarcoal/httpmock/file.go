package httpmock

import (
	"fmt"
	"io/ioutil"
)

// File is a file name. The contents of this file is loaded on demand
// by the following methods.
//
// Note that:
//   file := httpmock.File("file.txt")
//   fmt.Printf("file: %s\n", file)
//
// prints the content of file "file.txt" as String() method is used.
//
// To print the file name, and not its content, simply do:
//   file := httpmock.File("file.txt")
//   fmt.Printf("file: %s\n", string(file))
type File string

// MarshalJSON implements json.Marshaler.
//
// Useful to be used in conjunction with NewJsonResponse() or
// NewJsonResponder() as in:
//   httpmock.NewJsonResponder(200, httpmock.File("body.json"))
func (f File) MarshalJSON() ([]byte, error) {
	return f.bytes()
}

func (f File) bytes() ([]byte, error) {
	return ioutil.ReadFile(string(f))
}

// Bytes returns the content of file as a []byte. If an error occurs
// during the opening or reading of the file, it panics.
//
// Useful to be used in conjunction with NewBytesResponse() or
// NewBytesResponder() as in:
//   httpmock.NewBytesResponder(200, httpmock.File("body.raw").Bytes())
func (f File) Bytes() []byte {
	b, err := f.bytes()
	if err != nil {
		panic(fmt.Sprintf("Cannot read %s: %s", string(f), err))
	}
	return b
}

// String returns the content of file as a string. If an error occurs
// during the opening or reading of the file, it panics.
//
// Useful to be used in conjunction with NewStringResponse() or
// NewStringResponder() as in:
//   httpmock.NewStringResponder(200, httpmock.File("body.txt").String())
func (f File) String() string {
	return string(f.Bytes())
}

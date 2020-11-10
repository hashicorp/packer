package hcl2template

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func TestHCL2Formatter_Format(t *testing.T) {
	tt := []struct {
		Name           string
		Path           string
		FormatExpected bool
	}{
		{Name: "Unformatted file", Path: "testdata/format/unformatted.pkr.hcl", FormatExpected: true},
		{Name: "Formatted file", Path: "testdata/format/formatted.pkr.hcl"},
		{Name: "Directory", Path: "testdata/format", FormatExpected: true},
	}

	for _, tc := range tt {
		tc := tc
		var buf bytes.Buffer
		f := NewHCL2Formatter()
		f.Output = &buf
		_, diags := f.Format(tc.Path)
		if diags.HasErrors() {
			t.Fatalf("the call to Format failed unexpectedly %s", diags.Error())
		}

		if buf.String() != "" && tc.FormatExpected == false {
			t.Errorf("Format(%q) should contain the name of the formatted file(s), but got %q", tc.Path, buf.String())
		}

	}
}

func TestHCL2Formatter_processFile(t *testing.T) {

	var buf bytes.Buffer
	f := NewHCL2Formatter()
	f.Output = &buf
	data, err := f.processFile("testdata/format/unformatted.pkr.hcl")
	if err != nil {
		t.Fatalf("the call to processFile failed unexpectedly %s", err)
	}

	formattedData, err := ioutil.ReadFile("testdata/format/formatted.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to open the formatted fixture %s", err)
	}

	if !bytes.Equal(data, formattedData) {
		t.Errorf("failed to format file")
	}

}

func TestHCL2Formatter_processFile_Write(t *testing.T) {

	var buf bytes.Buffer
	f := NewHCL2Formatter()
	f.Output = &buf
	f.Write = true

	hcl2data := `
source "amazon-ebs" "test" {
  name ="testsource"
}
`
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create tempfile for test %s", err)
	}
	defer os.Remove(tf.Name())

	_, _ = tf.Write([]byte(hcl2data))
	tf.Close()

	formattedData, err := f.processFile(tf.Name())
	if err != nil {
		t.Fatalf("the call to processFile failed unexpectedly %s", err)
	}

	//lets re-read the tempfile which should now be formatted
	data, err := ioutil.ReadFile(tf.Name())
	if err != nil {
		t.Fatalf("failed to open the newly formatted fixture %s", err)
	}

	if !bytes.Equal(data, formattedData) {
		t.Errorf("failed to format file %s", buf.String())
	}
}

func TestHCL2Formatter_processFile_ShwoDiff(t *testing.T) {

	var buf bytes.Buffer
	f := NewHCL2Formatter()
	f.Output = &buf
	f.ShowDiff = true

	data := `
source "amazon-ebs" "test" {
  name ="testsource"
}
`
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatalf("failed to create tempfile for test %s", err)
	}
	defer os.Remove(tf.Name())

	_, _ = tf.Write([]byte(data))
	tf.Close()

	formattedData, err := f.processFile(tf.Name())
	if err != nil {
		t.Fatalf("the call to processFile failed unexpectedly %s", err)
	}

	if bytes.Equal([]byte(data), formattedData) {
		t.Errorf("failed to format file %s", buf.String())
	}

	if !strings.Contains(buf.String(), "@@ -1,4 +1,4 @@") {
		t.Errorf("expected buf to contain a file diff, but instead we got %s", buf.String())
	}

}

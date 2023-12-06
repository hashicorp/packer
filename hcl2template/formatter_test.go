// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHCL2Formatter_Format(t *testing.T) {
	tt := []struct {
		Name           string
		Path           string
		FormatExpected bool
	}{
		{Name: "Unformatted file", Path: "testdata/format/unformatted.pkr.hcl", FormatExpected: true},
		{Name: "Unformatted vars file", Path: "testdata/format/unformatted.pkrvars.hcl", FormatExpected: true},
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

func TestHCL2Formatter_Format_Write(t *testing.T) {

	var buf bytes.Buffer
	f := NewHCL2Formatter()
	f.Output = &buf
	f.Write = true

	unformattedData, err := os.ReadFile("testdata/format/unformatted.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to open the unformatted fixture %s", err)
	}

	tf, err := os.CreateTemp("", "*.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to create tempfile for test %s", err)
	}
	defer os.Remove(tf.Name())

	_, _ = tf.Write(unformattedData)
	tf.Close()

	_, diags := f.Format(tf.Name())
	if diags.HasErrors() {
		t.Fatalf("the call to Format failed unexpectedly %s", diags.Error())
	}

	//lets re-read the tempfile which should now be formatted
	data, err := os.ReadFile(tf.Name())
	if err != nil {
		t.Fatalf("failed to open the newly formatted fixture %s", err)
	}

	formattedData, err := os.ReadFile("testdata/format/formatted.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to open the formatted fixture %s", err)
	}

	if diff := cmp.Diff(string(data), string(formattedData)); diff != "" {
		t.Errorf("Unexpected format output %s", diff)
	}
}

func TestHCL2Formatter_Format_ShowDiff(t *testing.T) {

	if _, err := exec.LookPath("diff"); err != nil {
		t.Skip("Skipping test because diff is not in the executable PATH")
	}

	var buf bytes.Buffer
	f := HCL2Formatter{
		Output:   &buf,
		ShowDiff: true,
	}

	_, diags := f.Format("testdata/format/unformatted.pkr.hcl")
	if diags.HasErrors() {
		t.Fatalf("the call to Format failed unexpectedly %s", diags.Error())
	}

	diffHeader := `
--- old/testdata/format/unformatted.pkr.hcl
+++ new/testdata/format/unformatted.pkr.hcl
@@ -1,149 +1,149 @@
`
	if !strings.Contains(buf.String(), diffHeader) {
		t.Errorf("expected buf to contain a file diff, but instead we got %s", buf.String())
	}

}

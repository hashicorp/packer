package hcl2template

import (
	"bytes"
	"io/ioutil"
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

type FormatterRecursiveTestCase struct {
	TestCaseName             string
	Recursion                bool
	TopLevelFilePreFormat    []byte
	LowerLevelFilePreFormat  []byte
	TopLevelFilePostFormat   []byte
	LowerLevelFilePostFormat []byte
}

func TestHCL2Formatter_Recursive(t *testing.T) {
	unformattedData := []byte(`
// starts resources to provision them.
build {
    sources = [
        "source.amazon-ebs.ubuntu-1604",
        "source.virtualbox-iso.ubuntu-1204",
    ]

    provisioner "shell" {
        string  = coalesce(null, "", "string")
        int     = "${41 + 1}"
        int64   = "${42 + 1}"
        bool    = "true"
        trilean = true
        duration = "${9 + 1}s"
        map_string_string = {
            a = "b"
            c = "d"
        }
        slice_string = [
            "a",
            "b",
            "c",
        ]
        slice_slice_string = [
            ["a","b"],
            ["c","d"]
        ]

        nested {
            string  = "string"
            int     = 42
            int64   = 43
            bool    = true
            trilean = true
            duration = "10s"
            map_string_string = {
                a = "b"
                c = "d"
            }
            slice_string = [
                "a",
                "b",
                "c",
            ]
            slice_slice_string = [
                ["a","b"],
                ["c","d"]
            ]
        }

        nested_slice {
        }
    }

    provisioner "file" {
        string  = "string"
        int     = 42
        int64   = 43
        bool    = true
        trilean = true
        duration          = "10s"
        map_string_string = {
            a = "b"
            c = "d"
        }
        slice_string = [
            "a",
            "b",
            "c",
        ]
        slice_slice_string = [
            ["a","b"],
            ["c","d"]
        ]

        nested {
            string   = "string"
            int      = 42
            int64    = 43
            bool     = true
            trilean  = true
            duration = "10s"
            map_string_string = {
                a = "b"
                c = "d"
            }
            slice_string = [
                "a",
                "b",
                "c",
            ]
            slice_slice_string = [
                ["a","b"],
                ["c","d"]
            ]
        }

        nested_slice {
        }
    }

    post-processor "amazon-import" {
        string   = "string"
        int      = 42
        int64    = 43
        bool     = true
        trilean  = true
        duration = "10s"
        map_string_string = {
            a = "b"
            c = "d"
        }
        slice_string = [
            "a",
            "b",
            "c",
        ]
        slice_slice_string = [
            ["a","b"],
            ["c","d"]
        ]

        nested {
            string   = "string"
            int      = 42
            int64    = 43
            bool     = true
            trilean  = true
            duration = "10s"
            map_string_string = {
                a = "b"
                c = "d"
            }
            slice_string = [
                "a",
                "b",
                "c",
            ]
            slice_slice_string = [
                ["a","b"],
                ["c","d"]
            ]
        }

        nested_slice {
        }
    }
}
`)

	formattedData := []byte(`
// starts resources to provision them.
build {
  sources = [
    "source.amazon-ebs.ubuntu-1604",
    "source.virtualbox-iso.ubuntu-1204",
  ]

  provisioner "shell" {
    string   = coalesce(null, "", "string")
    int      = "${41 + 1}"
    int64    = "${42 + 1}"
    bool     = "true"
    trilean  = true
    duration = "${9 + 1}s"
    map_string_string = {
      a = "b"
      c = "d"
    }
    slice_string = [
      "a",
      "b",
      "c",
    ]
    slice_slice_string = [
      ["a", "b"],
      ["c", "d"]
    ]

    nested {
      string   = "string"
      int      = 42
      int64    = 43
      bool     = true
      trilean  = true
      duration = "10s"
      map_string_string = {
        a = "b"
        c = "d"
      }
      slice_string = [
        "a",
        "b",
        "c",
      ]
      slice_slice_string = [
        ["a", "b"],
        ["c", "d"]
      ]
    }

    nested_slice {
    }
  }

  provisioner "file" {
    string   = "string"
    int      = 42
    int64    = 43
    bool     = true
    trilean  = true
    duration = "10s"
    map_string_string = {
      a = "b"
      c = "d"
    }
    slice_string = [
      "a",
      "b",
      "c",
    ]
    slice_slice_string = [
      ["a", "b"],
      ["c", "d"]
    ]

    nested {
      string   = "string"
      int      = 42
      int64    = 43
      bool     = true
      trilean  = true
      duration = "10s"
      map_string_string = {
        a = "b"
        c = "d"
      }
      slice_string = [
        "a",
        "b",
        "c",
      ]
      slice_slice_string = [
        ["a", "b"],
        ["c", "d"]
      ]
    }

    nested_slice {
    }
  }

  post-processor "amazon-import" {
    string   = "string"
    int      = 42
    int64    = 43
    bool     = true
    trilean  = true
    duration = "10s"
    map_string_string = {
      a = "b"
      c = "d"
    }
    slice_string = [
      "a",
      "b",
      "c",
    ]
    slice_slice_string = [
      ["a", "b"],
      ["c", "d"]
    ]

    nested {
      string   = "string"
      int      = 42
      int64    = 43
      bool     = true
      trilean  = true
      duration = "10s"
      map_string_string = {
        a = "b"
        c = "d"
      }
      slice_string = [
        "a",
        "b",
        "c",
      ]
      slice_slice_string = [
        ["a", "b"],
        ["c", "d"]
      ]
    }

    nested_slice {
    }
  }
}
`)

	var buf bytes.Buffer
	f := NewHCL2Formatter()
	f.Output = &buf
	f.Write = true

	recursiveTestCases := []FormatterRecursiveTestCase{
		{
			TestCaseName:             "With Recursive flag on",
			Recursion:                true,
			TopLevelFilePreFormat:    unformattedData,
			LowerLevelFilePreFormat:  unformattedData,
			TopLevelFilePostFormat:   formattedData,
			LowerLevelFilePostFormat: formattedData,
		},
		{
			TestCaseName:             "With Recursive flag off",
			Recursion:                false,
			TopLevelFilePreFormat:    unformattedData,
			LowerLevelFilePreFormat:  unformattedData,
			TopLevelFilePostFormat:   formattedData,
			LowerLevelFilePostFormat: unformattedData,
		},
	}

	for _, tc := range recursiveTestCases {
		executeRecursiveTestCase(t, tc, f)
	}
}

func executeRecursiveTestCase(t *testing.T, tc FormatterRecursiveTestCase, f *HCL2Formatter) {
	f.Recursive = tc.Recursion

	var subDir string
	subDir, err := ioutil.TempDir("testdata/format", "sub_dir")
	if err != nil {
		t.Fatalf("failed to create sub level recurisve directory for test %s", err)
	}
	defer os.Remove(subDir)

	var superSubDir string
	superSubDir, err = ioutil.TempDir(subDir, "super_sub_dir")
	if err != nil {
		t.Fatalf("failed to create sub level recurisve directory for test %s", err)
	}
	defer os.Remove(superSubDir)

	tf, err := ioutil.TempFile(subDir, "*.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to create top level tempfile for test %s", err)
	}
	defer os.Remove(tf.Name())

	_, _ = tf.Write(tc.TopLevelFilePreFormat)
	tf.Close()

	subTf, err := ioutil.TempFile(superSubDir, "*.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to create sub level tempfile for test %s", err)
	}
	defer os.Remove(subTf.Name())

	_, _ = subTf.Write(tc.LowerLevelFilePreFormat)
	subTf.Close()

	_, diags := f.Format(subDir)
	if diags.HasErrors() {
		t.Fatalf("the call to Format failed unexpectedly %s", diags.Error())
	}

	validateFileIsFormatted(t, tc.TopLevelFilePostFormat, tf)
	validateFileIsFormatted(t, tc.LowerLevelFilePostFormat, subTf)
}

func validateFileIsFormatted(t *testing.T, formattedData []byte, testFile *os.File) {
	data, err := ioutil.ReadFile(testFile.Name())
	if err != nil {
		t.Fatalf("failed to open the newly formatted fixture %s", err)
	}

	if diff := cmp.Diff(string(data), string(formattedData)); diff != "" {
		t.Errorf("Unexpected format tfData output %s", diff)
	}
}

func TestHCL2Formatter_Format_Write(t *testing.T) {

	var buf bytes.Buffer
	f := NewHCL2Formatter()
	f.Output = &buf
	f.Write = true

	unformattedData, err := ioutil.ReadFile("testdata/format/unformatted.pkr.hcl")
	if err != nil {
		t.Fatalf("failed to open the unformatted fixture %s", err)
	}

	tf, err := ioutil.TempFile("", "*.pkr.hcl")
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
	data, err := ioutil.ReadFile(tf.Name())
	if err != nil {
		t.Fatalf("failed to open the newly formatted fixture %s", err)
	}

	formattedData, err := ioutil.ReadFile("testdata/format/formatted.pkr.hcl")
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

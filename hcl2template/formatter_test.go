package hcl2template

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

func TestHCL2Formatter_Recursive(t *testing.T) {
	unformattedData := `
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
`

	formattedData := `
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
`

	tests := []struct {
		name                  string
		recursive             bool
		alreadyPresentContent map[string]string
		expectedContent       map[string]string
	}{
		{
			name:      "nested formats recursively",
			recursive: true,
			alreadyPresentContent: map[string]string{
				"foo/bar/baz":     unformattedData,
				"foo/bar/baz/woo": unformattedData,
				"":                unformattedData,
			},
			expectedContent: map[string]string{
				"foo/bar/baz":     formattedData,
				"foo/bar/baz/woo": formattedData,
				"":                formattedData,
			},
		},
		{
			name:      "nested no recursive format",
			recursive: false,
			alreadyPresentContent: map[string]string{
				"foo/bar/baz":     unformattedData,
				"foo/bar/baz/woo": unformattedData,
				"":                unformattedData,
			},
			expectedContent: map[string]string{
				"foo/bar/baz":     unformattedData,
				"foo/bar/baz/woo": unformattedData,
				"":                formattedData,
			},
		},
	}

	var buf bytes.Buffer
	f := NewHCL2Formatter()
	f.Output = &buf
	f.Write = true

	testFileName := "test.pkrvars.hcl"

	for _, tt := range tests {
		// testFilePaths := make(map[string]string)
		topDir := filepath.Join("testdata/format", "temp-dir")
		err := os.Mkdir(topDir, 0777)
		// topDir, err := ioutil.TempDir("testdata/format", "temp-dir")
		if err != nil {
			t.Fatalf("Failed to create TopDir for test case: %s, error: %v", tt.name, err)
		}

		for testDir, content := range tt.alreadyPresentContent {
			dir := filepath.Join(topDir, testDir)
			err := os.MkdirAll(dir, 0777)
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf(
					"Failed to create subDir: %s\n\n, for test case: %s\n\n, error: %v",
					testDir,
					tt.name,
					err)
			}

			file, err := os.Create(filepath.Join(dir, testFileName))
			//file, err := ioutil.TempFile(dir, "*.pkrvars.hcl")
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("failed to create testfile at directory: %s\n\n, for test case: %s\n\n, error: %s",
					testDir,
					tt.name,
					err)
			}
			//testFilePaths[testDir] = file.Name()

			_, err = file.Write([]byte(content))
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("failed to write to testfile at directory: %s\n\n, for test case: %s\n\n, error: %s",
					testDir,
					tt.name,
					err)
			}

			err = file.Close()
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("failed to close testfile at directory: %s\n\n, for test case: %s\n\n, error: %s",
					testDir,
					tt.name,
					err)
			}
		}

		f.Recursive = tt.recursive
		_, diags := f.Format(topDir)
		if diags.HasErrors() {
			os.RemoveAll(topDir)
			t.Fatalf("the call to Format failed unexpectedly for test case: %s, errors: %s", tt.name, diags.Error())
		}

		for expectedPath, expectedContent := range tt.expectedContent {
			//testFilePath := testFilePaths[expectedPath]
			testFilePath := filepath.Join(topDir, expectedPath, testFileName)
			b, err := ioutil.ReadFile(testFilePath)
			if err != nil {
				os.RemoveAll(topDir)
				t.Fatalf("ReadFile failed for test case: %s, error : %v", tt.name, err)
			}
			got := string(b)
			if diff := cmp.Diff(got, expectedContent); diff != "" {
				os.RemoveAll(topDir)
				t.Errorf(
					"format dir, unexpected result for test case: %s, path: %s,  Expected: %s, Got: %s",
					tt.name,
					expectedPath,
					expectedContent,
					got)
			}
			err = os.Remove(testFilePath)
			if err != nil {
				t.Errorf(
					"Failed to delete test file: %s for test case: %s, please clean before another test run. Error: %s",
					testFilePath,
					tt.name,
					err)
			}
		}

		err = os.RemoveAll(topDir)
		if err != nil {
			t.Errorf(
				"Failed to delete top level test directory for test case: %s, please clean before another test run. Error: %s",
				tt.name,
				err)
		}
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

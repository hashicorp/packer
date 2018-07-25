package common

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestChooseString(t *testing.T) {
	cases := []struct {
		Input  []string
		Output string
	}{
		{
			[]string{"", "foo", ""},
			"foo",
		},
		{
			[]string{"", "foo", "bar"},
			"foo",
		},
		{
			[]string{"", "", ""},
			"",
		},
	}

	for _, tc := range cases {
		result := ChooseString(tc.Input...)
		if result != tc.Output {
			t.Fatalf("bad: %#v", tc.Input)
		}
	}
}

func TestValidatedURL(t *testing.T) {
	// Invalid URL: has hex code in host
	_, err := ValidatedURL("http://what%20.com")
	if err == nil {
		t.Fatalf("expected err : %s", err)
	}

	// Invalid: unsupported scheme
	_, err = ValidatedURL("ftp://host.com/path")
	if err == nil {
		t.Fatalf("expected err : %s", err)
	}

	// Valid: http
	u, err := ValidatedURL("HTTP://packer.io/path")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if u != "http://packer.io/path" {
		t.Fatalf("bad: %s", u)
	}

	cases := []struct {
		InputString string
		OutputURL   string
		ErrExpected bool
	}{
		// Invalid URL: has hex code in host
		{"http://what%20.com", "", true},
		// Valid: http
		{"HTTP://packer.io/path", "http://packer.io/path", false},
		// No path
		{"HTTP://packer.io", "http://packer.io", false},
		// Invalid: unsupported scheme
		{"ftp://host.com/path", "", true},
	}

	for _, tc := range cases {
		u, err := ValidatedURL(tc.InputString)
		if u != tc.OutputURL {
			t.Fatal(fmt.Sprintf("Error with URL %s: got %s but expected %s",
				tc.InputString, tc.OutputURL, u))
		}
		if (err != nil) != tc.ErrExpected {
			if tc.ErrExpected == true {
				t.Fatal(fmt.Sprintf("Error with URL %s: we expected "+
					"ValidatedURL to return an error but didn't get one.",
					tc.InputString))
			} else {
				t.Fatal(fmt.Sprintf("Error with URL %s: we did not expect an "+
					" error from ValidatedURL but we got: %s",
					tc.InputString, err))
			}
		}
	}
}

func GetNativePathToTestFixtures(t *testing.T) string {
	const path = "./test-fixtures"
	res, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("err converting test-fixtures path into an absolute path : %s", err)
	}
	return res
}

func GetPortablePathToTestFixtures(t *testing.T) string {
	res := GetNativePathToTestFixtures(t)
	return filepath.ToSlash(res)
}

func TestDownloadableURL_WindowsFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		portablepath := GetPortablePathToTestFixtures(t)
		nativepath := GetNativePathToTestFixtures(t)

		dirCases := []struct {
			InputString string
			OutputURL   string
			ErrExpected bool
		}{ // TODO: add different directories
			{
				fmt.Sprintf("%s\\SomeDir\\myfile.txt", nativepath),
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // without the drive makes this native path a relative file:// uri
				"test-fixtures\\SomeDir\\myfile.txt",
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // without the drive makes this native path a relative file:// uri
				"test-fixtures/SomeDir/myfile.txt",
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // UNC paths being promoted to smb:// uri scheme.
				fmt.Sprintf("\\\\localhost\\C$\\%s\\SomeDir\\myfile.txt", nativepath),
				fmt.Sprintf("smb://localhost/C$/%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // Absolute uri (incorrect slash type)
				fmt.Sprintf("file:///%s\\SomeDir\\myfile.txt", nativepath),
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // Absolute uri (existing and mis-spelled)
				fmt.Sprintf("file:///%s/Somedir/myfile.txt", nativepath),
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // Absolute path (non-existing)
				"\\absolute\\path\\to\\non-existing\\file.txt",
				"file:///absolute/path/to/non-existing/file.txt",
				false,
			},
			{ // Absolute paths (existing)
				fmt.Sprintf("%s/SomeDir/myfile.txt", nativepath),
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // Relative path (non-existing)
				"./nonexisting/relative/path/to/file.txt",
				"file://./nonexisting/relative/path/to/file.txt",
				false,
			},
			{ // Relative path (existing)
				"./test-fixtures/SomeDir/myfile.txt",
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // Absolute uri (existing and with `/` prefix)
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				fmt.Sprintf("file:///%s/SomeDir/myfile.txt", portablepath),
				false,
			},
			{ // Absolute uri (non-existing and with `/` prefix)
				"file:///path/to/non-existing/file.txt",
				"file:///path/to/non-existing/file.txt",
				false,
			},
			{ // Absolute uri (non-existing and missing `/` prefix)
				"file://path/to/non-existing/file.txt",
				"file://path/to/non-existing/file.txt",
				false,
			},
			{ // Absolute uri and volume (non-existing and with `/` prefix)
				"file:///T:/path/to/non-existing/file.txt",
				"file:///T:/path/to/non-existing/file.txt",
				false,
			},
			{ // Absolute uri and volume (non-existing and missing `/` prefix)
				"file://T:/path/to/non-existing/file.txt",
				"file://T:/path/to/non-existing/file.txt",
				false,
			},
		}
		// Run through test cases to make sure they all parse correctly
		for idx, tc := range dirCases {
			u, err := DownloadableURL(tc.InputString)
			if (err != nil) != tc.ErrExpected {
				t.Fatalf("Test Case %d failed: Expected err = %#v, err = %#v, input = %s",
					idx, tc.ErrExpected, err, tc.InputString)
			}
			if u != tc.OutputURL {
				t.Fatalf("Test Case %d failed: Expected %s but received %s from input %s",
					idx, tc.OutputURL, u, tc.InputString)
			}
		}
	}
}

func TestDownloadableURL_FilePaths(t *testing.T) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("tempfile err: %s", err)
	}
	defer os.Remove(tf.Name())
	tf.Close()

	tfPath, err := filepath.EvalSymlinks(tf.Name())
	if err != nil {
		t.Fatalf("tempfile err: %s", err)
	}

	tfPath = filepath.Clean(tfPath)
	filePrefix := "file://"

	// If we're running windows, then absolute URIs are `/`-prefixed.
	platformPrefix := ""
	if runtime.GOOS == "windows" {
		platformPrefix = "/"
	}

	// Relative filepath. We run this test in a func so that
	// the defers run right away.
	func() {
		wd, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd err: %s", err)
		}

		err = os.Chdir(filepath.Dir(tfPath))
		if err != nil {
			t.Fatalf("chdir err: %s", err)
		}
		defer os.Chdir(wd)

		filename := filepath.Base(tfPath)
		u, err := DownloadableURL(filename)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		expected := fmt.Sprintf("%s%s%s",
			filePrefix,
			platformPrefix,
			strings.Replace(tfPath, `\`, `/`, -1))
		if u != expected {
			t.Fatalf("unexpected: %#v != %#v", u, expected)
		}
	}()

	// Test some cases with and without a schema prefix
	for _, prefix := range []string{"", filePrefix + platformPrefix} {
		// Nonexistent file
		_, err = DownloadableURL(prefix + "i/dont/exist")
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		// Good file (absolute)
		u, err := DownloadableURL(prefix + tfPath)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		expected := fmt.Sprintf("%s%s%s",
			filePrefix,
			platformPrefix,
			strings.Replace(tfPath, `\`, `/`, -1))
		if u != expected {
			t.Fatalf("unexpected: %s != %s", u, expected)
		}
	}
}

func TestFileExistsLocally(t *testing.T) {
	portablepath := GetPortablePathToTestFixtures(t)

	dirCases := []struct {
		Input  string
		Output bool
	}{
		// file exists locally
		{fmt.Sprintf("file://%s/SomeDir/myfile.txt", portablepath), true},
		// remote protocols short-circuit and are considered to exist locally
		{"https://myfile.iso", true},
		// non-existent protocols do not exist and hence fail
		{"nonexistent-protocol://myfile.iso", false},
		// file does not exist locally
		{"file:///C/i/dont/exist", false},
	}
	// Run through test cases to make sure they all parse correctly
	for _, tc := range dirCases {
		fileOK := FileExistsLocally(tc.Input)
		if fileOK != tc.Output {
			t.Fatalf("Test Case failed: Expected %#v, received = %#v, input = %s",
				tc.Output, fileOK, tc.Input)
		}
	}
}

func TestScrubConfig(t *testing.T) {
	type Inner struct {
		Baz string
	}
	type Local struct {
		Foo string
		Bar string
		Inner
	}
	c := Local{"foo", "bar", Inner{"bar"}}
	expect := "Config: {Foo:foo Bar:<Filtered> Inner:{Baz:<Filtered>}}"
	conf := ScrubConfig(c, c.Bar)
	if conf != expect {
		t.Fatalf("got %s, expected %s", conf, expect)
	}
}

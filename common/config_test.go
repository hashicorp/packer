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

func TestDownloadableURL(t *testing.T) {

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
		u, err := DownloadableURL(tc.InputString)
		if u != tc.OutputURL {
			t.Fatal(fmt.Sprintf("Error with URL %s: got %s but expected %s",
				tc.InputString, tc.OutputURL, u))
		}
		if (err != nil) != tc.ErrExpected {
			if tc.ErrExpected == true {
				t.Fatal(fmt.Sprintf("Error with URL %s: we expected "+
					"DownloadableURL to return an error but didn't get one.",
					tc.InputString))
			} else {
				t.Fatal(fmt.Sprintf("Error with URL %s: we did not expect an "+
					" error from DownloadableURL but we got: %s",
					tc.InputString, err))
			}
		}
	}
}

func TestDownloadableURL_WindowsFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		dirCases := []struct {
			InputString string
			OutputURL   string
			ErrExpected bool
		}{ // TODO: add different directories
			{
				"C:\\Temp\\SomeDir\\myfile.txt",
				"file:///C:/Temp/SomeDir/myfile.txt",
				false,
			},
			{ // need windows drive
				"\\Temp\\SomeDir\\myfile.txt",
				"",
				true,
			},
			{ // need windows drive
				"/Temp/SomeDir/myfile.txt",
				"",
				true,
			},
			{ // UNC paths; why not?
				"\\\\?\\c:\\Temp\\SomeDir\\myfile.txt",
				"",
				true,
			},
			{
				"file:///C:\\Temp\\SomeDir\\myfile.txt",
				"file:///c:/Temp/SomeDir/myfile.txt",
				false,
			},
			{
				"file:///c:/Temp/Somedir/myfile.txt",
				"file:///c:/Temp/SomeDir/myfile.txt",
				false,
			},
		}
		// create absolute-pathed tempfile to play with
		err := os.Mkdir("C:\\Temp\\SomeDir", 0755)
		if err != nil {
			t.Fatalf("err creating test dir: %s", err)
		}
		fi, err := os.Create("C:\\Temp\\SomeDir\\myfile.txt")
		if err != nil {
			t.Fatalf("err creating test file: %s", err)
		}
		fi.Close()
		defer os.Remove("C:\\Temp\\SomeDir\\myfile.txt")
		defer os.Remove("C:\\Temp\\SomeDir")

		// Run through test cases to make sure they all parse correctly
		for _, tc := range dirCases {
			u, err := DownloadableURL(tc.InputString)
			if (err != nil) != tc.ErrExpected {
				t.Fatalf("Test Case failed: Expected err = %#v, err = %#v, input = %s",
					tc.ErrExpected, err, tc.InputString)
			}
			if u != tc.OutputURL {
				t.Fatalf("Test Case failed: Expected %s but received %s from input %s",
					tc.OutputURL, u, tc.InputString)
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
	if runtime.GOOS == "windows" {
		filePrefix += "/"
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

		expected := fmt.Sprintf("%s%s",
			filePrefix,
			strings.Replace(tfPath, `\`, `/`, -1))
		if u != expected {
			t.Fatalf("unexpected: %#v != %#v", u, expected)
		}
	}()

	// Test some cases with and without a schema prefix
	for _, prefix := range []string{"", filePrefix} {
		// Nonexistent file
		_, err = DownloadableURL(prefix + "i/dont/exist")
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		// Good file
		u, err := DownloadableURL(prefix + tfPath)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		expected := fmt.Sprintf("%s%s",
			filePrefix,
			strings.Replace(tfPath, `\`, `/`, -1))
		if u != expected {
			t.Fatalf("unexpected: %s != %s", u, expected)
		}
	}
}

func test_FileExistsLocally(t *testing.T) {
	if runtime.GOOS == "windows" {
		dirCases := []struct {
			Input  string
			Output bool
		}{
			// file exists locally
			{"file:///C:/Temp/SomeDir/myfile.txt", true},
			// file is not supposed to exist locally
			{"https://myfile.iso", true},
			// file does not exist locally
			{"file:///C/i/dont/exist", false},
		}
		// create absolute-pathed tempfile to play with
		err := os.Mkdir("C:\\Temp\\SomeDir", 0755)
		if err != nil {
			t.Fatalf("err creating test dir: %s", err)
		}
		fi, err := os.Create("C:\\Temp\\SomeDir\\myfile.txt")
		if err != nil {
			t.Fatalf("err creating test file: %s", err)
		}
		fi.Close()
		defer os.Remove("C:\\Temp\\SomeDir\\myfile.txt")
		defer os.Remove("C:\\Temp\\SomeDir")

		// Run through test cases to make sure they all parse correctly
		for _, tc := range dirCases {
			fileOK := FileExistsLocally(tc.Input)
			if !fileOK {
				t.Fatalf("Test Case failed: Expected %#v, received = %#v, input = %s",
					tc.Output, fileOK, tc.Input)
			}
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

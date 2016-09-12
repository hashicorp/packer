package common

import (
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"testing"
	"log"
	"strings"
	"fmt"
)

// utility function for returning a directory structure as a list of strings
func getDirectory(path string) []string {
	var result []string
	walk := func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && !strings.HasSuffix(path, "/") {
			path = path + "/"
		}
		result = append(result, filepath.ToSlash(path))
		return nil
	}
	filepath.Walk(path, walk)
	return result
}

// utility function for creating a directory structure
type createFileContents func(string) []byte
func createDirectory(path string, hier []string, fileContents createFileContents) error {
	if fileContents == nil {
		fileContents = func(string) []byte {
			return []byte{}
		}
	}
	for _,filename := range hier {
		p := filepath.Join(path, filename)
		if strings.HasSuffix(filename, "/") {
			err := os.MkdirAll(p, 0)
			if err != nil { return err }
			continue
		}
		f,err := os.Create(p)
		if err != nil { return err }
		_,err = f.Write(fileContents(filename))
		if err != nil { return err }
		err = f.Close()
		if err != nil { return err }
	}
	return nil
}

func TestStepCreateFloppy_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepCreateFloppy)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("StepCreateFloppy should be a step")
	}
}

func testStepCreateFloppyState(t *testing.T) multistep.StateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	return state
}

func TestStepCreateFloppy(t *testing.T) {
	state := testStepCreateFloppyState(t)
	step := new(StepCreateFloppy)

	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	count := 10
	expected := count
	files := make([]string, count)

	prefix := "exists"
	ext := ".tmp"

	for i := 0; i < expected; i++ {
		files[i] = path.Join(dir, prefix+strconv.Itoa(i)+ext)

		_, err := os.Create(files[i])
		if err != nil {
			t.Fatalf("err: %s", err)
		}
	}

	lists := [][]string{
		files,
		[]string{dir + string(os.PathSeparator) + prefix + "*" + ext},
		[]string{dir + string(os.PathSeparator) + prefix + "?" + ext},
		[]string{dir + string(os.PathSeparator) + prefix + "[0123456789]" + ext},
		[]string{dir + string(os.PathSeparator) + prefix + "[0-9]" + ext},
		[]string{dir + string(os.PathSeparator)},
		[]string{dir},
	}

	for _, step.Files = range lists {
		if action := step.Run(state); action != multistep.ActionContinue {
			t.Fatalf("bad action: %#v for %v", action, step.Files)
		}

		if _, ok := state.GetOk("error"); ok {
			t.Fatalf("state should be ok for %v", step.Files)
		}

		floppy_path := state.Get("floppy_path").(string)

		if _, err := os.Stat(floppy_path); err != nil {
			t.Fatalf("file not found: %s for %v", floppy_path, step.Files)
		}

		if len(step.FilesAdded) != expected {
			t.Fatalf("expected %d, found %d for %v", expected, len(step.FilesAdded), step.Files)
		}

		step.Cleanup(state)

		if _, err := os.Stat(floppy_path); err == nil {
			t.Fatalf("file found: %s for %v", floppy_path, step.Files)
		}
	}
}

func xxxTestStepCreateFloppy_missing(t *testing.T) {
	state := testStepCreateFloppyState(t)
	step := new(StepCreateFloppy)

	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	count := 2
	expected := 0
	files := make([]string, count)

	prefix := "missing"

	for i := 0; i < count; i++ {
		files[i] = path.Join(dir, prefix+strconv.Itoa(i))
	}

	lists := [][]string{
		files,
	}

	for _, step.Files = range lists {
		if action := step.Run(state); action != multistep.ActionHalt {
			t.Fatalf("bad action: %#v for %v", action, step.Files)
		}

		if _, ok := state.GetOk("error"); !ok {
			t.Fatalf("state should not be ok for %v", step.Files)
		}

		floppy_path := state.Get("floppy_path")

		if floppy_path != nil {
			t.Fatalf("floppy_path is not nil for %v", step.Files)
		}

		if len(step.FilesAdded) != expected {
			t.Fatalf("expected %d, found %d for %v", expected, len(step.FilesAdded), step.Files)
		}
	}
}

func xxxTestStepCreateFloppy_notfound(t *testing.T) {
	state := testStepCreateFloppyState(t)
	step := new(StepCreateFloppy)

	dir, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(dir)

	count := 2
	expected := 0
	files := make([]string, count)

	prefix := "notfound"

	for i := 0; i < count; i++ {
		files[i] = path.Join(dir, prefix+strconv.Itoa(i))
	}

	lists := [][]string{
		[]string{dir + string(os.PathSeparator) + prefix + "*"},
		[]string{dir + string(os.PathSeparator) + prefix + "?"},
		[]string{dir + string(os.PathSeparator) + prefix + "[0123456789]"},
		[]string{dir + string(os.PathSeparator) + prefix + "[0-9]"},
		[]string{dir + string(os.PathSeparator)},
		[]string{dir},
	}

	for _, step.Files = range lists {
		if action := step.Run(state); action != multistep.ActionContinue {
			t.Fatalf("bad action: %#v for %v", action, step.Files)
		}

		if _, ok := state.GetOk("error"); ok {
			t.Fatalf("state should be ok for %v", step.Files)
		}

		floppy_path := state.Get("floppy_path").(string)

		if _, err := os.Stat(floppy_path); err != nil {
			t.Fatalf("file not found: %s for %v", floppy_path, step.Files)
		}

		if len(step.FilesAdded) != expected {
			t.Fatalf("expected %d, found %d for %v", expected, len(step.FilesAdded), step.Files)
		}

		step.Cleanup(state)

		if _, err := os.Stat(floppy_path); err == nil {
			t.Fatalf("file found: %s for %v", floppy_path, step.Files)
		}
	}
}

func TestStepCreateFloppyContents(t *testing.T) {
	// file-system hierarchies
	hierarchies := [][]string{
		[]string{"file1", "file2", "file3"},
		[]string{"dir1/", "dir1/file1", "dir1/file2", "dir1/file3"},
		[]string{"dir1/", "dir1/file1", "dir1/subdir1/", "dir1/subdir1/file1", "dir1/subdir1/file2", "dir2/", "dir2/subdir1/", "dir2/subdir1/file1","dir2/subdir1/file2"},
	}

	type contentsTest struct {
		contents []string
		result []string
	}

	// keep in mind that .FilesAdded doesn't keep track of the target filename or directories, but rather the source filename.
	contents := [][]contentsTest{
		[]contentsTest{
			contentsTest{contents:[]string{"file1","file2","file3"},result:[]string{"file1","file2","file3"}},
			contentsTest{contents:[]string{"file?"},result:[]string{"file1","file2","file3"}},
			contentsTest{contents:[]string{"*"},result:[]string{"file1","file2","file3"}},
		},
		[]contentsTest{
			contentsTest{contents:[]string{"dir1"},result:[]string{"dir1/file1","dir1/file2","dir1/file3"}},
			contentsTest{contents:[]string{"dir1/file1","dir1/file2","dir1/file3"},result:[]string{"dir1/file1","dir1/file2","dir1/file3"}},
			contentsTest{contents:[]string{"*"},result:[]string{"dir1/file1","dir1/file2","dir1/file3"}},
			contentsTest{contents:[]string{"*/*"},result:[]string{"dir1/file1","dir1/file2","dir1/file3"}},
		},
		[]contentsTest{
			contentsTest{contents:[]string{"dir1"},result:[]string{"dir1/file1","dir1/subdir1/file1","dir1/subdir1/file2"}},
			contentsTest{contents:[]string{"dir2/*"},result:[]string{"dir2/subdir1/file1","dir2/subdir1/file2"}},
			contentsTest{contents:[]string{"dir2/subdir1"},result:[]string{"dir2/subdir1/file1","dir2/subdir1/file2"}},
			contentsTest{contents:[]string{"dir?"},result:[]string{"dir1/file1","dir1/subdir1/file1","dir1/subdir1/file2","dir2/subdir1/file1","dir2/subdir1/file2"}},
		},
	}

	// create the hierarchy for each file
	for i,hier := range hierarchies {
		t.Logf("Trying with hierarchy : %v",hier)

		// create the temp directory
		dir, err := ioutil.TempDir("", "packer")
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		// create the file contents
		err = createDirectory(dir, hier, nil)
		if err != nil { t.Fatalf("err: %s", err) }
		t.Logf("Making %v", hier)

		for _,test := range contents[i] {
			// createa new state and step
			state := testStepCreateFloppyState(t)
			step := new(StepCreateFloppy)

			// modify step.Contents with ones from testcase
			step.Contents = []string{}
			for _,c := range test.contents {
				step.Contents = append(step.Contents, filepath.Join(dir,filepath.FromSlash(c)))
			}
			log.Println(fmt.Sprintf("Trying against floppy_dirs : %v",step.Contents))

			// run the step
			if action := step.Run(state); action != multistep.ActionContinue {
				t.Fatalf("bad action: %#v for %v : %v", action, step.Contents, state.Get("error"))
			}

			if _, ok := state.GetOk("error"); ok {
				t.Fatalf("state should be ok for %v : %v", step.Contents, state.Get("error"))
			}

			floppy_path := state.Get("floppy_path").(string)
			if _, err := os.Stat(floppy_path); err != nil {
				t.Fatalf("file not found: %s for %v : %v", floppy_path, step.Contents, err)
			}

			// check the FilesAdded array to see if it matches
			for _,rpath := range test.result {
				fpath := filepath.Join(dir, filepath.FromSlash(rpath))
				if !step.FilesAdded[fpath] {
					t.Fatalf("unable to find file: %s for %v", fpath, step.Contents)
				}
			}

			// cleanup the step
			step.Cleanup(state)

			if _, err := os.Stat(floppy_path); err == nil {
				t.Fatalf("file found: %s for %v", floppy_path, step.Contents)
			}
		}
		// remove the temp directory
		os.RemoveAll(dir)
	}
}

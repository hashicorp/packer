package common

import (
	"bytes"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
)

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

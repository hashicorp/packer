package ovf

import (
	"fmt"
	"io/ioutil"
	"os"
	osexec "os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/mitchellh/packer/packer"
)

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"ssh_username":     "foo",
		"shutdown_command": "foo",
		"source_path":      "config_test.go",
	}
}

func getTempFile(t *testing.T, dir string) *os.File {
	tf, err := ioutil.TempFile(dir, "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()

	// don't forget to cleanup the file downstream:
	// defer os.Remove(tf.Name())

	return tf
}

func TestNewConfig_FloppyFiles(t *testing.T) {
	c := testConfig(t)
	floppies_path := "../../../common/test-fixtures/floppies"
	c["floppy_files"] = []string{fmt.Sprintf("%s/bar.bat", floppies_path), fmt.Sprintf("%s/foo.ps1", floppies_path)}
	_, _, err := NewConfig(c)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestNewConfig_InvalidFloppies(t *testing.T) {
	c := testConfig(t)
	c["floppy_files"] = []string{"nonexistant.bat", "nonexistant.ps1"}
	_, _, errs := NewConfig(c)
	if errs == nil {
		t.Fatalf("Non existant floppies should trigger multierror")
	}

	if len(errs.(*packer.MultiError).Errors) != 2 {
		t.Fatalf("Multierror should work and report 2 errors")
	}
}

func TestNewConfig_sourcePath(t *testing.T) {
	// Okay, because it gets caught during download
	c := testConfig(t)
	delete(c, "source_path")
	_, warns, err := NewConfig(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatalf("should error with empty `source_path`")
	}

	// Okay, because it gets caught during download
	c = testConfig(t)
	c["source_path"] = "/i/dont/exist"
	_, warns, err = NewConfig(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Bad
	c = testConfig(t)
	c["source_path"] = "ftp://i/dont/exist"
	_, warns, err = NewConfig(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}

	// Good
	tf := getTempFile(t, "")
	defer os.Remove(tf.Name())

	c = testConfig(t)
	c["source_path"] = tf.Name()
	_, warns, err = NewConfig(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Tests symlinks
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("%s", err)
	}
	linkName := time.Now().Format("20060102150405")
	err = os.Symlink(cwd, linkName)
	defer os.Remove(cwd + "/" + linkName)
	if err != nil {
		t.Fatalf("%s", err)
	}

	tf = getTempFile(t, cwd+"/"+linkName)
	defer os.Remove(tf.Name())

	c = testConfig(t)
	c["source_path"] = tf.Name()
	_, warns, err = NewConfig(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}

	// Test Windows-style(?) symlinks
	linkName = time.Now().Format("20060102150405")
	if runtime.GOOS == "windows" {
		// Create special symlink according to #4323
		_, err := osexec.Command("cmd", "/c", "mklink", "/J", linkName, "\\\\?\\"+cwd).CombinedOutput()
		defer os.Remove(cwd + "\\" + linkName)
		tf = getTempFile(t, cwd+"\\"+linkName)
		defer os.Remove(tf.Name())
		c = testConfig(t)
		c["source_path"] = tf.Name()
		_, warns, err = NewConfig(c)
		if len(warns) > 0 {
			t.Fatalf("bad: %#v", warns)
		}
		if err != nil {
			t.Fatalf("bad: %s", err)
		}
	}
}

func TestNewConfig_shutdown_timeout(t *testing.T) {
	c := testConfig(t)
	tf := getTempFile(t, "")
	defer os.Remove(tf.Name())

	// Expect this to fail
	c["source_path"] = tf.Name()
	c["shutdown_timeout"] = "NaN"
	_, warns, err := NewConfig(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}

	// Passes when given a valid time duration
	c["shutdown_timeout"] = "10s"
	_, warns, err = NewConfig(c)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

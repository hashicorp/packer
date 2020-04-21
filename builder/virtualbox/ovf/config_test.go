package ovf

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"ssh_username":     "foo",
		"shutdown_command": "foo",
		"source_path":      "config_test.go",
	}
}

func getTempFile(t *testing.T) *os.File {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()

	// don't forget to cleanup the file downstream:
	// defer os.Remove(tf.Name())

	return tf
}

func TestNewConfig_FloppyFiles(t *testing.T) {
	cfg := testConfig(t)
	floppies_path := "../../../common/test-fixtures/floppies"
	cfg["floppy_files"] = []string{fmt.Sprintf("%s/bar.bat", floppies_path), fmt.Sprintf("%s/foo.ps1", floppies_path)}
	var c Config
	_, err := c.Prepare(cfg)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestNewConfig_InvalidFloppies(t *testing.T) {
	cfg := testConfig(t)
	cfg["floppy_files"] = []string{"nonexistent.bat", "nonexistent.ps1"}
	var c Config
	_, errs := c.Prepare(cfg)
	if errs == nil {
		t.Fatalf("Nonexistent floppies should trigger multierror")
	}

	if len(errs.(*packer.MultiError).Errors) != 2 {
		t.Fatalf("Multierror should work and report 2 errors")
	}
}

func TestNewConfig_sourcePath(t *testing.T) {
	// Okay, because it gets caught during download
	cfg := testConfig(t)
	delete(cfg, "source_path")
	var c Config
	warns, err := c.Prepare(cfg)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatalf("should error with empty `source_path`")
	}

	// Good
	tf := getTempFile(t)
	defer os.Remove(tf.Name())

	cfg = testConfig(t)
	cfg["source_path"] = tf.Name()
	warns, err = c.Prepare(cfg)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

func TestNewConfig_shutdown_timeout(t *testing.T) {
	cfg := testConfig(t)
	tf := getTempFile(t)
	defer os.Remove(tf.Name())

	// Expect this to fail
	cfg["source_path"] = tf.Name()
	cfg["shutdown_timeout"] = "NaN"
	var c Config
	warns, err := c.Prepare(cfg)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
	}

	// Passes when given a valid time duration
	cfg["shutdown_timeout"] = "10s"
	warns, err = c.Prepare(cfg)
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

// TestChecksumFileNameMixedCaseBug reproduces Github issue #9049:
//	https://github.com/hashicorp/packer/issues/9049
func TestChecksumFileNameMixedCaseBug(t *testing.T) {

	tf, err := ioutil.TempFile("", "Packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()

	// Calculate sum of the tempfile and store in checksum file
	hash := sha256.New()
	if _, err := io.Copy(hash, tf); err != nil {
		t.Fatal(err)
	}
	sum := hash.Sum(nil)
	checksumFilepath := fmt.Sprintf("%s.sha256", tf.Name())
	err = ioutil.WriteFile(checksumFilepath, sum, 0644)
	if err != nil {
		t.Errorf("Failed to write checksum file to: %s", checksumFilepath)
		t.Fail()
	}

	cfg := testConfig(t)
	cfg["source_path"] = tf.Name()
	cfg["checksum_type"] = "file"
	cfg["checksum"] = checksumFilepath
	cfg["type"] = "virtualbox-ovf"
	cfg["guest_additions_mode"] = "disable"
	cfg["headless"] = false

	var c Config
	warns, err := c.Prepare(cfg)
	if err != nil {
		t.Errorf("config failed to Prepare, %s", err.Error())
		t.Fail()
	}

	if len(warns) != 0 {
		t.Errorf("Encountered warnings during config preparation: %s", warns)
	}

}

package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"destination": "something",
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a provisioner")
	}
}

func TestProvisionerPrepare_InvalidKey(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_InvalidSource(t *testing.T) {
	var p Provisioner
	config := testConfig()
	config["source"] = "/this/should/not/exist"

	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should require existing file")
	}
}

func TestProvisionerPrepare_ValidSource(t *testing.T) {
	var p Provisioner

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config := testConfig()
	config["source"] = tf.Name()

	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should allow valid file: %s", err)
	}
}

func TestProvisionerPrepare_EmptyDestination(t *testing.T) {
	var p Provisioner

	config := testConfig()
	delete(config, "destination")
	err := p.Prepare(config)
	if err == nil {
		t.Fatalf("should require destination path")
	}
}

type stubUi struct {
	sayMessages string
}

func (su *stubUi) Ask(string) (string, error) {
	return "", nil
}

func (su *stubUi) Error(string) {
}

func (su *stubUi) Machine(string, ...string) {
}

func (su *stubUi) Message(string) {
}

func (su *stubUi) Say(msg string) {
	su.sayMessages += msg
}

func TestProvisionerProvision_SendsFile(t *testing.T) {
	var p Provisioner
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	if _, err = tf.Write([]byte("hello")); err != nil {
		t.Fatalf("error writing tempfile: %s", err)
	}

	config := map[string]interface{}{
		"source":      tf.Name(),
		"destination": "something",
	}

	if err := p.Prepare(config); err != nil {
		t.Fatalf("err: %s", err)
	}

	ui := &stubUi{}
	comm := &packer.MockCommunicator{}
	err = p.Provision(ui, comm)
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(ui.sayMessages, tf.Name()) {
		t.Fatalf("should print source filename")
	}

	if !strings.Contains(ui.sayMessages, "something") {
		t.Fatalf("should print destination filename")
	}

	if comm.UploadPath != "something" {
		t.Fatalf("should upload to configured destination")
	}

	if comm.UploadData != "hello" {
		t.Fatalf("should upload with source file's data")
	}
}

func TestProvisionDownloadMkdirAll(t *testing.T) {
	tests := []struct {
		path string
	}{
		{"dir"},
		{"dir/"},
		{"dir/subdir"},
		{"dir/subdir/"},
		{"path/to/dir"},
		{"path/to/dir/"},
	}
	tmpDir, err := ioutil.TempDir("", "packer-file")
	if err != nil {
		t.Fatalf("error tempdir: %s", err)
	}
	defer os.RemoveAll(tmpDir)
	tf, err := ioutil.TempFile(tmpDir, "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config := map[string]interface{}{
		"source": tf.Name(),
	}
	var p Provisioner
	for _, test := range tests {
		path := filepath.Join(tmpDir, test.path)
		config["destination"] = filepath.Join(path, "something")
		if err := p.Prepare(config); err != nil {
			t.Fatalf("err: %s", err)
		}
		ui := &stubUi{}
		comm := &packer.MockCommunicator{}
		err = p.ProvisionDownload(ui, comm)
		if err != nil {
			t.Fatalf("should successfully provision: %s", err)
		}

		if !strings.Contains(ui.sayMessages, tf.Name()) {
			t.Fatalf("should print source filename")
		}

		if !strings.Contains(ui.sayMessages, "something") {
			t.Fatalf("should print destination filename")
		}

		if _, err := os.Stat(path); err != nil {
			t.Fatalf("stat of download dir should not error: %s", err)
		}
	}
}

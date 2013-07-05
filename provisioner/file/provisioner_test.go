package file

import (
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"
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

type stubUploadCommunicator struct {
	dest string
	data []byte
}

func (suc *stubUploadCommunicator) Download(src string, data io.Writer) error {
	return nil
}

func (suc *stubUploadCommunicator) Upload(dest string, data io.Reader) error {
	var err error
	suc.dest = dest
	suc.data, err = ioutil.ReadAll(data)
	return err
}

func (suc *stubUploadCommunicator) Start(cmd *packer.RemoteCmd) error {
	return nil
}

type stubUi struct {
	sayMessages string
}

func (su *stubUi) Ask(string) (string, error) {
	return "", nil
}

func (su *stubUi) Error(string) {
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
	comm := &stubUploadCommunicator{}
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

	if comm.dest != "something" {
		t.Fatalf("should upload to configured destination")
	}

	if string(comm.data) != "hello" {
		t.Fatalf("should upload with source file's data")
	}
}

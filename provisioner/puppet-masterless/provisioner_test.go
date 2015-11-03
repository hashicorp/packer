package puppetmasterless

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func testConfig() map[string]interface{} {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		panic(err)
	}

	return map[string]interface{}{
		"manifest_file": tf.Name(),
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_hieraConfigPath(t *testing.T) {
	config := testConfig()

	delete(config, "hiera_config_path")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a good one
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["hiera_config_path"] = tf.Name()
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_manifestFile(t *testing.T) {
	config := testConfig()

	delete(config, "manifest_file")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should be an error")
	}

	// Test with a good one
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["manifest_file"] = tf.Name()
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_manifestDir(t *testing.T) {
	config := testConfig()

	delete(config, "manifestdir")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a good one
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer os.RemoveAll(td)

	config["manifest_dir"] = td
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_modulePaths(t *testing.T) {
	config := testConfig()

	delete(config, "module_paths")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with bad paths
	config["module_paths"] = []string{"i-should-not-exist"}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should be an error")
	}

	// Test with a good one
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer os.RemoveAll(td)

	config["module_paths"] = []string{td}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_facterFacts(t *testing.T) {
	config := testConfig()

	delete(config, "facter")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with malformed fact
	config["facter"] = "fact=stringified"
	p = new(Provisioner)
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should be an error")
	}

	// Test with a good one
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("error: %s", err)
	}
	defer os.RemoveAll(td)

	facts := make(map[string]string)
	facts["fact_name"] = "fact_value"
	config["facter"] = facts

	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Make sure the default facts are present
	delete(config, "facter")
	p = new(Provisioner)
	err = p.Prepare(config)
	if p.config.Facter == nil {
		t.Fatalf("err: Default facts are not set in the Puppet provisioner!")
	}
}

func TestProvisionerPrepare_extraArguments(t *testing.T) {
	config := testConfig()

	// Test with missing parameter
	delete(config, "extra_arguments")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with malformed value
	config["extra_arguments"] = "{{}}"
	p = new(Provisioner)
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should be an error")
	}

	// Test with valid values
	config["extra_arguments"] = []string{
		"arg",
	}

	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerProvision_extraArguments(t *testing.T) {
	config := testConfig()
	ui := &packer.MachineReadableUi{
		Writer: ioutil.Discard,
	}
	comm := new(packer.MockCommunicator)

	extraArguments := []string{
		"--some-arg=yup",
		"--some-other-arg",
	}
	config["extra_arguments"] = extraArguments

	// Test with valid values
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = p.Provision(ui, comm)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	expectedArgs := strings.Join(extraArguments, " ")

	if !strings.Contains(comm.StartCmd.Command, expectedArgs) {
		t.Fatalf("Command %q doesn't contain the expected arguments %q", comm.StartCmd.Command, expectedArgs)
	}

	// Test with missing parameter
	delete(config, "extra_arguments")

	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	err = p.Provision(ui, comm)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Check the expected `extra_arguments` position for an empty value
	splitCommand := strings.Split(comm.StartCmd.Command, " ")
	if "" == splitCommand[len(splitCommand)-2] {
		t.Fatalf("Command %q contains an extra-space which may cause arg parsing issues", comm.StartCmd.Command)
	}
}

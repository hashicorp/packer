package puppetmasterless

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"testing"
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

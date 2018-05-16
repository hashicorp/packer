package puppetserver

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() map[string]interface{} {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		panic(err)
	}

	return map[string]interface{}{
		"puppet_server": tf.Name(),
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_puppetBinDir(t *testing.T) {
	config := testConfig()

	delete(config, "puppet_bin_dir")
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

	config["puppet_bin_dir"] = tf.Name()
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_clientPrivateKeyPath(t *testing.T) {
	config := testConfig()

	delete(config, "client_private_key_path")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with bad paths
	config["client_private_key_path"] = "i-should-not-exist"
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

	config["client_private_key_path"] = td
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_clientCertPath(t *testing.T) {
	config := testConfig()

	delete(config, "client_cert_path")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with bad paths
	config["client_cert_path"] = "i-should-not-exist"
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

	config["client_cert_path"] = td
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvisionerPrepare_executeCommand(t *testing.T) {
	config := testConfig()

	delete(config, "execute_command")
	p := new(Provisioner)
	err := p.Prepare(config)
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
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	if p.config.Facter == nil {
		t.Fatalf("err: Default facts are not set in the Puppet provisioner!")
	}

	if _, ok := p.config.Facter["packer_build_name"]; !ok {
		t.Fatalf("err: packer_build_name fact not set in the Puppet provisioner!")
	}

	if _, ok := p.config.Facter["packer_builder_type"]; !ok {
		t.Fatalf("err: packer_builder_type fact not set in the Puppet provisioner!")
	}
}

func TestProvisionerPrepare_stagingDir(t *testing.T) {
	config := testConfig()

	delete(config, "staging_dir")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Make sure the default staging directory is correct
	if p.config.StagingDir != "/tmp/packer-puppet-server" {
		t.Fatalf("err: Default staging_dir is not set in the Puppet provisioner!")
	}

	// Make sure default staging directory can be overridden
	config["staging_dir"] = "/tmp/override"
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.StagingDir != "/tmp/override" {
		t.Fatalf("err: Overridden staging_dir is not set correctly in the Puppet provisioner!")
	}
}

package chefclient

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/mitchellh/packer/packer"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"server_url": "foo",
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_configTemplate(t *testing.T) {
	var err error
	var p Provisioner

	// Test no config template
	config := testConfig()
	delete(config, "config_template")
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a file
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())

	config = testConfig()
	config["config_template"] = tf.Name()
	p = Provisioner{}
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Test with a directory
	td, err := ioutil.TempDir("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.RemoveAll(td)

	config = testConfig()
	config["config_template"] = td
	p = Provisioner{}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have err")
	}
}

func TestProvisionerPrepare_commands(t *testing.T) {
	commands := []string{
		"execute_command",
		"install_command",
	}

	for _, command := range commands {
		var p Provisioner

		// Test not set
		config := testConfig()
		delete(config, command)
		err := p.Prepare(config)
		if err != nil {
			t.Fatalf("err: %s", err)
		}

		// Test invalid template
		config = testConfig()
		config[command] = "{{if NOPE}}"
		err = p.Prepare(config)
		if err == nil {
			t.Fatal("should error")
		}

		// Test good template
		config = testConfig()
		config[command] = "{{.Foo}}"
		err = p.Prepare(config)
		if err != nil {
			t.Fatalf("err: %s", err)
		}
	}
}

func TestProvisionerPrepare_serverUrl(t *testing.T) {
	var p Provisioner

	// Test not set
	config := testConfig()
	delete(config, "server_url")
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should error")
	}

	// Test set
	config = testConfig()
	config["server_url"] = "foo"
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

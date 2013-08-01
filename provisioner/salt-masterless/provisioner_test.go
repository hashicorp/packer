package saltmasterless

import (
	"github.com/mitchellh/packer/packer"
	"os"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"local_state_tree": os.TempDir(),
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
	}
}

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.TempConfigDir != DefaultTempConfigDir {
		t.Errorf("unexpected temp config dir: %s", p.config.TempConfigDir)
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

func TestProvisionerPrepare_LocalStateTree(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["local_state_tree"] = "/i/dont/exist/i/think"
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	config["local_state_tree"] = os.TempDir()
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

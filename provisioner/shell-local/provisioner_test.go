package shelllocal

import (
	"github.com/mitchellh/packer/packer"
	"runtime"
	"strings"
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"inline": []interface{}{"foo", "bar"},
	}
}

func TestProvisioner_Impl(t *testing.T) {
	var raw interface{}
	raw = &Provisioner{}
	if _, ok := raw.(packer.Provisioner); !ok {
		t.Fatalf("must be a Provisioner")
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

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.InlineShebang != DefaultSheBang {
		t.Errorf("unexpected inline shebang: %s", p.config.InlineShebang)
	}
}

func TestProvisionerPrepare_InlineShebang(t *testing.T) {
	config := testConfig()

	delete(config, "inline_shebang")
	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.InlineShebang != "/bin/sh" {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}

	// Test with a good one
	config["inline_shebang"] = "foo"
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.InlineShebang != "foo" {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}
}

func TestProvisionerPrepare_Inline(t *testing.T) {
	var p Provisioner
	config := testConfig()

	delete(config, "inline")
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	config["inline"] = []interface{}{"foo"}
	err = p.Prepare(config)
	if err != nil {
		t.Fatal("should not have error %s", err)
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

func TestProvisionerProvision_Inline(t *testing.T) {
	var p Provisioner
	config := testConfig()

	config["inline"] = []interface{}{"echo \"Hello World\""}
	err := p.Prepare(config)
	if err != nil {
		t.Fatal("should not have error %s", err)
	}

	if runtime.GOOS == "windows" {
		//the rest of this test only runs on non-windows systems
		return
	}

	ui := &stubUi{}
	comm := &packer.MockCommunicator{}
	err = p.Provision(ui, comm)
	if err != nil {
		t.Fatalf("should successfully provision: %s", err)
	}

	if !strings.Contains(ui.sayMessages, "Provisioning with local shell script:") {
		t.Fatalf("should print Provisioning with local shell script")
	}

	if !strings.Contains(ui.sayMessages, "Hello World") {
		t.Fatalf("should print Hello")
	}

}

package shell

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
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

func TestProvisionerPrepare_Defaults(t *testing.T) {
	var p Provisioner
	config := testConfig()

	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if p.config.GuestOsType != defaultOsType {
		t.Errorf("defaultOsType: expecting %s, found %s", defaultOsType, p.config.GuestOsType)
	}

	if p.config.ExecuteCommand != guestOsConfigs[defaultOsType].executeCommand {
		t.Errorf("ExecuteCommand: expecting %s, found %s", guestOsConfigs[defaultOsType].executeCommand, p.config.ExecuteCommand)
	}

	if p.config.InlineShebang != guestOsConfigs[defaultOsType].inlineShebang {
		t.Errorf("InlineShebang: expecting %s, found %s", guestOsConfigs[defaultOsType].inlineShebang, p.config.InlineShebang)
	}

	if p.config.RemotePath != guestOsConfigs[defaultOsType].remotePath {
		t.Errorf("RemotePath: expecting %s, found %s", guestOsConfigs[defaultOsType].remotePath, p.config.RemotePath)
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

	if p.config.InlineShebang != guestOsConfigs[defaultOsType].inlineShebang {
		t.Errorf("InlineShebang: expecting %s, found %s", guestOsConfigs[defaultOsType].inlineShebang, p.config.InlineShebang)
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

func TestProvisionerPrepare_Script(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	config["script"] = "/this/should/not/exist"
	p := new(Provisioner)
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["script"] = tf.Name()
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestProvisionerPrepare_ScriptAndInline(t *testing.T) {
	var p Provisioner
	config := testConfig()

	delete(config, "inline")
	delete(config, "script")
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with both
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["inline"] = []interface{}{"foo"}
	config["script"] = tf.Name()
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_ScriptAndScripts(t *testing.T) {
	var p Provisioner
	config := testConfig()

	// Test with both
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["inline"] = []interface{}{"foo"}
	config["scripts"] = []string{tf.Name()}
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestProvisionerPrepare_Scripts(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	config["scripts"] = []string{}
	p := new(Provisioner)
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["scripts"] = []string{tf.Name()}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestProvisionerPrepare_EnvironmentVars(t *testing.T) {
	config := testConfig()

	// Test with a bad case
	config["environment_vars"] = []string{"badvar", "good=var"}
	p := new(Provisioner)
	err := p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a trickier case
	config["environment_vars"] = []string{"=bad"}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good case
	// Note: baz= is a real env variable, just empty
	config["environment_vars"] = []string{"FOO=bar", "baz="}
	p = new(Provisioner)
	err = p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestProvisioner_createFlattenedEnvVars_unix(t *testing.T) {
	config := testConfig()

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error preparing config: %s", err)
	}

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	// no user env var
	flattenedEnvVars, err := p.createFlattenedEnvVars()
	if err != nil {
		t.Fatalf("should not have error creating flattened env vars: %s", err)
	}
	if flattenedEnvVars != "PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware " {
		t.Fatalf("unexpected flattened env vars: %s", flattenedEnvVars)
	}

	// single user env var
	p.config.Vars = []string{"foo=bar"}

	flattenedEnvVars, err = p.createFlattenedEnvVars()
	if err != nil {
		t.Fatalf("should not have error creating flattened env vars: %s", err)
	}
	if flattenedEnvVars != "PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware foo=bar " {
		t.Fatalf("unexpected flattened env vars: %s", flattenedEnvVars)
	}

	// multiple user env vars
	p.config.Vars = []string{"FOO=bar", "BAZ=qux"}

	flattenedEnvVars, err = p.createFlattenedEnvVars()
	if err != nil {
		t.Fatalf("should not have error creating flattened env vars: %s", err)
	}
	if flattenedEnvVars != "BAZ=qux FOO=bar PACKER_BUILDER_TYPE=iso PACKER_BUILD_NAME=vmware " {
		t.Fatalf("unexpected flattened env vars: %s", flattenedEnvVars)
	}
}

func TestProvisioner_createFlattenedEnvVars_windows(t *testing.T) {
	config := testConfig()
	config["guest_os_type"] = "windows"

	p := new(Provisioner)
	err := p.Prepare(config)
	if err != nil {
		t.Fatalf("should not have error preparing config: %s", err)
	}

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	// no user env var
	flattenedEnvVars, err := p.createFlattenedEnvVars()
	if err != nil {
		t.Fatalf("should not have error creating flattened env vars: %s", err)
	}
	if flattenedEnvVars != "`$env:PACKER_BUILDER_TYPE='iso'; `$env:PACKER_BUILD_NAME='vmware'; " {
		t.Fatalf("unexpected flattened env vars: %s", flattenedEnvVars)
	}

	// single user env var
	p.config.Vars = []string{"FOO=bar"}

	flattenedEnvVars, err = p.createFlattenedEnvVars()
	if err != nil {
		t.Fatalf("should not have error creating flattened env vars: %s", err)
	}
	if flattenedEnvVars != "`$env:FOO='bar'; `$env:PACKER_BUILDER_TYPE='iso'; `$env:PACKER_BUILD_NAME='vmware'; " {
		t.Fatalf("unexpected flattened env vars: %s", flattenedEnvVars)
	}

	// multiple user env vars
	p.config.Vars = []string{"FOO=bar", "BAZ=qux"}

	flattenedEnvVars, err = p.createFlattenedEnvVars()
	if err != nil {
		t.Fatalf("should not have error creating flattened env vars: %s", err)
	}
	if flattenedEnvVars != "`$env:BAZ='qux'; `$env:FOO='bar'; `$env:PACKER_BUILDER_TYPE='iso'; `$env:PACKER_BUILD_NAME='vmware'; " {
		t.Fatalf("unexpected flattened env vars: %s", flattenedEnvVars)
	}
}

func TestProvisionerQuote_EnvironmentVars(t *testing.T) {
	config := testConfig()

	config["environment_vars"] = []string{"keyone=valueone", "keytwo=value\ntwo"}
	p := new(Provisioner)
	p.Prepare(config)

	expectedValue := "keyone='valueone'"
	if p.config.Vars[0] != expectedValue {
		t.Fatalf("%s should be equal to %s", p.config.Vars[0], expectedValue)
	}

	expectedValue = "keytwo='value\ntwo'"
	if p.config.Vars[1] != expectedValue {
		t.Fatalf("%s should be equal to %s", p.config.Vars[1], expectedValue)
	}
}

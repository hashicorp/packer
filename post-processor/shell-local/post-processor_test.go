package shell_local

import (
	"github.com/mitchellh/packer/packer"
	"io/ioutil"
	"os"
	"testing"
)

func TestPostProcessor_ImplementsPostProcessor(t *testing.T) {
	var _ packer.PostProcessor = new(PostProcessor)
}

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"inline": []interface{}{"foo", "bar"},
	}
}

func TestPostProcessor_Impl(t *testing.T) {
	var raw interface{}
	raw = &PostProcessor{}
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatalf("must be a post processor")
	}
}

func TestPostProcessorPrepare_Defaults(t *testing.T) {
	var p PostProcessor
	config := testConfig()

	err := p.Configure(config)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPostProcessorPrepare_InlineShebang(t *testing.T) {
	config := testConfig()

	delete(config, "inline_shebang")
	p := new(PostProcessor)
	err := p.Configure(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.InlineShebang != "/bin/sh -e" {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}

	// Test with a good one
	config["inline_shebang"] = "foo"
	p = new(PostProcessor)
	err = p.Configure(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.InlineShebang != "foo" {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}
}

func TestPostProcessorPrepare_InvalidKey(t *testing.T) {
	var p PostProcessor
	config := testConfig()

	// Add a random key
	config["i_should_not_be_valid"] = true
	err := p.Configure(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPostProcessorPrepare_Script(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	config["script"] = "/this/should/not/exist"
	p := new(PostProcessor)
	err := p.Configure(config)
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
	p = new(PostProcessor)
	err = p.Configure(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestPostProcessorPrepare_ScriptAndInline(t *testing.T) {
	var p PostProcessor
	config := testConfig()

	delete(config, "inline")
	delete(config, "script")
	err := p.Configure(config)
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
	err = p.Configure(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPostProcessorPrepare_ScriptAndScripts(t *testing.T) {
	var p PostProcessor
	config := testConfig()

	// Test with both
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	config["inline"] = []interface{}{"foo"}
	config["scripts"] = []string{tf.Name()}
	err = p.Configure(config)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPostProcessorPrepare_Scripts(t *testing.T) {
	config := testConfig()
	delete(config, "inline")

	config["scripts"] = []string{}
	p := new(PostProcessor)
	err := p.Configure(config)
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
	p = new(PostProcessor)
	err = p.Configure(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestPostProcessorPrepare_EnvironmentVars(t *testing.T) {
	config := testConfig()

	// Test with a bad case
	config["environment_vars"] = []string{"badvar", "good=var"}
	p := new(PostProcessor)
	err := p.Configure(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a trickier case
	config["environment_vars"] = []string{"=bad"}
	p = new(PostProcessor)
	err = p.Configure(config)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good case
	// Note: baz= is a real env variable, just empty
	config["environment_vars"] = []string{"FOO=bar", "baz="}
	p = new(PostProcessor)
	err = p.Configure(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test when the env variable value contains an equals sign
	config["environment_vars"] = []string{"good=withequals=true"}
	p = new(PostProcessor)
	err = p.Configure(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test when the env variable value starts with an equals sign
	config["environment_vars"] = []string{"good==true"}
	p = new(PostProcessor)
	err = p.Configure(config)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestPostProcessor_createFlattenedEnvVars(t *testing.T) {
	var flattenedEnvVars string
	config := testConfig()

	userEnvVarTests := [][]string{
		{},                     // No user env var
		{"FOO=bar"},            // Single user env var
		{"FOO=bar's"},          // User env var with single quote in value
		{"FOO=bar", "BAZ=qux"}, // Multiple user env vars
		{"FOO=bar=baz"},        // User env var with value containing equals
		{"FOO==bar"},           // User env var with value starting with equals
	}
	expected := []string{
		`PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`FOO='bar' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`FOO='bar'"'"'s' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`BAZ='qux' FOO='bar' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`FOO='bar=baz' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
		`FOO='=bar' PACKER_BUILDER_TYPE='iso' PACKER_BUILD_NAME='vmware' `,
	}

	p := new(PostProcessor)
	p.Configure(config)

	// Defaults provided by Packer
	p.config.PackerBuildName = "vmware"
	p.config.PackerBuilderType = "iso"

	for i, expectedValue := range expected {
		p.config.Vars = userEnvVarTests[i]
		flattenedEnvVars = p.createFlattenedEnvVars()
		if flattenedEnvVars != expectedValue {
			t.Fatalf("expected flattened env vars to be: %s, got %s.", expectedValue, flattenedEnvVars)
		}
	}
}

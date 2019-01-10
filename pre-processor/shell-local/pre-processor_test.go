package shell_local

import (
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func TestPreProcessor_ImplementsPreProcessor(t *testing.T) {
	var _ packer.PreProcessor = new(PreProcessor)
}

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"inline": []interface{}{"foo", "bar"},
	}
}

func TestPreProcessor_Impl(t *testing.T) {
	var raw interface{}
	raw = &PreProcessor{}
	if _, ok := raw.(packer.PreProcessor); !ok {
		t.Fatalf("must be a pre processor")
	}
}

func TestPreProcessorPrepare_Defaults(t *testing.T) {
	var p PreProcessor
	raws := testConfig()

	err := p.Configure(raws)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestPreProcessorPrepare_InlineShebang(t *testing.T) {
	raws := testConfig()

	delete(raws, "inline_shebang")
	p := new(PreProcessor)
	err := p.Configure(raws)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	expected := ""
	if runtime.GOOS != "windows" {
		expected = "/bin/sh -e"
	}
	if p.config.InlineShebang != expected {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}

	// Test with a good one
	raws["inline_shebang"] = "foo"
	p = new(PreProcessor)
	err = p.Configure(raws)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	if p.config.InlineShebang != "foo" {
		t.Fatalf("bad value: %s", p.config.InlineShebang)
	}
}

func TestPreProcessorPrepare_InvalidKey(t *testing.T) {
	var p PreProcessor
	raws := testConfig()

	// Add a random key
	raws["i_should_not_be_valid"] = true
	err := p.Configure(raws)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPreProcessorPrepare_Script(t *testing.T) {
	raws := testConfig()
	delete(raws, "inline")

	raws["script"] = "/this/should/not/exist"
	p := new(PreProcessor)
	err := p.Configure(raws)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	raws["script"] = tf.Name()
	p = new(PreProcessor)
	err = p.Configure(raws)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestPreProcessorPrepare_ExecuteCommand(t *testing.T) {
	// Check that passing a string will work (Backwards Compatibility)
	p := new(PreProcessor)
	raws := testConfig()
	raws["execute_command"] = "foo bar"
	err := p.Configure(raws)
	expected := []string{"sh", "-c", "foo bar"}
	if err != nil {
		t.Fatalf("should handle backwards compatibility: %s", err)
	}
	assert.Equal(t, p.config.ExecuteCommand, expected,
		"Did not get expected execute_command: expected: %#v; received %#v", expected, p.config.ExecuteCommand)

	// Check that passing a list will work
	p = new(PreProcessor)
	raws = testConfig()
	raws["execute_command"] = []string{"foo", "bar"}
	err = p.Configure(raws)
	if err != nil {
		t.Fatalf("should handle backwards compatibility: %s", err)
	}
	expected = []string{"foo", "bar"}
	assert.Equal(t, p.config.ExecuteCommand, expected,
		"Did not get expected execute_command: expected: %#v; received %#v", expected, p.config.ExecuteCommand)

	// Check that default is as expected
	raws = testConfig()
	delete(raws, "execute_command")
	p = new(PreProcessor)
	p.Configure(raws)
	if runtime.GOOS != "windows" {
		expected = []string{"/bin/sh", "-c", "{{.Vars}} {{.Script}}"}
	} else {
		expected = []string{"cmd", "/V", "/C", "{{.Vars}}", "call", "{{.Script}}"}
	}
	assert.Equal(t, p.config.ExecuteCommand, expected,
		"Did not get expected default: expected: %#v; received %#v", expected, p.config.ExecuteCommand)
}

func TestPreProcessorPrepare_ScriptAndInline(t *testing.T) {
	var p PreProcessor
	raws := testConfig()

	// Error if no scripts/inline commands provided
	delete(raws, "inline")
	delete(raws, "script")
	delete(raws, "command")
	delete(raws, "scripts")
	err := p.Configure(raws)
	if err == nil {
		t.Fatalf("should error when no scripts/inline commands are provided")
	}

	// Test with both
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	raws["inline"] = []interface{}{"foo"}
	raws["script"] = tf.Name()
	err = p.Configure(raws)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPreProcessorPrepare_ScriptAndScripts(t *testing.T) {
	var p PreProcessor
	raws := testConfig()

	// Test with both
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	raws["inline"] = []interface{}{"foo"}
	raws["scripts"] = []string{tf.Name()}
	err = p.Configure(raws)
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestPreProcessorPrepare_Scripts(t *testing.T) {
	raws := testConfig()
	delete(raws, "inline")

	raws["scripts"] = []string{}
	p := new(PreProcessor)
	err := p.Configure(raws)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good one
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())

	raws["scripts"] = []string{tf.Name()}
	p = new(PreProcessor)
	err = p.Configure(raws)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestPreProcessorPrepare_EnvironmentVars(t *testing.T) {
	raws := testConfig()

	// Test with a bad case
	raws["environment_vars"] = []string{"badvar", "good=var"}
	p := new(PreProcessor)
	err := p.Configure(raws)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a trickier case
	raws["environment_vars"] = []string{"=bad"}
	p = new(PreProcessor)
	err = p.Configure(raws)
	if err == nil {
		t.Fatal("should have error")
	}

	// Test with a good case
	// Note: baz= is a real env variable, just empty
	raws["environment_vars"] = []string{"FOO=bar", "baz="}
	p = new(PreProcessor)
	err = p.Configure(raws)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test when the env variable value contains an equals sign
	raws["environment_vars"] = []string{"good=withequals=true"}
	p = new(PreProcessor)
	err = p.Configure(raws)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test when the env variable value starts with an equals sign
	raws["environment_vars"] = []string{"good==true"}
	p = new(PreProcessor)
	err = p.Configure(raws)
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

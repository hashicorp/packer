package packer

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"os"
	"testing"
)

func testEnvironment() *Environment {
	config := &EnvironmentConfig{}
	config.ui = &ReaderWriterUi{
		new(bytes.Buffer),
		new(bytes.Buffer),
	}

	return NewEnvironment(config)
}

func TestEnvironment_Cli_CallsRun(t *testing.T) {
	//_ := asserts.NewTestingAsserts(t, true)

	// TODO: Test that the call to `Run` is done with
	// proper arguments and such.
}

func TestEnvironment_DefaultCli_Empty(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := testEnvironment()

	assert.Equal(defaultEnv.Cli([]string{}), 1, "CLI with no args")
}

func TestEnvironment_DefaultCli_Help(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := testEnvironment()

	// Test the basic version options
	assert.Equal(defaultEnv.Cli([]string{"--help"}), 1, "--help should print")
	assert.Equal(defaultEnv.Cli([]string{"-h"}), 1, "--help should print")
}

func TestEnvironment_DefaultCli_Version(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := testEnvironment()

	// Test the basic version options
	assert.Equal(defaultEnv.Cli([]string{"version"}), 0, "version should work")
	assert.Equal(defaultEnv.Cli([]string{"--version"}), 0, "--version should work")
	assert.Equal(defaultEnv.Cli([]string{"-v"}), 0, "-v should work")

	// Test the --version and -v can appear anywhere
	assert.Equal(defaultEnv.Cli([]string{"bad", "-v"}), 0, "-v should work anywhere")
	assert.Equal(defaultEnv.Cli([]string{"bad", "--version"}), 0, "--version should work anywhere")

	// Test that "version" can't appear anywhere
	assert.Equal(defaultEnv.Cli([]string{"bad", "version"}), 1, "version should NOT work anywhere")
}

func TestEnvironment_DefaultUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := NewEnvironment(nil)
	assert.NotNil(defaultEnv.Ui(), "default UI should not be nil")

	rwUi, ok := defaultEnv.Ui().(*ReaderWriterUi)
	assert.True(ok, "default UI should be ReaderWriterUi")
	assert.Equal(rwUi.Writer, os.Stdout, "default UI should go to stdout")
	assert.Equal(rwUi.Reader, os.Stdin, "default UI should read from stdin")
}

func TestEnvironment_PrintHelp(t *testing.T) {
	// Just call the function and verify that no panics occur
	testEnvironment().PrintHelp()
}

func TestEnvironment_SettingUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := &ReaderWriterUi{new(bytes.Buffer), new(bytes.Buffer)}

	config := &EnvironmentConfig{}
	config.ui = ui

	env := NewEnvironment(config)

	assert.Equal(env.Ui(), ui, "UIs should be equal")
}

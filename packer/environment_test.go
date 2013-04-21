package packer

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"os"
	"strings"
	"testing"
)

type TestCommand struct {
	runArgs []string
	runCalled bool
	runEnv *Environment
}

func (tc *TestCommand) Run(env *Environment, args []string) int {
	tc.runCalled = true
	tc.runArgs = args
	tc.runEnv = env
	return 0
}

func (tc *TestCommand) Synopsis() string {
	return ""
}

func testEnvironment() *Environment {
	config := &EnvironmentConfig{}
	config.Ui = &ReaderWriterUi{
		new(bytes.Buffer),
		new(bytes.Buffer),
	}

	return NewEnvironment(config)
}

func TestEnvironment_Cli_CallsRun(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	command := &TestCommand{}

	config := &EnvironmentConfig{}
	config.Command = make(map[string]Command)
	config.Command["foo"] = command

	env := NewEnvironment(config)
	assert.Equal(env.Cli([]string{"foo", "bar", "baz"}), 0, "runs foo command")
	assert.True(command.runCalled, "run should've been called")
	assert.Equal(command.runEnv, env, "should've ran with env")
	assert.Equal(command.runArgs, []string{"bar", "baz"}, "should have right args")
}

func TestEnvironment_DefaultCli_Empty(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := testEnvironment()

	assert.Equal(defaultEnv.Cli([]string{}), 1, "CLI with no args")
}

func TestEnvironment_DefaultCli_Help(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := testEnvironment()

	// A little lambda to help us test the output actually contains help
	testOutput := func() {
		buffer := defaultEnv.Ui().(*ReaderWriterUi).Writer.(*bytes.Buffer)
		output := buffer.String()
		buffer.Reset()
		assert.True(strings.Contains(output, "usage: packer"), "should print help")
	}

	// Test "--help"
	assert.Equal(defaultEnv.Cli([]string{"--help"}), 1, "--help should print")
	testOutput()

	// Test "-h"
	assert.Equal(defaultEnv.Cli([]string{"-h"}), 1, "--help should print")
	testOutput()
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
	config.Ui = ui

	env := NewEnvironment(config)

	assert.Equal(env.Ui(), ui, "UIs should be equal")
}

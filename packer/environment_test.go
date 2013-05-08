package packer

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"fmt"
	"os"
	"strings"
	"testing"
)

func testEnvironment() Environment {
	config := DefaultEnvironmentConfig()
	config.Ui = &ReaderWriterUi{
		new(bytes.Buffer),
		new(bytes.Buffer),
	}

	env, err := NewEnvironment(config)
	if err != nil {
		panic(err)
	}

	return env
}

func TestEnvironment_DefaultConfig_Commands(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	assert.Empty(config.Commands, "should have no commands")
}

func TestEnvironment_DefaultConfig_Ui(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	assert.NotNil(config.Ui, "default UI should not be nil")

	rwUi, ok := config.Ui.(*ReaderWriterUi)
	assert.True(ok, "default UI should be ReaderWriterUi")
	assert.Equal(rwUi.Writer, os.Stdout, "default UI should go to stdout")
	assert.Equal(rwUi.Reader, os.Stdin, "default UI should read from stdin")
}

func TestNewEnvironment_NoConfig(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	env, err := NewEnvironment(nil)
	assert.Nil(env, "env should be nil")
	assert.NotNil(err, "should be an error")
}

func TestEnvironment_Builder(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	builder := &TestBuilder{}
	builders := make(map[string]Builder)
	builders["foo"] = builder

	config := DefaultEnvironmentConfig()
	config.BuilderFunc = func(n string) (Builder, error) { return builders[n], nil }

	env, _ := NewEnvironment(config)
	returnedBuilder, err := env.Builder("foo")
	assert.Nil(err, "should be no error")
	assert.Equal(returnedBuilder, builder, "should return correct builder")
}

func TestEnvironment_Cli_CallsRun(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	command := &TestCommand{}
	commands := make(map[string]Command)
	commands["foo"] = command

	config := &EnvironmentConfig{}
	config.Commands = []string{"foo"}
	config.CommandFunc = func(n string) (Command, error) { return commands[n], nil }

	env, _ := NewEnvironment(config)
	exitCode, err := env.Cli([]string{"foo", "bar", "baz"})
	assert.Nil(err, "should be no error")
	assert.Equal(exitCode, 0, "runs foo command")
	assert.True(command.runCalled, "run should've been called")
	assert.Equal(command.runEnv, env, "should've ran with env")
	assert.Equal(command.runArgs, []string{"bar", "baz"}, "should have right args")
}

func TestEnvironment_DefaultCli_Empty(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := testEnvironment()

	exitCode, _ := defaultEnv.Cli([]string{})
	assert.Equal(exitCode, 1, "CLI with no args")
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
	exitCode, _ := defaultEnv.Cli([]string{"--help"})
	assert.Equal(exitCode, 1, "--help should print")
	testOutput()

	// Test "-h"
	exitCode, _ = defaultEnv.Cli([]string{"--help"})
	assert.Equal(exitCode, 1, "--help should print")
	testOutput()
}

func TestEnvironment_DefaultCli_Version(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	defaultEnv := testEnvironment()

	versionCommands := []string{"version", "--version", "-v"}
	for _, command := range versionCommands {
		exitCode, _ := defaultEnv.Cli([]string{command})
		assert.Equal(exitCode, 0, fmt.Sprintf("%s should work", command))

		// Test the --version and -v can appear anywhere
		exitCode, _ = defaultEnv.Cli([]string{"bad", command})

		if command != "version" {
			assert.Equal(exitCode, 0, fmt.Sprintf("%s should work anywhere", command))
		} else {
			assert.Equal(exitCode, 1, fmt.Sprintf("%s should NOT work anywhere", command))
		}
	}
}

func TestEnvironment_SettingUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := &ReaderWriterUi{new(bytes.Buffer), new(bytes.Buffer)}

	config := &EnvironmentConfig{}
	config.Ui = ui

	env, _ := NewEnvironment(config)

	assert.Equal(env.Ui(), ui, "UIs should be equal")
}

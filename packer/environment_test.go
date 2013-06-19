package packer

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

func init() {
	// Disable log output for tests
	log.SetOutput(ioutil.Discard)
}

func testEnvironment() Environment {
	config := DefaultEnvironmentConfig()
	config.Ui = &ReaderWriterUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
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

func TestEnvironment_NilComponents(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components = *new(ComponentFinder)

	env, err := NewEnvironment(config)
	assert.Nil(err, "should not have an error")

	// All of these should not cause panics... so we don't assert
	// anything but if there is a panic in the test then yeah, something
	// went wrong.
	env.Builder("foo")
	env.Cli([]string{"foo"})
	env.Hook("foo")
	env.PostProcessor("foo")
	env.Provisioner("foo")
}

func TestEnvironment_Builder(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	builder := &TestBuilder{}
	builders := make(map[string]Builder)
	builders["foo"] = builder

	config := DefaultEnvironmentConfig()
	config.Components.Builder = func(n string) (Builder, error) { return builders[n], nil }

	env, _ := NewEnvironment(config)
	returnedBuilder, err := env.Builder("foo")
	assert.Nil(err, "should be no error")
	assert.Equal(returnedBuilder, builder, "should return correct builder")
}

func TestEnvironment_Builder_NilError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.Builder = func(n string) (Builder, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returnedBuilder, err := env.Builder("foo")
	assert.NotNil(err, "should be an error")
	assert.Nil(returnedBuilder, "should be no builder")
}

func TestEnvironment_Builder_Error(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.Builder = func(n string) (Builder, error) { return nil, errors.New("foo") }

	env, _ := NewEnvironment(config)
	returnedBuilder, err := env.Builder("foo")
	assert.NotNil(err, "should be an error")
	assert.Equal(err.Error(), "foo", "should be correct error")
	assert.Nil(returnedBuilder, "should be no builder")
}

func TestEnvironment_Cache(t *testing.T) {
	config := DefaultEnvironmentConfig()
	env, _ := NewEnvironment(config)
	if env.Cache() == nil {
		t.Fatal("cache should not be nil")
	}
}

func TestEnvironment_Cli_Error(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.Command = func(n string) (Command, error) { return nil, errors.New("foo") }

	env, _ := NewEnvironment(config)
	_, err := env.Cli([]string{"foo"})
	assert.NotNil(err, "should be an error")
	assert.Equal(err.Error(), "foo", "should be correct error")
}

func TestEnvironment_Cli_CallsRun(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	command := &TestCommand{}
	commands := make(map[string]Command)
	commands["foo"] = command

	config := &EnvironmentConfig{}
	config.Commands = []string{"foo"}
	config.Components.Command = func(n string) (Command, error) { return commands[n], nil }

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

func TestEnvironment_Hook(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	hook := &TestHook{}
	hooks := make(map[string]Hook)
	hooks["foo"] = hook

	config := DefaultEnvironmentConfig()
	config.Components.Hook = func(n string) (Hook, error) { return hooks[n], nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Hook("foo")
	assert.Nil(err, "should be no error")
	assert.Equal(returned, hook, "should return correct hook")
}

func TestEnvironment_Hook_NilError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.Hook = func(n string) (Hook, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Hook("foo")
	assert.NotNil(err, "should be an error")
	assert.Nil(returned, "should be no hook")
}

func TestEnvironment_Hook_Error(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.Hook = func(n string) (Hook, error) { return nil, errors.New("foo") }

	env, _ := NewEnvironment(config)
	returned, err := env.Hook("foo")
	assert.NotNil(err, "should be an error")
	assert.Equal(err.Error(), "foo", "should be correct error")
	assert.Nil(returned, "should be no hook")
}

func TestEnvironment_PostProcessor(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	pp := &TestPostProcessor{}
	pps := make(map[string]PostProcessor)
	pps["foo"] = pp

	config := DefaultEnvironmentConfig()
	config.Components.PostProcessor = func(n string) (PostProcessor, error) { return pps[n], nil }

	env, _ := NewEnvironment(config)
	returned, err := env.PostProcessor("foo")
	assert.Nil(err, "should be no error")
	assert.Equal(returned, pp, "should return correct pp")
}

func TestEnvironment_PostProcessor_NilError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.PostProcessor = func(n string) (PostProcessor, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returned, err := env.PostProcessor("foo")
	assert.NotNil(err, "should be an error")
	assert.Nil(returned, "should be no pp")
}

func TestEnvironment_PostProcessor_Error(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.PostProcessor = func(n string) (PostProcessor, error) { return nil, errors.New("foo") }

	env, _ := NewEnvironment(config)
	returned, err := env.PostProcessor("foo")
	assert.NotNil(err, "should be an error")
	assert.Equal(err.Error(), "foo", "should be correct error")
	assert.Nil(returned, "should be no pp")
}

func TestEnvironmentProvisioner(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	p := &TestProvisioner{}
	ps := make(map[string]Provisioner)
	ps["foo"] = p

	config := DefaultEnvironmentConfig()
	config.Components.Provisioner = func(n string) (Provisioner, error) { return ps[n], nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Provisioner("foo")
	assert.Nil(err, "should be no error")
	assert.Equal(returned, p, "should return correct provisioner")
}

func TestEnvironmentProvisioner_NilError(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.Provisioner = func(n string) (Provisioner, error) { return nil, nil }

	env, _ := NewEnvironment(config)
	returned, err := env.Provisioner("foo")
	assert.NotNil(err, "should be an error")
	assert.Nil(returned, "should be no provisioner")
}

func TestEnvironmentProvisioner_Error(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	config := DefaultEnvironmentConfig()
	config.Components.Provisioner = func(n string) (Provisioner, error) {
		return nil, errors.New("foo")
	}

	env, _ := NewEnvironment(config)
	returned, err := env.Provisioner("foo")
	assert.NotNil(err, "should be an error")
	assert.Equal(err.Error(), "foo", "should be correct error")
	assert.Nil(returned, "should be no provisioner")
}

func TestEnvironment_SettingUi(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	ui := &ReaderWriterUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}

	config := &EnvironmentConfig{}
	config.Ui = ui

	env, _ := NewEnvironment(config)

	assert.Equal(env.Ui(), ui, "UIs should be equal")
}

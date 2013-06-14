package build

import (
	"bytes"
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testEnvironment() packer.Environment {
	config := packer.DefaultEnvironmentConfig()
	config.Ui = &packer.ReaderWriterUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	}

	env, err := packer.NewEnvironment(config)
	if err != nil {
		panic(err)
	}

	return env
}

func TestCommand_Implements(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var actual packer.Command
	assert.Implementor(new(Command), &actual, "should be a Command")
}

func TestCommand_Run_NoArgs(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	command := new(Command)
	result := command.Run(testEnvironment(), make([]string, 0))
	assert.Equal(result, 1, "no args should error")
}

func TestCommand_Run_MoreThanOneArg(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	command := new(Command)

	args := []string{"one", "two"}
	result := command.Run(testEnvironment(), args)
	assert.Equal(result, 1, "More than one arg should fail")
}

func TestCommand_Run_MissingFile(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)
	command := new(Command)

	args := []string{"i-better-not-exist"}
	result := command.Run(testEnvironment(), args)
	assert.Equal(result, 1, "a non-existent file should error")
}

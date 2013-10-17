package build

import (
	"bytes"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testEnvironment() packer.Environment {
	config := packer.DefaultEnvironmentConfig()
	config.Ui = &packer.BasicUi{
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
	var _ packer.Command = new(Command)
}

func TestCommand_Run_NoArgs(t *testing.T) {
	command := new(Command)
	result := command.Run(testEnvironment(), make([]string, 0))
	if result != 1 {
		t.Fatalf("bad: %d", result)
	}
}

func TestCommand_Run_MoreThanOneArg(t *testing.T) {
	command := new(Command)

	args := []string{"one", "two"}
	result := command.Run(testEnvironment(), args)
	if result != 1 {
		t.Fatalf("bad: %d", result)
	}
}

func TestCommand_Run_MissingFile(t *testing.T) {
	command := new(Command)

	args := []string{"i-better-not-exist"}
	result := command.Run(testEnvironment(), args)
	if result != 1 {
		t.Fatalf("bad: %d", result)
	}
}

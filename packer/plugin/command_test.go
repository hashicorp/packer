package plugin

import (
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"testing"
)

type helperCommand byte

func (helperCommand) Help() string {
	return "2"
}

func (helperCommand) Run(packer.Environment, []string) int {
	return 42
}

func (helperCommand) Synopsis() string {
	return "1"
}

func TestCommand_NoExist(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: exec.Command("i-should-not-exist")})
	defer c.Kill()

	_, err := c.Command()
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestCommand_Good(t *testing.T) {
	c := NewClient(&ClientConfig{Cmd: helperProcess("command")})
	defer c.Kill()

	command, err := c.Command()
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	result := command.Synopsis()
	if result != "1" {
		t.Errorf("synopsis not correct: %s", result)
	}

	result = command.Help()
	if result != "2" {
		t.Errorf("help not correct: %s", result)
	}
}

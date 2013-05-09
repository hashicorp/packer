package plugin

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"testing"
)

// TODO: Test command cleanup functionality
// TODO: Test timeout functionality

type helperCommand byte

func (helperCommand) Run(packer.Environment, []string) int {
	return 42
}

func (helperCommand) Synopsis() string {
	return "1"
}

func TestCommand_NoExist(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Command(exec.Command("i-should-never-ever-ever-exist"))
	assert.NotNil(err, "should have an error")
}

func TestCommand_Good(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	command, err := Command(helperProcess("command"))
	assert.Nil(err, "should start command properly")

	assert.NotNil(command, "should have a command")
	if command != nil {
		result := command.Synopsis()
		assert.Equal(result, "1", "should return result")
	}
}

func TestCommand_CommandExited(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Command(helperProcess("im-a-command-that-doesnt-work"))
	assert.NotNil(err, "should have an error")
	assert.Equal(err.Error(), "plugin exited before we could connect", "be correct error")
}

func TestCommand_BadRPC(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Command(helperProcess("invalid-rpc-address"))
	assert.NotNil(err, "should have an error")
	assert.Equal(err.Error(), "missing port in address lolinvalid", "be correct error")
}

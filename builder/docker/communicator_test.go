package docker

import (
	"github.com/mitchellh/packer/packer"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func TestCommunicator_impl(t *testing.T) {
	var _ packer.Communicator = new(Communicator)
}

func TestIsValidDockerCommand(t *testing.T) {
	assert := assert.New(t)

	assert.False(IsValidDockerCommand(exec.Command("THIS_COMMAND_WILL_FAIL")), "THIS_COMMAND_WILL_FAIL")
	assert.False(IsValidDockerCommand(exec.Command("docker", "attack")), "docker attack")
	assert.True(IsValidDockerCommand(exec.Command("docker", "attach")), "docker attach")
	assert.True(IsValidDockerCommand(exec.Command("docker", "exec")), "docker exec")
}

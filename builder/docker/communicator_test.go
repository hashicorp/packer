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

func TestIsValidDockerShellCommand(t *testing.T) {
	assert := assert.New(t)

	assert.False(IsValidDockerShellCommand(exec.Command("THIS_COMMAND_WILL_FAIL")), "THIS_COMMAND_WILL_FAIL")
	assert.False(IsValidDockerShellCommand(exec.Command("docker", "attack")), "docker attack")
	assert.True(IsValidDockerShellCommand(exec.Command("docker", "attach")), "docker attach")

	//This one depends on the version of docker.
	//assert.True(IsValidDockerShellCommand(exec.Command("docker", "exec")), "docker exec")
}

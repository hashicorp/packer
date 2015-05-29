package docker

import (
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"testing"
)

func TestCommunicator_impl(t *testing.T) {
	var _ packer.Communicator = new(Communicator)
}

func TestIsValidDockerShellCommand(t *testing.T) {
	if IsValidDockerShellCommand(exec.Command("THIS_COMMAND_WILL_FAIL")) == true {
		t.Error("THIS_COMMAND_WILL_FAIL should be invalid")
	}

	//Can only be tested for integration if docker is available
	//creating a mock feels like simulating failures
	if exec.Command("docker").Run() == nil {
		if IsValidDockerShellCommand(exec.Command("docker", "attack")) == true {
			t.Error("'docker attack' should be invalid")
		}
		if IsValidDockerShellCommand(exec.Command("docker", "attach")) != true {
			t.Error("'docker attach' should be valid")
		}

		//  This one depends on the version of docker.
		//	if IsValidDockerShellCommand(exec.Command("docker", "exec")) != true {
		//		t.Error("'docker exec' should be valid?")
		//	}
	}
}

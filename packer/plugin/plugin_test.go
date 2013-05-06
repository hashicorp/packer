package plugin

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"os"
	"os/exec"
	"testing"
)

type helperCommand byte

func (helperCommand) Run(packer.Environment, []string) int {
	return 42
}

func (helperCommand) Synopsis() string {
	return "1"
}

func helperProcess(s... string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, s...)
	env := []string{
		"GO_WANT_HELPER_PROCESS=1",
		"PACKER_PLUGIN_MIN_PORT=10000",
		"PACKER_PLUGIN_MAX_PORT=25000",
	}

	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = append(env, os.Environ()...)
	return cmd
}

func TestClient(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	command := Command(helperProcess("command"))
	result := command.Synopsis()

	assert.Equal(result, "1", "should return result")
}

// This is not a real test. This is just a helper process kicked off by
// tests.
func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	ServeCommand(new(helperCommand))
}

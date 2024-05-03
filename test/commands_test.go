package test

import (
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

type packerCommand struct {
	once       sync.Once
	packerPath string
	args       []string
	env        map[string]string
	stderr     *strings.Builder
	stdout     *strings.Builder
	err        error
}

// PackerCommand creates a skeleton of packer command with the ability to execute gadgets on the outputs of the command.
func (ts *PackerTestSuite) PackerCommand() *packerCommand {
	stderr := &strings.Builder{}
	stdout := &strings.Builder{}

	return &packerCommand{
		packerPath: ts.packerPath,
		env: map[string]string{
			"PACKER_LOG": "1",
		},
		stderr: stderr,
		stdout: stdout,
	}
}

// UsePluginDir sets the plugin directory in the environment to `dir`
func (pc *packerCommand) UsePluginDir(dir string) *packerCommand {
	return pc.AddEnv("PACKER_PLUGIN_PATH", dir)
}

func (pc *packerCommand) SetArgs(args ...string) *packerCommand {
	pc.args = args
	return pc
}

func (pc *packerCommand) AddEnv(key, val string) *packerCommand {
	pc.env[key] = val
	return pc
}

// Run executes the packer command with the args/env requested and returns the
// output streams (stdout, stderr)
//
// Note: "Run" will only execute the command once, and return the streams and
// error from the only execution for every subsequent call
func (pc *packerCommand) Run(t *testing.T) (string, string, error) {
	pc.once.Do(pc.doRun)

	if strings.Contains(pc.stdout.String(), "PACKER CRASH") || strings.Contains(pc.stderr.String(), "PACKER CRASH") {
		t.Fatalf("Packer has crashed while running the following command: packer %#v", pc.args)
	}

	return pc.stdout.String(), pc.stderr.String(), pc.err
}

func (pc *packerCommand) doRun() {
	cmd := exec.Command("packer", pc.args...)
	for key, val := range pc.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Stdout = pc.stdout
	cmd.Stderr = pc.stderr

	pc.err = cmd.Run()
}

func (pc *packerCommand) Assert(t *testing.T, checks ...Checker) {
	stdout, stderr, err := pc.Run(t)

	for _, check := range checks {
		checkErr := check.Check(stdout, stderr, err)
		if checkErr != nil {
			checkerName := InferName(check)
			t.Errorf("check %q failed: %s", checkerName, checkErr)
		}
	}
}

package packer_test

import (
	"fmt"
	"os"
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
	workdir    string
	err        error
	t          *testing.T
}

// PackerCommand creates a skeleton of packer command with the ability to execute gadgets on the outputs of the command.
func (ts *PackerTestSuite) PackerCommand() *packerCommand {
	stderr := &strings.Builder{}
	stdout := &strings.Builder{}

	return &packerCommand{
		packerPath: ts.packerPath,
		env: map[string]string{
			"PACKER_LOG": "1",
			// Required for Windows, otherwise since we overwrite all
			// the envvars for the test and Go relies on that envvar
			// being set in order to return another path than
			// C:\Windows by default
			//
			// If we don't have it, Packer immediately errors upon
			// invocation as the temporary logfile that we write in
			// case of Panic will fail to be created (unless tests
			// are running as Administrator, but please don't).
			"TMP": os.TempDir(),
		},
		stderr: stderr,
		stdout: stdout,
		t:      ts.T(),
	}
}

// NoVerbose removes the `PACKER_LOG=1` environment variable from the command
func (pc *packerCommand) NoVerbose() *packerCommand {
	_, ok := pc.env["PACKER_LOG"]
	if ok {
		delete(pc.env, "PACKER_LOG")
	}
	return pc
}

// SetWD changes the directory Packer is invoked from
func (pc *packerCommand) SetWD(dir string) *packerCommand {
	pc.workdir = dir
	return pc
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
func (pc *packerCommand) Run() (string, string, error) {
	pc.once.Do(pc.doRun)

	return pc.stdout.String(), pc.stderr.String(), pc.err
}

func (pc *packerCommand) doRun() {
	cmd := exec.Command(pc.packerPath, pc.args...)
	for key, val := range pc.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Stdout = pc.stdout
	cmd.Stderr = pc.stderr

	if pc.workdir != "" {
		cmd.Dir = pc.workdir
	}

	pc.err = cmd.Run()
}

func (pc *packerCommand) Assert(checks ...Checker) {
	stdout, stderr, err := pc.Run()

	checks = append(checks, PanicCheck{})

	for _, check := range checks {
		checkErr := check.Check(stdout, stderr, err)
		if checkErr != nil {
			checkerName := InferName(check)
			pc.t.Errorf("check %q failed: %s", checkerName, checkErr)
		}
	}
}

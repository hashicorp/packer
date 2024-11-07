package common

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer_test/common/check"
)

type packerCommand struct {
	runs         int
	packerPath   string
	args         []string
	env          map[string]string
	stdin        string
	stderr       *strings.Builder
	stdout       *strings.Builder
	workdir      string
	err          error
	t            *testing.T
	fatalfAssert bool
}

// PackerCommand creates a skeleton of packer command with the ability to execute gadgets on the outputs of the command.
func (ts *PackerTestSuite) PackerCommand() *packerCommand {
	return &packerCommand{
		packerPath: ts.packerPath,
		runs:       1,
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
			// Since those commands are used to run tests, we want to
			// make them as self-contained and quick as possible.
			// Removing telemetry here is probably for the best.
			"CHECKPOINT_DISABLE": "1",
		},
		t: ts.T(),
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
func (pc *packerCommand) UsePluginDir(dir *PluginDirSpec) *packerCommand {
	return pc.UseRawPluginDir(dir.dirPath)
}

// UseRawPluginDir is meant to be used for setting the plugin directory with a
// raw directory path instead of a PluginDirSpec.
func (pc *packerCommand) UseRawPluginDir(dirPath string) *packerCommand {
	return pc.AddEnv("PACKER_PLUGIN_PATH", dirPath)
}

func (pc *packerCommand) SetArgs(args ...string) *packerCommand {
	pc.args = args
	return pc
}

func (pc *packerCommand) AddEnv(key, val string) *packerCommand {
	pc.env[key] = val
	return pc
}

// Runs changes the number of times the command is run.
//
// This is useful for testing non-deterministic bugs, which we can reasonably
// execute multiple times and expose a dysfunctional run.
//
// This is not necessarily a guarantee that the code is sound, but so long as
// we run the test enough times, we can be decently confident the problem has
// been solved.
func (pc *packerCommand) Runs(runs int) *packerCommand {
	if runs <= 0 {
		panic(fmt.Sprintf("cannot set command runs to %d", runs))
	}

	pc.runs = runs
	return pc
}

// Stdin changes the contents of the stdin for the command.
//
// Each run will be populated with a copy of this string, and wait for the
// command to terminate.
//
// Note: this could lead to a deadlock if the command doesn't support stdin
// closing after it's finished feeding the inputs.
func (pc *packerCommand) Stdin(in string) *packerCommand {
	pc.stdin = in
	return pc
}

// SetAssertFatal allows changing how Assert behaves when reporting an error.
//
// By default Assert will invoke t.Errorf with the error details, but this can be
// changed to a t.Fatalf so that if the assertion fails, the test invoking it will
// also immediately fail and stop execution.
func (pc *packerCommand) SetAssertFatal() *packerCommand {
	pc.fatalfAssert = true
	return pc
}

// Run executes the packer command with the args/env requested and returns the
// output streams (stdout, stderr)
//
// Note: while originally "Run" was designed to be idempotent, with the
// introduction of multiple runs for a command, this is not the case anymore
// and the function should not be considered thread-safe anymore.
func (pc *packerCommand) run() (string, string, error) {
	if pc.runs <= 0 {
		return pc.stdout.String(), pc.stderr.String(), pc.err
	}
	pc.runs--

	pc.stdout = &strings.Builder{}
	pc.stderr = &strings.Builder{}

	cmd := exec.Command(pc.packerPath, pc.args...)
	for key, val := range pc.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Stdout = pc.stdout
	cmd.Stderr = pc.stderr

	if pc.stdin != "" {
		cmd.Stdin = strings.NewReader(pc.stdin)
	}

	if pc.workdir != "" {
		cmd.Dir = pc.workdir
	}

	pc.err = cmd.Run()

	// Check that the command didn't panic, and if it did, we can immediately error
	panicErr := check.PanicCheck{}.Check(pc.stdout.String(), pc.stderr.String(), pc.err)
	if panicErr != nil {
		pc.t.Fatalf("Packer panicked during execution: %s", panicErr)
	}

	return pc.stdout.String(), pc.stderr.String(), pc.err
}

// Output returns the results of the latest Run that was executed.
//
// In general there is only one run of the command, but as it can be changed
// through the Runs function, only the latest run will be returned.
//
// If the command was not run (through Assert), this will make the test fail
// immediately.
func (pc *packerCommand) Output() (string, string, error) {
	if pc.runs > 0 {
		pc.t.Fatalf("command was not run, invoke Assert first, then Output.")
	}

	return pc.stdout.String(), pc.stderr.String(), pc.err
}

func (pc *packerCommand) Assert(checks ...check.Checker) {
	attempt := 0
	for pc.runs > 0 {
		attempt++
		stdout, stderr, err := pc.run()

		for _, checker := range checks {
			checkErr := checker.Check(stdout, stderr, err)
			if checkErr != nil {
				checkerName := check.InferName(checker)

				pc.t.Errorf("check %q failed: %s", checkerName, checkErr)
			}
		}

		if pc.t.Failed() {
			pc.t.Errorf("attempt %d failed validation", attempt)

			pc.t.Logf("dumping stdout: %s", stdout)
			pc.t.Logf("dumping stdout: %s", stderr)

			if pc.fatalfAssert {
				pc.t.Fatalf("stopping test now because of failures reported")
			}

			break
		}
	}
}

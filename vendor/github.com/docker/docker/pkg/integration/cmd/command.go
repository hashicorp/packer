package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-check/check"
)

type testingT interface {
	Fatalf(string, ...interface{})
}

const (
	// None is a token to inform Result.Assert that the output should be empty
	None string = "<NOTHING>"
)

// GetExitCode returns the ExitStatus of the specified error if its type is
// exec.ExitError, returns 0 and an error otherwise.
func GetExitCode(err error) (int, error) {
	exitCode := 0
	if exiterr, ok := err.(*exec.ExitError); ok {
		if procExit, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return procExit.ExitStatus(), nil
		}
	}
	return exitCode, fmt.Errorf("failed to get exit code")
}

// ProcessExitCode process the specified error and returns the exit status code
// if the error was of type exec.ExitError, returns nothing otherwise.
func ProcessExitCode(err error) (exitCode int) {
	if err != nil {
		var exiterr error
		if exitCode, exiterr = GetExitCode(err); exiterr != nil {
			// TODO: Fix this so we check the error's text.
			// we've failed to retrieve exit code, so we set it to 127
			exitCode = 127
		}
	}
	return
}

type lockedBuffer struct {
	m   sync.RWMutex
	buf bytes.Buffer
}

func (buf *lockedBuffer) Write(b []byte) (int, error) {
	buf.m.Lock()
	defer buf.m.Unlock()
	return buf.buf.Write(b)
}

func (buf *lockedBuffer) String() string {
	buf.m.RLock()
	defer buf.m.RUnlock()
	return buf.buf.String()
}

// Result stores the result of running a command
type Result struct {
	Cmd      *exec.Cmd
	ExitCode int
	Error    error
	// Timeout is true if the command was killed because it ran for too long
	Timeout   bool
	outBuffer *lockedBuffer
	errBuffer *lockedBuffer
}

// Assert compares the Result against the Expected struct, and fails the test if
// any of the expcetations are not met.
func (r *Result) Assert(t testingT, exp Expected) {
	err := r.Compare(exp)
	if err == nil {
		return
	}

	_, file, line, _ := runtime.Caller(1)
	t.Fatalf("at %s:%d\n%s", filepath.Base(file), line, err.Error())
}

// Compare returns an formatted error with the command, stdout, stderr, exit
// code, and any failed expectations
func (r *Result) Compare(exp Expected) error {
	errors := []string{}
	add := func(format string, args ...interface{}) {
		errors = append(errors, fmt.Sprintf(format, args...))
	}

	if exp.ExitCode != r.ExitCode {
		add("ExitCode was %d expected %d", r.ExitCode, exp.ExitCode)
	}
	if exp.Timeout != r.Timeout {
		if exp.Timeout {
			add("Expected command to timeout")
		} else {
			add("Expected command to finish, but it hit the timeout")
		}
	}
	if !matchOutput(exp.Out, r.Stdout()) {
		add("Expected stdout to contain %q", exp.Out)
	}
	if !matchOutput(exp.Err, r.Stderr()) {
		add("Expected stderr to contain %q", exp.Err)
	}
	switch {
	// If a non-zero exit code is expected there is going to be an error.
	// Don't require an error message as well as an exit code because the
	// error message is going to be "exit status <code> which is not useful
	case exp.Error == "" && exp.ExitCode != 0:
	case exp.Error == "" && r.Error != nil:
		add("Expected no error")
	case exp.Error != "" && r.Error == nil:
		add("Expected error to contain %q, but there was no error", exp.Error)
	case exp.Error != "" && !strings.Contains(r.Error.Error(), exp.Error):
		add("Expected error to contain %q", exp.Error)
	}

	if len(errors) == 0 {
		return nil
	}
	return fmt.Errorf("%s\nFailures:\n%s\n", r, strings.Join(errors, "\n"))
}

func matchOutput(expected string, actual string) bool {
	switch expected {
	case None:
		return actual == ""
	default:
		return strings.Contains(actual, expected)
	}
}

func (r *Result) String() string {
	var timeout string
	if r.Timeout {
		timeout = " (timeout)"
	}

	return fmt.Sprintf(`
Command: %s
ExitCode: %d%s, Error: %s
Stdout: %v
Stderr: %v
`,
		strings.Join(r.Cmd.Args, " "),
		r.ExitCode,
		timeout,
		r.Error,
		r.Stdout(),
		r.Stderr())
}

// Expected is the expected output from a Command. This struct is compared to a
// Result struct by Result.Assert().
type Expected struct {
	ExitCode int
	Timeout  bool
	Error    string
	Out      string
	Err      string
}

// Success is the default expected result
var Success = Expected{}

// Stdout returns the stdout of the process as a string
func (r *Result) Stdout() string {
	return r.outBuffer.String()
}

// Stderr returns the stderr of the process as a string
func (r *Result) Stderr() string {
	return r.errBuffer.String()
}

// Combined returns the stdout and stderr combined into a single string
func (r *Result) Combined() string {
	return r.outBuffer.String() + r.errBuffer.String()
}

// SetExitError sets Error and ExitCode based on Error
func (r *Result) SetExitError(err error) {
	if err == nil {
		return
	}
	r.Error = err
	r.ExitCode = ProcessExitCode(err)
}

type matches struct{}

// Info returns the CheckerInfo
func (m *matches) Info() *check.CheckerInfo {
	return &check.CheckerInfo{
		Name:   "CommandMatches",
		Params: []string{"result", "expected"},
	}
}

// Check compares a result against the expected
func (m *matches) Check(params []interface{}, names []string) (bool, string) {
	result, ok := params[0].(*Result)
	if !ok {
		return false, fmt.Sprintf("result must be a *Result, not %T", params[0])
	}
	expected, ok := params[1].(Expected)
	if !ok {
		return false, fmt.Sprintf("expected must be an Expected, not %T", params[1])
	}

	err := result.Compare(expected)
	if err == nil {
		return true, ""
	}
	return false, err.Error()
}

// Matches is a gocheck.Checker for comparing a Result against an Expected
var Matches = &matches{}

// Cmd contains the arguments and options for a process to run as part of a test
// suite.
type Cmd struct {
	Command []string
	Timeout time.Duration
	Stdin   io.Reader
	Stdout  io.Writer
	Dir     string
	Env     []string
}

// RunCmd runs a command and returns a Result
func RunCmd(cmd Cmd) *Result {
	result := StartCmd(cmd)
	if result.Error != nil {
		return result
	}
	return WaitOnCmd(cmd.Timeout, result)
}

// RunCommand parses a command line and runs it, returning a result
func RunCommand(command string, args ...string) *Result {
	return RunCmd(Cmd{Command: append([]string{command}, args...)})
}

// StartCmd starts a command, but doesn't wait for it to finish
func StartCmd(cmd Cmd) *Result {
	result := buildCmd(cmd)
	if result.Error != nil {
		return result
	}
	result.SetExitError(result.Cmd.Start())
	return result
}

func buildCmd(cmd Cmd) *Result {
	var execCmd *exec.Cmd
	switch len(cmd.Command) {
	case 1:
		execCmd = exec.Command(cmd.Command[0])
	default:
		execCmd = exec.Command(cmd.Command[0], cmd.Command[1:]...)
	}
	outBuffer := new(lockedBuffer)
	errBuffer := new(lockedBuffer)

	execCmd.Stdin = cmd.Stdin
	execCmd.Dir = cmd.Dir
	execCmd.Env = cmd.Env
	if cmd.Stdout != nil {
		execCmd.Stdout = io.MultiWriter(outBuffer, cmd.Stdout)
	} else {
		execCmd.Stdout = outBuffer
	}
	execCmd.Stderr = errBuffer
	return &Result{
		Cmd:       execCmd,
		outBuffer: outBuffer,
		errBuffer: errBuffer,
	}
}

// WaitOnCmd waits for a command to complete. If timeout is non-nil then
// only wait until the timeout.
func WaitOnCmd(timeout time.Duration, result *Result) *Result {
	if timeout == time.Duration(0) {
		result.SetExitError(result.Cmd.Wait())
		return result
	}

	done := make(chan error, 1)
	// Wait for command to exit in a goroutine
	go func() {
		done <- result.Cmd.Wait()
	}()

	select {
	case <-time.After(timeout):
		killErr := result.Cmd.Process.Kill()
		if killErr != nil {
			fmt.Printf("failed to kill (pid=%d): %v\n", result.Cmd.Process.Pid, killErr)
		}
		result.Timeout = true
	case err := <-done:
		result.SetExitError(err)
	}
	return result
}

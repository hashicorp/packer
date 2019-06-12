// Shamelessly copied from the Terraform repo because it wasn't worth vendoring
// out two hundred lines of code so Packer could use it too.
//
// wrappedreadline is a package that has helpers for interacting with
// readline from a panicwrap executable.
//
// panicwrap overrides the standard file descriptors so that the child process
// no longer looks like a TTY. The helpers here access the extra file descriptors
// passed by panicwrap to fix that.
//
// panicwrap should be checked for with panicwrap.Wrapped before using this
// librar, since this library won't adapt if the binary is not wrapped.
package wrappedreadline

import (
	"os"
	"runtime"

	"github.com/chzyer/readline"
	"github.com/mitchellh/panicwrap"
)

// Override overrides the values in readline.Config that need to be
// set with wrapped values.
func Override(cfg *readline.Config) *readline.Config {
	cfg.Stdin = Stdin()
	cfg.Stdout = Stdout()
	cfg.Stderr = Stderr()

	cfg.FuncGetWidth = TerminalWidth
	cfg.FuncIsTerminal = IsTerminal

	rm := RawMode{StdinFd: int(Stdin().Fd())}
	cfg.FuncMakeRaw = rm.Enter
	cfg.FuncExitRaw = rm.Exit

	return cfg
}

// IsTerminal determines if this process is attached to a TTY.
func IsTerminal() bool {
	// Windows is always a terminal
	if runtime.GOOS == "windows" {
		return true
	}

	// Same implementation as readline but with our custom fds
	return readline.IsTerminal(int(Stdin().Fd())) &&
		(readline.IsTerminal(int(Stdout().Fd())) ||
			readline.IsTerminal(int(Stderr().Fd())))
}

// TerminalWidth gets the terminal width in characters.
func TerminalWidth() int {
	if runtime.GOOS == "windows" {
		return readline.GetScreenWidth()
	}

	return getWidth()
}

// RawMode is a helper for entering and exiting raw mode.
type RawMode struct {
	StdinFd int

	state *readline.State
}

func (r *RawMode) Enter() (err error) {
	r.state, err = readline.MakeRaw(r.StdinFd)
	return err
}

func (r *RawMode) Exit() error {
	if r.state == nil {
		return nil
	}

	return readline.Restore(r.StdinFd, r.state)
}

// Package provides access to the standard OS streams
// (stdin, stdout, stderr) even if wrapped under panicwrap.
// Stdin returns the true stdin of the process.
func Stdin() *os.File {
	stdin := os.Stdin
	if panicwrap.Wrapped(nil) {
		stdin = wrappedStdin
	}

	return stdin
}

// Stdout returns the true stdout of the process.
func Stdout() *os.File {
	stdout := os.Stdout
	if panicwrap.Wrapped(nil) {
		stdout = wrappedStdout
	}

	return stdout
}

// Stderr returns the true stderr of the process.
func Stderr() *os.File {
	stderr := os.Stderr
	if panicwrap.Wrapped(nil) {
		stderr = wrappedStderr
	}

	return stderr
}

// These are the wrapped standard streams. These are setup by the
// platform specific code in initPlatform.
var (
	wrappedStdin  *os.File
	wrappedStdout *os.File
	wrappedStderr *os.File
)

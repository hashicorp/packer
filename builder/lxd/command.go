package lxd

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// CommandWrapper is a type that given a command, will possibly modify that
// command in-flight. This might return an error.
type CommandWrapper func(string) (string, error)

// ShellCommand takes a command string and returns an *exec.Cmd to execute
// it within the context of a shell (/bin/sh).
func ShellCommand(command string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", command)
}

// Yeah...LXD calls `lxc` because the command line is different between the
// packages. This should also avoid a naming collision between the LXC builder.
func LXDCommand(args ...string) (string, error) {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing lxc command: %#v", args)
	cmd := exec.Command("lxc", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("LXD command error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return stdoutString, err
}

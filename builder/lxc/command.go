package lxc

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

func RunCommand(args ...string) error {
	var stdout, stderr bytes.Buffer

	log.Printf("Executing args: %#v", args)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	stdoutString := strings.TrimSpace(stdout.String())
	stderrString := strings.TrimSpace(stderr.String())

	if _, ok := err.(*exec.ExitError); ok {
		err = fmt.Errorf("Command error: %s", stderrString)
	}

	log.Printf("stdout: %s", stdoutString)
	log.Printf("stderr: %s", stderrString)

	return err
}

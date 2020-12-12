// Package common provides the PackerConfig structure that gets passed to every
// plugin and contains information populated by the Packer core. This config
// contains data about command line flags that were used, as well as template
// information and information about the Packer core's version. It also
// proivdes string constants to use to access that config.
package common

import (
	"os/exec"
)

// CommandWrapper is a type that given a command, will modify that
// command in-flight. This might return an error.
// For example, your command could be `foo` and your CommandWrapper could be
// func(s string) (string, error) {
//	 return fmt.Sprintf("/bin/sh/ %s", s)
// }
// Using the CommandWrapper, you can set environment variables or perform
// string interpolation once rather than many times, to save some lines of code
// if similar wrapping needs to be performed many times during a plugin run.
type CommandWrapper func(string) (string, error)

// ShellCommand takes a command string and returns an *exec.Cmd to execute
// it within the context of a shell (/bin/sh).
func ShellCommand(command string) *exec.Cmd {
	return exec.Command("/bin/sh", "-c", command)
}

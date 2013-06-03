package packer

import (
	"io"
)

// RemoteCmd represents a remote command being prepared or run.
type RemoteCmd struct {
	// Command is the command to run remotely. This is executed as if
	// it were a shell command, so you are expected to do any shell escaping
	// necessary.
	Command string

	// Stdin specifies the process's standard input. If Stdin is
	// nil, the process reads from an empty bytes.Buffer.
	Stdin io.Reader

	// Stdout and Stderr represent the process's standard output and
	// error.
	//
	// If either is nil, it will be set to ioutil.Discard.
	Stdout io.Writer
	Stderr io.Writer

	// This will be set to true when the remote command has exited. It
	// shouldn't be set manually by the user, but there is no harm in
	// doing so.
	Exited bool

	// Once Exited is true, this will contain the exit code of the process.
	ExitStatus int
}

// A Communicator is the interface used to communicate with the machine
// that exists that will eventually be packaged into an image. Communicators
// allow you to execute remote commands, upload files, etc.
//
// Communicators must be safe for concurrency, meaning multiple calls to
// Start or any other method may be called at the same time.
type Communicator interface {
	Start(*RemoteCmd) error
	Upload(string, io.Reader) error
	Download(string, io.Writer) error
}

package packer

import (
	"io"
	"time"
)

// A Communicator is the interface used to communicate with the machine
// that exists that will eventually be packaged into an image. Communicators
// allow you to execute remote commands, upload files, etc.
//
// Communicators must be safe for concurrency, meaning multiple calls to
// Start or any other method may be called at the same time.
type Communicator interface {
	Start(string) (*RemoteCommand, error)
	Upload(string, io.Reader)
	Download(string, io.Writer)
}

// This struct contains some information about the remote command being
// executed and can be used to wait for it to complete.
//
// Stdin, Stdout, Stderr are readers and writers to varios IO streams for
// the remote command.
//
// Exited is false until Wait is called. It can be used to check if Wait
// has already been called.
//
// ExitStatus is the exit code of the remote process. It is only available
// once Wait is called.
type RemoteCommand struct {
	Stdin  io.Writer
	Stdout io.Reader
	Stderr io.Reader
	Exited bool
	ExitStatus int
}

// Wait waits for the command to exit.
func (r *RemoteCommand) Wait() {
	// Busy wait on being exited. We put a sleep to be kind to the
	// Go scheduler, and because we don't really need smaller granularity.
	for !r.Exited {
		time.Sleep(10 * time.Millisecond)
	}
}

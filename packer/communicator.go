package packer

import "io"

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
type RemoteCommand struct {
	Stdin  io.Writer
	Stdout io.Reader
	Stderr io.Reader
	ExitStatus int
}

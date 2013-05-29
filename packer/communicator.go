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
	Upload(string, io.Reader) error
	Download(string, io.Writer) error
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
	Stdin      io.Writer
	Stdout     io.Reader
	Stderr     io.Reader
	Exited     bool
	ExitStatus int
}

// StdoutStream returns a channel that will be sent all the output
// of stdout as it comes. The output isn't guaranteed to be a full line.
// When the channel is closed, the process is exited.
func (r *RemoteCommand) StdoutChan() (<-chan string) {
	return nil
}

// ExitChan returns a channel that will be sent the exit status once
// the process exits. This can be used in cases such a select statement
// waiting on the process to end.
func (r *RemoteCommand) ExitChan() (<-chan int) {
	// TODO(mitchellh): lock
	// TODO(mitchellh): Something more efficient than multiple Wait() calls

	// Make a single buffered channel so that the send doesn't block.
	exitChan := make(chan int, 1)

	go func() {
		defer close(exitChan)
		r.Wait()
		exitChan <- r.ExitStatus
	}()

	return exitChan
}

// Wait waits for the command to exit.
func (r *RemoteCommand) Wait() {
	// Busy wait on being exited. We put a sleep to be kind to the
	// Go scheduler, and because we don't really need smaller granularity.
	for !r.Exited {
		time.Sleep(10 * time.Millisecond)
	}
}

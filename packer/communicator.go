package packer

import (
	"io"
	"time"
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
	// Start takes a RemoteCmd and starts it. The RemoteCmd must not be
	// modified after being used with Start, and it must not be used with
	// Start again. The Start method returns immediately once the command
	// is started. It does not wait for the command to complete. The
	// RemoteCmd.Exited field should be used for this.
	Start(*RemoteCmd) error

	// Upload uploads a file to the machine to the given path with the
	// contents coming from the given reader. This method will block until
	// it completes.
	Upload(string, io.Reader) error

	// Download downloads a file from the machine from the given remote path
	// with the contents writing to the given writer. This method will
	// block until it completes.
	Download(string, io.Writer) error
}

// Wait waits for the remote command to complete.
func (r *RemoteCmd) Wait() {
	for !r.Exited {
		time.Sleep(50 * time.Millisecond)
	}
}

package packer

import (
	"io"
	"os"
	"strings"
	"sync"

	"github.com/mitchellh/iochan"
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

	// Internal fields
	exitCh chan struct{}

	// This thing is a mutex, lock when making modifications concurrently
	sync.Mutex
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
	Upload(string, io.Reader, *os.FileInfo) error

	// UploadDir uploads the contents of a directory recursively to
	// the remote path. It also takes an optional slice of paths to
	// ignore when uploading.
	//
	// The folder name of the source folder should be created unless there
	// is a trailing slash on the source "/". For example: "/tmp/src" as
	// the source will create a "src" directory in the destination unless
	// a trailing slash is added. This is identical behavior to rsync(1).
	UploadDir(dst string, src string, exclude []string) error

	// Download downloads a file from the machine from the given remote path
	// with the contents writing to the given writer. This method will
	// block until it completes.
	Download(string, io.Writer) error

	DownloadDir(src string, dst string, exclude []string) error
}

// StartWithUi runs the remote command and streams the output to any
// configured Writers for stdout/stderr, while also writing each line
// as it comes to a Ui.
func (r *RemoteCmd) StartWithUi(c Communicator, ui Ui) error {
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()
	defer stdout_w.Close()
	defer stderr_w.Close()

	// Retain the original stdout/stderr that we can replace back in.
	originalStdout := r.Stdout
	originalStderr := r.Stderr
	defer func() {
		r.Lock()
		defer r.Unlock()

		r.Stdout = originalStdout
		r.Stderr = originalStderr
	}()

	// Set the writers for the output so that we get it streamed to us
	if r.Stdout == nil {
		r.Stdout = stdout_w
	} else {
		r.Stdout = io.MultiWriter(r.Stdout, stdout_w)
	}

	if r.Stderr == nil {
		r.Stderr = stderr_w
	} else {
		r.Stderr = io.MultiWriter(r.Stderr, stderr_w)
	}

	// Start the command
	if err := c.Start(r); err != nil {
		return err
	}

	// Create the channels we'll use for data
	exitCh := make(chan struct{})
	stdoutCh := iochan.DelimReader(stdout_r, '\n')
	stderrCh := iochan.DelimReader(stderr_r, '\n')

	// Start the goroutine to watch for the exit
	go func() {
		defer close(exitCh)
		defer stdout_w.Close()
		defer stderr_w.Close()
		r.Wait()
	}()

	// Loop and get all our output
OutputLoop:
	for {
		select {
		case output := <-stderrCh:
			if output != "" {
				ui.Message(r.cleanOutputLine(output))
			}
		case output := <-stdoutCh:
			if output != "" {
				ui.Message(r.cleanOutputLine(output))
			}
		case <-exitCh:
			break OutputLoop
		}
	}

	// Make sure we finish off stdout/stderr because we may have gotten
	// a message from the exit channel before finishing these first.
	for output := range stdoutCh {
		ui.Message(strings.TrimSpace(output))
	}

	for output := range stderrCh {
		ui.Message(strings.TrimSpace(output))
	}

	return nil
}

// SetExited is a helper for setting that this process is exited. This
// should be called by communicators who are running a remote command in
// order to set that the command is done.
func (r *RemoteCmd) SetExited(status int) {
	r.Lock()
	defer r.Unlock()

	if r.exitCh == nil {
		r.exitCh = make(chan struct{})
	}

	r.Exited = true
	r.ExitStatus = status
	close(r.exitCh)
}

// Wait waits for the remote command to complete.
func (r *RemoteCmd) Wait() {
	// Make sure our condition variable is initialized.
	r.Lock()
	if r.exitCh == nil {
		r.exitCh = make(chan struct{})
	}
	r.Unlock()

	<-r.exitCh
}

// cleanOutputLine cleans up a line so that '\r' don't muck up the
// UI output when we're reading from a remote command.
func (r *RemoteCmd) cleanOutputLine(line string) string {
	// Trim surrounding whitespace
	line = strings.TrimSpace(line)

	// Trim up to the first carriage return, since that text would be
	// lost anyways.
	idx := strings.LastIndex(line, "\r")
	if idx > -1 {
		line = line[idx+1:]
	}

	return line
}

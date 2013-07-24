package packer

import (
	"fmt"
	"github.com/mitchellh/iochan"
	"io"
	"log"
	"strings"
)

func StartWithUi(comm Communicator, ui Ui, r *RemoteCmd) error {
	if r.Stdout != nil || r.Stderr != nil {
		log.Printf("not logging remote command: %s", r.Command)
		return comm.Start(r)
	}

	// Setup the remote command
	stdout_r, stdout_w := io.Pipe()
	stderr_r, stderr_w := io.Pipe()

	r.Stdout = stdout_w
	r.Stderr = stderr_w

	if err := comm.Start(r); err != nil {
		return err
	}

	exitChan := make(chan int, 1)
	stdoutChan := iochan.DelimReader(stdout_r, '\n')
	stderrChan := iochan.DelimReader(stderr_r, '\n')

	go func() {
		defer stdout_w.Close()
		defer stderr_w.Close()

		r.Wait()
		exitChan <- r.ExitStatus
	}()

OutputLoop:
	for {
		select {
		case output := <-stderrChan:
			ui.Message(strings.TrimSpace(output))
		case output := <-stdoutChan:
			ui.Message(strings.TrimSpace(output))
		case exitStatus := <-exitChan:
			log.Printf("command exited with status %d", exitStatus)

			if exitStatus != 0 {
				err := fmt.Errorf("command exited with non-zero exit status: %d", exitStatus)
				return err
			}

			break OutputLoop
		}
	}

	// Make sure we finish off stdout/stderr because we may have gotten
	// a message from the exit channel first.
	for output := range stdoutChan {
		ui.Message(output)
	}

	for output := range stderrChan {
		ui.Message(output)
	}

	return nil
}

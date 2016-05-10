package shell_local

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/mitchellh/packer/packer"
)

type Communicator struct{}

func (c *Communicator) Start(cmd *packer.RemoteCmd) error {
	localCmd := exec.Command("sh", "-c", cmd.Command)
	localCmd.Stdin = cmd.Stdin
	localCmd.Stdout = cmd.Stdout
	localCmd.Stderr = cmd.Stderr

	// Start it. If it doesn't work, then error right away.
	if err := localCmd.Start(); err != nil {
		return err
	}

	// We've started successfully. Start a goroutine to wait for
	// it to complete and track exit status.
	go func() {
		var exitStatus int
		err := localCmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitStatus = 1

				// There is no process-independent way to get the REAL
				// exit status so we just try to go deeper.
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitStatus = status.ExitStatus()
				}
			}
		}

		cmd.SetExited(exitStatus)
	}()

	return nil
}

func (c *Communicator) Upload(string, io.Reader, *os.FileInfo) error {
	return fmt.Errorf("upload not supported")
}

func (c *Communicator) UploadDir(string, string, []string) error {
	return fmt.Errorf("uploadDir not supported")
}

func (c *Communicator) Download(string, io.Writer) error {
	return fmt.Errorf("download not supported")
}

func (c *Communicator) DownloadDir(src string, dst string, exclude []string) error {
	return fmt.Errorf("downloadDir not supported")
}

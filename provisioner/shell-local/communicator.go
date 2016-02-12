package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"

	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Communicator struct {
	ExecuteCommand []string
	Ctx            interpolate.Context
}

func (c *Communicator) Start(cmd *packer.RemoteCmd) error {
	// Render the template so that we know how to execute the command
	c.Ctx.Data = &ExecuteCommandTemplate{
		Command: cmd.Command,
	}
	for i, field := range c.ExecuteCommand {
		command, err := interpolate.Render(field, &c.Ctx)
		if err != nil {
			return fmt.Errorf("Error processing command: %s", err)
		}

		c.ExecuteCommand[i] = command
	}

	// Build the local command to execute
	localCmd := exec.Command(c.ExecuteCommand[0], c.ExecuteCommand[1:]...)
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

func (c *Communicator) DownloadDir(string, string, []string) error {
	return fmt.Errorf("downloadDir not supported")
}

type ExecuteCommandTemplate struct {
	Command string
}

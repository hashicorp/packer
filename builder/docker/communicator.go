package docker

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"syscall"
)

type Communicator struct {
	ContainerId  string
	HostDir      string
	ContainerDir string
}

func (c *Communicator) Start(remote *packer.RemoteCmd) error {
	cmd := exec.Command("docker", "attach", c.ContainerId)
	stdin_w, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	cmd.Stdout = remote.Stdout
	cmd.Stderr = remote.Stderr

	log.Printf("Executing in container %s: %#v", c.ContainerId, remote.Command)
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		defer stdin_w.Close()
		stdin_w.Write([]byte(remote.Command + "\n"))
	}()

	var exitStatus int = 0
	err = cmd.Wait()
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitStatus = 1

		// There is no process-independent way to get the REAL
		// exit status so we just try to go deeper.
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			exitStatus = status.ExitStatus()
		}
	}

	if exitStatus != 0 {
		return fmt.Errorf("Exit status: %d", exitStatus)
	}

	return nil
}

func (c *Communicator) Upload(dst string, src io.Reader) error {
	// Create a temporary file to store the upload
	tempfile, err := ioutil.TempFile(c.HostDir, "upload")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())

	// Copy the contents to the temporary file
	_, err = io.Copy(tempfile, src)
	tempfile.Close()
	if err != nil {
		return err
	}

	// TODO(mitchellh): Copy the file into place
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("cp %s %s", tempfile.Name(), dst),
	}

	if err := c.Start(cmd); err != nil {
		return err
	}

	// Wait for the copy to complete
	cmd.Wait()
	if cmd.ExitStatus != 0 {
		return fmt.Errorf("Upload failed with non-zero exit status: %d", cmd.ExitStatus)
	}

	return nil
}

func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	return nil
}

func (c *Communicator) Download(src string, dst io.Writer) error {
	return nil
}

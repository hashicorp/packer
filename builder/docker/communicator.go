package docker

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type Communicator struct {
	ContainerId  string
	HostDir      string
	ContainerDir string
}

func (c *Communicator) Start(remote *packer.RemoteCmd) error {
	// Create a temporary file to store the output. Because of a bug in
	// Docker, sometimes all the output doesn't properly show up. This
	// file will capture ALL of the output, and we'll read that.
	//
	// https://github.com/dotcloud/docker/issues/2625
	outputFile, err := ioutil.TempFile(c.HostDir, "cmd")
	if err != nil {
		return err
	}
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	// This file will store the exit code of the command once it is complete.
	exitCodePath := outputFile.Name() + "-exit"

	// Modify the remote command so that all the output of the commands
	// go to a single file and so that the exit code is redirected to
	// a single file. This lets us determine both when the command
	// is truly complete (because the file will have data), what the
	// exit status is (because Docker loses it because of the pty, not
	// Docker's fault), and get the output (Docker bug).
	remoteCmd := fmt.Sprintf("(%s) >%s 2>&1; echo $? >%s",
		remote.Command,
		filepath.Join(c.ContainerDir, filepath.Base(outputFile.Name())),
		filepath.Join(c.ContainerDir, filepath.Base(exitCodePath)))

	cmd := exec.Command("docker", "attach", c.ContainerId)
	stdin_w, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	log.Printf("Executing in container %s: %#v", c.ContainerId, remoteCmd)
	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		defer stdin_w.Close()

		// This sleep needs to be here because of the issue linked to below.
		// Basically, without it, Docker will hang on reading stdin forever,
		// and won't see what we write, for some reason.
		//
		// https://github.com/dotcloud/docker/issues/2628
		time.Sleep(2 * time.Second)

		stdin_w.Write([]byte(remoteCmd + "\n"))
	}()

	err = cmd.Wait()
	if exitErr, ok := err.(*exec.ExitError); ok {
		exitStatus := 1

		// There is no process-independent way to get the REAL
		// exit status so we just try to go deeper.
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			exitStatus = status.ExitStatus()
		}

		// Say that we ended, since if Docker itself failed, then
		// the command must've not run, or so we assume
		remote.SetExited(exitStatus)
		return nil
	}

	// Wait for the exit code to appear in our file...
	log.Println("Waiting for exit code to appear for remote command...")
	for {
		fi, err := os.Stat(exitCodePath)
		if err == nil && fi.Size() > 0 {
			break
		}

		time.Sleep(1 * time.Second)
	}

	// Read the exit code
	exitRaw, err := ioutil.ReadFile(exitCodePath)
	if err != nil {
		return err
	}

	exitStatus, err := strconv.ParseInt(string(bytes.TrimSpace(exitRaw)), 10, 0)
	if err != nil {
		return err
	}
	log.Printf("Executed command exit status: %d", exitStatus)

	// Read the output
	f, err := os.Open(outputFile.Name())
	if err != nil {
		return err
	}
	defer f.Close()

	if remote.Stdout != nil {
		io.Copy(remote.Stdout, f)
	} else {
		output, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}

		log.Printf("Command output: %s", string(output))
	}

	// Finally, we're done
	remote.SetExited(int(exitStatus))

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
		Command: fmt.Sprintf("cp %s/%s %s", c.ContainerDir,
			filepath.Base(tempfile.Name()), dst),
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

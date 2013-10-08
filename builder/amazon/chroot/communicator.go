package chroot

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// Communicator is a special communicator that works by executing
// commands locally but within a chroot.
type Communicator struct {
	Chroot     string
	CmdWrapper CommandWrapper
}

func (c *Communicator) Start(cmd *packer.RemoteCmd) error {
	command, err := c.CmdWrapper(
		fmt.Sprintf("chroot %s %s", c.Chroot, cmd.Command))
	if err != nil {
		return err
	}

	localCmd := ShellCommand(command)
	localCmd.Stdin = cmd.Stdin
	localCmd.Stdout = cmd.Stdout
	localCmd.Stderr = cmd.Stderr
	log.Printf("Executing: %s %#v", localCmd.Path, localCmd.Args)
	if err := localCmd.Start(); err != nil {
		return err
	}

	go func() {
		exitStatus := 0
		if err := localCmd.Wait(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitStatus = 1

				// There is no process-independent way to get the REAL
				// exit status so we just try to go deeper.
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
					exitStatus = status.ExitStatus()
				}
			}
		}

		log.Printf(
			"Chroot execution exited with '%d': '%s'",
			exitStatus, cmd.Command)
		cmd.SetExited(exitStatus)
	}()

	return nil
}

func (c *Communicator) Upload(dst string, r io.Reader) error {
	dst = filepath.Join(c.Chroot, dst)
	log.Printf("Uploading to chroot dir: %s", dst)
	tf, err := ioutil.TempFile("", "packer-amazon-chroot")
	if err != nil {
		return fmt.Errorf("Error preparing shell script: %s", err)
	}
	defer os.Remove(tf.Name())
	io.Copy(tf, r)

	cpCmd, err := c.CmdWrapper(fmt.Sprintf("cp %s %s", tf.Name(), dst))
	if err != nil {
		return err
	}

	return ShellCommand(cpCmd).Run()
}

func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	// TODO: remove any file copied if it appears in `exclude`
	chrootDest := filepath.Join(c.Chroot, dst)
	log.Printf("Uploading directory '%s' to '%s'", src, chrootDest)
	cpCmd, err := c.CmdWrapper(fmt.Sprintf("cp -R %s* %s", src, chrootDest))
	if err != nil {
		return err
	}

	return ShellCommand(cpCmd).Run()
}

func (c *Communicator) Download(src string, w io.Writer) error {
	src = filepath.Join(c.Chroot, src)
	log.Printf("Downloading from chroot dir: %s", src)
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(w, f); err != nil {
		return err
	}

	return nil
}

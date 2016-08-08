package lxc

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
	"strings"
)

type LxcAttachCommunicator struct {
	RootFs        string
	ContainerName string
	CmdWrapper    CommandWrapper
}

func (c *LxcAttachCommunicator) Start(cmd *packer.RemoteCmd) error {
	localCmd, err := c.Execute(cmd.Command)

	if err != nil {
		return err
	}

	localCmd.Stdin = cmd.Stdin
	localCmd.Stdout = cmd.Stdout
	localCmd.Stderr = cmd.Stderr
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
			"lxc-attach execution exited with '%d': '%s'",
			exitStatus, cmd.Command)
		cmd.SetExited(exitStatus)
	}()

	return nil
}

func (c *LxcAttachCommunicator) Upload(dst string, r io.Reader, fi *os.FileInfo) error {
	dst = filepath.Join(c.RootFs, dst)
	log.Printf("Uploading to rootfs: %s", dst)
	tf, err := ioutil.TempFile("", "packer-lxc-attach")
	if err != nil {
		return fmt.Errorf("Error uploading file to rootfs: %s", err)
	}
	defer os.Remove(tf.Name())
	io.Copy(tf, r)

	cpCmd, err := c.CmdWrapper(fmt.Sprintf("sudo cp %s %s", tf.Name(), dst))
	if err != nil {
		return err
	}

	log.Printf("Running copy command: %s", dst)

	return ShellCommand(cpCmd).Run()
}

func (c *LxcAttachCommunicator) UploadDir(dst string, src string, exclude []string) error {
	// TODO: remove any file copied if it appears in `exclude`
	dest := filepath.Join(c.RootFs, dst)
	log.Printf("Uploading directory '%s' to rootfs '%s'", src, dest)
	cpCmd, err := c.CmdWrapper(fmt.Sprintf("sudo cp -R %s/. %s", src, dest))
	if err != nil {
		return err
	}

	return ShellCommand(cpCmd).Run()
}

func (c *LxcAttachCommunicator) Download(src string, w io.Writer) error {
	src = filepath.Join(c.RootFs, src)
	log.Printf("Downloading from rootfs dir: %s", src)
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

func (c *LxcAttachCommunicator) Execute(commandString string) (*exec.Cmd, error) {
	log.Printf("Executing with lxc-attach in container: %s %s %s", c.ContainerName, c.RootFs, commandString)
	command, err := c.CmdWrapper(
		fmt.Sprintf("sudo lxc-attach --name %s -- /bin/sh -c \"%s\"", c.ContainerName, commandString))
	if err != nil {
		return nil, err
	}

	localCmd := ShellCommand(command)
	log.Printf("Executing lxc-attach: %s %#v", localCmd.Path, localCmd.Args)

	return localCmd, nil
}

func (c *LxcAttachCommunicator) CheckInit() (string, error) {
	log.Printf("Debug runlevel exec")
	localCmd, err := c.Execute("/sbin/runlevel")

	if err != nil {
		return "", err
	}

	pr, _ := localCmd.StdoutPipe()
	if err = localCmd.Start(); err != nil {
		return "", err
	}

	output, err := ioutil.ReadAll(pr)

	if err != nil {
		return "", err
	}

	err = localCmd.Wait()

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
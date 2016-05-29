package lxd

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Communicator struct {
	ContainerName string
	CmdWrapper    CommandWrapper
}

func (c *Communicator) Start(cmd *packer.RemoteCmd) error {
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
			"lxc exec execution exited with '%d': '%s'",
			exitStatus, cmd.Command)
		cmd.SetExited(exitStatus)
	}()

	return nil
}

func (c *Communicator) Upload(dst string, r io.Reader, fi *os.FileInfo) error {
	cpCmd, err := c.CmdWrapper(fmt.Sprintf("lxc file push - %s", filepath.Join(c.ContainerName, dst)))
	if err != nil {
		return err
	}

	log.Printf("Running copy command: %s", cpCmd)
	command := ShellCommand(cpCmd)
	command.Stdin = r

	return command.Run()
}

func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	// XXX FIXME. lxc file push doesn't yet support directory uploads

	// TODO: make a tar tmpfile of the source, upload it to the dest, then untar it on the remote.
	// This approach would work when the LXD host is not the localhost.
	dest := filepath.Join("/var/lib/lxd/containers/", c.ContainerName, "rootfs", dst)
	log.Printf("Uploading directory '%s' to rootfs '%s'", src, dest)
	cpCmd, err := c.CmdWrapper(fmt.Sprintf("sudo cp -R %s/. %s", src, dest))
	if err != nil {
		log.Printf("Error when uploading directory '%s' to rootfs '%s': %s", src, dest, err)
		return err
	}

	log.Printf("Running copy command: %s", cpCmd)
	return ShellCommand(cpCmd).Run()
}

func (c *Communicator) Download(src string, w io.Writer) error {
	cpCmd, err := c.CmdWrapper(fmt.Sprintf("lxc file pull %s -", filepath.Join(c.ContainerName, src)))
	if err != nil {
		return err
	}

	log.Printf("Running copy command: %s", cpCmd)
	command := ShellCommand(cpCmd)
	command.Stdout = w

	return command.Run()
}

func (c *Communicator) DownloadDir(src string, dst string, exclude []string) error {
	return fmt.Errorf("DownloadDir is not implemented for lxc")
}

func (c *Communicator) Execute(commandString string) (*exec.Cmd, error) {
	log.Printf("Executing with lxc exec in container: %s %s", c.ContainerName, commandString)
	command, err := c.CmdWrapper(
		fmt.Sprintf("sudo lxc exec %s -- /bin/sh -c \"%s\"", c.ContainerName, commandString))
	if err != nil {
		return nil, err
	}

	localCmd := ShellCommand(command)
	log.Printf("Executing lxc exec: %s %#v", localCmd.Path, localCmd.Args)

	return localCmd, nil
}

func (c *Communicator) CheckInit() (string, error) {
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

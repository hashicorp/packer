package lxd

import (
	"fmt"
	"github.com/mitchellh/packer/packer"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	// NOTE:lxc file push doesn't yet support directory uploads.
	// As a work around, we tar up the folder, upload it as a file, then extract it

	// Don't use 'z' flag as compressing may take longer and the transfer is likely local.
	// If this isn't the case, it is possible for the user to crompress in another step then transfer.
	// It wouldn't be possibe to disable compression, without exposing this option.
	tar, err := c.CmdWrapper(fmt.Sprintf("tar -cf - -C %s .", src))
	if err != nil {
		return err
	}

	cp, err := c.CmdWrapper(fmt.Sprintf("lxc exec %s -- tar -xf - -C %s", c.ContainerName, dst))
	if err != nil {
		return err
	}

	tarCmd := ShellCommand(tar)
	cpCmd := ShellCommand(cp)

	cpCmd.Stdin, _ = tarCmd.StdoutPipe()
	log.Printf("Starting tar command: %s", tar)
	err = tarCmd.Start()
	if err != nil {
		return err
	}

	log.Printf("Running cp command: %s", cp)
	err = cpCmd.Run()
	if err != nil {
		log.Printf("Error running cp command: %s", err)
		return err
	}

	err = tarCmd.Wait()
	if err != nil {
		log.Printf("Error running tar command: %s", err)
		return err
	}

	return nil
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
	// TODO This could probably be "lxc exec <container> -- cd <src> && tar -czf - | tar -xzf - -C <dst>"
	return fmt.Errorf("DownloadDir is not implemented for lxc")
}

func (c *Communicator) Execute(commandString string) (*exec.Cmd, error) {
	log.Printf("Executing with lxc exec in container: %s %s", c.ContainerName, commandString)
	command, err := c.CmdWrapper(
		fmt.Sprintf("lxc exec %s -- /bin/sh -c \"%s\"", c.ContainerName, commandString))
	if err != nil {
		return nil, err
	}

	localCmd := ShellCommand(command)
	log.Printf("Executing lxc exec: %s %#v", localCmd.Path, localCmd.Args)

	return localCmd, nil
}

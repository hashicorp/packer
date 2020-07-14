package lxd

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/packer"
)

type Communicator struct {
	ContainerName string
	CmdWrapper    CommandWrapper
	Client        lxdClient
}

func (c *Communicator) Start(ctx context.Context, cmd *packer.RemoteCmd) error {
	return c.Client.ExecuteContainer(c.ContainerName, c.CmdWrapper, cmd)
}

func (c *Communicator) Upload(dst string, r io.Reader, fi *os.FileInfo) error {
	ctx := context.TODO()

	fileDestination := filepath.Join(c.ContainerName, dst)
	// find out if the place we are pushing to is a directory
	testDirectoryCommand := fmt.Sprintf(`test -d "%s"`, dst)
	cmd := &packer.RemoteCmd{Command: testDirectoryCommand}
	err := c.Start(ctx, cmd)

	if err != nil {
		log.Printf("Unable to check whether remote path is a dir: %s", err)
		return err
	}
	cmd.Wait()

	if cmd.ExitStatus() == 0 {
		log.Printf("path is a directory; copying file into directory.")
		fileDestination = filepath.Join(c.ContainerName, dst, (*fi).Name())
	}

	cpCmd, err := c.CmdWrapper(fmt.Sprintf("lxc file push - %s", fileDestination))
	if err != nil {
		return err
	}

	log.Printf("Running copy command: %s", cpCmd)
	command := ShellCommand(cpCmd)
	command.Stdin = r

	return command.Run()
}

func (c *Communicator) UploadDir(dst string, src string, exclude []string) error {
	fileDestination := fmt.Sprintf("%s/%s", c.ContainerName, dst)
	pushCommand := fmt.Sprintf("lxc file push --debug -pr %s %s", src, fileDestination)
	log.Printf(pushCommand)
	cp, err := c.CmdWrapper(pushCommand)
	if err != nil {
		log.Printf("Error running cp command: %s", err)
		return err
	}

	cpCmd := ShellCommand(cp)

	log.Printf("Running cp command: %s", cp)
	err = cpCmd.Run()
	if err != nil {
		log.Printf("Error running cp command: %s", err)
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

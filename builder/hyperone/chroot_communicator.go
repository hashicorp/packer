package hyperone

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type CommandWrapper func(string) (string, error)

// ChrootCommunicator works as a wrapper on SSHCommunicator, modyfing paths in
// flight to be run in a chroot.
type ChrootCommunicator struct {
	Chroot     string
	CmdWrapper CommandWrapper
	Wrapped    packersdk.Communicator
}

func (c *ChrootCommunicator) Start(ctx context.Context, cmd *packersdk.RemoteCmd) error {
	command := strconv.Quote(cmd.Command)
	chrootCommand, err := c.CmdWrapper(
		fmt.Sprintf("sudo chroot %s /bin/sh -c %s", c.Chroot, command))
	if err != nil {
		return err
	}

	cmd.Command = chrootCommand

	return c.Wrapped.Start(ctx, cmd)
}

func (c *ChrootCommunicator) Upload(dst string, r io.Reader, fi *os.FileInfo) error {
	dst = filepath.Join(c.Chroot, dst)
	return c.Wrapped.Upload(dst, r, fi)
}

func (c *ChrootCommunicator) UploadDir(dst string, src string, exclude []string) error {
	dst = filepath.Join(c.Chroot, dst)
	return c.Wrapped.UploadDir(dst, src, exclude)
}

func (c *ChrootCommunicator) Download(src string, w io.Writer) error {
	src = filepath.Join(c.Chroot, src)
	return c.Wrapped.Download(src, w)
}

func (c *ChrootCommunicator) DownloadDir(src string, dst string, exclude []string) error {
	src = filepath.Join(c.Chroot, src)
	return c.Wrapped.DownloadDir(src, dst, exclude)
}

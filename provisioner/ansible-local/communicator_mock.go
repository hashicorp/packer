package ansiblelocal

import (
	"context"
	"io"
	"os"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

type communicatorMock struct {
	startCommand      []string
	uploadDestination []string
}

func (c *communicatorMock) Start(ctx context.Context, cmd *packersdk.RemoteCmd) error {
	c.startCommand = append(c.startCommand, cmd.Command)
	cmd.SetExited(0)
	return nil
}

func (c *communicatorMock) Upload(dst string, _ io.Reader, _ *os.FileInfo) error {
	c.uploadDestination = append(c.uploadDestination, dst)
	return nil
}

func (c *communicatorMock) UploadDir(dst, src string, exclude []string) error {
	return nil
}

func (c *communicatorMock) Download(src string, dst io.Writer) error {
	return nil
}

func (c *communicatorMock) DownloadDir(src, dst string, exclude []string) error {
	return nil
}

func (c *communicatorMock) verify() {
}

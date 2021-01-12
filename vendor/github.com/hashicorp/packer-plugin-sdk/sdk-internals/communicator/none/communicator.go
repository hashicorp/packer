// Package none implements the 'none' communicator. Plugin maintainers should not
// import this package directly, instead using the tooling in the
// "packer-plugin-sdk/communicator" module.
package none

import (
	"context"
	"errors"
	"io"
	"os"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type comm struct {
	config string
}

// Creates a null packersdk.Communicator implementation. This takes
// an already existing configuration.
func New(config string) (result *comm, err error) {
	// Establish an initial connection and connect
	result = &comm{
		config: config,
	}

	return
}

func (c *comm) Start(ctx context.Context, cmd *packersdk.RemoteCmd) (err error) {
	cmd.SetExited(0)
	return
}

func (c *comm) Upload(path string, input io.Reader, fi *os.FileInfo) error {
	return errors.New("Upload is not implemented when communicator = 'none'")
}

func (c *comm) UploadDir(dst string, src string, excl []string) error {
	return errors.New("UploadDir is not implemented when communicator = 'none'")
}

func (c *comm) Download(path string, output io.Writer) error {
	return errors.New("Download is not implemented when communicator = 'none'")
}

func (c *comm) DownloadDir(dst string, src string, excl []string) error {
	return errors.New("DownloadDir is not implemented when communicator = 'none'")
}

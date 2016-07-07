package none

import (
	"errors"
	"github.com/mitchellh/packer/packer"
	"io"
	"os"
)

type comm struct {
	config string
}

// Creates a null packer.Communicator implementation. This takes
// an already existing configuration.
func New(config string) (result *comm, err error) {
	// Establish an initial connection and connect
	result = &comm{
		config: config,
	}

	return
}

func (c *comm) Start(cmd *packer.RemoteCmd) (err error) {
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

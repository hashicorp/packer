package packer

import (
	"io"
)

// MockCommunicator is a valid Communicator implementation that can be
// used for tests.
type MockCommunicator struct {
	Stderr io.Reader
	Stdout io.Reader
}

func (c *MockCommunicator) Start(rc *RemoteCmd) error {
	go func() {
		rc.Lock()
		defer rc.Unlock()

		if rc.Stdout != nil && c.Stdout != nil {
			io.Copy(rc.Stdout, c.Stdout)
		}

		if rc.Stderr != nil && c.Stderr != nil {
			io.Copy(rc.Stderr, c.Stderr)
		}
	}()

	return nil
}

func (c *MockCommunicator) Upload(string, io.Reader) error {
	return nil
}

func (c *MockCommunicator) UploadDir(string, string, []string) error {
	return nil
}

func (c *MockCommunicator) Download(string, io.Writer) error {
	return nil
}

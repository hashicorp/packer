package docker

import (
	"io"
)

// Driver is the interface that has to be implemented to communicate with
// Docker. The Driver interface also allows the steps to be tested since
// a mock driver can be shimmed in.
type Driver interface {
	// Export exports the container with the given ID to the given writer.
	Export(id string, dst io.Writer) error

	// Pull should pull down the given image.
	Pull(image string) error

	// StartContainer starts a container and returns the ID for that container,
	// along with a potential error.
	StartContainer(*ContainerConfig) (string, error)

	// StopContainer forcibly stops a container.
	StopContainer(id string) error

	// Verify verifies that the driver can run
	Verify() error
}

// ContainerConfig is the configuration used to start a container.
type ContainerConfig struct {
	Image   string
	Volumes map[string]string
}

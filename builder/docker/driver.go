package docker

import (
	"io"
)

// Driver is the interface that has to be implemented to communicate with
// Docker. The Driver interface also allows the steps to be tested since
// a mock driver can be shimmed in.
type Driver interface {
	// Delete an image that is imported into Docker
	DeleteImage(id string) error

	// Export exports the container with the given ID to the given writer.
	Export(id string, dst io.Writer) error

	// Import imports a container from a tar file
	Import(path, repo string) (string, error)

	// Pull should pull down the given image.
	Pull(image string) error

	// Push pushes an image to a Docker index/registry.
	Push(name string) error

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
	Image      string
	RunCommand []string
	Volumes    map[string]string
}

// This is the template that is used for the RunCommand in the ContainerConfig.
type startContainerTemplate struct {
	Image   string
	Volumes string
}

package docker

// Driver is the interface that has to be implemented to communicate with
// Docker. The Driver interface also allows the steps to be tested since
// a mock driver can be shimmed in.
type Driver interface {
	// Pull should pull down the given image.
	Pull(image string) error

	// StartContainer starts a container and returns the ID for that container,
	// along with a potential error.
	StartContainer(*ContainerConfig) (string, error)

	// StopContainer forcibly stops a container.
	StopContainer(id string) error
}

// ContainerConfig is the configuration used to start a container.
type ContainerConfig struct {
	Image   string
	Volumes map[string]string
}

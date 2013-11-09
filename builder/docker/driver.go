package docker

// Driver is the interface that has to be implemented to communicate with
// Docker. The Driver interface also allows the steps to be tested since
// a mock driver can be shimmed in.
type Driver interface {
	// Pull should pull down the given image.
	Pull(image string) error
}

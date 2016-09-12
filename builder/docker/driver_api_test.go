package docker

import "testing"

func TestDockerApiDriver_impl(t *testing.T) {
	var _ Driver = new(DockerApiDriver)
}

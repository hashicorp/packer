package docker

import "testing"

func TestDockerDriver_impl(t *testing.T) {
	var _ Driver = new(DockerDriver)
}

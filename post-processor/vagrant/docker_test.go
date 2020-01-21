package vagrant

import (
	"testing"
)

func TestDockerProvider_impl(t *testing.T) {
	var _ Provider = new(DockerProvider)
}

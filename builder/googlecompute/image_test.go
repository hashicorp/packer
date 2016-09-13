package googlecompute

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func StubImage(name, project string, licenses []string, sizeGb int64) *Image {
	return &Image{
		Licenses:  licenses,
		Name:      name,
		ProjectId: project,
		SelfLink:  fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/images/%s", project, name),
		SizeGb:    sizeGb,
	}
}

func TestImage_IsWindows(t *testing.T) {
	i := StubImage("foo", "foo-project", []string{"license-foo", "license-bar"}, 100)
	assert.False(t, i.IsWindows())

	i = StubImage("foo", "foo-project", []string{"license-foo", "windows-license"}, 100)
	assert.True(t, i.IsWindows())
}

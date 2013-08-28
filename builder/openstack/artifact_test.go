package openstack

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestArtifact_Impl(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var actual packer.Artifact
	assert.Implementor(&Artifact{}, &actual, "should be an Artifact")
}

func TestArtifactId(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	expected := `b8cdf55b-c916-40bd-b190-389ec144c4ed`

	a := &Artifact{
		ImageId: "b8cdf55b-c916-40bd-b190-389ec144c4ed",
	}

	result := a.Id()
	assert.Equal(result, expected, "should match output")
}

func TestArtifactString(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	expected := "An image was created: b8cdf55b-c916-40bd-b190-389ec144c4ed"

	a := &Artifact{
		ImageId: "b8cdf55b-c916-40bd-b190-389ec144c4ed",
	}
	result := a.String()
	assert.Equal(result, expected, "should match output")
}

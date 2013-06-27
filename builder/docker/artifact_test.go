package docker

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

func TestArtifactString(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	expected := "Image ID: abc123"
	a := &Artifact{"abc123"}
	result := a.String()
	assert.Equal(result, expected, "should match output")
}

package ebs

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestArtifact_Impl(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var actual packer.Artifact
	assert.Implementor(&artifact{}, &actual, "should be an Artifact")
}

func TestArtifactId(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	expected := `east:foo,west:bar`

	amis := make(map[string]string)
	amis["east"] = "foo"
	amis["west"] = "bar"

	a := &artifact{amis, nil}
	result := a.Id()
	assert.Equal(result, expected, "should match output")
}

func TestArtifactString(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	expected := `AMIs were created:

east: foo
west: bar`

	amis := make(map[string]string)
	amis["east"] = "foo"
	amis["west"] = "bar"

	a := &artifact{amis, nil}
	result := a.String()
	assert.Equal(result, expected, "should match output")
}

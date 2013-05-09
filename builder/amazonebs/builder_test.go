package amazonebs

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"testing"
)

func TestBuilder_ImplementsBuilder(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	var actual packer.Builder
	assert.Implementor(&Builder{}, &actual, "should be a Builder")
}

package plugin

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"testing"
)

type helperBuilder byte

func (helperBuilder) Prepare(interface{}) {}

func (helperBuilder) Run(packer.Build, packer.Ui) {}

func TestBuilder_NoExist(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Builder(exec.Command("i-should-never-ever-ever-exist"))
	assert.NotNil(err, "should have an error")
}

func TestBuilder_Good(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Builder(helperProcess("builder"))
	assert.Nil(err, "should start builder properly")
}


package plugin

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"testing"
)

type helperHook byte

func (helperHook) Run(string, interface{}, packer.Ui) {}

func TestHook_NoExist(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Hook(exec.Command("i-should-never-ever-ever-exist"))
	assert.NotNil(err, "should have an error")
}

func TestHook_Good(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Hook(helperProcess("hook"))
	assert.Nil(err, "should start hook properly")
}

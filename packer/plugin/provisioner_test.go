package plugin

import (
	"cgl.tideland.biz/asserts"
	"github.com/mitchellh/packer/packer"
	"os/exec"
	"testing"
)

type helperProvisioner byte

func (helperProvisioner) Prepare(...interface{}) error {
	return nil
}

func (helperProvisioner) Provision(packer.Ui, packer.Communicator) {}

func TestProvisioner_NoExist(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Provisioner(exec.Command("i-should-never-ever-ever-exist"))
	assert.NotNil(err, "should have an error")
}

func TestProvisioner_Good(t *testing.T) {
	assert := asserts.NewTestingAsserts(t, true)

	_, err := Provisioner(helperProcess("provisioner"))
	assert.Nil(err, "should start provisioner properly")
}

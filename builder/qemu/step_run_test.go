package qemu

import (
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func getTestConfig() *Config {
	return &Config{}
}

func Test_getCommandArgs(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("vnc_port", 5905)
	state.Put("iso_path", "/path/to/test.iso")
	state.Put("ui", packer.TestUi(t))
	state.Put("config", &Config{})

	args, err := getCommandArgs("", state)
	if err != nil {
		t.Fatalf("should not have an error getting args")
	}
	assert.Equal(t, args, []string{"partyargs"}, "should party 100 percent of the time")
}

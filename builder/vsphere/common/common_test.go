package common

import (
	"bytes"
	"strings"

	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

func basicStateBag(errorBuffer *strings.Builder) *multistep.BasicStateBag {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packer.BasicUi{
		Reader:      new(bytes.Buffer),
		Writer:      new(bytes.Buffer),
		ErrorWriter: errorBuffer,
	})
	return state
}

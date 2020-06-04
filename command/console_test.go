package command

import (
	"strings"
	"testing"

	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func TestConsole_help(t *testing.T) {
	input := "help"
	p := helperCommand(t, "console")
	p.Stdin = strings.NewReader(input)
	bs, err := p.Output()
	if err != nil {
		t.Errorf("cat: %v", err)
	}
	assert.Equal(t, packer.ConsoleHelpMessage+"\n", string(bs))
}

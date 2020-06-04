package command

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func Test_piping(t *testing.T) {

	tc := []struct {
		piped    string
		command  []string
		env      []string
		expected string
	}{
		{"help", []string{"console"}, nil, packer.ConsoleHelp + "\n"},
		{"help", []string{"console", "--config-type=hcl2"}, nil, hcl2template.PackerConsoleHelp + "\n"},
		{"var.fruit", []string{"console", filepath.Join(testFixture("var-arg"), "fruit_builder.pkr.hcl")}, []string{"PKR_VAR_fruit=potato"}, "potato\n"},
		{"upper(var.fruit)", []string{"console", filepath.Join(testFixture("var-arg"), "fruit_builder.pkr.hcl")}, []string{"PKR_VAR_fruit=potato"}, "POTATO\n"},
	}

	for _, tc := range tc {
		t.Run(fmt.Sprintf("echo %q | packer %s", tc.piped, tc.command), func(t *testing.T) {
			p := helperCommand(t, tc.command...)
			p.Stdin = strings.NewReader(tc.piped)
			p.Env = append(p.Env, tc.env...)
			bs, err := p.Output()
			if err != nil {
				t.Fatalf("%v", err)
			}
			assert.Equal(t, tc.expected, string(bs))
		})
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/packer"
	"github.com/stretchr/testify/assert"
)

func Test_console(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}

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
		{"1 + 5", []string{"console", "--config-type=hcl2"}, nil, "6\n"},
		{"var.images", []string{"console", filepath.Join(testFixture("var-arg"), "map.pkr.hcl")}, nil, "{\n" + `  "key" = "value"` + "\n}\n"},
		{"path.cwd", []string{"console", filepath.Join(testFixture("var-arg"), "map.pkr.hcl")}, nil, strings.ReplaceAll(cwd, `\`, `/`) + "\n"},
		{"path.root", []string{"console", filepath.Join(testFixture("var-arg"), "map.pkr.hcl")}, nil, strings.ReplaceAll(testFixture("var-arg"), `\`, `/`) + "\n"},
		{"var.list_of_string[0]", []string{"console", `-var=list_of_string=["first"]`, filepath.Join(testFixture("hcl", "variables", "list_of_string"))}, nil, "first\n"},
		{"var.untyped[2]", []string{"console", filepath.Join(testFixture("hcl", "variables", "untyped_var"))}, nil, "strings\n"},
		{"var.untyped", []string{"console", `-var=untyped=just_a_string`, filepath.Join(testFixture("hcl", "variables", "untyped_var"))}, nil, "just_a_string\n"},
		{"var.untyped", []string{"console", filepath.Join(testFixture("hcl", "variables", "untyped_var", "var.pkr.hcl"))}, nil, "<unknown>\n"},
		{"var.untyped", []string{"console", filepath.Join(testFixture("hcl", "variables", "untyped_var", "var.pkr.hcl"))}, []string{"PKR_VAR_untyped=just_a_string"}, "just_a_string\n"},
	}

	for _, tc := range tc {
		t.Run(fmt.Sprintf("echo %q | packer %s", tc.piped, tc.command), func(t *testing.T) {
			p := helperCommand(t, tc.command...)
			p.Stdin = strings.NewReader(tc.piped)
			p.Env = append(p.Env, tc.env...)
			bs, err := p.Output()
			if err != nil {
				t.Fatalf("%v: %s", err, bs)
			}
			assert.Equal(t, tc.expected, string(bs))
		})
	}
}

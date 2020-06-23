package command

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_commands(t *testing.T) {

	tc := []struct {
		command  []string
		env      []string
		expected string
	}{
		{[]string{"inspect", "-var=fruit=banana", filepath.Join(testFixture("var-arg"), "fruit_builder.pkr.hcl")}, nil, `Packer Inspect: HCL2 mode

> input-variables:

var.fruit: "banana" [debug: {Type:cty.String,CmdValue:banana,VarfileValue:null,EnvValue:null,DefaultValue:null}]

> local-variables:

local.fruit: "banana"

> builds:

  > <unnamed build 0>:

    provisioners:

      shell-local

    post-processors:

      <no post-processor>

`},
	}

	for _, tc := range tc {
		t.Run(fmt.Sprintf("packer %s", tc.command), func(t *testing.T) {
			p := helperCommand(t, tc.command...)
			p.Env = append(p.Env, tc.env...)
			bs, err := p.Output()
			if err != nil {
				t.Fatalf("%v: %s", err, bs)
			}
			actual := string(bs)
			if diff := cmp.Diff(tc.expected, actual); diff != "" {
				t.Fatalf("unexpected ouput %s", diff)
			}
		})
	}
}

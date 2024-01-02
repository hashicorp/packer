// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

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
		{[]string{"inspect", filepath.Join(testFixture("var-arg"), "fruit_builder.pkr.hcl")}, nil, `Packer Inspect: HCL2 mode

> input-variables:

var.fruit: "<unknown>"

> local-variables:

local.fruit: "<unknown>"

> builds:

  > <unnamed build 0>:

    sources:

      null.builder

    provisioners:

      shell-local

    post-processors:

      <no post-processor>

`},
		{[]string{"inspect", "-var=fruit=banana", filepath.Join(testFixture("var-arg"), "fruit_builder.pkr.hcl")}, nil, `Packer Inspect: HCL2 mode

> input-variables:

var.fruit: "banana"

> local-variables:

local.fruit: "banana"

> builds:

  > <unnamed build 0>:

    sources:

      null.builder

    provisioners:

      shell-local

    post-processors:

      <no post-processor>

`},
		{[]string{"inspect", "-var=fruit=peach",
			"-var=unknown_string=also_peach",
			`-var=unknown_unknown=["peach_too"]`,
			`-var=unknown_list_of_string=["first_peach", "second_peach"]`,
			filepath.Join(testFixture("hcl"), "inspect", "fruit_string.pkr.hcl")}, nil,
			`Packer Inspect: HCL2 mode

> input-variables:

var.default_from_env: ""
var.fruit: "peach"
var.other_default_from_env: ""
var.unknown_list_of_string: "[\n  \"first_peach\",\n  \"second_peach\",\n]"
var.unknown_string: "also_peach"
var.unknown_unknown: "[\"peach_too\"]"

> local-variables:


> builds:

`},
		{[]string{"inspect", "-var=fruit=peach", "-var=other_default_from_env=apple", filepath.Join(testFixture("hcl"), "inspect")}, []string{"DEFAULT_FROM_ENV=cherry"}, `Packer Inspect: HCL2 mode

> input-variables:

var.default_from_env: "cherry"
var.fruit: "peach"
var.other_default_from_env: "apple"
var.unknown_list_of_string: "<unknown>"
var.unknown_string: "<unknown>"
var.unknown_unknown: "<unknown>"

> local-variables:


> builds:

  > aws_example_builder:

  > Description: The builder of clouds !!

Use it at will.


    sources:

      amazon-ebs.example-1

      amazon-ebs.example-2

    provisioners:

      shell

    post-processors:

      0:
        manifest

      1:
        shell-local

      2:
        manifest
        shell-local

`},
		{[]string{"inspect", filepath.Join(testFixture("inspect"), "unset_var.json")}, nil, `Packer Inspect: JSON mode
Required variables:

  something

Optional variables and their defaults:


Builders:

  <No builders>

Provisioners:

  <No provisioners>

Note: If your build names contain user variables or template
functions such as 'timestamp', these are processed at build time,
and therefore only show in their raw form here.
`},
		{
			[]string{
				"inspect", filepath.Join(testFixture("hcl-inspect-with-sensitive-vars")),
			},
			nil,
			testFixtureContent("hcl-inspect-with-sensitive-vars", "expected-output.txt"),
		},
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

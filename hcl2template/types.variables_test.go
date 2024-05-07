// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package hcl2template

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/null"
	. "github.com/hashicorp/packer/hcl2template/internal"
	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"
)

func TestParse_variables(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic variables",
			defaultParser,
			parseTestArgs{"testdata/variables/basic.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"image_name": &Variable{
						Name:   "image_name",
						Type:   cty.String,
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("foo-image-{{user `my_secret`}}")}},
					},
					"key": &Variable{
						Name:   "key",
						Type:   cty.String,
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("value")}},
					},
					"my_secret": &Variable{
						Name:   "my_secret",
						Type:   cty.String,
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("foo")}},
					},
					"image_id": &Variable{
						Name:   "image_id",
						Type:   cty.String,
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("image-id-default")}},
					},
					"port": &Variable{
						Name:   "port",
						Type:   cty.Number,
						Values: []VariableAssignment{{From: "default", Value: cty.NumberIntVal(42)}},
					},
					"availability_zone_names": &Variable{
						Name: "availability_zone_names",
						Values: []VariableAssignment{{
							From: "default",
							Value: cty.ListVal([]cty.Value{
								cty.StringVal("us-west-1a"),
							}),
						}},
						Type:        cty.List(cty.String),
						Description: fmt.Sprintln("Describing is awesome ;D"),
					},
					"super_secret_password": &Variable{
						Name:      "super_secret_password",
						Sensitive: true,
						Values: []VariableAssignment{{
							From:  "default",
							Value: cty.NullVal(cty.String),
						}},
						Type:        cty.String,
						Description: fmt.Sprintln("Handle with care plz"),
					},
				},
				LocalVariables: Variables{
					"owner": &Variable{
						Name: "owner",
						Values: []VariableAssignment{{
							From:  "default",
							Value: cty.StringVal("Community Team"),
						}},
						Type: cty.String,
					},
					"service_name": &Variable{
						Name: "service_name",
						Values: []VariableAssignment{{
							From:  "default",
							Value: cty.StringVal("forum"),
						}},
						Type: cty.String,
					},
					"supersecret": &Variable{
						Name: "supersecret",
						Values: []VariableAssignment{{
							From:  "default",
							Value: cty.StringVal("secretvar"),
						}},
						Type:      cty.String,
						Sensitive: true,
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:           "null.test",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
				},
			},
			false,
		},
		{"duplicate variable",
			defaultParser,
			parseTestArgs{"testdata/variables/duplicate_variable.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"boolean_value": &Variable{
						Name: "boolean_value",
						Values: []VariableAssignment{{
							From:  "default",
							Value: cty.BoolVal(false),
						}},
						Type: cty.Bool,
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			false,
		},
		{"duplicate variable in variables",
			defaultParser,
			parseTestArgs{"testdata/variables/duplicate_variables.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"boolean_value": &Variable{
						Name: "boolean_value",
						Values: []VariableAssignment{{
							From:  "default",
							Value: cty.BoolVal(false),
						}},
						Type: cty.Bool,
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			false,
		},
		{"duplicate local block",
			defaultParser,
			parseTestArgs{"testdata/variables/duplicate_locals", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 "testdata/variables/duplicate_locals",
				LocalVariables: Variables{
					"sensible": &Variable{
						Values: []VariableAssignment{
							{
								From:  "default",
								Value: cty.StringVal("something"),
							},
						},
						Type: cty.String,
						Name: "sensible",
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			false,
		},
		{"invalid default type",
			defaultParser,
			parseTestArgs{"testdata/variables/invalid_default.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"broken_type": &Variable{
						Name: "broken_type",
						Values: []VariableAssignment{{
							From:  "default",
							Value: cty.UnknownVal(cty.DynamicPseudoType),
						}},
						Type: cty.List(cty.String),
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			false,
		},

		{"unknown key",
			defaultParser,
			parseTestArgs{"testdata/variables/unknown_key.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"broken_variable": &Variable{
						Name:   "broken_variable",
						Values: []VariableAssignment{{From: "default", Value: cty.BoolVal(true)}},
						Type:   cty.Bool,
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			false,
		},

		{"unset used variable",
			defaultParser,
			parseTestArgs{"testdata/variables/unset_used_string_variable.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"foo": &Variable{
						Name: "foo",
						Type: cty.String,
					},
				},
			},
			true, true,
			[]packersdk.Build{},
			true,
		},

		{"unset unused variable",
			defaultParser,
			parseTestArgs{"testdata/variables/unset_unused_string_variable.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"foo": &Variable{
						Name: "foo",
						Type: cty.String,
					},
				},
				Sources: map[SourceRef]SourceBlock{
					SourceRef{Type: "null", Name: "null-builder"}: SourceBlock{
						Name: "null-builder",
						Type: "null",
					},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{Type: "null", Name: "null-builder"},
							},
						},
					},
				},
			},
			true, true,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:           "null",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
				},
			},
			false,
		},

		{"locals within another locals usage in different files",
			defaultParser,
			parseTestArgs{"testdata/variables/complicated", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
				Basedir: "testdata/variables/complicated",
				InputVariables: Variables{
					"name_prefix": &Variable{
						Name:   "name_prefix",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("foo")}},
						Type:   cty.String,
					},
				},
				LocalVariables: Variables{
					"name_prefix": &Variable{
						Name:   "name_prefix",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("foo")}},
						Type:   cty.String,
					},
					"foo": &Variable{
						Name:   "foo",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("foo")}},
						Type:   cty.String,
					},
					"bar": &Variable{
						Name:   "bar",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("foo")}},
						Type:   cty.String,
					},
					"for_var": &Variable{
						Name:   "for_var",
						Values: []VariableAssignment{{From: "default", Value: cty.StringVal("foo")}},
						Type:   cty.String,
					},
					"bar_var": &Variable{
						Name: "bar_var",
						Values: []VariableAssignment{{
							From: "default",
							Value: cty.TupleVal([]cty.Value{
								cty.StringVal("foo"),
								cty.StringVal("foo"),
								cty.StringVal("foo"),
							}),
						}},
						Type: cty.Tuple([]cty.Type{
							cty.String,
							cty.String,
							cty.String,
						}),
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:           "null.test",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
				},
			},
			false,
		},
		{"recursive locals",
			defaultParser,
			parseTestArgs{"testdata/variables/recursive_locals.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				LocalVariables:          Variables{},
			},
			true, true,
			[]packersdk.Build{},
			false,
		},

		{"set variable from var-file",
			defaultParser,
			parseTestArgs{"testdata/variables/foo-string.variable.pkr.hcl", nil, []string{"testdata/variables/set-foo-too-wee.hcl"}},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
				InputVariables: Variables{
					"foo": &Variable{
						Name: "foo",
						Values: []VariableAssignment{
							VariableAssignment{"default", cty.StringVal("bar"), nil},
							VariableAssignment{"varfile", cty.StringVal("wee"), nil},
						},
						Type: cty.String,
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:           "null.test",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
				},
			},
			false,
		},

		{"unknown variable from var-file",
			defaultParser,
			parseTestArgs{"testdata/variables/empty.pkr.hcl", nil, []string{"testdata/variables/set-foo-too-wee.hcl"}},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
				Basedir: filepath.Join("testdata", "variables"),
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:           "null.test",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
				},
			},
			false,
		},

		{"provisioner variable decoding",
			defaultParser,
			parseTestArgs{"testdata/variables/provisioner_variable_decoding.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"max_retries": &Variable{
						Name:   "max_retries",
						Values: []VariableAssignment{{"default", cty.StringVal("1"), nil}},
						Type:   cty.String,
					},
					"max_retries_int": &Variable{
						Name:   "max_retries_int",
						Values: []VariableAssignment{{"default", cty.NumberIntVal(1), nil}},
						Type:   cty.Number,
					},
				},
				Sources: map[SourceRef]SourceBlock{
					SourceRef{Type: "null", Name: "null-builder"}: SourceBlock{
						Name: "null-builder",
						Type: "null",
					},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{Type: "null", Name: "null-builder"},
							},
						},
						ProvisionerBlocks: []*ProvisionerBlock{
							{
								PType:      "shell",
								MaxRetries: 1,
							},
							{
								PType:      "shell",
								MaxRetries: 1,
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{&packer.CoreBuild{
				Type:        "null.null-builder",
				BuilderType: "null",
				Prepared:    true,
				Builder:     &null.Builder{},
				Provisioners: []packer.CoreBuildProvisioner{
					{
						PType: "shell",
						Provisioner: &packer.RetriedProvisioner{
							MaxRetries: 1,
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{
											Tags: []MockTag{},
										},
										NestedSlice: []NestedMockConfig{},
									},
								},
							},
						},
					},
					{
						PType: "shell",
						Provisioner: &packer.RetriedProvisioner{
							MaxRetries: 1,
							Provisioner: &HCL2Provisioner{
								Provisioner: &MockProvisioner{
									Config: MockConfig{
										NestedMockConfig: NestedMockConfig{
											Tags: []MockTag{},
										},
										NestedSlice: []NestedMockConfig{},
									},
								},
							},
						},
					},
				},
				PostProcessors: [][]packer.CoreBuildPostProcessor{},
			},
			},
			false,
		},

		{"valid validation block",
			defaultParser,
			parseTestArgs{"testdata/variables/validation/valid.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables", "validation"),
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceUseBlock{
							{
								SourceRef: SourceRef{
									Type: "null",
									Name: "test",
								},
							},
						},
					},
				},
				Sources: map[SourceRef]SourceBlock{
					{
						Type: "null",
						Name: "test",
					}: {
						Type: "null",
						Name: "test",
					},
				},
				InputVariables: Variables{
					"image_id": &Variable{
						Values: []VariableAssignment{
							{"default", cty.StringVal("ami-something-something"), nil},
						},
						Name: "image_id",
						Type: cty.String,
						Validations: []*VariableValidation{
							&VariableValidation{
								ErrorMessage: `The image_id value must be a valid AMI id, starting with "ami-".`,
							},
						},
					},
				},
			},
			false, false,
			[]packersdk.Build{
				&packer.CoreBuild{
					Type:           "null.test",
					BuilderType:    "null",
					Builder:        &null.Builder{},
					Provisioners:   []packer.CoreBuildProvisioner{},
					PostProcessors: [][]packer.CoreBuildPostProcessor{},
					Prepared:       true,
				},
			},
			false,
		},

		{"valid validation block - invalid default",
			defaultParser,
			parseTestArgs{"testdata/variables/validation/invalid_default.pkr.hcl", nil, nil},
			&PackerConfig{
				CorePackerVersionString: lockedVersion,
				Basedir:                 filepath.Join("testdata", "variables", "validation"),
				InputVariables: Variables{
					"image_id": &Variable{
						Values: []VariableAssignment{{"default", cty.StringVal("potato"), nil}},
						Name:   "image_id",
						Type:   cty.String,
						Validations: []*VariableValidation{
							&VariableValidation{
								ErrorMessage: `The image_id value must be a valid AMI id, starting with "ami-".`,
							},
						},
					},
				},
			},
			true, true,
			nil,
			false,
		},
	}
	testParse(t, tests)
}

func TestVariables_collectVariableValues(t *testing.T) {
	type args struct {
		env      []string
		hclFiles []string
		argv     map[string]string
	}
	tests := []struct {
		name              string
		variables         Variables
		validationOptions ValidationOptions
		args              args
		wantDiags         bool
		wantDiagsHasError bool
		wantVariables     Variables
		wantValues        map[string]cty.Value
	}{

		{name: "string",
			variables: Variables{"used_string": &Variable{
				Values: []VariableAssignment{
					{"default", cty.StringVal("default_value"), nil},
				},
				Type: cty.String,
			}},
			args: args{
				env: []string{`PKR_VAR_used_string=env_value`},
				hclFiles: []string{
					`used_string="xy"`,
					`used_string="varfile_value"`,
				},
				argv: map[string]string{
					"used_string": `cmd_value`,
				},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"used_string": &Variable{
					Type: cty.String,
					Values: []VariableAssignment{
						{"default", cty.StringVal(`default_value`), nil},
						{"env", cty.StringVal(`env_value`), nil},
						{"varfile", cty.StringVal(`xy`), nil},
						{"varfile", cty.StringVal(`varfile_value`), nil},
						{"cmd", cty.StringVal(`cmd_value`), nil},
					},
				},
			},
			wantValues: map[string]cty.Value{
				"used_string": cty.StringVal("cmd_value"),
			},
		},

		{name: "quoted string",
			variables: Variables{"quoted_string": &Variable{
				Values: []VariableAssignment{
					{"default", cty.StringVal(`"default_value"`), nil},
				},
				Type: cty.String,
			}},
			args: args{
				env: []string{`PKR_VAR_quoted_string="env_value"`},
				hclFiles: []string{
					`quoted_string="\"xy\""`,
					`quoted_string="\"varfile_value\""`,
				},
				argv: map[string]string{
					"quoted_string": `"cmd_value"`,
				},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"quoted_string": &Variable{
					Type: cty.String,
					Values: []VariableAssignment{
						{"default", cty.StringVal(`"default_value"`), nil},
						{"env", cty.StringVal(`"env_value"`), nil},
						{"varfile", cty.StringVal(`"xy"`), nil},
						{"varfile", cty.StringVal(`"varfile_value"`), nil},
						{"cmd", cty.StringVal(`"cmd_value"`), nil},
					},
				},
			},
			wantValues: map[string]cty.Value{
				"quoted_string": cty.StringVal(`"cmd_value"`),
			},
		},

		{name: "array of strings",
			variables: Variables{"used_strings": &Variable{
				Values: []VariableAssignment{
					{"default", stringListVal("default_value_1"), nil},
				},
				Type: cty.List(cty.String),
			}},
			args: args{
				env: []string{`PKR_VAR_used_strings=["env_value_1", "env_value_2"]`},
				hclFiles: []string{
					`used_strings=["xy"]`,
					`used_strings=["varfile_value_1"]`,
				},
				argv: map[string]string{
					"used_strings": `["cmd_value_1"]`,
				},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"used_strings": &Variable{
					Type: cty.List(cty.String),
					Values: []VariableAssignment{
						{"default", stringListVal("default_value_1"), nil},
						{"env", stringListVal("env_value_1", "env_value_2"), nil},
						{"varfile", stringListVal("xy"), nil},
						{"varfile", stringListVal("varfile_value_1"), nil},
						{"cmd", stringListVal("cmd_value_1"), nil},
					},
				},
			},
			wantValues: map[string]cty.Value{
				"used_strings": stringListVal("cmd_value_1"),
			},
		},

		{name: "bool",
			variables: Variables{"enabled": &Variable{
				Values: []VariableAssignment{{"default", cty.False, nil}},
				Type:   cty.Bool,
			}},
			args: args{
				env: []string{`PKR_VAR_enabled=true`},
				hclFiles: []string{
					`enabled="false"`,
				},
				argv: map[string]string{
					"enabled": `true`,
				},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"enabled": &Variable{
					Type: cty.Bool,
					Values: []VariableAssignment{
						{"default", cty.False, nil},
						{"env", cty.True, nil},
						{"varfile", cty.False, nil},
						{"cmd", cty.True, nil},
					},
				},
			},
			wantValues: map[string]cty.Value{
				"enabled": cty.True,
			},
		},

		{name: "invalid env var",
			variables: Variables{"used_string": &Variable{
				Values: []VariableAssignment{{"default", cty.StringVal("default_value"), nil}},
				Type:   cty.String,
			}},
			args: args{
				env: []string{`PKR_VAR_used_string`},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:   cty.String,
					Values: []VariableAssignment{{"default", cty.StringVal("default_value"), nil}},
				},
			},
			wantValues: map[string]cty.Value{
				"used_string": cty.StringVal("default_value"),
			},
		},

		{name: "undefined but set value - pkrvar file - normal mode",
			variables: Variables{},
			args: args{
				hclFiles: []string{`undefined_string="value"`},
			},

			// output
			wantDiags:         false,
			wantDiagsHasError: false,
			wantVariables:     Variables{},
			wantValues:        map[string]cty.Value{},
		},

		{name: "undefined but set value - pkrvar file - strict mode",
			variables: Variables{},
			validationOptions: ValidationOptions{
				WarnOnUndeclaredVar: true,
			},
			args: args{
				hclFiles: []string{`undefined_string="value"`},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: false,
			wantVariables:     Variables{},
			wantValues:        map[string]cty.Value{},
		},

		{name: "undefined but set value - env",
			variables: Variables{},
			args: args{
				env: []string{`PKR_VAR_undefined_string=value`},
			},

			// output
			wantDiags:     false,
			wantVariables: Variables{},
			wantValues:    map[string]cty.Value{},
		},

		{name: "undefined but set value - argv",
			variables: Variables{},
			args: args{
				argv: map[string]string{
					"undefined_string": "value",
				},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: true,
			wantVariables:     Variables{},
			wantValues:        map[string]cty.Value{},
		},

		{name: "value not corresponding to type - env",
			variables: Variables{
				"used_string": &Variable{
					Type: cty.List(cty.String),
				},
			},
			args: args{
				env: []string{`PKR_VAR_used_string="string"`},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: true,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:   cty.List(cty.String),
					Values: []VariableAssignment{{"env", cty.DynamicVal, nil}},
				},
			},
			wantValues: map[string]cty.Value{
				"used_string": cty.DynamicVal,
			},
		},

		{name: "value not corresponding to type - cfg file",
			variables: Variables{
				"used_string": &Variable{
					Type: cty.Bool,
				},
			},
			args: args{
				hclFiles: []string{`used_string=["string"]`},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: true,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:   cty.Bool,
					Values: []VariableAssignment{{"varfile", cty.DynamicVal, nil}},
				},
			},
			wantValues: map[string]cty.Value{
				"used_string": cty.DynamicVal,
			},
		},

		{name: "value not corresponding to type - argv",
			variables: Variables{
				"used_string": &Variable{
					Type: cty.Bool,
				},
			},
			args: args{
				argv: map[string]string{
					"used_string": `["true"]`,
				},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: true,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:   cty.Bool,
					Values: []VariableAssignment{{"cmd", cty.DynamicVal, nil}},
				},
			},
			wantValues: map[string]cty.Value{
				"used_string": cty.DynamicVal,
			},
		},

		{name: "defining a variable block in a variables file is invalid ",
			variables: Variables{},
			args: args{
				hclFiles: []string{`variable "something" {}`},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: true,
			wantVariables:     Variables{},
			wantValues:        map[string]cty.Value{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var files []*hcl.File
			parser := getBasicParser()
			for i, hclContent := range tt.args.hclFiles {
				file, diags := parser.ParseHCL([]byte(hclContent), fmt.Sprintf("test_file_%d_*"+hcl2AutoVarFileExt, i))
				if diags != nil {
					t.Fatalf("ParseHCLFile %d: %v", i, diags)
				}
				files = append(files, file)
			}
			cfg := &PackerConfig{
				InputVariables:    tt.variables,
				ValidationOptions: tt.validationOptions,
			}
			gotDiags := cfg.collectInputVariableValues(tt.args.env, files, tt.args.argv)
			if (gotDiags == nil) == tt.wantDiags {
				t.Fatalf("Variables.collectVariableValues() = %v, want %v", gotDiags, tt.wantDiags)
			}
			if tt.wantDiagsHasError != gotDiags.HasErrors() {
				t.Fatalf("Variables.collectVariableValues() unexpected diagnostics HasErrors. %s", gotDiags)
			}
			if diff := cmp.Diff(tt.wantVariables, tt.variables, cmpOpts...); diff != "" {
				t.Fatalf("didn't get expected variables: %s", diff)
			}
			values := map[string]cty.Value{}
			for k, v := range tt.variables {
				value, diag := v.Value(), v.ValidateValue()
				if diag != nil {
					t.Fatalf("Value %s: %v", k, diag)
				}
				values[k] = value
			}
			if diff := cmp.Diff(fmt.Sprintf("%#v", values), fmt.Sprintf("%#v", tt.wantValues)); diff != "" {
				t.Fatalf("didn't get expected values: %s", diff)
			}
		})
	}
}

func stringListVal(strings ...string) cty.Value {
	values := []cty.Value{}
	for _, str := range strings {
		values = append(values, cty.StringVal(str))
	}
	list, err := convert.Convert(cty.ListVal(values), cty.List(cty.String))
	if err != nil {
		panic(err)
	}
	return list
}

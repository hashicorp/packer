package hcl2template

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/packer"
)

func TestParse_variables(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic variables",
			defaultParser,
			parseTestArgs{"testdata/variables/basic.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"image_name": &Variable{
						DefaultValue: cty.StringVal("foo-image-{{user `my_secret`}}"),
					},
					"key": &Variable{
						DefaultValue: cty.StringVal("value"),
					},
					"my_secret": &Variable{
						DefaultValue: cty.StringVal("foo"),
					},
					"image_id": &Variable{
						DefaultValue: cty.StringVal("image-id-default"),
					},
					"port": &Variable{
						DefaultValue: cty.NumberIntVal(42),
					},
					"availability_zone_names": &Variable{
						DefaultValue: cty.ListVal([]cty.Value{
							cty.StringVal("us-west-1a"),
						}),
						Description: fmt.Sprintln("Describing is awesome ;D"),
					},
					"super_secret_password": &Variable{
						Sensitive:   true,
						Description: fmt.Sprintln("Handle with care plz"),
					},
				},
				LocalVariables: Variables{
					"owner": &Variable{
						DefaultValue: cty.StringVal("Community Team"),
					},
					"service_name": &Variable{
						DefaultValue: cty.StringVal("forum"),
					},
				},
			},
			false, false,
			[]packer.Build{},
			false,
		},
		{"duplicate variable",
			defaultParser,
			parseTestArgs{"testdata/variables/duplicate_variable.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"boolean_value": &Variable{},
				},
			},
			true, true,
			[]packer.Build{},
			false,
		},
		{"duplicate variable in variables",
			defaultParser,
			parseTestArgs{"testdata/variables/duplicate_variables.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"boolean_value": &Variable{},
				},
			},
			true, true,
			[]packer.Build{},
			false,
		},
		{"invalid default type",
			defaultParser,
			parseTestArgs{"testdata/variables/invalid_default.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"broken_type": &Variable{},
				},
			},
			true, true,
			[]packer.Build{},
			false,
		},
		{"invalid default type",
			defaultParser,
			parseTestArgs{"testdata/variables/unknown_key.pkr.hcl", nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"broken_type": &Variable{},
				},
			},
			true, false,
			[]packer.Build{},
			false,
		},
		{"locals within another locals usage in different files",
			defaultParser,
			parseTestArgs{"testdata/variables/complicated", nil},
			&PackerConfig{
				Basedir: "testdata/variables/complicated",
				InputVariables: Variables{
					"name_prefix": &Variable{
						DefaultValue: cty.StringVal("foo"),
					},
				},
				LocalVariables: Variables{
					"name_prefix": &Variable{
						DefaultValue: cty.StringVal("foo"),
					},
					"foo": &Variable{
						DefaultValue: cty.StringVal("foo"),
					},
					"bar": &Variable{
						DefaultValue: cty.StringVal("foo"),
					},
					"for_var": &Variable{
						DefaultValue: cty.StringVal("foo"),
					},
					"bar_var": &Variable{
						DefaultValue: cty.TupleVal([]cty.Value{
							cty.StringVal("foo"),
							cty.StringVal("foo"),
							cty.StringVal("foo"),
						}),
					},
				},
			},
			false, false,
			[]packer.Build{},
			false,
		},
		{"recursive locals",
			defaultParser,
			parseTestArgs{"testdata/variables/recursive_locals.pkr.hcl", nil},
			&PackerConfig{
				Basedir:        filepath.Join("testdata", "variables"),
				LocalVariables: Variables{},
			},
			true, true,
			[]packer.Build{},
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
		name          string
		variables     Variables
		args          args
		wantDiags     bool
		wantVariables Variables
		wantValues    map[string]cty.Value
	}{

		{name: "string",
			variables: Variables{"used_string": &Variable{DefaultValue: cty.StringVal("default_value")}},
			args: args{
				env: []string{`PKR_VAR_used_string="env_value"`},
				hclFiles: []string{
					`used_string="xy"`,
					`used_string="varfile_value"`,
				},
				argv: map[string]string{
					"used_string": `"cmd_value"`,
				},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"used_string": &Variable{
					CmdValue:     cty.StringVal("cmd_value"),
					VarfileValue: cty.StringVal("varfile_value"),
					EnvValue:     cty.StringVal("env_value"),
					DefaultValue: cty.StringVal("default_value"),
				},
			},
			wantValues: map[string]cty.Value{
				"used_string": cty.StringVal("cmd_value"),
			},
		},

		{name: "array of strings",
			variables: Variables{"used_strings": &Variable{
				DefaultValue: stringListVal("default_value_1"),
				Type:         cty.List(cty.String),
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
					Type:         cty.List(cty.String),
					CmdValue:     stringListVal("cmd_value_1"),
					VarfileValue: stringListVal("varfile_value_1"),
					EnvValue:     stringListVal("env_value_1", "env_value_2"),
					DefaultValue: stringListVal("default_value_1"),
				},
			},
			wantValues: map[string]cty.Value{
				"used_strings": stringListVal("cmd_value_1"),
			},
		},

		{name: "invalid env var",
			variables: Variables{"used_string": &Variable{DefaultValue: cty.StringVal("default_value")}},
			args: args{
				env: []string{`PKR_VAR_used_string`},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"used_string": &Variable{
					DefaultValue: cty.StringVal("default_value"),
				},
			},
			wantValues: map[string]cty.Value{
				"used_string": cty.StringVal("default_value"),
			},
		},

		{name: "undefined but set value",
			variables: Variables{},
			args: args{
				env:      []string{`PKR_VAR_unused_string=value`},
				hclFiles: []string{`unused_string="value"`},
			},

			// output
			wantDiags:     false,
			wantVariables: Variables{},
			wantValues:    map[string]cty.Value{},
		},

		{name: "undefined but set value - args",
			variables: Variables{},
			args: args{
				argv: map[string]string{
					"unused_string": "value",
				},
			},

			// output
			wantDiags:     true,
			wantVariables: Variables{},
			wantValues:    map[string]cty.Value{},
		},

		{name: "value not corresponding to type - env",
			variables: Variables{
				"used_string": &Variable{
					Type: cty.Bool,
				},
			},
			args: args{
				env: []string{`PKR_VAR_used_string=["string"]`},
			},

			// output
			wantDiags: true,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:     cty.Bool,
					EnvValue: cty.DynamicVal,
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
			wantDiags: true,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:         cty.Bool,
					VarfileValue: cty.DynamicVal,
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
					"used_string": `["string"]`,
				},
			},

			// output
			wantDiags: true,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:     cty.Bool,
					CmdValue: cty.DynamicVal,
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
			wantDiags:     true,
			wantVariables: Variables{},
			wantValues:    map[string]cty.Value{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var files []*hcl.File
			parser := getBasicParser()
			for i, hclContent := range tt.args.hclFiles {
				file, diags := parser.ParseHCL([]byte(hclContent), fmt.Sprintf("test_file_%d_*"+hcl2VarFileExt, i))
				if diags != nil {
					t.Fatalf("ParseHCLFile %d: %v", i, diags)
				}
				files = append(files, file)
			}
			if gotDiags := tt.variables.collectVariableValues(tt.args.env, files, tt.args.argv); (gotDiags == nil) == tt.wantDiags {
				t.Fatalf("Variables.collectVariableValues() = %v, want %v", gotDiags, tt.wantDiags)
			}
			if diff := cmp.Diff(fmt.Sprintf("%#v", tt.wantVariables), fmt.Sprintf("%#v", tt.variables)); diff != "" {
				t.Fatalf("didn't get expected variables: %s", diff)
			}
			values := map[string]cty.Value{}
			for k, v := range tt.variables {
				value, diag := v.Value()
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

package hcl2template

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/convert"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/packer/builder/null"
	"github.com/hashicorp/packer/packer"
)

func TestParse_variables(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic variables",
			defaultParser,
			parseTestArgs{"testdata/variables/basic.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"image_name": &Variable{
						Name:         "image_name",
						DefaultValue: cty.StringVal("foo-image-{{user `my_secret`}}"),
					},
					"key": &Variable{
						Name:         "key",
						DefaultValue: cty.StringVal("value"),
					},
					"my_secret": &Variable{
						Name:         "my_secret",
						DefaultValue: cty.StringVal("foo"),
					},
					"image_id": &Variable{
						Name:         "image_id",
						DefaultValue: cty.StringVal("image-id-default"),
					},
					"port": &Variable{
						Name:         "port",
						DefaultValue: cty.NumberIntVal(42),
					},
					"availability_zone_names": &Variable{
						Name: "availability_zone_names",
						DefaultValue: cty.ListVal([]cty.Value{
							cty.StringVal("us-west-1a"),
						}),
						Description: fmt.Sprintln("Describing is awesome ;D"),
					},
					"super_secret_password": &Variable{
						Name:         "super_secret_password",
						Sensitive:    true,
						DefaultValue: cty.NullVal(cty.String),
						Description:  fmt.Sprintln("Handle with care plz"),
					},
				},
				LocalVariables: Variables{
					"owner": &Variable{
						Name:         "owner",
						DefaultValue: cty.StringVal("Community Team"),
					},
					"service_name": &Variable{
						Name:         "service_name",
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
			parseTestArgs{"testdata/variables/duplicate_variable.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"boolean_value": &Variable{
						Name: "boolean_value",
					},
				},
			},
			true, true,
			[]packer.Build{},
			false,
		},
		{"duplicate variable in variables",
			defaultParser,
			parseTestArgs{"testdata/variables/duplicate_variables.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"boolean_value": &Variable{
						Name: "boolean_value",
					},
				},
			},
			true, true,
			[]packer.Build{},
			false,
		},
		{"invalid default type",
			defaultParser,
			parseTestArgs{"testdata/variables/invalid_default.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"broken_type": &Variable{
						Name: "broken_type",
					},
				},
			},
			true, true,
			[]packer.Build{},
			false,
		},

		{"unknown key",
			defaultParser,
			parseTestArgs{"testdata/variables/unknown_key.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"broken_variable": &Variable{
						Name:         "broken_variable",
						DefaultValue: cty.BoolVal(true),
					},
				},
			},
			true, false,
			[]packer.Build{},
			false,
		},

		{"unset used variable",
			defaultParser,
			parseTestArgs{"testdata/variables/unset_used_string_variable.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"foo": &Variable{
						Name: "foo",
					},
				},
			},
			true, true,
			[]packer.Build{},
			true,
		},

		{"unset unused variable",
			defaultParser,
			parseTestArgs{"testdata/variables/unset_unused_string_variable.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"foo": &Variable{
						Name: "foo",
					},
				},
				Sources: map[SourceRef]*SourceBlock{
					SourceRef{"null", "null-builder"}: &SourceBlock{
						Name: "null-builder",
						Type: "null",
					},
				},
				Builds: Builds{
					&BuildBlock{
						Sources: []SourceRef{SourceRef{"null", "null-builder"}},
					},
				},
			},
			true, true,
			[]packer.Build{
				&packer.CoreBuild{
					Type:           "null",
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
				Basedir: "testdata/variables/complicated",
				InputVariables: Variables{
					"name_prefix": &Variable{
						Name:         "name_prefix",
						DefaultValue: cty.StringVal("foo"),
					},
				},
				LocalVariables: Variables{
					"name_prefix": &Variable{
						Name:         "name_prefix",
						DefaultValue: cty.StringVal("foo"),
					},
					"foo": &Variable{
						Name:         "foo",
						DefaultValue: cty.StringVal("foo"),
					},
					"bar": &Variable{
						Name:         "bar",
						DefaultValue: cty.StringVal("foo"),
					},
					"for_var": &Variable{
						Name:         "for_var",
						DefaultValue: cty.StringVal("foo"),
					},
					"bar_var": &Variable{
						Name: "bar_var",
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
			parseTestArgs{"testdata/variables/recursive_locals.pkr.hcl", nil, nil},
			&PackerConfig{
				Basedir:        filepath.Join("testdata", "variables"),
				LocalVariables: Variables{},
			},
			true, true,
			[]packer.Build{},
			false,
		},

		{"set variable from var-file",
			defaultParser,
			parseTestArgs{"testdata/variables/foo-string.variable.pkr.hcl", nil, []string{"testdata/variables/set-foo-too-wee.hcl"}},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
				InputVariables: Variables{
					"foo": &Variable{
						DefaultValue: cty.StringVal("bar"),
						Name:         "foo",
						VarfileValue: cty.StringVal("wee"),
					},
				},
			},
			false, false,
			[]packer.Build{},
			false,
		},

		{"unknown variable from var-file",
			defaultParser,
			parseTestArgs{"testdata/variables/empty.pkr.hcl", nil, []string{"testdata/variables/set-foo-too-wee.hcl"}},
			&PackerConfig{
				Basedir: filepath.Join("testdata", "variables"),
			},
			true, false,
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
				DefaultValue: cty.StringVal("default_value"),
				Type:         cty.String,
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
					Type:         cty.String,
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

		{name: "quoted string",
			variables: Variables{"quoted_string": &Variable{
				DefaultValue: cty.StringVal(`"default_value"`),
				Type:         cty.String,
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
					Type:         cty.String,
					CmdValue:     cty.StringVal(`"cmd_value"`),
					VarfileValue: cty.StringVal(`"varfile_value"`),
					EnvValue:     cty.StringVal(`"env_value"`),
					DefaultValue: cty.StringVal(`"default_value"`),
				},
			},
			wantValues: map[string]cty.Value{
				"quoted_string": cty.StringVal(`"cmd_value"`),
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

		{name: "bool",
			variables: Variables{"enabled": &Variable{
				DefaultValue: cty.False,
				Type:         cty.Bool,
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
					Type:         cty.Bool,
					CmdValue:     cty.True,
					VarfileValue: cty.False,
					EnvValue:     cty.True,
					DefaultValue: cty.False,
				},
			},
			wantValues: map[string]cty.Value{
				"enabled": cty.True,
			},
		},

		{name: "invalid env var",
			variables: Variables{"used_string": &Variable{
				DefaultValue: cty.StringVal("default_value"),
				Type:         cty.String,
			}},
			args: args{
				env: []string{`PKR_VAR_used_string`},
			},

			// output
			wantDiags: false,
			wantVariables: Variables{
				"used_string": &Variable{
					Type:         cty.String,
					DefaultValue: cty.StringVal("default_value"),
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
			wantDiags:         true,
			wantDiagsHasError: false,
			wantVariables:     Variables{},
			wantValues:        map[string]cty.Value{},
		},

		{name: "undefined but set value - pkrvar file - strict mode",
			variables: Variables{},
			validationOptions: ValidationOptions{
				Strict: true,
			},
			args: args{
				hclFiles: []string{`undefined_string="value"`},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: true,
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
					Type:     cty.List(cty.String),
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
			wantDiags:         true,
			wantDiagsHasError: true,
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
					"used_string": `["true"]`,
				},
			},

			// output
			wantDiags:         true,
			wantDiagsHasError: true,
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
				file, diags := parser.ParseHCL([]byte(hclContent), fmt.Sprintf("test_file_%d_*"+hcl2VarFileExt, i))
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

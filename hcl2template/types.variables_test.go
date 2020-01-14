package hcl2template

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"

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
				InputVariables: Variables{
					"image_name": &Variable{},
					"key":        &Variable{},
					"my_secret":  &Variable{},
					"image_id":   &Variable{},
					"port":       &Variable{},
					"availability_zone_names": &Variable{
						Description: fmt.Sprintln("Describing is awesome ;D"),
					},
					"super_secret_password": &Variable{
						Sensible:    true,
						Description: fmt.Sprintln("Handle with care plz"),
					},
				},
				LocalVariables: Variables{
					"owner":        &Variable{},
					"service_name": &Variable{},
				},
			},
			false, false,
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
		want          hcl.Diagnostics
		wantVariables Variables
		wantValues    map[string]cty.Value
	}{
		{"string", Variables{"used_string": &Variable{DefaultValue: cty.StringVal("default_value")}},
			args{
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
			nil,
			Variables{
				"used_string": &Variable{
					CmdValue:     cty.StringVal("cmd_value"),
					VarfileValue: cty.StringVal("varfile_value"),
					EnvValue:     cty.StringVal("env_value"),
					DefaultValue: cty.StringVal("default_value"),
				},
			},
			map[string]cty.Value{
				"used_string": cty.StringVal("cmd_value"),
			},
		},
		{"array of strings", Variables{"used_strings": &Variable{DefaultValue: cty.TupleVal([]cty.Value{cty.StringVal("default_value_1")})}},
			args{
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
			nil,
			Variables{
				"used_strings": &Variable{
					CmdValue:     cty.TupleVal([]cty.Value{cty.StringVal("cmd_value_1")}),
					VarfileValue: cty.TupleVal([]cty.Value{cty.StringVal("varfile_value_1")}),
					EnvValue:     cty.TupleVal([]cty.Value{cty.StringVal("env_value_1"), cty.StringVal("env_value_2")}),
					DefaultValue: cty.TupleVal([]cty.Value{cty.StringVal("default_value_1")}),
				},
			},
			map[string]cty.Value{
				"used_strings": cty.TupleVal([]cty.Value{cty.StringVal("cmd_value_1")}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var files []*hcl.File
			parser := getBasicParser()
			for i, hclContent := range tt.args.hclFiles {
				tmpFile, err := ioutil.TempFile("", fmt.Sprintf("test_file_%d_*"+hcl2VarFileExt, i))
				if err != nil {
					t.Fatal("ioutil.TempFile", err)
				}
				_, err = tmpFile.Write([]byte(hclContent))
				if err != nil {
					t.Fatal("tmpFile.Write", err)
				}
				tmpFileName := tmpFile.Name()
				tmpFile.Close()
				file, diags := parser.ParseHCLFile(tmpFileName)
				if diags != nil {
					t.Fatalf("ParseHCLFile %d: %v", i, diags)
				}
				files = append(files, file)
				defer os.Remove(tmpFileName)
			}
			if got := tt.variables.collectVariableValues(tt.args.env, files, tt.args.argv); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Variables.collectVariableValues() = %v, want %v", got, tt.want)
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

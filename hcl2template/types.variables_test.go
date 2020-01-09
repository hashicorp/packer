package hcl2template

import (
	"testing"

	"github.com/hashicorp/packer/packer"
	"github.com/zclconf/go-cty/cty"
)

func TestParse_variables(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic variables",
			defaultParser,
			parseTestArgs{"testdata/variables/basic.pkr.hcl"},
			&PackerConfig{
				Variables: PackerV1Variables{
					"image_name": cty.StringVal("foo-image-{{user `my_secret`}}"),
					"key":        cty.StringVal("value"),
					"my_secret":  cty.StringVal("foo"),
				},
			},
			false, false,
			[]packer.Build{},
			false,
		},
	}
	testParse(t, tests)
}

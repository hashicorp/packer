package hcl2template

import (
	"testing"

	"github.com/hashicorp/packer/packer"
)

func TestParse_variables(t *testing.T) {
	defaultParser := getBasicParser()

	tests := []parseTest{
		{"basic variables",
			defaultParser,
			parseTestArgs{"testdata/variables/basic.pkr.hcl"},
			&PackerConfig{
				InputVariables: InputVariables{
					"image_name":              InputVariable{},
					"key":                     InputVariable{},
					"my_secret":               InputVariable{},
					"image_id":                InputVariable{},
					"port":                    InputVariable{},
					"availability_zone_names": InputVariable{},
				},
			},
			false, false,
			[]packer.Build{},
			false,
		},
	}
	testParse(t, tests)
}

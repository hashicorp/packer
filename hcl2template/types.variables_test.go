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
				InputVariables: Variables{
					"image_name": Variable{},
					"key":        Variable{},
					"my_secret":  Variable{},
					"image_id":   Variable{},
					"port":       Variable{},
					"availability_zone_names": Variable{
						Description: "Describing is awesome ;D\n",
					},
				},
				LocalVariables: Variables{
					"owner":        Variable{},
					"service_name": Variable{},
				},
			},
			false, false,
			[]packer.Build{},
			false,
		},
	}
	testParse(t, tests)
}

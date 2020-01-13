package hcl2template

import (
	"fmt"
	"testing"

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

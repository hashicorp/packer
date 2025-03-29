// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type DatasourceOutput,Config
package password

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/zclconf/go-cty/cty"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	// The desired length of password.
	Length int64 `mapstructure:"length" required:"true"`
	// Include special characters in the result. These are `!@#$%&*()-_=+[]{}<>:?`. Default value is true
	Special bool `mapstructure:"special" required:"false"`
	// Include uppercase alphabet characters in the result. Default value is true
	Upper bool `mapstructure:"upper" required:"false"`
	// Include lowercase alphabet characters in the result. Default value is true
	Lower bool `mapstructure:"lower" required:"false"`
	// Include numeric characters in the result. Default value is true
	Numeric bool `mapstructure:"numeric" required:"false"`
	// Minimum number of numeric characters in the result. Default value is 0
	MinNumeric int64 `mapstructure:"min_numeric" required:"false"`
	// Minimum number of uppercase alphabet characters in the result. Default value is 0
	MinUpper int64 `mapstructure:"min_upper" required:"false"`
	// Minimum number of lowercase alphabet characters in the result. Default value is 0
	MinLower int64 `mapstructure:"min_lower" required:"false"`
	// Minimum number of special characters in the result. Default value is 0
	MinSpecial int64 `mapstructure:"min_special" required:"false"`
	// Supply your own list of special characters to use for string generation.
	// This overrides the default character list in the special argument.
	// The `special` argument must still be set to true for any overwritten characters to be used in generation.
	OverrideSpecial string `mapstructure:"override_special" required:"false"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	// The result of the password generation. This is the final password string.
	Result string `mapstructure:"result"`
	// A bcrypt hash of the generated random string
	// **NOTE**: If the generated random string is greater than 72 bytes in length,
	// `bcrypt_hash` will contain a hash of the first 72 bytes
	BcryptHash string `mapstructure:"bcrypt_hash"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {

	d.config = fetchDefaultPasswordParameters()

	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	var errs *packersdk.MultiError

	if d.config.Length < 1 ||
		d.config.Length < (d.config.MinLower+d.config.MinUpper+d.config.MinNumeric+d.config.MinSpecial) {

		errs = packersdk.MultiErrorAppend(
			errs,
			fmt.Errorf("the minimum value for length is 1 and, "+
				"length must also be >= (`min_upper` + `min_lower` + `min_numeric` + `min_special`)"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	params := StringParams{
		Length:          d.config.Length,
		Upper:           d.config.Upper,
		MinUpper:        d.config.MinUpper,
		Lower:           d.config.Lower,
		MinLower:        d.config.MinLower,
		Numeric:         d.config.Numeric,
		MinNumeric:      d.config.MinNumeric,
		Special:         d.config.Special,
		MinSpecial:      d.config.MinSpecial,
		OverrideSpecial: d.config.OverrideSpecial,
	}

	result, err := CreateString(params)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf(
			"while attempting to generate a random value a read error was generated: %s",
			err.Error(),
		)
	}

	hash, err := generateHash(result)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf(
			"while attempting to generate a hash from the password an error occurred: %s",
			err.Error(),
		)
	}

	output := DatasourceOutput{
		Result:     result,
		BcryptHash: hash,
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

func fetchDefaultPasswordParameters() Config {
	return Config{
		Special:    true,
		Lower:      true,
		Upper:      true,
		Numeric:    true,
		MinLower:   int64(0),
		MinUpper:   int64(0),
		MinNumeric: int64(0),
		MinSpecial: int64(0),
	}
}

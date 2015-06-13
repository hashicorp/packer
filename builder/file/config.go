package file

import (
	"fmt"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Filename string `mapstructure:"filename"`
	Contents string `mapstructure:"contents"`
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	warnings := []string{}

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return nil, warnings, err
	}

	var errs *packer.MultiError

	if c.Filename == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("filename is required"))
	}

	if c.Contents == "" {
		warnings = append(warnings, "contents is empty")
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}

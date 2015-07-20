package file

import (
	"fmt"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

var ErrTargetRequired = fmt.Errorf("target required")
var ErrContentSourceConflict = fmt.Errorf("Cannot specify source file AND content")

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Source  string `mapstructure:"source"`
	Target  string `mapstructure:"target"`
	Content string `mapstructure:"content"`
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

	if c.Target == "" {
		errs = packer.MultiErrorAppend(errs, ErrTargetRequired)
	}

	if c.Content == "" && c.Source == "" {
		warnings = append(warnings, "Both source file and contents are blank; target will have no content")
	}

	if c.Content != "" && c.Source != "" {
		errs = packer.MultiErrorAppend(errs, ErrContentSourceConflict)
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}

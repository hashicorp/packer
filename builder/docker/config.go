package docker

import (
	"fmt"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/packer"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	ExportPath string `mapstructure:"export_path"`
	Image      string

	tpl *packer.ConfigTemplate
}

func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := new(Config)
	md, err := common.DecodeConfig(c, raws...)
	if err != nil {
		return nil, nil, err
	}

	c.tpl, err = packer.NewConfigTemplate()
	if err != nil {
		return nil, nil, err
	}

	errs := common.CheckUnusedConfig(md)

	templates := map[string]*string{
		"export_path": &c.ExportPath,
		"image":       &c.Image,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.ExportPath == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("export_path must be specified"))
	}

	if c.Image == "" {
		errs = packer.MultiErrorAppend(errs,
			fmt.Errorf("image must be specified"))
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return c, nil, nil
}

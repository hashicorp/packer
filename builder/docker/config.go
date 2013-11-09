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

func (c *Config) Prepare() ([]string, []error) {
	errs := make([]error, 0)

	templates := map[string]*string{
		"export_path": &c.ExportPath,
		"image":       &c.Image,
	}

	for n, ptr := range templates {
		var err error
		*ptr, err = c.tpl.Process(*ptr, nil)
		if err != nil {
			errs = append(
				errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if c.ExportPath == "" {
		errs = append(errs, fmt.Errorf("export_path must be specified"))
	}

	if c.Image == "" {
		errs = append(errs, fmt.Errorf("image must be specified"))
	}

	return nil, errs
}

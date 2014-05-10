package common

import (
	"fmt"
	"text/template"

	"github.com/mitchellh/packer/packer"
)

type ToolsConfig struct {
	ToolsUploadFlavor string `mapstructure:"tools_upload_flavor"`
	ToolsUploadPath   string `mapstructure:"tools_upload_path"`
}

func (c *ToolsConfig) Prepare(t *packer.ConfigTemplate) []error {
	if c.ToolsUploadPath == "" {
		c.ToolsUploadPath = "{{ .Flavor }}.iso"
	}

	templates := map[string]*string{
		"tools_upload_flavor": &c.ToolsUploadFlavor,
	}

	var err error
	errs := make([]error, 0)
	for n, ptr := range templates {
		*ptr, err = t.Process(*ptr, nil)
		if err != nil {
			errs = append(errs, fmt.Errorf("Error processing %s: %s", n, err))
		}
	}

	if _, err := template.New("path").Parse(c.ToolsUploadPath); err != nil {
		errs = append(errs, fmt.Errorf("tools_upload_path invalid: %s", err))
	}

	return errs
}

package common

import (
	"github.com/mitchellh/packer/template/interpolate"
)

type ToolsConfig struct {
	ToolsUploadFlavor string `mapstructure:"tools_upload_flavor"`
	ToolsUploadPath   string `mapstructure:"tools_upload_path"`
}

func (c *ToolsConfig) Prepare(ctx *interpolate.Context) []error {
	if c.ToolsUploadPath == "" {
		c.ToolsUploadPath = "{{ .Flavor }}.iso"
	}

	return nil
}

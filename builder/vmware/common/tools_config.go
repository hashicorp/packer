package common

import (
	"github.com/hashicorp/packer/template/interpolate"
)

type ToolsConfig struct {
	// The flavor of the VMware Tools ISO to
    // upload into the VM. Valid values are darwin, linux, and windows. By
    // default, this is empty, which means VMware tools won't be uploaded.
	ToolsUploadFlavor string `mapstructure:"tools_upload_flavor" required:"false"`
	// The path in the VM to upload the
    // VMware tools. This only takes effect if tools_upload_flavor is non-empty.
    // This is a configuration
    // template that has a single
    // valid variable: Flavor, which will be the value of tools_upload_flavor.
    // By default the upload path is set to {{.Flavor}}.iso. This setting is not
    // used when remote_type is esx5.
	ToolsUploadPath   string `mapstructure:"tools_upload_path" required:"false"`
}

func (c *ToolsConfig) Prepare(ctx *interpolate.Context) []error {
	if c.ToolsUploadPath == "" {
		c.ToolsUploadPath = "{{ .Flavor }}.iso"
	}

	return nil
}

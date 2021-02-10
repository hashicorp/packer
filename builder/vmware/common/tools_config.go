//go:generate struct-markdown

package common

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type ToolsConfig struct {
	// The flavor of the VMware Tools ISO to
	// upload into the VM. Valid values are darwin, linux, and windows. By
	// default, this is empty, which means VMware tools won't be uploaded.
	ToolsUploadFlavor string `mapstructure:"tools_upload_flavor" required:"false"`
	// The path in the VM to upload the VMware tools. This only takes effect if
	// `tools_upload_flavor` is non-empty. This is a [configuration
	// template](/docs/templates/legacy_json_templates/engine) that has a single valid variable:
	// `Flavor`, which will be the value of `tools_upload_flavor`. By default
	// the upload path is set to `{{.Flavor}}.iso`. This setting is not used
	// when `remote_type` is `esx5`.
	ToolsUploadPath string `mapstructure:"tools_upload_path" required:"false"`
	// The path on your local machine to fetch the vmware tools from. If this
	// is not set but the tools_upload_flavor is set, then Packer will try to
	// load the VMWare tools from the VMWare installation directory.
	ToolsSourcePath string `mapstructure:"tools_source_path" required:"false"`
}

func (c *ToolsConfig) Prepare(ctx *interpolate.Context) []error {
	errs := []error{}
	if c.ToolsUploadPath == "" {
		if c.ToolsSourcePath != "" && c.ToolsUploadFlavor == "" {
			errs = append(errs, fmt.Errorf("If you provide a "+
				"tools_source_path, you must also provide either a "+
				"tools_upload_flavor or a tools_upload_path."))
		}
		c.ToolsUploadPath = "{{ .Flavor }}.iso"
	}

	return errs
}

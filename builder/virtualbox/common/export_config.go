//go:generate struct-markdown

package common

import (
	"errors"

	"github.com/hashicorp/packer/template/interpolate"
)

type ExportConfig struct {
	// Either ovf or ova, this specifies the output format
	// of the exported virtual machine. This defaults to ovf.
	Format string `mapstructure:"format" required:"false"`
	// Additional options to pass to the [VBoxManage
	// export](https://www.virtualbox.org/manual/ch09.html#vboxmanage-export).
	// This can be useful for passing product information to include in the
	// resulting appliance file. Packer JSON configuration file example:
	//
	// ```json
	// {
	//   "type": "virtualbox-iso",
	//   "export_opts":
	//   [
	//     "--manifest",
	//     "--vsys", "0",
	//     "--description", "{{user `vm_description`}}",
	//     "--version", "{{user `vm_version`}}"
	//   ],
	//   "format": "ova",
	// }
	// ```
	//
	// A VirtualBox [VM
	// description](https://www.virtualbox.org/manual/ch09.html#vboxmanage-export-ovf)
	// may contain arbitrary strings; the GUI interprets HTML formatting. However,
	// the JSON format does not allow arbitrary newlines within a value. Add a
	// multi-line description by preparing the string in the shell before the
	// packer call like this (shell `>` continuation character snipped for easier
	// copy & paste):
	//
	// ``` {.shell}
	//
	// vm_description='some
	// multiline
	// description'
	//
	// vm_version='0.2.0'
	//
	// packer build \
	//     -var "vm_description=${vm_description}" \
	//     -var "vm_version=${vm_version}"         \
	//     "packer_conf.json"
	// ```
	ExportOpts []string `mapstructure:"export_opts" required:"false"`
}

func (c *ExportConfig) Prepare(ctx *interpolate.Context) []error {
	if c.Format == "" {
		c.Format = "ovf"
	}

	var errs []error
	if c.Format != "ovf" && c.Format != "ova" {
		errs = append(errs,
			errors.New("invalid format, only 'ovf' or 'ova' are allowed"))
	}

	if c.ExportOpts == nil {
		c.ExportOpts = make([]string, 0)
	}

	return errs
}

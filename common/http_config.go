//go:generate struct-markdown

package common

import (
	"errors"

	"github.com/hashicorp/packer/template/interpolate"
)

// Packer will create an http server serving `http_directory` when it is set, a
// random free port will be selected and the architecture of the directory
// referenced will be available in your builder.
//
// Example usage from a builder:
//
//   `wget http://{{ .HTTPIP }}:{{ .HTTPPort }}/foo/bar/preseed.cfg`
type HTTPConfig struct {
	// Path to a directory to serve using an HTTP server. The files in this
	// directory will be available over HTTP that will be requestable from the
	// virtual machine. This is useful for hosting kickstart files and so on.
	// By default this is an empty string, which means no HTTP server will be
	// started. The address and port of the HTTP server will be available as
	// variables in `boot_command`. This is covered in more detail below.
	HTTPDir string `mapstructure:"http_directory"`
	// These are the minimum and maximum port to use for the HTTP server
	// started to serve the `http_directory`. Because Packer often runs in
	// parallel, Packer will choose a randomly available port in this range to
	// run the HTTP server. If you want to force the HTTP server to be on one
	// port, make this minimum and maximum port the same. By default the values
	// are `8000` and `9000`, respectively.
	HTTPPortMin int `mapstructure:"http_port_min"`
	HTTPPortMax int `mapstructure:"http_port_max"`
}

func (c *HTTPConfig) Prepare(ctx *interpolate.Context) []error {
	// Validation
	var errs []error

	if c.HTTPPortMin == 0 {
		c.HTTPPortMin = 8000
	}

	if c.HTTPPortMax == 0 {
		c.HTTPPortMax = 9000
	}

	if c.HTTPPortMin > c.HTTPPortMax {
		errs = append(errs,
			errors.New("http_port_min must be less than http_port_max"))
	}

	return errs
}

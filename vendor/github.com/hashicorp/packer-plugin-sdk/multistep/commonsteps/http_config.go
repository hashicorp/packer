//go:generate struct-markdown

package commonsteps

import (
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
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
	// This is the bind address for the HTTP server. Defaults to 0.0.0.0 so that
	// it will work with any network interface.
	HTTPAddress string `mapstructure:"http_bind_address"`
	// This is the bind interface for the HTTP server. Defaults to the first
	// interface with a non-loopback address. Either `http_bind_address` or
	// `http_interface` can be specified.
	HTTPInterface string `mapstructure:"http_interface" undocumented:"true"`
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

	if c.HTTPInterface != "" && c.HTTPAddress != "" {
		errs = append(errs,
			errors.New("either http_interface or http_bind_address can be specified"))
	}

	if c.HTTPAddress == "" {
		c.HTTPAddress = "0.0.0.0"
	}

	if c.HTTPPortMin > c.HTTPPortMax {
		errs = append(errs,
			errors.New("http_port_min must be less than http_port_max"))
	}

	return errs
}

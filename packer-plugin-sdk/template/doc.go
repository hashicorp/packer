/*
Package template helps plugins parse the Packer template into golang structures.

This package should be imported and used by all plugins. It implements the
golang template engines that Packer documentes on its website, along with input
validation, custom type decoding, and template variable interpolation.

A simple usage example that defines a config and then unpacks a user-provided
json template into the provided config:

  import (
  	// ...
  	"github.com/hashicorp/packer-plugin-sdk/template/config"
  	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
  )

  type Config struct {
  	Field1   string `mapstructure:"field_1"`
  	Field2   bool   `mapstructure:"field_2"`
  	Field3   bool   `mapstructure:"field_3"`

  	ctx interpolate.Context
  }

  type Provisioner struct {
  	config Config
  }

  func (p *CommentProvisioner) Prepare(raws ...interface{}) error {
  	err := config.Decode(&p.config, &config.DecodeOpts{
  		Interpolate:        true,
  		InterpolateContext: &p.config.ctx,
  	}, raws...)
  	if err != nil {
  		return err
  	}

  	return nil
  }

More implementation details for plugins can be found in the
[extending packer](https://www.packer.io/docs/extending) section of the website.
*/
package template

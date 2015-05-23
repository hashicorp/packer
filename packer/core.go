package packer

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/packer/template"
)

// Core is the main executor of Packer. If Packer is being used as a
// library, this is the struct you'll want to instantiate to get anything done.
type Core struct {
	cache      Cache
	components ComponentFinder
	ui         Ui
	template   *template.Template
	variables  map[string]string
}

// CoreConfig is the structure for initializing a new Core. Once a CoreConfig
// is used to initialize a Core, it shouldn't be re-used or modified again.
type CoreConfig struct {
	Cache      Cache
	Components ComponentFinder
	Ui         Ui
	Template   *template.Template
	Variables  map[string]string
}

// NewCore creates a new Core.
func NewCore(c *CoreConfig) (*Core, error) {
	if c.Ui == nil {
		c.Ui = &BasicUi{
			Reader:      os.Stdin,
			Writer:      os.Stdout,
			ErrorWriter: os.Stdout,
		}
	}

	return &Core{
		cache:      c.Cache,
		components: c.Components,
		ui:         c.Ui,
		template:   c.Template,
		variables:  c.Variables,
	}, nil
}

// Build returns the Build object for the given name.
func (c *Core) Build(n string) (Build, error) {
	// Setup the builder
	configBuilder, ok := c.template.Builders[n]
	if !ok {
		return nil, fmt.Errorf("no such build found: %s", n)
	}
	builder, err := c.components.Builder(configBuilder.Type)
	if err != nil {
		return nil, fmt.Errorf(
			"error initializing builder '%s': %s",
			configBuilder.Type, err)
	}
	if builder == nil {
		return nil, fmt.Errorf(
			"builder type not found: %s", configBuilder.Type)
	}

	// TODO: template process name

	return &coreBuild{
		name:          n,
		builder:       builder,
		builderConfig: configBuilder.Config,
		builderType:   configBuilder.Type,
		variables:     c.variables,
	}, nil
}

// Validate does a full validation of the template.
//
// This will automatically call template.Validate() in addition to doing
// richer semantic checks around variables and so on.
func (c *Core) Validate() error {
	// First validate the template in general, we can't do anything else
	// unless the template itself is valid.
	if err := c.template.Validate(); err != nil {
		return err
	}

	// Validate variables are set
	var err error
	for n, v := range c.template.Variables {
		if v.Required {
			if _, ok := c.variables[n]; !ok {
				err = multierror.Append(err, fmt.Errorf(
					"required variable not set: %s", n))
			}
		}
	}

	// TODO: validate all builders exist
	// TODO: ^^ provisioner
	// TODO: ^^ post-processor

	return err
}

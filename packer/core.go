package packer

import (
	"fmt"
	"os"
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/packer/template"
	"github.com/mitchellh/packer/template/interpolate"
)

// Core is the main executor of Packer. If Packer is being used as a
// library, this is the struct you'll want to instantiate to get anything done.
type Core struct {
	cache      Cache
	components ComponentFinder
	ui         Ui
	template   *template.Template
	variables  map[string]string
	builds     map[string]*template.Builder
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

// The function type used to lookup Builder implementations.
type BuilderFunc func(name string) (Builder, error)

// The function type used to lookup Hook implementations.
type HookFunc func(name string) (Hook, error)

// The function type used to lookup PostProcessor implementations.
type PostProcessorFunc func(name string) (PostProcessor, error)

// The function type used to lookup Provisioner implementations.
type ProvisionerFunc func(name string) (Provisioner, error)

// ComponentFinder is a struct that contains the various function
// pointers necessary to look up components of Packer such as builders,
// commands, etc.
type ComponentFinder struct {
	Builder       BuilderFunc
	Hook          HookFunc
	PostProcessor PostProcessorFunc
	Provisioner   ProvisionerFunc
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

	// Go through and interpolate all the build names. We shuld be able
	// to do this at this point with the variables.
	builds := make(map[string]*template.Builder)
	for _, b := range c.Template.Builders {
		v, err := interpolate.Render(b.Name, &interpolate.Context{
			UserVariables: c.Variables,
		})
		if err != nil {
			return nil, fmt.Errorf(
				"Error interpolating builder '%s': %s",
				b.Name, err)
		}

		builds[v] = b
	}

	return &Core{
		cache:      c.Cache,
		components: c.Components,
		ui:         c.Ui,
		template:   c.Template,
		variables:  c.Variables,
		builds:     builds,
	}, nil
}

// BuildNames returns the builds that are available in this configured core.
func (c *Core) BuildNames() []string {
	r := make([]string, 0, len(c.builds))
	for n, _ := range c.builds {
		r = append(r, n)
	}
	sort.Strings(r)

	return r
}

// Build returns the Build object for the given name.
func (c *Core) Build(n string) (Build, error) {
	// Setup the builder
	configBuilder, ok := c.builds[n]
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

	// rawName is the uninterpolated name that we use for various lookups
	rawName := configBuilder.Name

	// Setup the provisioners for this build
	provisioners := make([]coreBuildProvisioner, 0, len(c.template.Provisioners))
	for _, rawP := range c.template.Provisioners {
		// If we're skipping this, then ignore it
		if rawP.Skip(rawName) {
			continue
		}

		// Get the provisioner
		provisioner, err := c.components.Provisioner(rawP.Type)
		if err != nil {
			return nil, fmt.Errorf(
				"error initializing provisioner '%s': %s",
				rawP.Type, err)
		}
		if provisioner == nil {
			return nil, fmt.Errorf(
				"provisioner type not found: %s", rawP.Type)
		}

		// Get the configuration
		config := make([]interface{}, 1, 2)
		config[0] = rawP.Config
		if rawP.Override != nil {
			if override, ok := rawP.Override[rawName]; ok {
				config = append(config, override)
			}
		}

		// If we're pausing, we wrap the provisioner in a special pauser.
		if rawP.PauseBefore > 0 {
			provisioner = &PausedProvisioner{
				PauseBefore: rawP.PauseBefore,
				Provisioner: provisioner,
			}
		}

		provisioners = append(provisioners, coreBuildProvisioner{
			provisioner: provisioner,
			config:      config,
		})
	}

	// Setup the post-processors
	postProcessors := make([][]coreBuildPostProcessor, 0, len(c.template.PostProcessors))
	for _, rawPs := range c.template.PostProcessors {
		current := make([]coreBuildPostProcessor, 0, len(rawPs))
		for _, rawP := range rawPs {
			// If we skip, ignore
			if rawP.Skip(rawName) {
				continue
			}

			// Get the post-processor
			postProcessor, err := c.components.PostProcessor(rawP.Type)
			if err != nil {
				return nil, fmt.Errorf(
					"error initializing post-processor '%s': %s",
					rawP.Type, err)
			}
			if postProcessor == nil {
				return nil, fmt.Errorf(
					"post-processor type not found: %s", rawP.Type)
			}

			current = append(current, coreBuildPostProcessor{
				processor:         postProcessor,
				processorType:     rawP.Type,
				config:            rawP.Config,
				keepInputArtifact: rawP.KeepInputArtifact,
			})
		}

		// If we have no post-processors in this chain, just continue.
		if len(current) == 0 {
			continue
		}

		postProcessors = append(postProcessors, current)
	}

	// TODO hooks one day

	return &coreBuild{
		name:           n,
		builder:        builder,
		builderConfig:  configBuilder.Config,
		builderType:    configBuilder.Type,
		postProcessors: postProcessors,
		provisioners:   provisioners,
		variables:      c.variables,
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

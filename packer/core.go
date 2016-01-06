package packer

import (
	"fmt"
	"sort"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/packer/template"
	"github.com/mitchellh/packer/template/interpolate"
)

// Core is the main executor of Packer. If Packer is being used as a
// library, this is the struct you'll want to instantiate to get anything done.
type Core struct {
	Template *template.Template

	components ComponentFinder
	variables  map[string]string
	builds     map[string]*template.Builder
	version    string
}

// CoreConfig is the structure for initializing a new Core. Once a CoreConfig
// is used to initialize a Core, it shouldn't be re-used or modified again.
type CoreConfig struct {
	Components ComponentFinder
	Template   *template.Template
	Variables  map[string]string
	Version    string
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
	result := &Core{
		Template:   c.Template,
		components: c.Components,
		variables:  c.Variables,
		version:    c.Version,
	}
	if err := result.validate(); err != nil {
		return nil, err
	}
	if err := result.init(); err != nil {
		return nil, err
	}

	// Go through and interpolate all the build names. We shuld be able
	// to do this at this point with the variables.
	result.builds = make(map[string]*template.Builder)
	for _, b := range c.Template.Builders {
		v, err := interpolate.Render(b.Name, result.Context())
		if err != nil {
			return nil, fmt.Errorf(
				"Error interpolating builder '%s': %s",
				b.Name, err)
		}

		result.builds[v] = b
	}

	return result, nil
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
	provisioners := make([]coreBuildProvisioner, 0, len(c.Template.Provisioners))
	for _, rawP := range c.Template.Provisioners {
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
	postProcessors := make([][]coreBuildPostProcessor, 0, len(c.Template.PostProcessors))
	for _, rawPs := range c.Template.PostProcessors {
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
		templatePath:   c.Template.Path,
		variables:      c.variables,
	}, nil
}

// Context returns an interpolation context.
func (c *Core) Context() *interpolate.Context {
	return &interpolate.Context{
		TemplatePath:  c.Template.Path,
		UserVariables: c.variables,
	}
}

// validate does a full validation of the template.
//
// This will automatically call template.validate() in addition to doing
// richer semantic checks around variables and so on.
func (c *Core) validate() error {
	// First validate the template in general, we can't do anything else
	// unless the template itself is valid.
	if err := c.Template.Validate(); err != nil {
		return err
	}

	// Validate the minimum version is satisfied
	if c.Template.MinVersion != "" {
		versionActual, err := version.NewVersion(c.version)
		if err != nil {
			// This shouldn't happen since we set it via the compiler
			panic(err)
		}

		versionMin, err := version.NewVersion(c.Template.MinVersion)
		if err != nil {
			return fmt.Errorf(
				"min_version is invalid: %s", err)
		}

		if versionActual.LessThan(versionMin) {
			return fmt.Errorf(
				"This template requires Packer version %s or higher; using %s",
				versionMin,
				versionActual)
		}
	}

	// Validate variables are set
	var err error
	for n, v := range c.Template.Variables {
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

func (c *Core) init() error {
	if c.variables == nil {
		c.variables = make(map[string]string)
	}

	// Go through the variables and interpolate the environment variables
	ctx := c.Context()
	ctx.EnableEnv = true
	ctx.UserVariables = nil
	for k, v := range c.Template.Variables {
		// Ignore variables that are required
		if v.Required {
			continue
		}

		// Ignore variables that have a value
		if _, ok := c.variables[k]; ok {
			continue
		}

		// Interpolate the default
		def, err := interpolate.Render(v.Default, ctx)
		if err != nil {
			return fmt.Errorf(
				"error interpolating default value for '%s': %s",
				k, err)
		}

		c.variables[k] = def
	}

	// Interpolate the push configuration
	if _, err := interpolate.RenderInterface(&c.Template.Push, c.Context()); err != nil {
		return fmt.Errorf("Error interpolating 'push': %s", err)
	}

	return nil
}

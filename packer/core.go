package packer

import (
	"fmt"
	"sort"
	"strings"

	ttmp "text/template"

	multierror "github.com/hashicorp/go-multierror"
	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/packer/template"
	"github.com/hashicorp/packer/template/interpolate"
)

// Core is the main executor of Packer. If Packer is being used as a
// library, this is the struct you'll want to instantiate to get anything done.
type Core struct {
	Template *template.Template

	components ComponentFinder
	variables  map[string]string
	builds     map[string]*template.Builder
	version    string
	secrets    []string

	except []string
	only   []string
}

// CoreConfig is the structure for initializing a new Core. Once a CoreConfig
// is used to initialize a Core, it shouldn't be re-used or modified again.
type CoreConfig struct {
	Components         ComponentFinder
	Template           *template.Template
	Variables          map[string]string
	SensitiveVariables []string
	Version            string

	// These are set by command-line flags
	Except []string
	Only   []string
}

// The function type used to lookup Builder implementations.
type BuilderFunc func(name string) (Builder, error)

// The function type used to lookup Hook implementations.
type HookFunc func(name string) (Hook, error)

// The function type used to lookup PostProcessor implementations.
type PostProcessorFunc func(name string) (PostProcessor, error)

// The function type used to lookup Provisioner implementations.
type ProvisionerFunc func(name string) (Provisioner, error)

type BasicStore interface {
	Has(name string) bool
	List() (names []string)
}

type BuilderStore interface {
	BasicStore
	Start(name string) (Builder, error)
}

type ProvisionerStore interface {
	BasicStore
	Start(name string) (Provisioner, error)
}

type PostProcessorStore interface {
	BasicStore
	Start(name string) (PostProcessor, error)
}

// ComponentFinder is a struct that contains the various function
// pointers necessary to look up components of Packer such as builders,
// commands, etc.
type ComponentFinder struct {
	Hook HookFunc

	// For HCL2
	BuilderStore       BuilderStore
	ProvisionerStore   ProvisionerStore
	PostProcessorStore PostProcessorStore
}

// NewCore creates a new Core.
func NewCore(c *CoreConfig) (*Core, error) {
	result := &Core{
		Template:   c.Template,
		components: c.Components,
		variables:  c.Variables,
		version:    c.Version,
		only:       c.Only,
		except:     c.Except,
	}

	if err := result.validate(); err != nil {
		return nil, err
	}
	if err := result.init(); err != nil {
		return nil, err
	}
	for _, secret := range result.secrets {
		LogSecretFilter.Set(secret)
	}

	// Go through and interpolate all the build names. We should be able
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
	for n := range c.builds {
		r = append(r, n)
	}
	sort.Strings(r)

	return r
}

func (c *Core) generateCoreBuildProvisioner(rawP *template.Provisioner, rawName string) (CoreBuildProvisioner, error) {
	// Get the provisioner
	cbp := CoreBuildProvisioner{}
	provisioner, err := c.components.ProvisionerStore.Start(rawP.Type)
	if err != nil {
		return cbp, fmt.Errorf(
			"error initializing provisioner '%s': %s",
			rawP.Type, err)
	}
	if provisioner == nil {
		return cbp, fmt.Errorf(
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
	if rawP.PauseBefore != 0 {
		provisioner = &PausedProvisioner{
			PauseBefore: rawP.PauseBefore,
			Provisioner: provisioner,
		}
	} else if rawP.Timeout != 0 {
		provisioner = &TimeoutProvisioner{
			Timeout:     rawP.Timeout,
			Provisioner: provisioner,
		}
	}
	cbp = CoreBuildProvisioner{
		PType:       rawP.Type,
		Provisioner: provisioner,
		config:      config,
	}

	return cbp, nil
}

// Build returns the Build object for the given name.
func (c *Core) Build(n string) (Build, error) {
	// Setup the builder
	configBuilder, ok := c.builds[n]
	if !ok {
		return nil, fmt.Errorf("no such build found: %s", n)
	}
	builder, err := c.components.BuilderStore.Start(configBuilder.Type)
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
	provisioners := make([]CoreBuildProvisioner, 0, len(c.Template.Provisioners))
	for _, rawP := range c.Template.Provisioners {
		// If we're skipping this, then ignore it
		if rawP.OnlyExcept.Skip(rawName) {
			continue
		}
		cbp, err := c.generateCoreBuildProvisioner(rawP, rawName)
		if err != nil {
			return nil, err
		}

		provisioners = append(provisioners, cbp)
	}

	var cleanupProvisioner CoreBuildProvisioner
	if c.Template.CleanupProvisioner != nil {
		// This is a special instantiation of the shell-local provisioner that
		// is only run on error at end of provisioning step before other step
		// cleanup occurs.
		cleanupProvisioner, err = c.generateCoreBuildProvisioner(c.Template.CleanupProvisioner, rawName)
		if err != nil {
			return nil, err
		}
	}

	// Setup the post-processors
	postProcessors := make([][]CoreBuildPostProcessor, 0, len(c.Template.PostProcessors))
	for _, rawPs := range c.Template.PostProcessors {
		current := make([]CoreBuildPostProcessor, 0, len(rawPs))
		for _, rawP := range rawPs {
			if rawP.Skip(rawName) {
				continue
			}
			// -except skips post-processor & build
			foundExcept := false
			for _, except := range c.except {
				if except != "" && except == rawP.Name {
					foundExcept = true
				}
			}
			if foundExcept {
				continue
			}

			// Get the post-processor
			postProcessor, err := c.components.PostProcessorStore.Start(rawP.Type)
			if err != nil {
				return nil, fmt.Errorf(
					"error initializing post-processor '%s': %s",
					rawP.Type, err)
			}
			if postProcessor == nil {
				return nil, fmt.Errorf(
					"post-processor type not found: %s", rawP.Type)
			}

			current = append(current, CoreBuildPostProcessor{
				PostProcessor:     postProcessor,
				PType:             rawP.Type,
				PName:             rawP.Name,
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

	return &CoreBuild{
		Type:               n,
		Builder:            builder,
		BuilderConfig:      configBuilder.Config,
		BuilderType:        configBuilder.Type,
		PostProcessors:     postProcessors,
		Provisioners:       provisioners,
		CleanupProvisioner: cleanupProvisioner,
		TemplatePath:       c.Template.Path,
		Variables:          c.variables,
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
	// Go through the variables and interpolate the environment and
	// user variables

	ctx := c.Context()
	ctx.EnableEnv = true
	ctx.UserVariables = make(map[string]string)
	shouldRetry := true
	changed := false
	failedInterpolation := ""

	// Why this giant loop?  User variables can be recursively defined. For
	// example:
	// "variables": {
	//    	"foo":  "bar",
	//	 	"baz":  "{{user `foo`}}baz",
	// 		"bang": "bang{{user `baz`}}"
	// },
	// In this situation, we cannot guarantee that we've added "foo" to
	// UserVariables before we try to interpolate "baz" the first time. We need
	// to have the option to loop back over in order to add the properly
	// interpolated "baz" to the UserVariables map.
	// Likewise, we'd need to loop up to two times to properly add "bang",
	// since that depends on "baz" being set, which depends on "foo" being set.

	// We break out of the while loop either if all our variables have been
	// interpolated or if after 100 loops we still haven't succeeded in
	// interpolating them.  Please don't actually nest your variables in 100
	// layers of other variables. Please.

	// c.Template.Variables is populated by variables defined within the Template
	// itself
	// c.variables is populated by variables read in from the command line and
	// var-files.
	// We need to read the keys from both, then loop over all of them to figure
	// out the appropriate interpolations.

	allVariables := make(map[string]string)
	// load in template variables
	for k, v := range c.Template.Variables {
		allVariables[k] = v.Default
	}

	// overwrite template variables with command-line-read variables
	for k, v := range c.variables {
		allVariables[k] = v
	}

	// Regex to exclude any build function variable or template variable
	// from interpolating earlier
	// E.g.: {{ .HTTPIP }}  won't interpolate now
	renderFilter := "{{(\\s|)\\.(.*?)(\\s|)}}"

	for i := 0; i < 100; i++ {
		shouldRetry = false
		// First, loop over the variables in the template
		for k, v := range allVariables {
			// Interpolate the default
			renderedV, err := interpolate.RenderRegex(v, ctx, renderFilter)
			switch err.(type) {
			case nil:
				// We only get here if interpolation has succeeded, so something is
				// different in this loop than in the last one.
				changed = true
				c.variables[k] = renderedV
				ctx.UserVariables = c.variables
			case ttmp.ExecError:
				castError := err.(ttmp.ExecError)
				if strings.Contains(castError.Error(), interpolate.ErrVariableNotSetString) {
					shouldRetry = true
					failedInterpolation = fmt.Sprintf(`"%s": "%s"; error: %s`, k, v, err)
				} else {
					return err
				}
			default:
				return fmt.Errorf(
					// unexpected interpolation error: abort the run
					"error interpolating default value for '%s': %s",
					k, err)
			}
		}
		if !shouldRetry {
			break
		}
	}

	if !changed && shouldRetry {
		return fmt.Errorf("Failed to interpolate %s: Please make sure that "+
			"the variable you're referencing has been defined; Packer treats "+
			"all variables used to interpolate other user varaibles as "+
			"required.", failedInterpolation)
	}

	for _, v := range c.Template.SensitiveVariables {
		secret := ctx.UserVariables[v.Key]
		c.secrets = append(c.secrets, secret)
	}

	return nil
}

package packer

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	ttmp "text/template"

	multierror "github.com/hashicorp/go-multierror"
	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl/v2"
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
func (c *Core) BuildNames(only, except []string) []string {

	sort.Strings(only)
	sort.Strings(except)

	r := make([]string, 0, len(c.builds))
	for n := range c.builds {
		onlyPos := sort.SearchStrings(only, n)
		foundInOnly := onlyPos < len(only) && only[onlyPos] == n
		if len(only) > 0 && !foundInOnly {
			continue
		}

		if pos := sort.SearchStrings(except, n); pos < len(except) && except[pos] == n {
			continue
		}
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
	if rawP.MaxRetries != 0 {
		provisioner = &RetriedProvisioner{
			MaxRetries:  rawP.MaxRetries,
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

// This is used for json templates to launch the build plugins.
// They will be prepared via b.Prepare() later.
func (c *Core) GetBuilds(opts GetBuildsOptions) ([]Build, hcl.Diagnostics) {
	buildNames := c.BuildNames(opts.Only, opts.Except)
	builds := []Build{}
	diags := hcl.Diagnostics{}
	for _, n := range buildNames {
		b, err := c.Build(n)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to initialize build %q", n),
				Detail:   err.Error(),
			})
			continue
		}
		builds = append(builds, b)

		// Now that build plugin has been launched, call Prepare()
		log.Printf("Preparing build: %s", b.Name())
		b.SetDebug(opts.Debug)
		b.SetForce(opts.Force)
		b.SetOnError(opts.OnError)

		warnings, err := b.Prepare()
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to prepare build: %q", n),
				Detail:   err.Error(),
			})
			continue
		}

		if len(warnings) > 0 {
			for _, warning := range warnings {
				diags = append(diags, &hcl.Diagnostic{
					Severity: hcl.DiagWarning,
					Summary:  fmt.Sprintf("Warning when preparing build: %q", n),
					Detail:   warning,
				})
			}
		}
	}
	return builds, diags
}

// Build returns the Build object for the given name.
func (c *Core) Build(n string) (Build, error) {
	// Setup the builder
	configBuilder, ok := c.builds[n]
	if !ok {
		return nil, fmt.Errorf("no such build found: %s", n)
	}
	// BuilderStore = config.Builders, gathered in loadConfig() in main.go
	// For reference, the builtin BuilderStore is generated in
	// packer/config.go in the Discover() func.

	// the Start command launches the builder plugin of the given type without
	// calling Prepare() or passing any build-specific details.
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

	// Return a structure that contains the plugins, their types, variables, and
	// the raw builder config loaded from the json template
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

func isDoneInterpolating(v string) (bool, error) {
	// Check for whether the var contains any more references to `user`, wrapped
	// in interpolation syntax.
	filter := `{{\s*user\s*\x60.*\x60\s*}}`
	matched, err := regexp.MatchString(filter, v)
	if err != nil {
		return false, fmt.Errorf("Can't tell if interpolation is done: %s", err)
	}
	if matched {
		// not done interpolating; there's still a call to "user" in a template
		// engine
		return false, nil
	}
	// No more calls to "user" as a template engine, so we're done.
	return true, nil
}

func (c *Core) renderVarsRecursively() (*interpolate.Context, error) {
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

	repeatMap := make(map[string]string)
	allKeys := make([]string, 0)

	// load in template variables
	for k, v := range c.Template.Variables {
		repeatMap[k] = v.Default
		allKeys = append(allKeys, k)
	}

	// overwrite template variables with command-line-read variables
	for k, v := range c.variables {
		repeatMap[k] = v
		allKeys = append(allKeys, k)
	}

	// sort map to force the following loop to be deterministic.
	sort.Strings(allKeys)
	type keyValue struct {
		Key   string
		Value string
	}
	sortedMap := make([]keyValue, len(repeatMap))
	for _, k := range allKeys {
		sortedMap = append(sortedMap, keyValue{k, repeatMap[k]})
	}

	// Regex to exclude any build function variable or template variable
	// from interpolating earlier
	// E.g.: {{ .HTTPIP }}  won't interpolate now
	renderFilter := "{{(\\s|)\\.(.*?)(\\s|)}}"

	for i := 0; i < 100; i++ {
		shouldRetry = false
		changed = false
		deleteKeys := []string{}
		// First, loop over the variables in the template
		for _, kv := range sortedMap {
			// Interpolate the default
			renderedV, err := interpolate.RenderRegex(kv.Value, ctx, renderFilter)
			switch err.(type) {
			case nil:
				// We only get here if interpolation has succeeded, so something is
				// different in this loop than in the last one.
				changed = true
				c.variables[kv.Key] = renderedV
				ctx.UserVariables = c.variables
				// Remove fully-interpolated variables from the map, and flag
				// variables that still need interpolating for a repeat.
				done, err := isDoneInterpolating(kv.Value)
				if err != nil {
					return ctx, err
				}
				if done {
					deleteKeys = append(deleteKeys, kv.Key)
				} else {
					shouldRetry = true
				}
			case ttmp.ExecError:
				castError := err.(ttmp.ExecError)
				if strings.Contains(castError.Error(), interpolate.ErrVariableNotSetString) {
					shouldRetry = true
					failedInterpolation = fmt.Sprintf(`"%s": "%s"; error: %s`, kv.Key, kv.Value, err)
				} else {
					return ctx, err
				}
			default:
				return ctx, fmt.Errorf(
					// unexpected interpolation error: abort the run
					"error interpolating default value for '%s': %s",
					kv.Key, err)
			}
		}
		if !shouldRetry {
			break
		}

		// Clear completed vars from sortedMap before next loop. Do this one
		// key at a time because the indices are gonna change ever time you
		// delete from the map.
		for _, k := range deleteKeys {
			for ind, kv := range sortedMap {
				if kv.Key == k {
					log.Printf("Deleting kv.Value: %s", kv.Value)
					sortedMap = append(sortedMap[:ind], sortedMap[ind+1:]...)
					break
				}
			}
		}
		deleteKeys = []string{}
	}

	if !changed && shouldRetry {
		return ctx, fmt.Errorf("Failed to interpolate %s: Please make sure that "+
			"the variable you're referencing has been defined; Packer treats "+
			"all variables used to interpolate other user varaibles as "+
			"required.", failedInterpolation)
	}

	return ctx, nil
}

func (c *Core) init() error {
	if c.variables == nil {
		c.variables = make(map[string]string)
	}
	// Go through the variables and interpolate the environment and
	// user variables
	ctx, err := c.renderVarsRecursively()
	if err != nil {
		return err
	}
	for _, v := range c.Template.SensitiveVariables {
		secret := ctx.UserVariables[v.Key]
		c.secrets = append(c.secrets, secret)
	}

	return nil
}

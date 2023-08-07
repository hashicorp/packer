package json

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/hashicorp/hcl/v2"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/packer"
)

func (s *JSONSequentialScheduler) EvaluateBuilds() hcl.Diagnostics {
	var diags hcl.Diagnostics

	// Go through and interpolate all the build names. We should be able
	// to do this at this point with the variables.
	s.config.Builds = make(map[string]*template.Builder)
	for _, b := range s.config.Template.Builders {
		v, err := interpolate.Render(b.Name, s.config.Context())
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  "Build interpolation failure",
				Detail: fmt.Sprintf("Error interpolating builder '%s': %s",
					b.Name, err),
			})
		}

		s.config.Builds[v] = b
	}

	return diags
}

// This is used for json templates to launch the build plugins.
// They will be prepared via b.Prepare() later.
func (s *JSONSequentialScheduler) GetBuilds() ([]packersdk.Build, hcl.Diagnostics) {
	buildNames := s.BuildNames()
	builds := []packersdk.Build{}
	diags := hcl.Diagnostics{}
	for _, n := range buildNames {
		b, err := s.Build(n)
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to initialize build %q", n),
				Detail:   err.Error(),
			})
			continue
		}

		// Now that build plugin has been launched, call Prepare()
		log.Printf("Preparing build: %s", b.Name())
		b.SetDebug(s.opts.Debug)
		b.SetForce(s.opts.Force)
		b.SetOnError(s.opts.OnError)

		warnings, err := b.Prepare()
		if err != nil {
			diags = append(diags, &hcl.Diagnostic{
				Severity: hcl.DiagError,
				Summary:  fmt.Sprintf("Failed to prepare build: %q", n),
				Detail:   err.Error(),
			})
			continue
		}

		// Only append builds to list if the Prepare() is successful.
		builds = append(builds, b)

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

// BuildNames returns the builds that are available in this configured core.
func (s *JSONSequentialScheduler) BuildNames() []string {

	only := s.opts.Only
	except := s.opts.Except

	sort.Strings(only)
	sort.Strings(except)
	s.config.Except = except
	s.config.Only = only

	r := make([]string, 0, len(s.config.Builds))
	for n := range s.config.Builds {
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

func (s *JSONSequentialScheduler) generateCoreBuildProvisioner(rawP *template.Provisioner, rawName string) (packer.CoreBuildProvisioner, error) {
	// Get the provisioner
	cbp := packer.CoreBuildProvisioner{}
	provisioner, err := s.config.Components.PluginConfig.Provisioners.Start(rawP.Type)
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
		provisioner = &packer.PausedProvisioner{
			PauseBefore: rawP.PauseBefore,
			Provisioner: provisioner,
		}
	} else if rawP.Timeout != 0 {
		provisioner = &packer.TimeoutProvisioner{
			Timeout:     rawP.Timeout,
			Provisioner: provisioner,
		}
	}
	maxRetries := 0
	if rawP.MaxRetries != "" {
		renderedMaxRetries, err := interpolate.Render(rawP.MaxRetries, s.config.Context())
		if err != nil {
			return cbp, fmt.Errorf("failed to interpolate `max_retries`: %s", err.Error())
		}
		maxRetries, err = strconv.Atoi(renderedMaxRetries)
		if err != nil {
			return cbp, fmt.Errorf("`max_retries` must be a valid integer: %s", err.Error())
		}
	}
	if maxRetries != 0 {
		provisioner = &packer.RetriedProvisioner{
			MaxRetries:  maxRetries,
			Provisioner: provisioner,
		}
	}
	cbp = packer.CoreBuildProvisioner{
		PType:       rawP.Type,
		Provisioner: provisioner,
		Config:      config,
	}

	return cbp, nil
}

// Build returns the Build object for the given name.
func (s *JSONSequentialScheduler) Build(n string) (packersdk.Build, error) {
	// Setup the builder
	configBuilder, ok := s.config.Builds[n]
	if !ok {
		return nil, fmt.Errorf("no such build found: %s", n)
	}
	// BuilderStore = config.Builders, gathered in loadConfig() in main.go
	// For reference, the builtin BuilderStore is generated in
	// packer/config.go in the Discover() func.

	// the Start command launches the builder plugin of the given type without
	// calling Prepare() or passing any build-specific details.
	builder, err := s.config.Components.PluginConfig.Builders.Start(configBuilder.Type)
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
	provisioners := make([]packer.CoreBuildProvisioner, 0, len(s.config.Template.Provisioners))
	for _, rawP := range s.config.Template.Provisioners {
		// If we're skipping this, then ignore it
		if rawP.OnlyExcept.Skip(rawName) {
			continue
		}
		cbp, err := s.generateCoreBuildProvisioner(rawP, rawName)
		if err != nil {
			return nil, err
		}

		provisioners = append(provisioners, cbp)
	}

	var cleanupProvisioner packer.CoreBuildProvisioner
	if s.config.Template.CleanupProvisioner != nil {
		// This is a special instantiation of the shell-local provisioner that
		// is only run on error at end of provisioning step before other step
		// cleanup occurs.
		cleanupProvisioner, err = s.generateCoreBuildProvisioner(s.config.Template.CleanupProvisioner, rawName)
		if err != nil {
			return nil, err
		}
	}

	// Setup the post-processors
	postProcessors := make([][]packer.CoreBuildPostProcessor, 0, len(s.config.Template.PostProcessors))
	for _, rawPs := range s.config.Template.PostProcessors {
		current := make([]packer.CoreBuildPostProcessor, 0, len(rawPs))
		for _, rawP := range rawPs {
			if rawP.Skip(rawName) {
				continue
			}
			// -except skips post-processor & build
			foundExcept := false
			for _, except := range s.config.Except {
				if except != "" && except == rawP.Name {
					foundExcept = true
				}
			}
			if foundExcept {
				break
			}

			// Get the post-processor
			postProcessor, err := s.config.Components.PluginConfig.PostProcessors.Start(rawP.Type)
			if err != nil {
				return nil, fmt.Errorf(
					"error initializing post-processor '%s': %s",
					rawP.Type, err)
			}
			if postProcessor == nil {
				return nil, fmt.Errorf(
					"post-processor type not found: %s", rawP.Type)
			}

			current = append(current, packer.CoreBuildPostProcessor{
				PostProcessor:     postProcessor,
				PType:             rawP.Type,
				PName:             rawP.Name,
				Config:            rawP.Config,
				KeepInputArtifact: rawP.KeepInputArtifact,
			})
		}

		// If we have no post-processors in this chain, just continue.
		if len(current) == 0 {
			continue
		}

		postProcessors = append(postProcessors, current)
	}

	// Return a structure that contains the plugins, their types, variables, and
	// the raw builder config loaded from the json template
	cb := &packer.CoreBuild{
		Type:               n,
		Builder:            builder,
		BuilderConfig:      configBuilder.Config,
		BuilderType:        configBuilder.Type,
		PostProcessors:     postProcessors,
		Provisioners:       provisioners,
		CleanupProvisioner: cleanupProvisioner,
		TemplatePath:       s.config.Template.Path,
		Variables:          s.config.Variables,
	}

	//configBuilder.Name is left uninterpolated so we must check against
	// the interpolated name.
	if configBuilder.Type != configBuilder.Name {
		cb.BuildName = configBuilder.Type
	}

	return cb, nil
}

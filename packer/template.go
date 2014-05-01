package packer

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/go-version"
	"github.com/mitchellh/mapstructure"
	jsonutil "github.com/mitchellh/packer/common/json"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"text/template"
	"time"
)

// The rawTemplate struct represents the structure of a template read
// directly from a file. The builders and other components map just to
// "interface{}" pointers since we actually don't know what their contents
// are until we read the "type" field.
type rawTemplate struct {
	MinimumPackerVersion string `mapstructure:"min_packer_version"`

	Description    string
	Builders       []map[string]interface{}
	Hooks          map[string][]string
	PostProcessors []interface{} `mapstructure:"post-processors"`
	Provisioners   []map[string]interface{}
	Variables      map[string]interface{}
}

// The Template struct represents a parsed template, parsed into the most
// completed form it can be without additional processing by the caller.
type Template struct {
	Description    string
	Variables      map[string]RawVariable
	Builders       map[string]RawBuilderConfig
	Hooks          map[string][]string
	PostProcessors [][]RawPostProcessorConfig
	Provisioners   []RawProvisionerConfig
}

// The RawBuilderConfig struct represents a raw, unprocessed builder
// configuration. It contains the name of the builder as well as the
// raw configuration. If requested, this is used to compile into a full
// builder configuration at some point.
type RawBuilderConfig struct {
	Name string
	Type string

	RawConfig interface{}
}

// RawPostProcessorConfig represents a raw, unprocessed post-processor
// configuration. It contains the type of the post processor as well as the
// raw configuration that is handed to the post-processor for it to process.
type RawPostProcessorConfig struct {
	TemplateOnlyExcept `mapstructure:",squash"`

	Type              string
	KeepInputArtifact bool `mapstructure:"keep_input_artifact"`
	RawConfig         map[string]interface{}
}

// RawProvisionerConfig represents a raw, unprocessed provisioner configuration.
// It contains the type of the provisioner as well as the raw configuration
// that is handed to the provisioner for it to process.
type RawProvisionerConfig struct {
	TemplateOnlyExcept `mapstructure:",squash"`

	Type           string
	Override       map[string]interface{}
	RawPauseBefore string `mapstructure:"pause_before"`

	RawConfig interface{}

	pauseBefore time.Duration
}

// RawVariable represents a variable configuration within a template.
type RawVariable struct {
	Default  string // The default value for this variable
	Required bool   // If the variable is required or not
	Value    string // The set value for this variable
	HasValue bool   // True if the value was set
}

// ParseTemplate takes a byte slice and parses a Template from it, returning
// the template and possibly errors while loading the template. The error
// could potentially be a MultiError, representing multiple errors. Knowing
// and checking for this can be useful, if you wish to format it in a certain
// way.
//
// The second parameter, vars, are the values for a set of user variables.
func ParseTemplate(data []byte, vars map[string]string) (t *Template, err error) {
	var rawTplInterface interface{}
	err = jsonutil.Unmarshal(data, &rawTplInterface)
	if err != nil {
		return
	}

	// Decode the raw template interface into the actual rawTemplate
	// structure, checking for any extranneous keys along the way.
	var md mapstructure.Metadata
	var rawTpl rawTemplate
	decoderConfig := &mapstructure.DecoderConfig{
		Metadata: &md,
		Result:   &rawTpl,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return
	}

	err = decoder.Decode(rawTplInterface)
	if err != nil {
		return
	}

	if rawTpl.MinimumPackerVersion != "" {
		vCur, err := version.NewVersion(Version)
		if err != nil {
			panic(err)
		}
		vReq, err := version.NewVersion(rawTpl.MinimumPackerVersion)
		if err != nil {
			return nil, fmt.Errorf(
				"'minimum_packer_version' error: %s", err)
		}

		if vCur.LessThan(vReq) {
			return nil, fmt.Errorf(
				"Template requires Packer version %s. "+
					"Running version is %s.",
				vReq, vCur)
		}
	}

	errors := make([]error, 0)

	if len(md.Unused) > 0 {
		sort.Strings(md.Unused)
		for _, unused := range md.Unused {
			errors = append(
				errors, fmt.Errorf("Unknown root level key in template: '%s'", unused))
		}
	}

	t = &Template{}
	t.Description = rawTpl.Description
	t.Variables = make(map[string]RawVariable)
	t.Builders = make(map[string]RawBuilderConfig)
	t.Hooks = rawTpl.Hooks
	t.PostProcessors = make([][]RawPostProcessorConfig, len(rawTpl.PostProcessors))
	t.Provisioners = make([]RawProvisionerConfig, len(rawTpl.Provisioners))

	// Gather all the variables
	for k, v := range rawTpl.Variables {
		var variable RawVariable
		variable.Required = v == nil

		// Create a new mapstructure decoder in order to decode the default
		// value since this is the only value in the regular template that
		// can be weakly typed.
		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Result:           &variable.Default,
			WeaklyTypedInput: true,
		})
		if err != nil {
			// This should never happen.
			panic(err)
		}

		err = decoder.Decode(v)
		if err != nil {
			errors = append(errors,
				fmt.Errorf("Error decoding default value for user var '%s': %s", k, err))
			continue
		}

		// Set the value of this variable if we have it
		if val, ok := vars[k]; ok {
			variable.HasValue = true
			variable.Value = val
			delete(vars, k)
		}

		t.Variables[k] = variable
	}

	// Gather all the builders
	for i, v := range rawTpl.Builders {
		var raw RawBuilderConfig
		if err := mapstructure.Decode(v, &raw); err != nil {
			if merr, ok := err.(*mapstructure.Error); ok {
				for _, err := range merr.Errors {
					errors = append(errors, fmt.Errorf("builder %d: %s", i+1, err))
				}
			} else {
				errors = append(errors, fmt.Errorf("builder %d: %s", i+1, err))
			}

			continue
		}

		if raw.Type == "" {
			errors = append(errors, fmt.Errorf("builder %d: missing 'type'", i+1))
			continue
		}

		// Attempt to get the name of the builder. If the "name" key
		// missing, use the "type" field, which is guaranteed to exist
		// at this point.
		if raw.Name == "" {
			raw.Name = raw.Type
		}

		// Check if we already have a builder with this name and error if so
		if _, ok := t.Builders[raw.Name]; ok {
			errors = append(errors, fmt.Errorf("builder with name '%s' already exists", raw.Name))
			continue
		}

		// Now that we have the name, remove it from the config - as the builder
		// itself doesn't know about, and it will cause a validation error.
		delete(v, "name")

		raw.RawConfig = v

		t.Builders[raw.Name] = raw
	}

	// Gather all the post-processors. This is a complicated process since there
	// are actually three different formats that the user can use to define
	// a post-processor.
	for i, rawV := range rawTpl.PostProcessors {
		rawPP, err := parsePostProcessor(i, rawV)
		if err != nil {
			errors = append(errors, err...)
			continue
		}

		configs := make([]RawPostProcessorConfig, 0, len(rawPP))
		for j, pp := range rawPP {
			var config RawPostProcessorConfig
			if err := mapstructure.Decode(pp, &config); err != nil {
				if merr, ok := err.(*mapstructure.Error); ok {
					for _, err := range merr.Errors {
						errors = append(errors,
							fmt.Errorf("Post-processor #%d.%d: %s", i+1, j+1, err))
					}
				} else {
					errors = append(errors,
						fmt.Errorf("Post-processor %d.%d: %s", i+1, j+1, err))
				}

				continue
			}

			if config.Type == "" {
				errors = append(errors,
					fmt.Errorf("Post-processor %d.%d: missing 'type'", i+1, j+1))
				continue
			}

			// Remove the input keep_input_artifact option
			config.TemplateOnlyExcept.Prune(pp)
			delete(pp, "keep_input_artifact")

			// Verify that the only settings are good
			if errs := config.TemplateOnlyExcept.Validate(t.Builders); len(errs) > 0 {
				for _, err := range errs {
					errors = append(errors,
						fmt.Errorf("Post-processor %d.%d: %s", i+1, j+1, err))
				}

				continue
			}

			config.RawConfig = pp

			// Add it to the list of configs
			configs = append(configs, config)
		}

		t.PostProcessors[i] = configs
	}

	// Gather all the provisioners
	for i, v := range rawTpl.Provisioners {
		raw := &t.Provisioners[i]
		if err := mapstructure.Decode(v, raw); err != nil {
			if merr, ok := err.(*mapstructure.Error); ok {
				for _, err := range merr.Errors {
					errors = append(errors, fmt.Errorf("provisioner %d: %s", i+1, err))
				}
			} else {
				errors = append(errors, fmt.Errorf("provisioner %d: %s", i+1, err))
			}

			continue
		}

		if raw.Type == "" {
			errors = append(errors, fmt.Errorf("provisioner %d: missing 'type'", i+1))
			continue
		}

		// Delete the keys that we used
		raw.TemplateOnlyExcept.Prune(v)
		delete(v, "override")

		// Verify that the override keys exist...
		for name, _ := range raw.Override {
			if _, ok := t.Builders[name]; !ok {
				errors = append(
					errors, fmt.Errorf("provisioner %d: build '%s' not found for override", i+1, name))
			}
		}

		// Verify that the only settings are good
		if errs := raw.TemplateOnlyExcept.Validate(t.Builders); len(errs) > 0 {
			for _, err := range errs {
				errors = append(errors,
					fmt.Errorf("provisioner %d: %s", i+1, err))
			}
		}

		// Setup the pause settings
		if raw.RawPauseBefore != "" {
			duration, err := time.ParseDuration(raw.RawPauseBefore)
			if err != nil {
				errors = append(
					errors, fmt.Errorf(
						"provisioner %d: pause_before invalid: %s",
						i+1, err))
			}

			raw.pauseBefore = duration
		}

		// Remove the pause_before setting if it is there so that we don't
		// get template validation errors later.
		delete(v, "pause_before")

		raw.RawConfig = v
	}

	if len(t.Builders) == 0 {
		errors = append(errors, fmt.Errorf("No builders are defined in the template."))
	}

	// Verify that all the variable sets were for real variables.
	for k, _ := range vars {
		errors = append(errors, fmt.Errorf("Unknown user variables: %s", k))
	}

	// If there were errors, we put it into a MultiError and return
	if len(errors) > 0 {
		err = &MultiError{errors}
		t = nil
		return
	}

	return
}

// ParseTemplateFile takes the given template file and parses it into
// a single template.
func ParseTemplateFile(path string, vars map[string]string) (*Template, error) {
	var data []byte

	if path == "-" {
		// Read from stdin...
		buf := new(bytes.Buffer)
		_, err := io.Copy(buf, os.Stdin)
		if err != nil {
			return nil, err
		}

		data = buf.Bytes()
	} else {
		var err error
		data, err = ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}
	}

	return ParseTemplate(data, vars)
}

func parsePostProcessor(i int, rawV interface{}) (result []map[string]interface{}, errors []error) {
	switch v := rawV.(type) {
	case string:
		result = []map[string]interface{}{
			{"type": v},
		}
	case map[string]interface{}:
		result = []map[string]interface{}{v}
	case []interface{}:
		result = make([]map[string]interface{}, len(v))
		errors = make([]error, 0)
		for j, innerRawV := range v {
			switch innerV := innerRawV.(type) {
			case string:
				result[j] = map[string]interface{}{"type": innerV}
			case map[string]interface{}:
				result[j] = innerV
			case []interface{}:
				errors = append(
					errors,
					fmt.Errorf("Post-processor %d.%d: sequences not allowed to be nested in sequences", i+1, j+1))
			default:
				errors = append(errors, fmt.Errorf("Post-processor %d.%d is in a bad format.", i+1, j+1))
			}
		}

		if len(errors) == 0 {
			errors = nil
		}
	default:
		result = nil
		errors = []error{fmt.Errorf("Post-processor %d is in a bad format.", i+1)}
	}

	return
}

// BuildNames returns a slice of the available names of builds that
// this template represents.
func (t *Template) BuildNames() []string {
	names := make([]string, 0, len(t.Builders))
	for name, _ := range t.Builders {
		names = append(names, name)
	}

	return names
}

// Build returns a Build for the given name.
//
// If the build does not exist as part of this template, an error is
// returned.
func (t *Template) Build(name string, components *ComponentFinder) (b Build, err error) {
	// Setup the Builder
	builderConfig, ok := t.Builders[name]
	if !ok {
		err = fmt.Errorf("No such build found in template: %s", name)
		return
	}

	// We panic if there is no builder function because this is really
	// an internal bug that always needs to be fixed, not an error.
	if components.Builder == nil {
		panic("no builder function")
	}

	// Panic if there are provisioners on the template but no provisioner
	// component finder. This is always an internal error, so we panic.
	if len(t.Provisioners) > 0 && components.Provisioner == nil {
		panic("no provisioner function")
	}

	builder, err := components.Builder(builderConfig.Type)
	if err != nil {
		return
	}

	if builder == nil {
		err = fmt.Errorf("Builder type not found: %s", builderConfig.Type)
		return
	}

	// Prepare the variable template processor, which is a bit unique
	// because we don't allow user variable usage and we add a function
	// to read from the environment.
	varTpl, err := NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	varTpl.Funcs(template.FuncMap{
		"env":  templateEnv,
		"user": templateDisableUser,
	})

	// Prepare the variables
	var varErrors []error
	variables := make(map[string]string)
	for k, v := range t.Variables {
		if v.Required && !v.HasValue {
			varErrors = append(varErrors,
				fmt.Errorf("Required user variable '%s' not set", k))
		}

		var val string
		if v.HasValue {
			val = v.Value
		} else {
			val, err = varTpl.Process(v.Default, nil)
			if err != nil {
				varErrors = append(varErrors,
					fmt.Errorf("Error processing user variable '%s': %s'", k, err))
			}
		}

		variables[k] = val
	}

	if len(varErrors) > 0 {
		return nil, &MultiError{varErrors}
	}

	// Process the name
	tpl, err := NewConfigTemplate()
	if err != nil {
		return nil, err
	}
	tpl.UserVars = variables

	name, err = tpl.Process(name, nil)
	if err != nil {
		return nil, err
	}

	// Gather the Hooks
	hooks := make(map[string][]Hook)
	for tplEvent, tplHooks := range t.Hooks {
		curHooks := make([]Hook, 0, len(tplHooks))

		for _, hookName := range tplHooks {
			var hook Hook
			hook, err = components.Hook(hookName)
			if err != nil {
				return
			}

			if hook == nil {
				err = fmt.Errorf("Hook not found: %s", hookName)
				return
			}

			curHooks = append(curHooks, hook)
		}

		hooks[tplEvent] = curHooks
	}

	// Prepare the post-processors
	postProcessors := make([][]coreBuildPostProcessor, 0, len(t.PostProcessors))
	for _, rawPPs := range t.PostProcessors {
		current := make([]coreBuildPostProcessor, 0, len(rawPPs))
		for _, rawPP := range rawPPs {
			if rawPP.TemplateOnlyExcept.Skip(name) {
				continue
			}

			pp, err := components.PostProcessor(rawPP.Type)
			if err != nil {
				return nil, err
			}

			if pp == nil {
				return nil, fmt.Errorf("PostProcessor type not found: %s", rawPP.Type)
			}

			current = append(current, coreBuildPostProcessor{
				processor:         pp,
				processorType:     rawPP.Type,
				config:            rawPP.RawConfig,
				keepInputArtifact: rawPP.KeepInputArtifact,
			})
		}

		// If we have no post-processors in this chain, just continue.
		// This can happen if the post-processors skip certain builds.
		if len(current) == 0 {
			continue
		}

		postProcessors = append(postProcessors, current)
	}

	// Prepare the provisioners
	provisioners := make([]coreBuildProvisioner, 0, len(t.Provisioners))
	for _, rawProvisioner := range t.Provisioners {
		if rawProvisioner.TemplateOnlyExcept.Skip(name) {
			continue
		}

		var provisioner Provisioner
		provisioner, err = components.Provisioner(rawProvisioner.Type)
		if err != nil {
			return
		}

		if provisioner == nil {
			err = fmt.Errorf("Provisioner type not found: %s", rawProvisioner.Type)
			return
		}

		configs := make([]interface{}, 1, 2)
		configs[0] = rawProvisioner.RawConfig

		if rawProvisioner.Override != nil {
			if override, ok := rawProvisioner.Override[name]; ok {
				configs = append(configs, override)
			}
		}

		if rawProvisioner.pauseBefore > 0 {
			provisioner = &PausedProvisioner{
				PauseBefore: rawProvisioner.pauseBefore,
				Provisioner: provisioner,
			}
		}

		coreProv := coreBuildProvisioner{provisioner, configs}
		provisioners = append(provisioners, coreProv)
	}

	b = &coreBuild{
		name:           name,
		builder:        builder,
		builderConfig:  builderConfig.RawConfig,
		builderType:    builderConfig.Type,
		hooks:          hooks,
		postProcessors: postProcessors,
		provisioners:   provisioners,
		variables:      variables,
	}

	return
}

// TemplateOnlyExcept contains the logic required for "only" and "except"
// meta-parameters.
type TemplateOnlyExcept struct {
	Only   []string
	Except []string
}

// Prune will prune out the used values from the raw map.
func (t *TemplateOnlyExcept) Prune(raw map[string]interface{}) {
	delete(raw, "except")
	delete(raw, "only")
}

// Skip tests if we should skip putting this item onto a build.
func (t *TemplateOnlyExcept) Skip(name string) bool {
	if len(t.Only) > 0 {
		onlyFound := false
		for _, n := range t.Only {
			if n == name {
				onlyFound = true
				break
			}
		}

		if !onlyFound {
			// Skip this provisioner
			return true
		}
	}

	// If the name is in the except list, then skip that
	for _, n := range t.Except {
		if n == name {
			return true
		}
	}

	return false
}

// Validates the only/except parameters.
func (t *TemplateOnlyExcept) Validate(b map[string]RawBuilderConfig) (e []error) {
	if len(t.Only) > 0 && len(t.Except) > 0 {
		e = append(e,
			fmt.Errorf("Only one of 'only' or 'except' may be specified."))
	}

	if len(t.Only) > 0 {
		for _, n := range t.Only {
			if _, ok := b[n]; !ok {
				e = append(e,
					fmt.Errorf("'only' specified builder '%s' not found", n))
			}
		}
	}

	for _, n := range t.Except {
		if _, ok := b[n]; !ok {
			e = append(e,
				fmt.Errorf("'except' specified builder '%s' not found", n))
		}
	}

	return
}

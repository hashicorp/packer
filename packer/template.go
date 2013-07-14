package packer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"sort"
)

// The rawTemplate struct represents the structure of a template read
// directly from a file. The builders and other components map just to
// "interface{}" pointers since we actually don't know what their contents
// are until we read the "type" field.
type rawTemplate struct {
	Builders       []map[string]interface{}
	Hooks          map[string][]string
	Provisioners   []map[string]interface{}
	PostProcessors []interface{} `mapstructure:"post-processors"`
}

// The Template struct represents a parsed template, parsed into the most
// completed form it can be without additional processing by the caller.
type Template struct {
	Builders       map[string]rawBuilderConfig
	Hooks          map[string][]string
	PostProcessors [][]rawPostProcessorConfig
	Provisioners   []rawProvisionerConfig
}

// The rawBuilderConfig struct represents a raw, unprocessed builder
// configuration. It contains the name of the builder as well as the
// raw configuration. If requested, this is used to compile into a full
// builder configuration at some point.
type rawBuilderConfig struct {
	Name string
	Type string

	rawConfig interface{}
}

// rawPostProcessorConfig represents a raw, unprocessed post-processor
// configuration. It contains the type of the post processor as well as the
// raw configuration that is handed to the post-processor for it to process.
type rawPostProcessorConfig struct {
	Type              string
	KeepInputArtifact bool `mapstructure:"keep_input_artifact"`
	rawConfig         interface{}
}

// rawProvisionerConfig represents a raw, unprocessed provisioner configuration.
// It contains the type of the provisioner as well as the raw configuration
// that is handed to the provisioner for it to process.
type rawProvisionerConfig struct {
	Type     string
	Override map[string]interface{}

	rawConfig interface{}
}

// ParseTemplate takes a byte slice and parses a Template from it, returning
// the template and possibly errors while loading the template. The error
// could potentially be a MultiError, representing multiple errors. Knowing
// and checking for this can be useful, if you wish to format it in a certain
// way.
func ParseTemplate(data []byte) (t *Template, err error) {
	var rawTplInterface interface{}
	err = json.Unmarshal(data, &rawTplInterface)
	if err != nil {
		syntaxErr, ok := err.(*json.SyntaxError)
		if !ok {
			return
		}

		// We have a syntax error. Extract out the line number and friends.
		// https://groups.google.com/forum/#!topic/golang-nuts/fizimmXtVfc
		newline := []byte{'\x0a'}

		// Calculate the start/end position of the line where the error is
		start := bytes.LastIndex(data[:syntaxErr.Offset], newline) + 1
		end := len(data)
		if idx := bytes.Index(data[start:], newline); idx >= 0 {
			end = start + idx
		}

		// Count the line number we're on plus the offset in the line
		line := bytes.Count(data[:start], newline) + 1
		pos := int(syntaxErr.Offset) - start - 1

		err = fmt.Errorf("Error in line %d, char %d: %s\n%s",
			line, pos, syntaxErr, data[start:end])

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

	errors := make([]error, 0)

	if len(md.Unused) > 0 {
		sort.Strings(md.Unused)
		for _, unused := range md.Unused {
			errors = append(
				errors, fmt.Errorf("Unknown root level key in template: '%s'", unused))
		}
	}

	t = &Template{}
	t.Builders = make(map[string]rawBuilderConfig)
	t.Hooks = rawTpl.Hooks
	t.PostProcessors = make([][]rawPostProcessorConfig, len(rawTpl.PostProcessors))
	t.Provisioners = make([]rawProvisionerConfig, len(rawTpl.Provisioners))

	// Gather all the builders
	for i, v := range rawTpl.Builders {
		var raw rawBuilderConfig
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

		raw.rawConfig = v

		t.Builders[raw.Name] = raw
	}

	// Gather all the post-processors. This is a complicated process since there
	// are actually three different formats that the user can use to define
	// a post-processor.
	for i, rawV := range rawTpl.PostProcessors {
		rawPP, err := parsePostProvisioner(i, rawV)
		if err != nil {
			errors = append(errors, err...)
			continue
		}

		t.PostProcessors[i] = make([]rawPostProcessorConfig, len(rawPP))
		configs := t.PostProcessors[i]
		for j, pp := range rawPP {
			config := &configs[j]
			if err := mapstructure.Decode(pp, config); err != nil {
				if merr, ok := err.(*mapstructure.Error); ok {
					for _, err := range merr.Errors {
						errors = append(errors, fmt.Errorf("Post-processor #%d.%d: %s", i+1, j+1, err))
					}
				} else {
					errors = append(errors, fmt.Errorf("Post-processor %d.%d: %s", i+1, j+1, err))
				}

				continue
			}

			if config.Type == "" {
				errors = append(errors, fmt.Errorf("Post-processor %d.%d: missing 'type'", i+1, j+1))
				continue
			}

			config.rawConfig = pp
		}
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

		// The provisioners not only don't need or want the override settings
		// (as they are processed as part of the preparation below), but will
		// actively reject them as invalid configuration.
		delete(v, "override")

		raw.rawConfig = v
	}

	if len(t.Builders) == 0 {
		errors = append(errors, fmt.Errorf("No builders are defined in the template."))
	}

	// If there were errors, we put it into a MultiError and return
	if len(errors) > 0 {
		err = &MultiError{errors}
		t = nil
		return
	}

	return
}

func parsePostProvisioner(i int, rawV interface{}) (result []map[string]interface{}, errors []error) {
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
		current := make([]coreBuildPostProcessor, len(rawPPs))
		for i, rawPP := range rawPPs {
			pp, err := components.PostProcessor(rawPP.Type)
			if err != nil {
				return nil, err
			}

			if pp == nil {
				return nil, fmt.Errorf("PostProcessor type not found: %s", rawPP.Type)
			}

			current[i] = coreBuildPostProcessor{
				processor:         pp,
				processorType:     rawPP.Type,
				config:            rawPP.rawConfig,
				keepInputArtifact: rawPP.KeepInputArtifact,
			}
		}

		postProcessors = append(postProcessors, current)
	}

	// Prepare the provisioners
	provisioners := make([]coreBuildProvisioner, 0, len(t.Provisioners))
	for _, rawProvisioner := range t.Provisioners {
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
		configs[0] = rawProvisioner.rawConfig

		if rawProvisioner.Override != nil {
			if override, ok := rawProvisioner.Override[name]; ok {
				configs = append(configs, override)
			}
		}

		coreProv := coreBuildProvisioner{provisioner, configs}
		provisioners = append(provisioners, coreProv)
	}

	b = &coreBuild{
		name:           name,
		builder:        builder,
		builderConfig:  builderConfig.rawConfig,
		builderType:    builderConfig.Type,
		hooks:          hooks,
		postProcessors: postProcessors,
		provisioners:   provisioners,
	}

	return
}

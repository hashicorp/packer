package packer

import (
	"encoding/json"
	"fmt"
)

// The rawTemplate struct represents the structure of a template read
// directly from a file. The builders and other components map just to
// "interface{}" pointers since we actually don't know what their contents
// are until we read the "type" field.
type rawTemplate struct {
	Name         string
	Builders     []map[string]interface{}
	Hooks        map[string][]string
	Provisioners []map[string]interface{}
	Outputs      []map[string]interface{}
}

// The Template struct represents a parsed template, parsed into the most
// completed form it can be without additional processing by the caller.
type Template struct {
	Name         string
	Builders     map[string]rawBuilderConfig
	Hooks        map[string][]string
	Provisioners []rawProvisionerConfig
}

// The rawBuilderConfig struct represents a raw, unprocessed builder
// configuration. It contains the name of the builder as well as the
// raw configuration. If requested, this is used to compile into a full
// builder configuration at some point.
type rawBuilderConfig struct {
	builderType string
	rawConfig   interface{}
}

// rawProvisionerConfig represents a raw, unprocessed provisioner configuration.
// It contains the type of the provisioner as well as the raw configuration
// that is handed to the provisioner for it to process.
type rawProvisionerConfig struct {
	pType     string
	rawConfig interface{}
}

// ParseTemplate takes a byte slice and parses a Template from it, returning
// the template and possibly errors while loading the template. The error
// could potentially be a MultiError, representing multiple errors. Knowing
// and checking for this can be useful, if you wish to format it in a certain
// way.
func ParseTemplate(data []byte) (t *Template, err error) {
	var rawTpl rawTemplate
	err = json.Unmarshal(data, &rawTpl)
	if err != nil {
		return
	}

	t = &Template{}
	t.Name = rawTpl.Name
	t.Builders = make(map[string]rawBuilderConfig)
	t.Hooks = rawTpl.Hooks
	t.Provisioners = make([]rawProvisionerConfig, len(rawTpl.Provisioners))

	errors := make([]error, 0)

	// Gather all the builders
	for i, v := range rawTpl.Builders {
		rawType, ok := v["type"]
		if !ok {
			errors = append(errors, fmt.Errorf("builder %d: missing 'type'", i+1))
			continue
		}

		// Attempt to get the name of the builder. If the "name" key
		// missing, use the "type" field, which is guaranteed to exist
		// at this point.
		rawName, ok := v["name"]
		if !ok {
			rawName = v["type"]
		}

		// Attempt to convert the name/type to strings, but error if we can't
		name, ok := rawName.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("builder %d: name must be a string", i+1))
			continue
		}

		typeName, ok := rawType.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("builder %d: type must be a string", i+1))
			continue
		}

		// Check if we already have a builder with this name and error if so
		if _, ok := t.Builders[name]; ok {
			errors = append(errors, fmt.Errorf("builder with name '%s' already exists", name))
			continue
		}

		t.Builders[name] = rawBuilderConfig{
			typeName,
			v,
		}
	}

	// Gather all the provisioners
	for i, v := range rawTpl.Provisioners {
		rawType, ok := v["type"]
		if !ok {
			errors = append(errors, fmt.Errorf("provisioner %d: missing 'type'", i+1))
			continue
		}

		typeName, ok := rawType.(string)
		if !ok {
			errors = append(errors, fmt.Errorf("provisioner %d: type must be a string", i+1))
			continue
		}

		t.Provisioners[i] = rawProvisionerConfig{
			typeName,
			v,
		}
	}

	// If there were errors, we put it into a MultiError and return
	if len(errors) > 0 {
		err = &MultiError{errors}
		return
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

	if len(t.Provisioners) > 0 && components.Provisioner == nil {
		panic("no provisioner function")
	}

	builder, err := components.Builder(builderConfig.builderType)
	if err != nil {
		return
	}

	if builder == nil {
		err = fmt.Errorf("Builder type not found: %s", builderConfig.builderType)
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

	// Prepare the provisioners
	provisioners := make([]coreBuildProvisioner, 0, len(t.Provisioners))
	for _, rawProvisioner := range t.Provisioners {
		var provisioner Provisioner
		provisioner, err = components.Provisioner(rawProvisioner.pType)
		if err != nil {
			return
		}

		if provisioner == nil {
			err = fmt.Errorf("Provisioner type not found: %s", rawProvisioner.pType)
			return
		}

		coreProv := coreBuildProvisioner{provisioner, rawProvisioner.rawConfig}
		provisioners = append(provisioners, coreProv)
	}

	b = &coreBuild{
		name:          name,
		builder:       builder,
		builderConfig: builderConfig.rawConfig,
		hooks:         hooks,
		provisioners:  provisioners,
	}

	return
}

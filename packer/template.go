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
	Name     string
	Builders map[string]rawBuilderConfig
	Hooks    map[string][]string
}

// The rawBuilderConfig struct represents a raw, unprocessed builder
// configuration. It contains the name of the builder as well as the
// raw configuration. If requested, this is used to compile into a full
// builder configuration at some point.
type rawBuilderConfig struct {
	builderName string
	rawConfig   interface{}
}

// ParseTemplate takes a byte slice and parses a Template from it, returning
// the template and possibly errors while loading the template.
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

	for _, v := range rawTpl.Builders {
		rawType, ok := v["type"]
		if !ok {
			// TODO: Missing type error
			return
		}

		// Attempt to get the name of the builder. If the "name" key
		// missing, use the "type" field, which is guaranteed to exist
		// at this point.
		rawName, ok := v["name"]
		if !ok {
			rawName = v["type"]
		}

		// TODO: Error checking if we can't convert
		name := rawName.(string)
		typeName := rawType.(string)

		// Check if we already have a builder with this name and record
		// an error.
		_, ok = t.Builders[name]
		if ok {
			// TODO: We already have a builder with this name
			return
		}

		t.Builders[name] = rawBuilderConfig{typeName, v}
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
	builderConfig, ok := t.Builders[name]
	if !ok {
		err = fmt.Errorf("No such build found in template: %s", name)
		return
	}

	builder, err := components.Builder(builderConfig.builderName)
	if err != nil {
		return
	}

	if builder == nil {
		err = fmt.Errorf("Builder not found: %s", name)
		return
	}

	b = &coreBuild{
		name:      name,
		builder:   builder,
		rawConfig: builderConfig.rawConfig,
	}

	return
}

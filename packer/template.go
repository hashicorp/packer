package packer

import "encoding/json"

// The rawTemplate struct represents the structure of a template read
// directly from a file. The builders and other components map just to
// "interface{}" pointers since we actually don't know what their contents
// are until we read the "type" field.
type rawTemplate struct {
	Name         string
	Builders     []map[string]interface{}
	Provisioners []map[string]interface{}
	Outputs      []map[string]interface{}
}

type Template struct {
	Name     string
	Builders map[string]rawBuilderConfig
}

// The rawBuilderConfig struct represents a raw, unprocessed builder
// configuration. It contains the name of the builder as well as the
// raw configuration. If requested, this is used to compile into a full
// builder configuration at some point.
type rawBuilderConfig struct {
	builderName string
	rawConfig   interface{}
}

func ParseTemplate(data []byte) (t *Template, err error) {
	var rawTpl rawTemplate
	err = json.Unmarshal(data, &rawTpl)
	if err != nil {
		return
	}

	t = &Template{}
	t.Name = rawTpl.Name
	t.Builders = make(map[string]rawBuilderConfig)

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

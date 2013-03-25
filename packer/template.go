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

func parseTemplate(data []byte) error {
	var rawTpl rawTemplate
	err := json.Unmarshal(data, &rawTpl)
	if err != nil {
		return err
	}

	return nil
}

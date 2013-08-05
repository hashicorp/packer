package fix

import (
	"github.com/mitchellh/mapstructure"
	"regexp"
)

// FixerGlobalTemplates is a Fixer that replaces the "{{.CreateTime}}"
// variable within the snapshot_name of a DigitalOcean builder with the
// new "{{timestamp}}" format.
type FixerGlobalTemplates struct{}

func (FixerGlobalTemplates) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// Our template type we'll use for this fixer only
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each builder and replace the iso_md5 if we can
	for _, builder := range tpl.Builders {
		builderTypeRaw, ok := builder["type"]
		if !ok {
			continue
		}

		builderType, ok := builderTypeRaw.(string)
		if !ok {
			// Non-string "type", odd, ignore.
			continue
		}

		if builderType != "digitalocean" {
			continue
		}

		snapshotNameRaw, ok := builder["snapshot_name"]
		if !ok {
			continue
		}

		snapshotName, ok := snapshotNameRaw.(string)
		if !ok {
			continue
		}

		re := regexp.MustCompile(`(?i){{\.CreateTime}}`)
		builder["snapshot_name"] = re.ReplaceAllString(snapshotName, "{{timestamp}}")
	}

	input["builders"] = tpl.Builders
	return input, nil
}

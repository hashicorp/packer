package fix

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/mitchellh/packer/common"
	"log"
	"regexp"
)

// FixerGlobalTemplates is a Fixer that replaces the "{{.CreateTime}}"
// variable within the snapshot_name of a DigitalOcean builder with the
// new "{{timestamp}}" format.
type FixerGlobalTemplates struct{}

func (f FixerGlobalTemplates) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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
	for i, builder := range tpl.Builders {
		builderTypeRaw, ok := builder["type"]
		if !ok {
			continue
		}

		builderType, ok := builderTypeRaw.(string)
		if !ok {
			// Non-string "type", odd, ignore.
			continue
		}

		// Convert the timestamps in all templates
		timestampRe := regexp.MustCompile(`(?i){{\s*\.CreateTime\s*}}`)
		err := common.TraverseStrings(&builder, func(n string, v string) string {
			return timestampRe.ReplaceAllString(v, "{{timestamp}}")
		})
		if err != nil {
			return nil, err
		}

		// Builder-specific replacements
		log.Printf("Fixing templates in type '%s'", builderType)
		switch builderType {
		case "amazon-chroot":
			builder = f.fixAmazonChroot(builder)
		case "digitalocean":
			builder = f.fixDigitalOcean(builder)
		case "virtualbox":
			builder = f.fixVirtualBox(builder)
		case "vmware":
			builder = f.fixVMware(builder)
		default:
		}

		tpl.Builders[i] = builder
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerGlobalTemplates) fixAmazonChroot(builder map[string]interface{}) map[string]interface{} {
	builderVars := map[string]string{
		"Device": "device",
	}

	err := common.TraverseStrings(&builder, func(n string, v string) string {
		for orig, replacement := range builderVars {
			re := regexp.MustCompile(fmt.Sprintf(`(?i){{\s*\.%s\s*}}`, orig))
			v = re.ReplaceAllString(v, fmt.Sprintf(`{{builder "%s"}}`, replacement))
		}

		return v
	})
	if err != nil {
		panic(err)
	}

	return builder
}

func (FixerGlobalTemplates) fixDigitalOcean(builder map[string]interface{}) map[string]interface{} {
	return builder
}

func (FixerGlobalTemplates) fixVirtualBox(builder map[string]interface{}) map[string]interface{} {
	builderVars := map[string]string{
		"HTTPIP":   "http_ip",
		"HTTPPort": "http_port",
		"Name":     "vm_name",
		"Version":  "vbox_version",
	}

	err := common.TraverseStrings(&builder, func(n string, v string) string {
		for orig, replacement := range builderVars {
			re := regexp.MustCompile(fmt.Sprintf(`(?i){{\s*\.%s\s*}}`, orig))
			v = re.ReplaceAllString(v, fmt.Sprintf(`{{builder "%s"}}`, replacement))
		}

		return v
	})
	if err != nil {
		panic(err)
	}

	return builder
}

func (FixerGlobalTemplates) fixVMware(builder map[string]interface{}) map[string]interface{} {
	builderVars := map[string]string{
		"HTTPIP":   "http_ip",
		"HTTPPort": "http_port",
		"Flavor":   "flavor",
		"Name":     "vm_name",
	}

	err := common.TraverseStrings(&builder, func(n string, v string) string {
		for orig, replacement := range builderVars {
			re := regexp.MustCompile(fmt.Sprintf(`(?i){{\s*\.%s\s*}}`, orig))
			v = re.ReplaceAllString(v, fmt.Sprintf(`{{builder "%s"}}`, replacement))
		}

		return v
	})
	if err != nil {
		panic(err)
	}

	return builder
}

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerISOMD5 is a Fixer that replaces the "iso_md5" configuration key
// with the newer "iso_checksum" and "iso_checksum_type" within builders.
type FixerISOMD5 struct{}

func (FixerISOMD5) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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
		md5raw, ok := builder["iso_md5"]
		if !ok {
			continue
		}

		md5, ok := md5raw.(string)
		if !ok {
			// TODO: error?
			continue
		}

		delete(builder, "iso_md5")
		builder["iso_checksum"] = md5
		builder["iso_checksum_type"] = "md5"
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerISOMD5) Synopsis() string {
	return `Replaces "iso_md5" in builders with "iso_checksum"`
}

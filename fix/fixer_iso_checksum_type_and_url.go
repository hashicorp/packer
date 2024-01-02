// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerISOChecksumTypeAndURL is a Fixer that remove the "iso_checksum_url" and
// "iso_checksum_type" to put everything in the checksum field.
type FixerISOChecksumTypeAndURL struct{}

func (FixerISOChecksumTypeAndURL) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"*": []string{"iso_checksum_url", "iso_checksum_type"},
	}
}

func (FixerISOChecksumTypeAndURL) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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
		checksumUrl := stringValue(builder["iso_checksum_url"])
		checksumType := stringValue(builder["iso_checksum_type"])
		checksum := stringValue(builder["iso_checksum"])
		delete(builder, "iso_checksum_url")
		delete(builder, "iso_checksum_type")
		if checksum == "" && checksumUrl == "" {
			continue
		}
		if checksumUrl != "" {
			checksum = "file:" + checksumUrl
		} else if checksumType != "" {
			checksum = checksumType + ":" + checksum
		}

		builder["iso_checksum"] = checksum
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func stringValue(v interface{}) string {
	switch rfl := v.(type) {
	case string:
		return rfl
	default:
		return ""
	}
}

func (FixerISOChecksumTypeAndURL) Synopsis() string {
	return `Puts content of potential "iso_checksum_url" and "iso_checksum_url" in "iso_checksum"`
}

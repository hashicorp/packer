// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import (
	"strings"

	"github.com/mitchellh/mapstructure"
)

// FixerDockerTagtoTags renames tag to tags
type FixerDockerTagtoTags struct{}

func (FixerDockerTagtoTags) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"packer.post-processor.docker-tag": []string{"tag"},
	}
}

func (FixerDockerTagtoTags) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	if input["post-processors"] == nil {
		return input, nil
	}

	// Our template type we'll use for this fixer only
	type template struct {
		PP `mapstructure:",squash"`
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each post-processor and get out all the complex configs
	pps := tpl.ppList()

	for _, pp := range pps {
		ppTypeRaw, ok := pp["type"]
		if !ok {
			continue
		}

		if ppType, ok := ppTypeRaw.(string); !ok {
			continue
		} else if ppType != "docker-tag" {
			continue
		}

		// Create a []string to hold tag and tags values
		allTags := []string{}

		tagRaw, ok := pp["tag"]
		if ok {
			// Gather all "tag" into the []string
			switch t := tagRaw.(type) {
			case []interface{}:
				for _, tag := range t {
					allTags = append(allTags, tag.(string))
				}
			case []string:
				allTags = append(allTags, t...)
			case string:
				tList := strings.Split(t, ",")
				for _, tag := range tList {
					allTags = append(allTags, strings.TrimSpace(tag))
				}
			}
		}

		// Now check to see if they already have the "tags" field
		tagsRaw, ok := pp["tags"]
		if ok {
			// Gather all "tag" into the []string
			switch t := tagsRaw.(type) {
			case []interface{}:
				for _, tag := range t {
					allTags = append(allTags, tag.(string))
				}
			case []string:
				allTags = append(allTags, t...)
			case string:
				tList := strings.Split(t, ",")
				for _, tag := range tList {
					allTags = append(allTags, strings.TrimSpace(tag))
				}
			}
		}

		// Now deduplicate the tags in the list so we don't tag the same thing
		// multiple times.
		deduplicater := map[string]bool{}
		finalTags := []string{}
		for _, tag := range allTags {
			if found := deduplicater[tag]; found {
				continue
			}
			deduplicater[tag] = true
			finalTags = append(finalTags, tag)
		}

		// Delete tag key, and set tags key to the final deduplicated list.
		delete(pp, "tag")
		pp["tags"] = finalTags
	}

	input["post-processors"] = tpl.PostProcessors
	return input, nil
}

func (FixerDockerTagtoTags) Synopsis() string {
	return `Updates "docker" post-processor so any "tag" field is renamed to "tags".`
}

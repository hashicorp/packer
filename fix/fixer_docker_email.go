// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package fix

import "github.com/mitchellh/mapstructure"

type FixerDockerEmail struct{}

func (FixerDockerEmail) DeprecatedOptions() map[string][]string {
	return map[string][]string{
		"packer.docker":                       []string{"login_email"},
		"packer.post-processor.docker-import": []string{"login_email"},
	}

}

func (FixerDockerEmail) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	if input["post-processors"] == nil {
		return input, nil
	}

	// Our template type we'll use for this fixer only
	type template struct {
		Builders []map[string]interface{}
		PP       `mapstructure:",squash"`
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	// Go through each builder and delete `docker_login` if present
	for _, builder := range tpl.Builders {
		_, ok := builder["login_email"]
		if !ok {
			continue
		}
		delete(builder, "login_email")
	}

	// Go through each post-processor and delete `docker_login` if present
	pps := tpl.ppList()

	for _, pp := range pps {
		_, ok := pp["login_email"]
		if !ok {
			continue
		}
		delete(pp, "login_email")
	}

	input["builders"] = tpl.Builders
	input["post-processors"] = tpl.PostProcessors
	return input, nil
}

func (FixerDockerEmail) Synopsis() string {
	return `Removes "login_email" from the Docker builder.`
}

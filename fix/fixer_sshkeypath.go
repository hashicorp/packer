package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FixerSSHKeyPath changes the "ssh_key_path" of a template
// to "ssh_private_key_file".
type FixerSSHKeyPath struct{}

func (FixerSSHKeyPath) Fix(input map[string]interface{}) (map[string]interface{}, error) {
	// The type we'll decode into; we only care about builders
	type template struct {
		Builders []map[string]interface{}
	}

	// Decode the input into our structure, if we can
	var tpl template
	if err := mapstructure.Decode(input, &tpl); err != nil {
		return nil, err
	}

	for _, builder := range tpl.Builders {
		sshKeyPathRaw, ok := builder["ssh_key_path"]
		if !ok {
			continue
		}

		sshKeyPath, ok := sshKeyPathRaw.(string)
		if !ok {
			continue
		}

		// only assign to ssh_private_key_file if it doesn't
		// already exist; otherwise we'll just ignore ssh_key_path
		_, sshPrivateIncluded := builder["ssh_private_key_file"]
		if !sshPrivateIncluded {
			builder["ssh_private_key_file"] = sshKeyPath
		}

		delete(builder, "ssh_key_path")
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FixerSSHKeyPath) Synopsis() string {
	return `Updates builders using "ssh_key_path" to use "ssh_private_key_file"`
}

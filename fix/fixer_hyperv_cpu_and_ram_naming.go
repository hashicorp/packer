package fix

import (
	"github.com/mitchellh/mapstructure"
)

// FizerHypervCPUandRAM fixes the typo in "clone_from_vmxc_path" replacing
// it with "clone_from_vmcx_path" in Hyper-V VMCX builder templates
type FizerHypervCPUandRAM struct{}

func (FizerHypervCPUandRAM) Fix(input map[string]interface{}) (map[string]interface{}, error) {
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
		builderTypeRaw, ok := builder["type"]
		if !ok {
			continue
		}

		builderType, ok := builderTypeRaw.(string)
		if !ok {
			continue
		}

		if builderType != "hyperv-vmcx" && builderType != "hyperv-iso" {
			continue
		}

		ncpus, ok := builder["cpu"]
		if ok {
			delete(builder, "cpu")
			builder["cpus"] = ncpus
		}

		memory, ok := builder["ram_size"]
		if ok {
			delete(builder, "ram_size")
			builder["memory"] = memory
		}
	}

	input["builders"] = tpl.Builders
	return input, nil
}

func (FizerHypervCPUandRAM) Synopsis() string {
	return `Replaces "cpu" with "cpus" and "ram_size" with "memory"` +
		`in Hyper-V VMCX builder templates`
}

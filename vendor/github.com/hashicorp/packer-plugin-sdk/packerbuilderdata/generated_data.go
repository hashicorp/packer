// Package packerbuilderdata provides tooling for setting and getting special
// builder-generated data that will be passed to the provisioners. This data
// should be limited to runtime data like instance id, ip address, and other
// relevant details that provisioning scripts may need access to.
package packerbuilderdata

import "github.com/hashicorp/packer-plugin-sdk/multistep"

// This is used in the BasicPlaceholderData() func in the packer/provisioner.go
// To force users to access generated data via the "generated" func.
const PlaceholderMsg = "To set this dynamically in the Packer template, " +
	"you must use the `build` function"

// GeneratedData manages variables created and exported by a builder after
// it starts, so that provisioners and post-processors can have access to
// build data generated at runtime -- for example, instance ID or instance IP
// address. Internally, it uses the builder's multistep.StateBag. The user
// must make sure that the State field is not is not nil before calling Put().
type GeneratedData struct {
	// The builder's StateBag
	State multistep.StateBag
}

func (gd *GeneratedData) Put(key string, data interface{}) {
	genData := make(map[string]interface{})
	if _, ok := gd.State.GetOk("generated_data"); ok {
		genData = gd.State.Get("generated_data").(map[string]interface{})
	}
	genData[key] = data
	gd.State.Put("generated_data", genData)
}

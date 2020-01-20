package builder

import "github.com/hashicorp/packer/helper/multistep"

// GeneratedData manages variables exported by a builder after
// it started. It uses the builder's multistep.StateBag internally, make sure it
// is not nil before calling any functions.
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

package builder

import "github.com/hashicorp/packer/helper/multistep"

// GeneratedData manages the generated_data inside
// the StateBag to make sure the data won't be overwritten
// It should be used when a builder adds multiple custom template engine variables
type GeneratedData struct {
	// Just the builder StateBag
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

package common

import "github.com/hashicorp/packer/helper/multistep"

type GeneratedData struct {
	Data  map[string]interface{}
	State multistep.StateBag
}

func (gd *GeneratedData) Put(key string, data interface{}) {
	if gd.Data == nil {
		gd.Data = make(map[string]interface{})
	}
	gd.Data[key] = data
	gd.State.Put("generated_data", gd.Data)
}

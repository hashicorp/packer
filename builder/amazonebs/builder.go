package amazonebs

import (
	"encoding/json"
	"errors"
	"github.com/mitchellh/packer/packer"
)

type config struct {
	AccessKey string `json:"access_key"`
	Region    string
	SecretKey string `json:"secret_key"`
	SourceAmi string `json:"source_ami"`
}

type Builder struct {
	config config
}

func (b *Builder) Prepare(raw interface{}) (err error) {
	_, ok := raw.(map[string]interface{})
	if !ok {
		err = errors.New("configuration isn't a valid map")
		return
	}

	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonBytes, &b.config)
	if err != nil {
		return
	}

	return
}

func (*Builder) Run(packer.Build, packer.Ui) {}

// The amazonebs package contains a packer.Builder implementation that
// builds AMIs for Amazon EC2.
//
// In general, there are two types of AMIs that can be created: ebs-backed or
// instance-store. This builder _only_ builds ebs-backed images.
package amazonebs

import (
	"encoding/json"
	"github.com/mitchellh/packer/packer"
	"log"
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
	// Marshal and unmarshal the raw configuration as a way to get it
	// into our "config" struct.
	// TODO: Use the reflection package and provide this as an API for
	// better error messages
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return
	}

	err = json.Unmarshal(jsonBytes, &b.config)
	if err != nil {
		return
	}

	log.Printf("Config: %+v\n", b.config)

	// TODO: Validate the configuration
	return
}

func (*Builder) Run(b packer.Build, ui packer.Ui) {
	ui.Say("Building an AWS image...\n")
}

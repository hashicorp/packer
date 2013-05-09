package amazonebs

import (
	"github.com/mitchellh/packer/packer"
)

type config struct {
	AccessKey string
	Region    string
	SecretKey string
	SourceAmi string
}

type Builder struct {
	config config
}

func (*Builder) Prepare(interface{}) {}

func (*Builder) Run(packer.Build, packer.Ui) {}

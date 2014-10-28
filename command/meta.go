package command

import (
	"github.com/mitchellh/cli"
	"github.com/mitchellh/packer/packer"
)

type Meta struct {
	EnvConfig *packer.EnvironmentConfig
	Ui        cli.Ui
}

func (m *Meta) Environment() (packer.Environment, error) {
	return packer.NewEnvironment(m.EnvConfig)
}

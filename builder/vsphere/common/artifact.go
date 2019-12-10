package common

import (
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

const BuilderId = "jetbrains.vsphere"

type Artifact struct {
	Name string
	VM   *driver.VirtualMachine
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
	return a.Name
}

func (a *Artifact) String() string {
	return a.Name
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return a.VM.Destroy()
}

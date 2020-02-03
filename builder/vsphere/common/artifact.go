package common

import (
	"github.com/hashicorp/packer/builder/vsphere/driver"
)

const BuilderId = "jetbrains.vsphere"

type Artifact struct {
	Name string
	VM   *driver.VirtualMachine

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
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
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return a.VM.Destroy()
}

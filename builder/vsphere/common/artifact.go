package common

import (
	"os"

	"github.com/hashicorp/packer/builder/vsphere/driver"
)

const BuilderId = "jetbrains.vsphere"

type Artifact struct {
	Outconfig *OutputConfig
	Name      string
	VM        *driver.VirtualMachineDriver

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	if a.Outconfig != nil {
		files, _ := a.Outconfig.ListFiles()
		return files
	}
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
	if a.Outconfig != nil {
		os.RemoveAll(a.Outconfig.OutputDir)
	}
	return a.VM.Destroy()
}

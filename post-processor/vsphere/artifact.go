package vsphere

import (
	"fmt"
)

const BuilderId = "packer.post-processor.vsphere"

type Artifact struct {
	files     []string
	datastore string
	vmfolder  string
	vmname    string
}

func NewArtifact(datastore, vmfolder, vmname string, files []string) *Artifact {
	return &Artifact{
		files:     files,
		datastore: datastore,
		vmfolder:  vmfolder,
		vmname:    vmname,
	}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.files
}

func (a *Artifact) Id() string {
	return fmt.Sprintf("%s::%s::%s", a.datastore, a.vmfolder, a.vmname)
}

func (a *Artifact) String() string {
	return fmt.Sprintf("VM: %s Folder: %s Datastore: %s", a.vmname, a.vmfolder, a.datastore)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}

package common

// BuilderID is the Packer id for nutanix
const BuilderID = "nutanix"

// Artifact contains the unique keys for the nutanix artifact produced from Packer
type Artifact struct {
	Name string
	UUID string
	//VM   *driver.VirtualMachine
}

// BuilderId will return the unique builder id
func (a *Artifact) BuilderId() string {
	return BuilderID
}

// Files will return the files from the builder
func (a *Artifact) Files() []string {
	return []string{}
}

// Id returns the UUID for the saved image
func (a *Artifact) Id() string {
	return a.UUID
}

// String returns a String name of the artifact
func (a *Artifact) String() string {
	return a.Name
}

// State returns nothing important right now
func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy returns nothing important right now
func (a *Artifact) Destroy() error {
	return nil
	//return a.VM.Destroy()
}

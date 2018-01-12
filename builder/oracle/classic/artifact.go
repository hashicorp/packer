package classic

// Artifact is an artifact implementation that contains a built Custom Image.
type Artifact struct {
}

// BuilderId uniquely identifies the builder.
func (a *Artifact) BuilderId() string {
	return BuilderId
}

// Files lists the files associated with an artifact. We don't have any files
// as the custom image is stored server side.
func (a *Artifact) Files() []string {
	return nil
}

// Id returns the OCID of the associated Image.
func (a *Artifact) Id() string {
	return ""
}

func (a *Artifact) String() string {
	return ""
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

// Destroy deletes the custom image associated with the artifact.
func (a *Artifact) Destroy() error {
	return nil
}

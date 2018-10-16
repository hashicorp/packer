package vmware_vmx

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer/packer"
)

const BuilderId = "packer.post-processor.vmware-vmx"

// An Artifact is the result of a build, and is the metadata that documents
// what a builder actually created. The exact meaning of the contents is
// specific to each builder, but this interface is used to communicate back
// to the user the result of a build.
type Artifact struct {
	files []string
}

func NewArtifact(files []string) *Artifact {
	return &Artifact{files: files}
}

// Returns the ID of the builder that was used to create this artifact.
// This is the internal ID of the builder and should be unique to every
// builder. This can be used to identify what the contents of the
// artifact actually are.
func (a *Artifact) BuilderId() string {
	return BuilderId
}

// Returns the set of files that comprise this artifact. If an
// artifact is not made up of files, then this will be empty.
func (a *Artifact) Files() []string {
	return a.files
}

// The ID for the artifact, if it has one. This is not guaranteed to
// be unique every run (like a GUID), but simply provide an identifier
// for the artifact that may be meaningful in some way. For example,
// for Amazon EC2, this value might be the AMI ID.
func (*Artifact) Id() string {
	return ""
}

// Returns human-readable output that describes the artifact created.
// This is used for UI output. It can be multiple lines.
func (a *Artifact) String() string {
	return fmt.Sprintf("Transformed VMX from a VMWare builder: %#v", a.files)
}

// State allows the caller to ask for builder specific state information
// relating to the artifact instance.
func (*Artifact) State(name string) interface{} {
	return nil
}

// Destroy deletes the artifact. Packer calls this for various reasons,
// such as if a post-processor has processed this artifact and it is
// no longer needed.
func (a *Artifact) Destroy() error {
	errs := new(packer.MultiError)

	// Remove all the artifact files
	for _, f := range a.files {
		if err := os.RemoveAll(f); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("Unable to remove file %s: %s", f, err))
		}
	}

	// Return any errors that were aggregated
	if len(errs.Errors) > 0 {
		return errs
	}
	return nil
}

package packer

// An Artifact is the result of a build, and is the metadata that documents
// what a builder actually created. The exact meaning of the contents is
// specific to each builder, but this interface is used to communicate back
// to the user the result of a build.
type Artifact interface {
	// Returns the ID of the builder that was used to create this artifact.
	// This is the internal ID of the builder and should be unique to every
	// builder. This can be used to identify what the contents of the
	// artifact actually are.
	BuilderId() string

	// Returns the set of files that comprise this artifact. If an
	// artifact is not made up of files, then this will be empty.
	Files() []string

	// The ID for the artifact, if it has one. This is not guaranteed to
	// be unique every run (like a GUID), but simply provide an identifier
	// for the artifact that may be meaningful in some way. For example,
	// for Amazon EC2, this value might be the AMI ID.
	Id() string

	// Returns human-readable output that describes the artifact created.
	// This is used for UI output. It can be multiple lines.
	String() string

	// State allows the caller to ask for builder specific state information
	// relating to the artifact instance.
	State(name string) interface{}

	// Destroy deletes the artifact. Packer calls this for various reasons,
	// such as if a post-processor has processed this artifact and it is
	// no longer needed.
	Destroy() error
}

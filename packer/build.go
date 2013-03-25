package packer

// A build struct represents a single build job, the result of which should
// be a single machine image artifact. This artifact may be comprised of
// multiple files, of course, but it should be for only a single provider
// (such as VirtualBox, EC2, etc.).
type Build struct {
	name    string
	builder Builder
}

type Builder interface {
	Prepare()
}

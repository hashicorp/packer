package common

// OutputDir is an interface type that abstracts the creation and handling
// of the output directory for VMware-based products. The abstraction is made
// so that the output directory can be properly made on remote (ESXi) based
// VMware products as well as local.
type OutputDir interface {
	DirExists() (bool, error)
	ListFiles() ([]string, error)
	MkdirAll() error
	Remove(string) error
	RemoveAll() error
	SetOutputDir(string)
	String() string
}

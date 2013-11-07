package vmware

import (
	"os"
)

// OutputDir is an interface type that abstracts the creation and handling
// of the output directory for VMware-based products. The abstraction is made
// so that the output directory can be properly made on remote (ESXi) based
// VMware products as well as local.
type OutputDir interface {
	DirExists(string) (bool, error)
	MkdirAll(string) error
	RemoveAll(string) error
}

// localOutputDir is an OutputDir implementation where the directory
// is on the local machine.
type localOutputDir struct{}

func (localOutputDir) DirExists(path string) (bool, error) {
	_, err := os.Stat(path)
	return err == nil, err
}

func (localOutputDir) MkdirAll(path string) error {
	return os.MkdirAll(path, 0755)
}

func (localOutputDir) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

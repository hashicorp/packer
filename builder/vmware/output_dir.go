package vmware

import (
	"os"
)

// OutputDir is an interface type that abstracts the creation and handling
// of the output directory for VMware-based products. The abstraction is made
// so that the output directory can be properly made on remote (ESXi) based
// VMware products as well as local.
type OutputDir interface {
	DirExists() (bool, error)
	MkdirAll() error
	RemoveAll() error
	SetOutputDir(string)
}

// localOutputDir is an OutputDir implementation where the directory
// is on the local machine.
type localOutputDir struct {
	dir string
}

func (d *localOutputDir) DirExists() (bool, error) {
	_, err := os.Stat(d.dir)
	return err == nil, err
}

func (d *localOutputDir) MkdirAll() error {
	return os.MkdirAll(d.dir, 0755)
}

func (d *localOutputDir) RemoveAll() error {
	return os.RemoveAll(d.dir)
}

func (d *localOutputDir) SetOutputDir(path string) {
	d.dir = path
}

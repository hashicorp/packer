package iso

import (
	"os"
	"path/filepath"
)

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
}

// localOutputDir is an OutputDir implementation where the directory
// is on the local machine.
type localOutputDir struct {
	dir string
}

func (d *localOutputDir) DirExists() (bool, error) {
	_, err := os.Stat(d.dir)
	return err == nil, nil
}

func (d *localOutputDir) ListFiles() ([]string, error) {
	files := make([]string, 0, 10)

	visit := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}

	return files, filepath.Walk(d.dir, visit)
}

func (d *localOutputDir) MkdirAll() error {
	return os.MkdirAll(d.dir, 0755)
}

func (d *localOutputDir) Remove(path string) error {
	return os.Remove(path)
}

func (d *localOutputDir) RemoveAll() error {
	return os.RemoveAll(d.dir)
}

func (d *localOutputDir) SetOutputDir(path string) {
	d.dir = path
}

func (d *localOutputDir) String() string {
	return d.dir
}

package common

import (
	"os"
	"path/filepath"
)

// LocalOutputDir is an OutputDir implementation where the directory
// is on the local machine.
type LocalOutputDir struct {
	Dir string
}

func (d *LocalOutputDir) DirExists() (bool, error) {
	_, err := os.Stat(d.Dir)
	return err == nil, nil
}

func (d *LocalOutputDir) ListFiles() ([]string, error) {
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

	return files, filepath.Walk(d.Dir, visit)
}

func (d *LocalOutputDir) MkdirAll() error {
	return os.MkdirAll(d.Dir, 0755)
}

func (d *LocalOutputDir) Remove(path string) error {
	return os.Remove(path)
}

func (d *LocalOutputDir) RemoveAll() error {
	return os.RemoveAll(d.Dir)
}

func (d *LocalOutputDir) SetOutputDir(path string) {
	d.Dir = path
}

func (d *LocalOutputDir) String() string {
	return d.Dir
}

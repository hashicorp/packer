package common

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mitchellh/packer/packer"
)

// BuilderId is the common builder ID to all of these artifacts.
const BuilderId = "packer.parallels"

// These are the extensions of files and directories that are unnecessary for the function
// of a Parallels virtual machine.
var unnecessaryFiles = []string{"\\.log$", "\\.backup$", "\\.Backup$", "\\.app"}

// Artifact is the result of running the parallels builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	dir string
	f   []string
}

// NewArtifact returns a Parallels artifact containing the files
// in the given directory.
func NewArtifact(dir string) (packer.Artifact, error) {
	files := make([]string, 0, 5)
	visit := func(path string, info os.FileInfo, err error) error {
		for _, unnecessaryFile := range unnecessaryFiles {
			if unnecessary, _ := regexp.MatchString(unnecessaryFile, path); unnecessary {
				return os.RemoveAll(path)
			}
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return err
	}

	if err := filepath.Walk(dir, visit); err != nil {
		return nil, err
	}

	return &artifact{
		dir: dir,
		f:   files,
	}, nil
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return a.f
}

func (*artifact) Id() string {
	return "VM"
}

func (a *artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *artifact) State(name string) interface{} {
	return nil
}

func (a *artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}

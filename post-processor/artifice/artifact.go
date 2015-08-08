package artifice

import (
	"fmt"
	"os"
	"strings"
)

const BuilderId = "packer.post-processor.artifice"

type Artifact struct {
	files []string
}

func NewArtifact(files []string) (*Artifact, error) {
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			return nil, err
		}
	}
	artifact := &Artifact{
		files: files,
	}
	return artifact, nil
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return a.files
}

func (a *Artifact) Id() string {
	return ""
}

func (a *Artifact) String() string {
	files := strings.Join(a.files, ", ")
	return fmt.Sprintf("Created artifact from files: %s", files)
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	for _, f := range a.files {
		err := os.RemoveAll(f)
		if err != nil {
			return err
		}
	}
	return nil
}

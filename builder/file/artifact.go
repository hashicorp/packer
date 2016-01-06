package file

import (
	"fmt"
	"log"
	"os"
)

type FileArtifact struct {
	filename string
}

func (*FileArtifact) BuilderId() string {
	return BuilderId
}

func (a *FileArtifact) Files() []string {
	return []string{a.filename}
}

func (a *FileArtifact) Id() string {
	return "File"
}

func (a *FileArtifact) String() string {
	return fmt.Sprintf("Stored file: %s", a.filename)
}

func (a *FileArtifact) State(name string) interface{} {
	return nil
}

func (a *FileArtifact) Destroy() error {
	log.Printf("Deleting %s", a.filename)
	return os.Remove(a.filename)
}

package file

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/hashicorp/packer-plugin-sdk/packer/registryimage"
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
	if name == registryimage.ArtifactStateURI {
		return registryimage.FromArtifact(a,
			registryimage.WithID(path.Base(a.filename)),
			registryimage.WithRegion(path.Dir(a.filename)),
		)
	}

	return nil
}

func (a *FileArtifact) Destroy() error {
	log.Printf("Deleting %s", a.filename)
	return os.Remove(a.filename)
}

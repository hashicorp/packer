// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package file

import (
	"fmt"
	"log"
	"os"
	"path"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

type FileArtifact struct {
	source   string
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
		img, err := registryimage.FromArtifact(a,
			registryimage.WithID(path.Base(a.filename)),
			registryimage.WithRegion(path.Dir(a.filename)),
			registryimage.WithSourceID(a.source),
		)

		if err != nil {
			log.Printf("[DEBUG] error encountered when creating a registry image %v", err)
			return nil
		}

		return img
	}

	return nil
}

func (a *FileArtifact) Destroy() error {
	log.Printf("Deleting %s", a.filename)
	return os.Remove(a.filename)
}

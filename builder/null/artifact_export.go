// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package null

import (
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
)

// dummy Artifact implementation - does nothing
type NullArtifact struct {
}

func (*NullArtifact) BuilderId() string {
	return BuilderId
}

func (a *NullArtifact) Files() []string {
	return []string{}
}

func (*NullArtifact) Id() string {
	return "Null"
}

func (a *NullArtifact) String() string {
	return "Did not export anything. This is the null builder"
}

func (a *NullArtifact) State(name string) interface{} {
	switch name {
	case registryimage.ArtifactStateURI:
		img, _ := registryimage.FromArtifact(a,
			registryimage.WithID(a.Id()),
			registryimage.WithProvider("null"),
			registryimage.WithRegion("null"),
			registryimage.WithSourceID("null"),
		)
		return img
	case "generated_data":
		return map[interface{}]interface{}{
			"ID": "Null",
		}
	default:
		return nil
	}
}

func (a *NullArtifact) Destroy() error {
	return nil
}

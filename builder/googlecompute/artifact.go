// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package googlecompute

import (
	"fmt"
	"log"
)

// Artifact represents a GCE image as the result of a Packer build.
type Artifact struct {
	imageName string
	client    *GoogleComputeClient
}

// BuilderId returns the builder Id.
func (*Artifact) BuilderId() string {
	return BuilderId
}

// Destroy destroys the GCE image represented by the artifact.
func (a *Artifact) Destroy() error {
	log.Printf("Destroying image: %s", a.imageName)
	// Ignore the operation result as we are not waiting until it completes.
	_, err := a.client.DeleteImage(a.imageName)
	if err != nil {
		return err
	}
	return nil
}

// Files returns the files represented by the artifact.
func (*Artifact) Files() []string {
	return nil
}

// Id returns the GCE image name.
func (a *Artifact) Id() string {
	return a.imageName
}

// String returns the string representation of the artifact.
func (a *Artifact) String() string {
	return fmt.Sprintf("A disk image was created: %v", a.imageName)
}

// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/packer/packer"
)

// This is the common builder ID to all of these artifacts.
const BuilderId = "MSOpenTech.hyperv"

// Artifact is the result of running the VirtualBox builder, namely a set
// of files associated with the resulting machine.
type artifact struct {
	dir string
	f   []string
}

// NewArtifact returns a VirtualBox artifact containing the files
// in the given directory.
func NewArtifact(dir string) (packer.Artifact, error) {
	files := make([]string, 0, 5)

	// we need to store output dir path to get rel path to keep dir tree :)
	// to not modify interface - put it as the first array element
	files = append(files, dir)

	visit := func(path string, info os.FileInfo, err error) error {
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

func (a *artifact) State(name string) interface{} {
	return "Not implemented"
}

func (a *artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", a.dir)
}

func (a *artifact) Destroy() error {
	return os.RemoveAll(a.dir)
}

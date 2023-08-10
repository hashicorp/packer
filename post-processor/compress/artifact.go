// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package compress

import (
	"fmt"
	"os"
)

const BuilderId = "packer.post-processor.compress"

type Artifact struct {
	Path string
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Id() string {
	return ""
}

func (a *Artifact) Files() []string {
	return []string{a.Path}
}

func (a *Artifact) String() string {
	return fmt.Sprintf("compressed artifacts in: %s", a.Path)
}

func (*Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	return os.Remove(a.Path)
}

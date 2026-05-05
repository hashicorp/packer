// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

//go:build netbsd || openbsd || solaris

package sbom

import (
	"context"
	"fmt"
	"runtime"
)

// Generate returns an error on platforms where the Syft SDK cannot be built.
func (g *Generator) Generate(ctx context.Context) ([]byte, error) {
	_ = ctx
	return nil, fmt.Errorf("sbom generation is not supported on %s builds", runtime.GOOS)
}

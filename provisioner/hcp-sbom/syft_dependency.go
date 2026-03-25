// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: BUSL-1.1

// Exclude platforms where containerd (a transitive dependency of syft) doesn't compile.
// Containerd has platform-specific code that lacks support for NetBSD, OpenBSD, and Solaris.
// This exclusion only affects dependency tracking; the hcp-sbom provisioner downloads
// pre-built syft binaries at runtime and works on all platforms where those binaries exist.
//go:build !netbsd && !openbsd && !solaris

package hcp_sbom

import (
	// Blank import to register Syft as a dependency
	// This file exists to declare Syft as a dependency for license and security scanning purposes.
	// While Packer downloads and executes Syft binaries at runtime, this import ensures
	// the Syft project appears in dependency analysis tools and SBOMs generated for Packer itself.
	_ "github.com/anchore/syft/syft"
)

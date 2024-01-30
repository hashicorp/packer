// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !go1.10
// +build !go1.10

package compress

import "archive/tar"

func setHeaderFormat(header *tar.Header) {
	// no-op
}

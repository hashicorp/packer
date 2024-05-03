// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

//go:build go1.10
// +build go1.10

package compress

import (
	"archive/tar"
	"time"
)

func setHeaderFormat(header *tar.Header) {
	// We have to set the Format explicitly for the googlecompute-import
	// post-processor. Google Cloud only allows importing GNU tar format.
	header.Format = tar.FormatGNU
	header.AccessTime = time.Time{}
	header.ModTime = time.Time{}
	header.ChangeTime = time.Time{}
}

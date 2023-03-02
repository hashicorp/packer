// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package packer

import (
	"path/filepath"
)

const FixtureDir = "./test-fixtures"

func fixtureDir(n string) string {
	return filepath.Join(FixtureDir, n)
}

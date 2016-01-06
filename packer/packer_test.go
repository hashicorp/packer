package packer

import (
	"path/filepath"
)

const FixtureDir = "./test-fixtures"

func fixtureDir(n string) string {
	return filepath.Join(FixtureDir, n)
}

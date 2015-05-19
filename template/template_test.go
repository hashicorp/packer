package template

import (
	"path/filepath"
)

const FixturesDir = "./test-fixtures"

// fixtureDir returns the path to a test fixtures directory
func fixtureDir(n string) string {
	return filepath.Join(FixturesDir, n)
}

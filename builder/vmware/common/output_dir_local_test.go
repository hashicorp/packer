package common

import (
	"testing"
)

func TestLocalOutputDir_impl(t *testing.T) {
	var _ OutputDir = new(LocalOutputDir)
}

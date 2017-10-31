package common

import (
	"testing"
)

func TestLocalOuputDir_impl(t *testing.T) {
	var _ OutputDir = new(LocalOutputDir)
}

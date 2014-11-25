package common

import (
	"github.com/mitchellh/packer/packer"
	"testing"
)

func testConfigTemplate(t *testing.T) *packer.ConfigTemplate {
	result, err := packer.NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return result
}

package virtualbox_acc

// This is the code necessary for running the provisioner acceptance tests.
// It provides the builder config and cleans up created resource.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/command"
	"github.com/hashicorp/packer/packer"

	testshelper "github.com/hashicorp/packer/helper/tests"
)

type VirtualBoxISOAccTest struct{}

func (v *VirtualBoxISOAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("../../builder/virtualbox/iso/acceptance/test-fixtures/", "virtualbox-iso.txt")
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), nil
}

func (v *VirtualBoxISOAccTest) CleanUp() error {
	testshelper.CleanupFiles("virtualbox-iso-packer-acc-test")
	testshelper.CleanupFiles("packer_cache")
	return nil
}

func (v *VirtualBoxISOAccTest) GetBuilderStore() packer.MapOfBuilder {
	return packer.MapOfBuilder{
		"virtualbox-iso": func() (packer.Builder, error) { return command.Builders["virtualbox-iso"], nil },
	}
}

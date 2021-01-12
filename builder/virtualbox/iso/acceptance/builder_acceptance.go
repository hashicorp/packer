package virtualbox_acc

// This is the code necessary for running the provisioner acceptance tests.
// It provides the builder config and cleans up created resource.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/packer/builder/virtualbox/iso"

	"github.com/hashicorp/packer-plugin-sdk/acctest/testutils"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type VirtualBoxISOAccTest struct{}

func (v *VirtualBoxISOAccTest) GetConfigs() (map[string]string, error) {
	filePath := filepath.Join("../../builder/virtualbox/iso/acceptance/test-fixtures/", "virtualbox-iso.txt")
	config, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	if err != nil {
		return nil, fmt.Errorf("Unable to read %s", filePath)
	}
	return map[string]string{"linux": string(file)}, nil
}

func (v *VirtualBoxISOAccTest) CleanUp() error {
	testutils.CleanupFiles("virtualbox-iso-packer-acc-test")
	testutils.CleanupFiles("packer_cache")
	return nil
}

func (v *VirtualBoxISOAccTest) GetBuilderStore() packersdk.MapOfBuilder {
	return packersdk.MapOfBuilder{
		"virtualbox-iso": func() (packersdk.Builder, error) { return &iso.Builder{}, nil },
	}
}

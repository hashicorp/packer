package virtualbox_acceptance

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	testshelper "github.com/hashicorp/packer/helper/tests"
)

type VirtualBoxISOAccTest struct{}

func (v *VirtualBoxISOAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("../../../builder/virtualbox/iso/acceptance/test-fixtures/", "virtualbox-iso.txt")
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

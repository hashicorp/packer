package amazon_acc

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

type AmazonEBSAccTest struct{}

func (s *AmazonEBSAccTest) GetConfigs() (map[string]string, error) {
	filePath := filepath.Join("../../builder/amazon/ebs/acceptance/test-fixtures/", "amazon-ebs.txt")
	config, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	if err != nil {
		return nil, fmt.Errorf("Uneble to read %s", filePath)
	}
	return map[string]string{"linux": string(file)}, nil
}

func (s *AmazonEBSAccTest) CleanUp() error {
	helper := testshelper.AWSHelper{
		Region:  "us-east-1",
		AMIName: "packer-acc-test",
	}
	return helper.CleanUpAmi()
}

func (s *AmazonEBSAccTest) GetBuilderStore() packer.MapOfBuilder {
	return packer.MapOfBuilder{
		"amazon-ebs": func() (packer.Builder, error) { return command.Builders["amazon-ebs"], nil },
	}
}

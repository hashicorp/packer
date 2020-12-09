package amazon_acc

// This is the code necessary for running the provisioner acceptance tests.
// It provides the builder config and cleans up created resource.

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	amazonebsbuilder "github.com/hashicorp/packer/builder/amazon/ebs"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type AmazonEBSAccTest struct{}

func (s *AmazonEBSAccTest) GetConfigs() (map[string]string, error) {
	fixtures := map[string]string{
		"linux":   "amazon-ebs.txt",
		"windows": "amazon-ebs_windows.txt",
	}

	configs := make(map[string]string)

	for distro, fixture := range fixtures {
		fileName := fixture
		filePath := filepath.Join("../../builder/amazon/ebs/acceptance/test-fixtures/", fileName)
		config, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("Expected to find %s", filePath)
		}
		defer config.Close()

		file, err := ioutil.ReadAll(config)
		if err != nil {
			return nil, fmt.Errorf("Unable to read %s", filePath)
		}

		configs[distro] = string(file)

	}
	return configs, nil
}

func (s *AmazonEBSAccTest) CleanUp() error {
	helper := AWSHelper{
		Region:  "us-east-1",
		AMIName: "packer-acc-test",
	}
	return helper.CleanUpAmi()
}

func (s *AmazonEBSAccTest) GetBuilderStore() packersdk.MapOfBuilder {
	return packersdk.MapOfBuilder{
		"amazon-ebs": func() (packersdk.Builder, error) { return &amazonebsbuilder.Builder{}, nil },
	}
}

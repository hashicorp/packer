package amazon_acceptance

import (
	"fmt"
	testshelper "github.com/hashicorp/packer/helper/tests"
	"io/ioutil"
	"os"
	"path/filepath"
)

type AmazonEBSAccTest struct {}

func (s *AmazonEBSAccTest) GetConfig() (string, error) {
	filePath := filepath.Join("../../../builder/amazon/ebs/acceptance/test-fixtures/", "amazon-ebs.txt")
	config, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("Expected to find %s", filePath)
	}
	defer config.Close()

	file, err := ioutil.ReadAll(config)
	return string(file), nil
}

func (s *AmazonEBSAccTest) CleanUp() error {
	helper := testshelper.AWSHelper{
		Region:  "us-east-1",
		AMIName: "packer-acc-test",
	}
	return helper.CleanUpAmi()
}

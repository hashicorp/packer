package jdcloud

import (
	"fmt"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	vm "github.com/jdcloud-api/jdcloud-sdk-go/services/vm/client"
	vpc "github.com/jdcloud-api/jdcloud-sdk-go/services/vpc/client"
	"os"
)

type JDCloudCredentialConfig struct {
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	RegionId  string `mapstructure:"region_id"`
	Az        string `mapstructure:"az"`
}

func (jd *JDCloudCredentialConfig) Prepare(ctx *interpolate.Context) []error {

	errorArray := []error{}

	if jd == nil {
		return append(errorArray, fmt.Errorf("[PRE-FLIGHT] Empty JDCloudCredentialConfig detected"))
	}

	if err := jd.ValidateKeyPair(); err != nil {
		errorArray = append(errorArray, err)
	}

	if err := jd.validateRegion(); err != nil {
		errorArray = append(errorArray, err)
	}

	if err := jd.validateAz(); err != nil {
		errorArray = append(errorArray, err)
	}

	if len(errorArray) != 0 {
		return errorArray
	}

	credential := core.NewCredentials(jd.AccessKey, jd.SecretKey)
	VmClient = vm.NewVmClient(credential)
	VpcClient = vpc.NewVpcClient(credential)
	Region = jd.RegionId

	return nil
}

func (jd *JDCloudCredentialConfig) ValidateKeyPair() error {

	if jd.AccessKey == "" {
		jd.AccessKey = os.Getenv("JDCLOUD_ACCESS_KEY")
	}

	if jd.SecretKey == "" {
		jd.SecretKey = os.Getenv("JDCLOUD_SECRET_KEY")
	}

	if jd.AccessKey == "" || jd.SecretKey == "" {
		return fmt.Errorf("[PRE-FLIGHT] We can't find your key pairs," +
			"write them here {access_key=xxx , secret_key=xxx} " +
			"or export them as env-variable, {export JDCLOUD_ACCESS_KEY=xxx, export JDCLOUD_SECRET_KEY=xxx} ")
	}

	return nil
}

func (config *JDCloudCredentialConfig) validateRegion() error {
	regionArray := []string{"cn-north-1", "cn-south-1", "cn-east-1", "cn-east-2"}
	for _, item := range regionArray {
		if item == config.RegionId {
			return nil
		}
	}
	return fmt.Errorf("[PRE-FLIGHT] Invalid RegionId:%s. "+
		"Legit RegionId are: {cn-north-1, cn-south-1, cn-east-1, cn-east-2}", config.RegionId)
}

func (config *JDCloudCredentialConfig) validateAz() error {
	if len(config.Az) == 0 {
		return fmt.Errorf("[PRE-FLIGHT] az info missing")
	}
	return nil
}

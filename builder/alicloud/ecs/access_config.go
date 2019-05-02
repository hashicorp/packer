package ecs

import (
	"fmt"
	"os"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/hashicorp/packer/version"
)

// Config of alicloud
type AlicloudAccessConfig struct {
	AlicloudAccessKey      string `mapstructure:"access_key"`
	AlicloudSecretKey      string `mapstructure:"secret_key"`
	AlicloudRegion         string `mapstructure:"region"`
	AlicloudSkipValidation bool   `mapstructure:"skip_region_validation"`
	SecurityToken          string `mapstructure:"security_token"`

	client *ClientWrapper
}

const Packer = "HashiCorp-Packer"
const DefaultRequestReadTimeout = 10 * time.Second

// Client for AlicloudClient
func (c *AlicloudAccessConfig) Client() (*ClientWrapper, error) {
	if c.client != nil {
		return c.client, nil
	}
	if c.SecurityToken == "" {
		c.SecurityToken = os.Getenv("SECURITY_TOKEN")
	}

	client, err := ecs.NewClientWithStsToken(c.AlicloudRegion, c.AlicloudAccessKey,
		c.AlicloudSecretKey, c.SecurityToken)
	if err != nil {
		return nil, err
	}

	client.AppendUserAgent(Packer, version.FormattedVersion())
	client.SetReadTimeout(DefaultRequestReadTimeout)
	c.client = &ClientWrapper{client}

	return c.client, nil
}

func (c *AlicloudAccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if err := c.Config(); err != nil {
		errs = append(errs, err)
	}

	if c.AlicloudRegion == "" {
		c.AlicloudRegion = os.Getenv("ALICLOUD_REGION")
	}

	if c.AlicloudRegion == "" {
		errs = append(errs, fmt.Errorf("region option or ALICLOUD_REGION must be provided in template file or environment variables."))
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func (c *AlicloudAccessConfig) Config() error {
	if c.AlicloudAccessKey == "" {
		c.AlicloudAccessKey = os.Getenv("ALICLOUD_ACCESS_KEY")
	}
	if c.AlicloudSecretKey == "" {
		c.AlicloudSecretKey = os.Getenv("ALICLOUD_SECRET_KEY")
	}
	if c.AlicloudAccessKey == "" || c.AlicloudSecretKey == "" {
		return fmt.Errorf("ALICLOUD_ACCESS_KEY and ALICLOUD_SECRET_KEY must be set in template file or environment variables.")
	}
	return nil

}

func (c *AlicloudAccessConfig) ValidateRegion(region string) error {

	supportedRegions, err := c.getSupportedRegions()
	if err != nil {
		return err
	}

	for _, supportedRegion := range supportedRegions {
		if region == supportedRegion {
			return nil
		}
	}

	return fmt.Errorf("Not a valid alicloud region: %s", region)
}

func (c *AlicloudAccessConfig) getSupportedRegions() ([]string, error) {
	client, err := c.Client()
	if err != nil {
		return nil, err
	}

	regionsRequest := ecs.CreateDescribeRegionsRequest()
	regionsResponse, err := client.DescribeRegions(regionsRequest)
	if err != nil {
		return nil, err
	}

	validRegions := make([]string, len(regionsResponse.Regions.Region))
	for _, valid := range regionsResponse.Regions.Region {
		validRegions = append(validRegions, valid.RegionId)
	}

	return validRegions, nil
}

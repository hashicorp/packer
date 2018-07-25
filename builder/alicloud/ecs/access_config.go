package ecs

import (
	"fmt"
	"os"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/hashicorp/packer/template/interpolate"
)

// Config of alicloud
type AlicloudAccessConfig struct {
	AlicloudAccessKey      string `mapstructure:"access_key"`
	AlicloudSecretKey      string `mapstructure:"secret_key"`
	AlicloudRegion         string `mapstructure:"region"`
	AlicloudSkipValidation bool   `mapstructure:"skip_region_validation"`
	SecurityToken          string `mapstructure:"security_token"`
}

// Client for AlicloudClient
func (c *AlicloudAccessConfig) Client() (*ecs.Client, error) {
	if err := c.loadAndValidate(); err != nil {
		return nil, err
	}
	if c.SecurityToken == "" {
		c.SecurityToken = os.Getenv("SECURITY_TOKEN")
	}
	client := ecs.NewECSClientWithSecurityToken(c.AlicloudAccessKey, c.AlicloudSecretKey,
		c.SecurityToken, common.Region(c.AlicloudRegion))

	client.SetBusinessInfo("Packer")
	if _, err := client.DescribeRegions(); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *AlicloudAccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error
	if err := c.Config(); err != nil {
		errs = append(errs, err)
	}

	if c.AlicloudRegion != "" && !c.AlicloudSkipValidation {
		if c.validateRegion() != nil {
			errs = append(errs, fmt.Errorf("Unknown alicloud region: %s", c.AlicloudRegion))
		}
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

func (c *AlicloudAccessConfig) loadAndValidate() error {
	if err := c.validateRegion(); err != nil {
		return err
	}

	return nil
}

func (c *AlicloudAccessConfig) validateRegion() error {

	for _, valid := range common.ValidRegions {
		if c.AlicloudRegion == string(valid) {
			return nil
		}
	}

	return fmt.Errorf("Not a valid alicloud region: %s", c.AlicloudRegion)
}

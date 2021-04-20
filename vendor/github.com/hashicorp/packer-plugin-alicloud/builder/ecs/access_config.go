//go:generate packer-sdc struct-markdown

package ecs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/packer-plugin-alicloud/version"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mitchellh/go-homedir"
)

// Config of alicloud
type AlicloudAccessConfig struct {
	// Alicloud access key must be provided unless `profile` is set, but it can
	// also be sourced from the `ALICLOUD_ACCESS_KEY` environment variable.
	AlicloudAccessKey string `mapstructure:"access_key" required:"true"`
	// Alicloud secret key must be provided unless `profile` is set, but it can
	// also be sourced from the `ALICLOUD_SECRET_KEY` environment variable.
	AlicloudSecretKey string `mapstructure:"secret_key" required:"true"`
	// Alicloud region must be provided unless `profile` is set, but it can
	// also be sourced from the `ALICLOUD_REGION` environment variable.
	AlicloudRegion string `mapstructure:"region" required:"true"`
	// The region validation can be skipped if this value is true, the default
	// value is false.
	AlicloudSkipValidation bool `mapstructure:"skip_region_validation" required:"false"`
	// The image validation can be skipped if this value is true, the default
	// value is false.
	AlicloudSkipImageValidation bool `mapstructure:"skip_image_validation" required:"false"`
	// Alicloud profile must be set unless `access_key` is set; it can also be
	// sourced from the `ALICLOUD_PROFILE` environment variable.
	AlicloudProfile string `mapstructure:"profile" required:"false"`
	// Alicloud shared credentials file path. If this file exists, access and
	// secret keys will be read from this file.
	AlicloudSharedCredentialsFile string `mapstructure:"shared_credentials_file" required:"false"`
	// STS access token, can be set through template or by exporting as
	// environment variable such as `export SECURITY_TOKEN=value`.
	SecurityToken string `mapstructure:"security_token" required:"false"`

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

	var getProviderConfig = func(str string, key string) string {
		value, err := getConfigFromProfile(c, key)
		if err == nil && value != nil {
			str = value.(string)
		}
		return str
	}

	if c.AlicloudAccessKey == "" || c.AlicloudSecretKey == "" {
		c.AlicloudAccessKey = getProviderConfig(c.AlicloudAccessKey, "access_key_id")
		c.AlicloudSecretKey = getProviderConfig(c.AlicloudSecretKey, "access_key_secret")
		c.AlicloudRegion = getProviderConfig(c.AlicloudRegion, "region_id")
		c.SecurityToken = getProviderConfig(c.SecurityToken, "sts_token")
	}

	client, err := ecs.NewClientWithStsToken(c.AlicloudRegion, c.AlicloudAccessKey, c.AlicloudSecretKey, c.SecurityToken)
	if err != nil {
		return nil, err
	}

	client.AppendUserAgent(Packer, version.PluginVersion.FormattedVersion())
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
	if c.AlicloudProfile == "" {
		c.AlicloudProfile = os.Getenv("ALICLOUD_PROFILE")
	}
	if c.AlicloudSharedCredentialsFile == "" {
		c.AlicloudSharedCredentialsFile = os.Getenv("ALICLOUD_SHARED_CREDENTIALS_FILE")
	}
	if (c.AlicloudAccessKey == "" || c.AlicloudSecretKey == "") && c.AlicloudProfile == "" {
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

func getConfigFromProfile(c *AlicloudAccessConfig, ProfileKey string) (interface{}, error) {
	providerConfig := make(map[string]interface{})
	current := c.AlicloudProfile
	if current != "" {
		profilePath, err := homedir.Expand(c.AlicloudSharedCredentialsFile)
		if err != nil {
			return nil, err
		}
		if profilePath == "" {
			profilePath = fmt.Sprintf("%s/.aliyun/config.json", os.Getenv("HOME"))
			if runtime.GOOS == "windows" {
				profilePath = fmt.Sprintf("%s/.aliyun/config.json", os.Getenv("USERPROFILE"))
			}
		}
		_, err = os.Stat(profilePath)
		if !os.IsNotExist(err) {
			data, err := ioutil.ReadFile(profilePath)
			if err != nil {
				return nil, err
			}
			config := map[string]interface{}{}
			err = json.Unmarshal(data, &config)
			if err != nil {
				return nil, err
			}
			for _, v := range config["profiles"].([]interface{}) {
				if current == v.(map[string]interface{})["name"] {
					providerConfig = v.(map[string]interface{})
				}
			}
		}
	}
	mode := ""
	if v, ok := providerConfig["mode"]; ok {
		mode = v.(string)
	} else {
		return v, nil
	}
	switch ProfileKey {
	case "access_key_id", "access_key_secret":
		if mode == "EcsRamRole" {
			return "", nil
		}
	case "ram_role_name":
		if mode != "EcsRamRole" {
			return "", nil
		}
	case "sts_token":
		if mode != "StsToken" {
			return "", nil
		}
	case "ram_role_arn", "ram_session_name":
		if mode != "RamRoleArn" {
			return "", nil
		}
	case "expired_seconds":
		if mode != "RamRoleArn" {
			return float64(0), nil
		}
	}
	return providerConfig[ProfileKey], nil
}

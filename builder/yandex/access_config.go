//go:generate struct-markdown

package yandex

import (
	"errors"
	"fmt"
	"os"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"
	"github.com/yandex-cloud/go-sdk/iamkey"
)

const defaultEndpoint = "api.cloud.yandex.net:443"

// AccessConfig is for common configuration related to Yandex.Cloud API access
type AccessConfig struct {
	// Non standard API endpoint. Default is `api.cloud.yandex.net:443`.
	Endpoint string `mapstructure:"endpoint" required:"false"`
	// Path to file with Service Account key in json format. This
	// is an alternative method to authenticate to Yandex.Cloud. Alternatively you may set environment variable
	// `YC_SERVICE_ACCOUNT_KEY_FILE`.
	ServiceAccountKeyFile string `mapstructure:"service_account_key_file" required:"false"`
	// [OAuth token](https://cloud.yandex.com/docs/iam/concepts/authorization/oauth-token)
	// or [IAM token](https://cloud.yandex.com/docs/iam/concepts/authorization/iam-token)
	// to use to authenticate to Yandex.Cloud. Alternatively you may set
	// value by environment variable `YC_TOKEN`.
	Token string `mapstructure:"token" required:"true"`
	// The maximum number of times an API request is being executed.
	MaxRetries int `mapstructure:"max_retries"`
}

func (c *AccessConfig) Prepare(ctx *interpolate.Context) []error {
	var errs []error

	if c.Endpoint == "" {
		c.Endpoint = defaultEndpoint
	}

	// provision config by OS environment variables
	if c.Token == "" {
		c.Token = os.Getenv("YC_TOKEN")
	}

	if c.ServiceAccountKeyFile == "" {
		c.ServiceAccountKeyFile = os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE")
	}

	if c.Token != "" && c.ServiceAccountKeyFile != "" {
		errs = append(errs, errors.New("one of token or service account key file must be specified, not both"))
	}

	if c.Token != "" {
		packersdk.LogSecretFilter.Set(c.Token)
	}

	if c.ServiceAccountKeyFile != "" {
		if _, err := iamkey.ReadFromJSONFile(c.ServiceAccountKeyFile); err != nil {
			errs = append(errs, fmt.Errorf("fail to read service account key file: %s", err))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

//go:generate struct-markdown
package yandexexport

import (
	"fmt"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

type ExchangeConfig struct {
	// Service Account ID with proper permission to modify an instance, create and attach disk and
	// make upload to specific Yandex Object Storage paths.
	ServiceAccountID string `mapstructure:"service_account_id" required:"true"`
}

func (c *ExchangeConfig) Prepare(errs *packersdk.MultiError) *packersdk.MultiError {
	if c.ServiceAccountID == "" {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("service_account_id must be specified"))
	}

	return errs
}

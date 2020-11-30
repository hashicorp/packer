//go:generate struct-markdown
package yandexexport

import (
	"fmt"

	"github.com/hashicorp/packer/packer"
)

type ExchangeConfig struct {
	// Service Account ID with proper permission to modify an instance, create and attach disk and
	// make upload to specific Yandex Object Storage paths.
	ServiceAccountID string `mapstructure:"service_account_id" required:"true"`
}

func (c *ExchangeConfig) Prepare(errs *packer.MultiError) *packer.MultiError {
	if c.ServiceAccountID == "" {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("service_account_id must be specified"))
	}

	return errs
}

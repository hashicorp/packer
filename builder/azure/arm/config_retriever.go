package arm

import (
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/mitchellh/packer/builder/azure/common"
)

type configRetriever struct {
	// test seams
	findTenantID func(azure.Environment, string) (string, error)
}

func newConfigRetriever() configRetriever {
	return configRetriever{common.FindTenantID}
}

func (cr configRetriever) FillParameters(c *Config) error {
	if c.TenantID == "" {
		tenantID, err := cr.findTenantID(*c.cloudEnvironment, c.SubscriptionID)
		if err != nil {
			return err
		}
		c.TenantID = tenantID
	}
	return nil
}

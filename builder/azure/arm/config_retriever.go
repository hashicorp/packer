package arm

// Method to resolve information about the user so that a client can be
// constructed to communicated with Azure.
//
// The following data are resolved.
//
// 1. TenantID

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/builder/azure/common"
)

type configRetriever struct {
	// test seams
	findTenantID func(azure.Environment, string) (string, error)
}

func newConfigRetriever() configRetriever {
	return configRetriever{
		common.FindTenantID,
	}
}

func (cr configRetriever) FillParameters(c *Config) error {
	if c.SubscriptionID == "" {
		subscriptionID, err := cr.getSubscriptionFromIMDS()
		if err != nil {
			return err
		}
		c.SubscriptionID = subscriptionID
	}

	if c.TenantID == "" {
		tenantID, err := cr.findTenantID(*c.cloudEnvironment, c.SubscriptionID)
		if err != nil {
			return err
		}
		c.TenantID = tenantID
	}

	return nil
}

func (cr configRetriever) getSubscriptionFromIMDS() (string, error) {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", "http://169.254.169.254/metadata/instance/compute", nil)
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api-version", "2017-08-01")

	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	resp_body, _ := ioutil.ReadAll(resp.Body)
	result := map[string]string{}
	err = json.Unmarshal(resp_body, &result)
	if err != nil {
		return "", err
	}

	return result["subscriptionId"], nil
}

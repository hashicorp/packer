package chroot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
)

// DefaultMetadataClient is the default instance metadata client for Azure. Replace this variable for testing purposes only
var DefaultMetadataClient = NewMetadataClient()

// MetadataClient holds methods that Packer uses to get information about the current VM
type MetadataClient interface {
	VMResourceID() (string, error)
}

// metadataClient implements MetadataClient
type metadataClient struct{}

const imdsURL = "http://169.254.169.254/metadata/instance?api-version=2017-08-01"

// VMResourceID returns the resource ID of the current VM
func (metadataClient) VMResourceID() (string, error) {
	wc := retryablehttp.NewClient()
	wc.RetryMax = 5

	req, err := retryablehttp.NewRequest(http.MethodGet, imdsURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Metadata", "true")

	res, err := wc.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	d, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var vminfo struct {
		Compute struct {
			Name              string
			ResourceGroupName string
			SubscriptionID    string
		}
	}

	err = json.Unmarshal(d, &vminfo)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s",
		vminfo.Compute.Name,
		vminfo.Compute.ResourceGroupName,
		vminfo.Compute.SubscriptionID,
	), nil

}

// NewMetadataClient creates a new instance metadata client
func NewMetadataClient() MetadataClient {
	return metadataClient{}
}

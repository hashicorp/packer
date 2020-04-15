package client

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

// DefaultMetadataClient is the default instance metadata client for Azure. Replace this variable for testing purposes only
var DefaultMetadataClient = NewMetadataClient()

// MetadataClientAPI holds methods that Packer uses to get information about the current VM
type MetadataClientAPI interface {
	GetComputeInfo() (*ComputeInfo, error)
}

// MetadataClientStub is an easy way to put a test hook in DefaultMetadataClient
type MetadataClientStub struct {
	ComputeInfo
}

//GetComputeInfo implements MetadataClientAPI
func (s MetadataClientStub) GetComputeInfo() (*ComputeInfo, error) {
	return &s.ComputeInfo, nil
}

// ComputeInfo defines the Azure VM metadata that is used in Packer
type ComputeInfo struct {
	Name              string
	ResourceGroupName string
	SubscriptionID    string
	Location          string
}

// metadataClient implements MetadataClient
type metadataClient struct {
	autorest.Sender
	UserAgent string
}

var _ MetadataClientAPI = metadataClient{}

const imdsURL = "http://169.254.169.254/metadata/instance?api-version=2017-08-01"

// VMResourceID returns the resource ID of the current VM
func (client metadataClient) GetComputeInfo() (*ComputeInfo, error) {
	req, err := autorest.CreatePreparer(
		autorest.AsGet(),
		autorest.WithHeader("Metadata", "true"),
		autorest.WithUserAgent(client.UserAgent),
		autorest.WithBaseURL(imdsURL),
	).Prepare((&http.Request{}))
	if err != nil {
		return nil, err
	}

	res, err := autorest.SendWithSender(client, req,
		autorest.DoRetryForDuration(1*time.Minute, 5*time.Second))
	if err != nil {
		return nil, err
	}

	var vminfo struct {
		ComputeInfo `json:"compute"`
	}

	err = autorest.Respond(
		res,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&vminfo),
		autorest.ByClosing())
	if err != nil {
		return nil, err
	}
	return &vminfo.ComputeInfo, nil
}

func (ci ComputeInfo) ResourceID() string {
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s",
		ci.SubscriptionID,
		ci.ResourceGroupName,
		ci.Name,
	)
}

// NewMetadataClient creates a new instance metadata client
func NewMetadataClient() MetadataClientAPI {
	return metadataClient{
		Sender: autorest.CreateSender(),
	}
}

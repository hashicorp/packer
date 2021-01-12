package client

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/useragent"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute/computeapi"
	"github.com/Azure/go-autorest/autorest"
	version "github.com/hashicorp/packer/builder/azure/version"
)

type AzureClientSet interface {
	MetadataClient() MetadataClientAPI

	DisksClient() computeapi.DisksClientAPI
	SnapshotsClient() computeapi.SnapshotsClientAPI
	ImagesClient() computeapi.ImagesClientAPI

	GalleryImagesClient() computeapi.GalleryImagesClientAPI
	GalleryImageVersionsClient() computeapi.GalleryImageVersionsClientAPI

	VirtualMachinesClient() computeapi.VirtualMachinesClientAPI
	VirtualMachineImagesClient() VirtualMachineImagesClientAPI

	PollClient() autorest.Client

	// SubscriptionID returns the subscription ID that this client set was created for
	SubscriptionID() string
}

var subscriptionPathRegex = regexp.MustCompile(`/subscriptions/([[:xdigit:]]{8}(-[[:xdigit:]]{4}){3}-[[:xdigit:]]{12})`)

var _ AzureClientSet = &azureClientSet{}

type azureClientSet struct {
	sender         autorest.Sender
	authorizer     autorest.Authorizer
	subscriptionID string
	PollingDelay   time.Duration
}

func New(c Config, say func(string)) (AzureClientSet, error) {
	return new(c, say)
}

func new(c Config, say func(string)) (*azureClientSet, error) {
	token, err := c.GetServicePrincipalToken(say, c.CloudEnvironment().ResourceManagerEndpoint)
	if err != nil {
		return nil, err
	}
	return &azureClientSet{
		authorizer:     autorest.NewBearerAuthorizer(token),
		subscriptionID: c.SubscriptionID,
		sender:         http.DefaultClient,
		PollingDelay:   time.Second,
	}, nil
}

func (s azureClientSet) SubscriptionID() string {
	return s.subscriptionID
}

func (s azureClientSet) configureAutorestClient(c *autorest.Client) {
	err := c.AddToUserAgent(useragent.String(version.AzurePluginVersion.FormattedVersion()))
	if err != nil {
		log.Printf("Error appending Packer plugin version to user agent.")
	}
	c.Authorizer = s.authorizer
	c.Sender = s.sender
}

func (s azureClientSet) MetadataClient() MetadataClientAPI {
	return metadataClient{
		s.sender,
		useragent.String(version.AzurePluginVersion.FormattedVersion()),
	}
}

func (s azureClientSet) DisksClient() computeapi.DisksClientAPI {
	c := compute.NewDisksClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) SnapshotsClient() computeapi.SnapshotsClientAPI {
	c := compute.NewSnapshotsClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) ImagesClient() computeapi.ImagesClientAPI {
	c := compute.NewImagesClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) VirtualMachinesClient() computeapi.VirtualMachinesClientAPI {
	c := compute.NewVirtualMachinesClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) VirtualMachineImagesClient() VirtualMachineImagesClientAPI {
	c := compute.NewVirtualMachineImagesClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return VirtualMachineImagesClient{c}
}

func (s azureClientSet) GalleryImagesClient() computeapi.GalleryImagesClientAPI {
	c := compute.NewGalleryImagesClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) GalleryImageVersionsClient() computeapi.GalleryImageVersionsClientAPI {
	c := compute.NewGalleryImageVersionsClient(s.subscriptionID)
	s.configureAutorestClient(&c.Client)
	c.PollingDelay = s.PollingDelay
	return c
}

func (s azureClientSet) PollClient() autorest.Client {
	c := autorest.NewClientWithUserAgent("Packer-Azure-ClientSet")
	s.configureAutorestClient(&c)
	c.PollingDelay = time.Second * 5
	return c
}

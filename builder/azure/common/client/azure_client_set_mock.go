package client

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute/computeapi"
	"github.com/Azure/go-autorest/autorest"
)

var _ AzureClientSet = &AzureClientSetMock{}

// AzureClientSetMock provides a generic mock for AzureClientSet
type AzureClientSetMock struct {
	DisksClientMock                computeapi.DisksClientAPI
	SnapshotsClientMock            computeapi.SnapshotsClientAPI
	ImagesClientMock               computeapi.ImagesClientAPI
	VirtualMachineImagesClientMock VirtualMachineImagesClientAPI
	VirtualMachinesClientMock      computeapi.VirtualMachinesClientAPI
	GalleryImagesClientMock        computeapi.GalleryImagesClientAPI
	GalleryImageVersionsClientMock computeapi.GalleryImageVersionsClientAPI
	PollClientMock                 autorest.Client
	MetadataClientMock             MetadataClientAPI
	SubscriptionIDMock             string
}

// DisksClient returns a DisksClientAPI
func (m *AzureClientSetMock) DisksClient() computeapi.DisksClientAPI {
	return m.DisksClientMock
}

// SnapshotsClient returns a SnapshotsClientAPI
func (m *AzureClientSetMock) SnapshotsClient() computeapi.SnapshotsClientAPI {
	return m.SnapshotsClientMock
}

// ImagesClient returns a ImagesClientAPI
func (m *AzureClientSetMock) ImagesClient() computeapi.ImagesClientAPI {
	return m.ImagesClientMock
}

// VirtualMachineImagesClient returns a VirtualMachineImagesClientAPI
func (m *AzureClientSetMock) VirtualMachineImagesClient() VirtualMachineImagesClientAPI {
	return m.VirtualMachineImagesClientMock
}

// VirtualMachinesClient returns a VirtualMachinesClientAPI
func (m *AzureClientSetMock) VirtualMachinesClient() computeapi.VirtualMachinesClientAPI {
	return m.VirtualMachinesClientMock
}

// GalleryImagesClient returns a GalleryImagesClientAPI
func (m *AzureClientSetMock) GalleryImagesClient() computeapi.GalleryImagesClientAPI {
	return m.GalleryImagesClientMock
}

// GalleryImageVersionsClient returns a GalleryImageVersionsClientAPI
func (m *AzureClientSetMock) GalleryImageVersionsClient() computeapi.GalleryImageVersionsClientAPI {
	return m.GalleryImageVersionsClientMock
}

// PollClient returns an autorest Client that can be used for polling async requests
func (m *AzureClientSetMock) PollClient() autorest.Client {
	return m.PollClientMock
}

// MetadataClient returns a MetadataClientAPI
func (m *AzureClientSetMock) MetadataClient() MetadataClientAPI {
	return m.MetadataClientMock
}

// SubscriptionID returns SubscriptionIDMock
func (m *AzureClientSetMock) SubscriptionID() string {
	return m.SubscriptionIDMock
}

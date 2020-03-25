package client

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-07-01/compute/computeapi"
	"github.com/Azure/go-autorest/autorest"
)

// AzureClientSetMock provides a generic mock for AzureClientSet
type AzureClientSetMock struct {
	DisksClientMock                computeapi.DisksClientAPI
	ImagesClientMock               computeapi.ImagesClientAPI
	VirtualMachineImagesClientMock VirtualMachineImagesClientAPI
	VirtualMachinesClientMock      computeapi.VirtualMachinesClientAPI
	PollClientMock                 autorest.Client
	MetadataClientMock             MetadataClientAPI
}

// DisksClient returns a DisksClientAPI
func (m *AzureClientSetMock) DisksClient() computeapi.DisksClientAPI {
	return m.DisksClientMock
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

// PollClient returns an autorest Client that can be used for polling async requests
func (m *AzureClientSetMock) PollClient() autorest.Client {
	return m.PollClientMock
}

// MetadataClient returns a MetadataClientAPI
func (m *AzureClientSetMock) MetadataClient() MetadataClientAPI {
	return m.MetadataClientMock
}

// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Core Services API
//
// APIs for Networking Service, Compute Service, and Block Volume Service.
//

package core

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//BlockstorageClient a client for Blockstorage
type BlockstorageClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewBlockstorageClientWithConfigurationProvider Creates a new default Blockstorage client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewBlockstorageClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client BlockstorageClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	client = BlockstorageClient{BaseClient: baseClient}
	client.BasePath = "20160918"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *BlockstorageClient) SetRegion(region string) {
	client.Host = fmt.Sprintf(common.DefaultHostURLTemplate, "iaas", region)
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *BlockstorageClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
	if ok, err := common.IsConfigurationProviderValid(configProvider); !ok {
		return err
	}

	// Error has been checked already
	region, _ := configProvider.Region()
	client.config = &configProvider
	client.SetRegion(region)
	return nil
}

// ConfigurationProvider the ConfigurationProvider used in this client, or null if none set
func (client *BlockstorageClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// CreateVolume Creates a new volume in the specified compartment. Volumes can be created in sizes ranging from
// 50 GB (51200 MB) to 16 TB (16777216 MB), in 1 GB (1024 MB) increments. By default, volumes are 1 TB (1048576 MB).
// For general information about block volumes, see
// Overview of Block Volume Service (https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/overview.htm).
// A volume and instance can be in separate compartments but must be in the same Availability Domain.
// For information about access control and compartments, see
// Overview of the IAM Service (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/overview.htm). For information about
// Availability Domains, see Regions and Availability Domains (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/regions.htm).
// To get a list of Availability Domains, use the `ListAvailabilityDomains` operation
// in the Identity and Access Management Service API.
// You may optionally specify a *display name* for the volume, which is simply a friendly name or
// description. It does not have to be unique, and you can change it. Avoid entering confidential information.
func (client BlockstorageClient) CreateVolume(ctx context.Context, request CreateVolumeRequest) (response CreateVolumeResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodPost, "/volumes", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// CreateVolumeBackup Creates a new backup of the specified volume. For general information about volume backups,
// see Overview of Block Volume Service Backups (https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Concepts/blockvolumebackups.htm)
// When the request is received, the backup object is in a REQUEST_RECEIVED state.
// When the data is imaged, it goes into a CREATING state.
// After the backup is fully uploaded to the cloud, it goes into an AVAILABLE state.
func (client BlockstorageClient) CreateVolumeBackup(ctx context.Context, request CreateVolumeBackupRequest) (response CreateVolumeBackupResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodPost, "/volumeBackups", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// DeleteBootVolume Deletes the specified boot volume. The volume cannot have an active connection to an instance.
// To disconnect the boot volume from a connected instance, see
// Disconnecting From a Boot Volume (https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Tasks/deletingbootvolume.htm).
// **Warning:** All data on the boot volume will be permanently lost when the boot volume is deleted.
func (client BlockstorageClient) DeleteBootVolume(ctx context.Context, request DeleteBootVolumeRequest) (response DeleteBootVolumeResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodDelete, "/bootVolumes/{bootVolumeId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// DeleteVolume Deletes the specified volume. The volume cannot have an active connection to an instance.
// To disconnect the volume from a connected instance, see
// Disconnecting From a Volume (https://docs.us-phoenix-1.oraclecloud.com/Content/Block/Tasks/disconnectingfromavolume.htm).
// **Warning:** All data on the volume will be permanently lost when the volume is deleted.
func (client BlockstorageClient) DeleteVolume(ctx context.Context, request DeleteVolumeRequest) (response DeleteVolumeResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodDelete, "/volumes/{volumeId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// DeleteVolumeBackup Deletes a volume backup.
func (client BlockstorageClient) DeleteVolumeBackup(ctx context.Context, request DeleteVolumeBackupRequest) (response DeleteVolumeBackupResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodDelete, "/volumeBackups/{volumeBackupId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// GetBootVolume Gets information for the specified boot volume.
func (client BlockstorageClient) GetBootVolume(ctx context.Context, request GetBootVolumeRequest) (response GetBootVolumeResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodGet, "/bootVolumes/{bootVolumeId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// GetVolume Gets information for the specified volume.
func (client BlockstorageClient) GetVolume(ctx context.Context, request GetVolumeRequest) (response GetVolumeResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodGet, "/volumes/{volumeId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// GetVolumeBackup Gets information for the specified volume backup.
func (client BlockstorageClient) GetVolumeBackup(ctx context.Context, request GetVolumeBackupRequest) (response GetVolumeBackupResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodGet, "/volumeBackups/{volumeBackupId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// ListBootVolumes Lists the boot volumes in the specified compartment and Availability Domain.
func (client BlockstorageClient) ListBootVolumes(ctx context.Context, request ListBootVolumesRequest) (response ListBootVolumesResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodGet, "/bootVolumes", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// ListVolumeBackups Lists the volume backups in the specified compartment. You can filter the results by volume.
func (client BlockstorageClient) ListVolumeBackups(ctx context.Context, request ListVolumeBackupsRequest) (response ListVolumeBackupsResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodGet, "/volumeBackups", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// ListVolumes Lists the volumes in the specified compartment and Availability Domain.
func (client BlockstorageClient) ListVolumes(ctx context.Context, request ListVolumesRequest) (response ListVolumesResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodGet, "/volumes", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// UpdateBootVolume Updates the specified boot volume's display name.
func (client BlockstorageClient) UpdateBootVolume(ctx context.Context, request UpdateBootVolumeRequest) (response UpdateBootVolumeResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodPut, "/bootVolumes/{bootVolumeId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// UpdateVolume Updates the specified volume's display name.
// Avoid entering confidential information.
func (client BlockstorageClient) UpdateVolume(ctx context.Context, request UpdateVolumeRequest) (response UpdateVolumeResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodPut, "/volumes/{volumeId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

// UpdateVolumeBackup Updates the display name for the specified volume backup.
// Avoid entering confidential information.
func (client BlockstorageClient) UpdateVolumeBackup(ctx context.Context, request UpdateVolumeBackupRequest) (response UpdateVolumeBackupResponse, err error) {
	httpRequest, err := common.MakeDefaultHTTPRequestWithTaggedStruct(http.MethodPut, "/volumeBackups/{volumeBackupId}", request)
	if err != nil {
		return
	}

	httpResponse, err := client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return
}

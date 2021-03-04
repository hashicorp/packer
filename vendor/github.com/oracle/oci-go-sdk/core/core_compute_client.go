// Copyright (c) 2016, 2018, 2020, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"net/http"
)

//ComputeClient a client for Compute
type ComputeClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewComputeClientWithConfigurationProvider Creates a new default Compute client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewComputeClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client ComputeClient, err error) {
	baseClient, err := common.NewClientWithConfig(configProvider)
	if err != nil {
		return
	}

	return newComputeClientFromBaseClient(baseClient, configProvider)
}

// NewComputeClientWithOboToken Creates a new default Compute client with the given configuration provider.
// The obotoken will be added to default headers and signed; the configuration provider will be used for the signer
//  as well as reading the region
func NewComputeClientWithOboToken(configProvider common.ConfigurationProvider, oboToken string) (client ComputeClient, err error) {
	baseClient, err := common.NewClientWithOboToken(configProvider, oboToken)
	if err != nil {
		return
	}

	return newComputeClientFromBaseClient(baseClient, configProvider)
}

func newComputeClientFromBaseClient(baseClient common.BaseClient, configProvider common.ConfigurationProvider) (client ComputeClient, err error) {
	client = ComputeClient{BaseClient: baseClient}
	client.BasePath = "20160918"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *ComputeClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).EndpointForTemplate("iaas", "https://iaas.{region}.{secondLevelDomain}")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *ComputeClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
	if ok, err := common.IsConfigurationProviderValid(configProvider); !ok {
		return err
	}

	// Error has been checked already
	region, _ := configProvider.Region()
	client.SetRegion(region)
	client.config = &configProvider
	return nil
}

// ConfigurationProvider the ConfigurationProvider used in this client, or null if none set
func (client *ComputeClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// AddImageShapeCompatibilityEntry Adds a shape to the compatible shapes list for the image.
func (client ComputeClient) AddImageShapeCompatibilityEntry(ctx context.Context, request AddImageShapeCompatibilityEntryRequest) (response AddImageShapeCompatibilityEntryResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.addImageShapeCompatibilityEntry, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AddImageShapeCompatibilityEntryResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AddImageShapeCompatibilityEntryResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AddImageShapeCompatibilityEntryResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AddImageShapeCompatibilityEntryResponse")
	}
	return
}

// addImageShapeCompatibilityEntry implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) addImageShapeCompatibilityEntry(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/images/{imageId}/shapes/{shapeName}")
	if err != nil {
		return nil, err
	}

	var response AddImageShapeCompatibilityEntryResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// AttachBootVolume Attaches the specified boot volume to the specified instance.
func (client ComputeClient) AttachBootVolume(ctx context.Context, request AttachBootVolumeRequest) (response AttachBootVolumeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.attachBootVolume, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AttachBootVolumeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AttachBootVolumeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AttachBootVolumeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AttachBootVolumeResponse")
	}
	return
}

// attachBootVolume implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) attachBootVolume(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/bootVolumeAttachments")
	if err != nil {
		return nil, err
	}

	var response AttachBootVolumeResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// AttachVnic Creates a secondary VNIC and attaches it to the specified instance.
// For more information about secondary VNICs, see
// Virtual Network Interface Cards (VNICs) (https://docs.cloud.oracle.com/Content/Network/Tasks/managingVNICs.htm).
func (client ComputeClient) AttachVnic(ctx context.Context, request AttachVnicRequest) (response AttachVnicResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.attachVnic, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AttachVnicResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AttachVnicResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AttachVnicResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AttachVnicResponse")
	}
	return
}

// attachVnic implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) attachVnic(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vnicAttachments")
	if err != nil {
		return nil, err
	}

	var response AttachVnicResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// AttachVolume Attaches the specified storage volume to the specified instance.
func (client ComputeClient) AttachVolume(ctx context.Context, request AttachVolumeRequest) (response AttachVolumeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.attachVolume, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AttachVolumeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AttachVolumeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AttachVolumeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AttachVolumeResponse")
	}
	return
}

// attachVolume implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) attachVolume(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/volumeAttachments")
	if err != nil {
		return nil, err
	}

	var response AttachVolumeResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponseWithPolymorphicBody(httpResponse, &response, &volumeattachment{})
	return response, err
}

// CaptureConsoleHistory Captures the most recent serial console data (up to a megabyte) for the
// specified instance.
// The `CaptureConsoleHistory` operation works with the other console history operations
// as described below.
// 1. Use `CaptureConsoleHistory` to request the capture of up to a megabyte of the
// most recent console history. This call returns a `ConsoleHistory`
// object. The object will have a state of REQUESTED.
// 2. Wait for the capture operation to succeed by polling `GetConsoleHistory` with
// the identifier of the console history metadata. The state of the
// `ConsoleHistory` object will go from REQUESTED to GETTING-HISTORY and
// then SUCCEEDED (or FAILED).
// 3. Use `GetConsoleHistoryContent` to get the actual console history data (not the
// metadata).
// 4. Optionally, use `DeleteConsoleHistory` to delete the console history metadata
// and the console history data.
func (client ComputeClient) CaptureConsoleHistory(ctx context.Context, request CaptureConsoleHistoryRequest) (response CaptureConsoleHistoryResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.captureConsoleHistory, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CaptureConsoleHistoryResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CaptureConsoleHistoryResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CaptureConsoleHistoryResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CaptureConsoleHistoryResponse")
	}
	return
}

// captureConsoleHistory implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) captureConsoleHistory(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/instanceConsoleHistories")
	if err != nil {
		return nil, err
	}

	var response CaptureConsoleHistoryResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ChangeComputeImageCapabilitySchemaCompartment Moves a compute image capability schema into a different compartment within the same tenancy.
// For information about moving resources between compartments, see
//         Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
func (client ComputeClient) ChangeComputeImageCapabilitySchemaCompartment(ctx context.Context, request ChangeComputeImageCapabilitySchemaCompartmentRequest) (response ChangeComputeImageCapabilitySchemaCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeComputeImageCapabilitySchemaCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeComputeImageCapabilitySchemaCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeComputeImageCapabilitySchemaCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeComputeImageCapabilitySchemaCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeComputeImageCapabilitySchemaCompartmentResponse")
	}
	return
}

// changeComputeImageCapabilitySchemaCompartment implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) changeComputeImageCapabilitySchemaCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/computeImageCapabilitySchemas/{computeImageCapabilitySchemaId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeComputeImageCapabilitySchemaCompartmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ChangeDedicatedVmHostCompartment Moves a dedicated virtual machine host from one compartment to another.
func (client ComputeClient) ChangeDedicatedVmHostCompartment(ctx context.Context, request ChangeDedicatedVmHostCompartmentRequest) (response ChangeDedicatedVmHostCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeDedicatedVmHostCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeDedicatedVmHostCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeDedicatedVmHostCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeDedicatedVmHostCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeDedicatedVmHostCompartmentResponse")
	}
	return
}

// changeDedicatedVmHostCompartment implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) changeDedicatedVmHostCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/dedicatedVmHosts/{dedicatedVmHostId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeDedicatedVmHostCompartmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ChangeImageCompartment Moves an image into a different compartment within the same tenancy. For information about moving
// resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
func (client ComputeClient) ChangeImageCompartment(ctx context.Context, request ChangeImageCompartmentRequest) (response ChangeImageCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeImageCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeImageCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeImageCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeImageCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeImageCompartmentResponse")
	}
	return
}

// changeImageCompartment implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) changeImageCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/images/{imageId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeImageCompartmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ChangeInstanceCompartment Moves an instance into a different compartment within the same tenancy. For information about
// moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
// When you move an instance to a different compartment, associated resources such as boot volumes and VNICs
// are not moved.
func (client ComputeClient) ChangeInstanceCompartment(ctx context.Context, request ChangeInstanceCompartmentRequest) (response ChangeInstanceCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeInstanceCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeInstanceCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeInstanceCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeInstanceCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeInstanceCompartmentResponse")
	}
	return
}

// changeInstanceCompartment implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) changeInstanceCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/instances/{instanceId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeInstanceCompartmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateAppCatalogSubscription Create a subscription for listing resource version for a compartment. It will take some time to propagate to all regions.
func (client ComputeClient) CreateAppCatalogSubscription(ctx context.Context, request CreateAppCatalogSubscriptionRequest) (response CreateAppCatalogSubscriptionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createAppCatalogSubscription, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateAppCatalogSubscriptionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateAppCatalogSubscriptionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateAppCatalogSubscriptionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateAppCatalogSubscriptionResponse")
	}
	return
}

// createAppCatalogSubscription implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) createAppCatalogSubscription(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/appCatalogSubscriptions")
	if err != nil {
		return nil, err
	}

	var response CreateAppCatalogSubscriptionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateComputeImageCapabilitySchema Creates compute image capability schema.
func (client ComputeClient) CreateComputeImageCapabilitySchema(ctx context.Context, request CreateComputeImageCapabilitySchemaRequest) (response CreateComputeImageCapabilitySchemaResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createComputeImageCapabilitySchema, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateComputeImageCapabilitySchemaResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateComputeImageCapabilitySchemaResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateComputeImageCapabilitySchemaResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateComputeImageCapabilitySchemaResponse")
	}
	return
}

// createComputeImageCapabilitySchema implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) createComputeImageCapabilitySchema(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/computeImageCapabilitySchemas")
	if err != nil {
		return nil, err
	}

	var response CreateComputeImageCapabilitySchemaResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateDedicatedVmHost Creates a new dedicated virtual machine host in the specified compartment and the specified availability domain.
// Dedicated virtual machine hosts enable you to run your Compute virtual machine (VM) instances on dedicated servers
// that are a single tenant and not shared with other customers.
// For more information, see Dedicated Virtual Machine Hosts (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/dedicatedvmhosts.htm).
func (client ComputeClient) CreateDedicatedVmHost(ctx context.Context, request CreateDedicatedVmHostRequest) (response CreateDedicatedVmHostResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createDedicatedVmHost, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateDedicatedVmHostResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateDedicatedVmHostResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateDedicatedVmHostResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateDedicatedVmHostResponse")
	}
	return
}

// createDedicatedVmHost implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) createDedicatedVmHost(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/dedicatedVmHosts")
	if err != nil {
		return nil, err
	}

	var response CreateDedicatedVmHostResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateImage Creates a boot disk image for the specified instance or imports an exported image from the Oracle Cloud Infrastructure Object Storage service.
// When creating a new image, you must provide the OCID of the instance you want to use as the basis for the image, and
// the OCID of the compartment containing that instance. For more information about images,
// see Managing Custom Images (https://docs.cloud.oracle.com/Content/Compute/Tasks/managingcustomimages.htm).
// When importing an exported image from Object Storage, you specify the source information
// in ImageSourceDetails.
// When importing an image based on the namespace, bucket name, and object name,
// use ImageSourceViaObjectStorageTupleDetails.
// When importing an image based on the Object Storage URL, use
// ImageSourceViaObjectStorageUriDetails.
// See Object Storage URLs (https://docs.cloud.oracle.com/Content/Compute/Tasks/imageimportexport.htm#URLs) and Using Pre-Authenticated Requests (https://docs.cloud.oracle.com/Content/Object/Tasks/usingpreauthenticatedrequests.htm)
// for constructing URLs for image import/export.
// For more information about importing exported images, see
// Image Import/Export (https://docs.cloud.oracle.com/Content/Compute/Tasks/imageimportexport.htm).
// You may optionally specify a *display name* for the image, which is simply a friendly name or description.
// It does not have to be unique, and you can change it. See UpdateImage.
// Avoid entering confidential information.
func (client ComputeClient) CreateImage(ctx context.Context, request CreateImageRequest) (response CreateImageResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createImage, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateImageResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateImageResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateImageResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateImageResponse")
	}
	return
}

// createImage implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) createImage(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/images")
	if err != nil {
		return nil, err
	}

	var response CreateImageResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// CreateInstanceConsoleConnection Creates a new console connection to the specified instance.
// After the console connection has been created and is available,
// you connect to the console using SSH.
// For more information about console access, see Accessing the Console (https://docs.cloud.oracle.com/Content/Compute/References/serialconsole.htm).
func (client ComputeClient) CreateInstanceConsoleConnection(ctx context.Context, request CreateInstanceConsoleConnectionRequest) (response CreateInstanceConsoleConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createInstanceConsoleConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateInstanceConsoleConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateInstanceConsoleConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateInstanceConsoleConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateInstanceConsoleConnectionResponse")
	}
	return
}

// createInstanceConsoleConnection implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) createInstanceConsoleConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/instanceConsoleConnections")
	if err != nil {
		return nil, err
	}

	var response CreateInstanceConsoleConnectionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteAppCatalogSubscription Delete a subscription for a listing resource version for a compartment.
func (client ComputeClient) DeleteAppCatalogSubscription(ctx context.Context, request DeleteAppCatalogSubscriptionRequest) (response DeleteAppCatalogSubscriptionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteAppCatalogSubscription, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteAppCatalogSubscriptionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteAppCatalogSubscriptionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteAppCatalogSubscriptionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteAppCatalogSubscriptionResponse")
	}
	return
}

// deleteAppCatalogSubscription implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) deleteAppCatalogSubscription(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/appCatalogSubscriptions")
	if err != nil {
		return nil, err
	}

	var response DeleteAppCatalogSubscriptionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteComputeImageCapabilitySchema Deletes the specified Compute Image Capability Schema
func (client ComputeClient) DeleteComputeImageCapabilitySchema(ctx context.Context, request DeleteComputeImageCapabilitySchemaRequest) (response DeleteComputeImageCapabilitySchemaResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteComputeImageCapabilitySchema, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteComputeImageCapabilitySchemaResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteComputeImageCapabilitySchemaResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteComputeImageCapabilitySchemaResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteComputeImageCapabilitySchemaResponse")
	}
	return
}

// deleteComputeImageCapabilitySchema implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) deleteComputeImageCapabilitySchema(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/computeImageCapabilitySchemas/{computeImageCapabilitySchemaId}")
	if err != nil {
		return nil, err
	}

	var response DeleteComputeImageCapabilitySchemaResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteConsoleHistory Deletes the specified console history metadata and the console history data.
func (client ComputeClient) DeleteConsoleHistory(ctx context.Context, request DeleteConsoleHistoryRequest) (response DeleteConsoleHistoryResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteConsoleHistory, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteConsoleHistoryResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteConsoleHistoryResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteConsoleHistoryResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteConsoleHistoryResponse")
	}
	return
}

// deleteConsoleHistory implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) deleteConsoleHistory(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/instanceConsoleHistories/{instanceConsoleHistoryId}")
	if err != nil {
		return nil, err
	}

	var response DeleteConsoleHistoryResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteDedicatedVmHost Deletes the specified dedicated virtual machine host.
// If any VM instances are assigned to the dedicated virtual machine host,
// the delete operation will fail and the service will return a 409 response code.
func (client ComputeClient) DeleteDedicatedVmHost(ctx context.Context, request DeleteDedicatedVmHostRequest) (response DeleteDedicatedVmHostResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteDedicatedVmHost, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteDedicatedVmHostResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteDedicatedVmHostResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteDedicatedVmHostResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteDedicatedVmHostResponse")
	}
	return
}

// deleteDedicatedVmHost implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) deleteDedicatedVmHost(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/dedicatedVmHosts/{dedicatedVmHostId}")
	if err != nil {
		return nil, err
	}

	var response DeleteDedicatedVmHostResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteImage Deletes an image.
func (client ComputeClient) DeleteImage(ctx context.Context, request DeleteImageRequest) (response DeleteImageResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteImage, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteImageResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteImageResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteImageResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteImageResponse")
	}
	return
}

// deleteImage implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) deleteImage(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/images/{imageId}")
	if err != nil {
		return nil, err
	}

	var response DeleteImageResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DeleteInstanceConsoleConnection Deletes the specified instance console connection.
func (client ComputeClient) DeleteInstanceConsoleConnection(ctx context.Context, request DeleteInstanceConsoleConnectionRequest) (response DeleteInstanceConsoleConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteInstanceConsoleConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteInstanceConsoleConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteInstanceConsoleConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteInstanceConsoleConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteInstanceConsoleConnectionResponse")
	}
	return
}

// deleteInstanceConsoleConnection implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) deleteInstanceConsoleConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/instanceConsoleConnections/{instanceConsoleConnectionId}")
	if err != nil {
		return nil, err
	}

	var response DeleteInstanceConsoleConnectionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DetachBootVolume Detaches a boot volume from an instance. You must specify the OCID of the boot volume attachment.
// This is an asynchronous operation. The attachment's `lifecycleState` will change to DETACHING temporarily
// until the attachment is completely removed.
func (client ComputeClient) DetachBootVolume(ctx context.Context, request DetachBootVolumeRequest) (response DetachBootVolumeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.detachBootVolume, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DetachBootVolumeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DetachBootVolumeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DetachBootVolumeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DetachBootVolumeResponse")
	}
	return
}

// detachBootVolume implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) detachBootVolume(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/bootVolumeAttachments/{bootVolumeAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response DetachBootVolumeResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DetachVnic Detaches and deletes the specified secondary VNIC.
// This operation cannot be used on the instance's primary VNIC.
// When you terminate an instance, all attached VNICs (primary
// and secondary) are automatically detached and deleted.
// **Important:** If the VNIC has a
// PrivateIp that is the
// target of a route rule (https://docs.cloud.oracle.com/Content/Network/Tasks/managingroutetables.htm#privateip),
// deleting the VNIC causes that route rule to blackhole and the traffic
// will be dropped.
func (client ComputeClient) DetachVnic(ctx context.Context, request DetachVnicRequest) (response DetachVnicResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.detachVnic, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DetachVnicResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DetachVnicResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DetachVnicResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DetachVnicResponse")
	}
	return
}

// detachVnic implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) detachVnic(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/vnicAttachments/{vnicAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response DetachVnicResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// DetachVolume Detaches a storage volume from an instance. You must specify the OCID of the volume attachment.
// This is an asynchronous operation. The attachment's `lifecycleState` will change to DETACHING temporarily
// until the attachment is completely removed.
func (client ComputeClient) DetachVolume(ctx context.Context, request DetachVolumeRequest) (response DetachVolumeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.detachVolume, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DetachVolumeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DetachVolumeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DetachVolumeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DetachVolumeResponse")
	}
	return
}

// detachVolume implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) detachVolume(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/volumeAttachments/{volumeAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response DetachVolumeResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ExportImage Exports the specified image to the Oracle Cloud Infrastructure Object Storage service. You can use the Object Storage URL,
// or the namespace, bucket name, and object name when specifying the location to export to.
// For more information about exporting images, see Image Import/Export (https://docs.cloud.oracle.com/Content/Compute/Tasks/imageimportexport.htm).
// To perform an image export, you need write access to the Object Storage bucket for the image,
// see Let Users Write Objects to Object Storage Buckets (https://docs.cloud.oracle.com/Content/Identity/Concepts/commonpolicies.htm#Let4).
// See Object Storage URLs (https://docs.cloud.oracle.com/Content/Compute/Tasks/imageimportexport.htm#URLs) and Using Pre-Authenticated Requests (https://docs.cloud.oracle.com/Content/Object/Tasks/usingpreauthenticatedrequests.htm)
// for constructing URLs for image import/export.
func (client ComputeClient) ExportImage(ctx context.Context, request ExportImageRequest) (response ExportImageResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.exportImage, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ExportImageResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ExportImageResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ExportImageResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ExportImageResponse")
	}
	return
}

// exportImage implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) exportImage(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/images/{imageId}/actions/export")
	if err != nil {
		return nil, err
	}

	var response ExportImageResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetAppCatalogListing Gets the specified listing.
func (client ComputeClient) GetAppCatalogListing(ctx context.Context, request GetAppCatalogListingRequest) (response GetAppCatalogListingResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getAppCatalogListing, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetAppCatalogListingResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetAppCatalogListingResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetAppCatalogListingResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetAppCatalogListingResponse")
	}
	return
}

// getAppCatalogListing implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getAppCatalogListing(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/appCatalogListings/{listingId}")
	if err != nil {
		return nil, err
	}

	var response GetAppCatalogListingResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetAppCatalogListingAgreements Retrieves the agreements for a particular resource version of a listing.
func (client ComputeClient) GetAppCatalogListingAgreements(ctx context.Context, request GetAppCatalogListingAgreementsRequest) (response GetAppCatalogListingAgreementsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getAppCatalogListingAgreements, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetAppCatalogListingAgreementsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetAppCatalogListingAgreementsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetAppCatalogListingAgreementsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetAppCatalogListingAgreementsResponse")
	}
	return
}

// getAppCatalogListingAgreements implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getAppCatalogListingAgreements(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/appCatalogListings/{listingId}/resourceVersions/{resourceVersion}/agreements")
	if err != nil {
		return nil, err
	}

	var response GetAppCatalogListingAgreementsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetAppCatalogListingResourceVersion Gets the specified listing resource version.
func (client ComputeClient) GetAppCatalogListingResourceVersion(ctx context.Context, request GetAppCatalogListingResourceVersionRequest) (response GetAppCatalogListingResourceVersionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getAppCatalogListingResourceVersion, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetAppCatalogListingResourceVersionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetAppCatalogListingResourceVersionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetAppCatalogListingResourceVersionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetAppCatalogListingResourceVersionResponse")
	}
	return
}

// getAppCatalogListingResourceVersion implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getAppCatalogListingResourceVersion(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/appCatalogListings/{listingId}/resourceVersions/{resourceVersion}")
	if err != nil {
		return nil, err
	}

	var response GetAppCatalogListingResourceVersionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetBootVolumeAttachment Gets information about the specified boot volume attachment.
func (client ComputeClient) GetBootVolumeAttachment(ctx context.Context, request GetBootVolumeAttachmentRequest) (response GetBootVolumeAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getBootVolumeAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetBootVolumeAttachmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetBootVolumeAttachmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetBootVolumeAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetBootVolumeAttachmentResponse")
	}
	return
}

// getBootVolumeAttachment implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getBootVolumeAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/bootVolumeAttachments/{bootVolumeAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response GetBootVolumeAttachmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetComputeGlobalImageCapabilitySchema Gets the specified Compute Global Image Capability Schema
func (client ComputeClient) GetComputeGlobalImageCapabilitySchema(ctx context.Context, request GetComputeGlobalImageCapabilitySchemaRequest) (response GetComputeGlobalImageCapabilitySchemaResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getComputeGlobalImageCapabilitySchema, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetComputeGlobalImageCapabilitySchemaResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetComputeGlobalImageCapabilitySchemaResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetComputeGlobalImageCapabilitySchemaResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetComputeGlobalImageCapabilitySchemaResponse")
	}
	return
}

// getComputeGlobalImageCapabilitySchema implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getComputeGlobalImageCapabilitySchema(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/computeGlobalImageCapabilitySchemas/{computeGlobalImageCapabilitySchemaId}")
	if err != nil {
		return nil, err
	}

	var response GetComputeGlobalImageCapabilitySchemaResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetComputeGlobalImageCapabilitySchemaVersion Gets the specified Compute Global Image Capability Schema Version
func (client ComputeClient) GetComputeGlobalImageCapabilitySchemaVersion(ctx context.Context, request GetComputeGlobalImageCapabilitySchemaVersionRequest) (response GetComputeGlobalImageCapabilitySchemaVersionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getComputeGlobalImageCapabilitySchemaVersion, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetComputeGlobalImageCapabilitySchemaVersionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetComputeGlobalImageCapabilitySchemaVersionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetComputeGlobalImageCapabilitySchemaVersionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetComputeGlobalImageCapabilitySchemaVersionResponse")
	}
	return
}

// getComputeGlobalImageCapabilitySchemaVersion implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getComputeGlobalImageCapabilitySchemaVersion(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/computeGlobalImageCapabilitySchemas/{computeGlobalImageCapabilitySchemaId}/versions/{computeGlobalImageCapabilitySchemaVersionName}")
	if err != nil {
		return nil, err
	}

	var response GetComputeGlobalImageCapabilitySchemaVersionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetComputeImageCapabilitySchema Gets the specified Compute Image Capability Schema
func (client ComputeClient) GetComputeImageCapabilitySchema(ctx context.Context, request GetComputeImageCapabilitySchemaRequest) (response GetComputeImageCapabilitySchemaResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getComputeImageCapabilitySchema, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetComputeImageCapabilitySchemaResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetComputeImageCapabilitySchemaResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetComputeImageCapabilitySchemaResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetComputeImageCapabilitySchemaResponse")
	}
	return
}

// getComputeImageCapabilitySchema implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getComputeImageCapabilitySchema(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/computeImageCapabilitySchemas/{computeImageCapabilitySchemaId}")
	if err != nil {
		return nil, err
	}

	var response GetComputeImageCapabilitySchemaResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetConsoleHistory Shows the metadata for the specified console history.
// See CaptureConsoleHistory
// for details about using the console history operations.
func (client ComputeClient) GetConsoleHistory(ctx context.Context, request GetConsoleHistoryRequest) (response GetConsoleHistoryResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getConsoleHistory, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetConsoleHistoryResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetConsoleHistoryResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetConsoleHistoryResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetConsoleHistoryResponse")
	}
	return
}

// getConsoleHistory implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getConsoleHistory(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instanceConsoleHistories/{instanceConsoleHistoryId}")
	if err != nil {
		return nil, err
	}

	var response GetConsoleHistoryResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetConsoleHistoryContent Gets the actual console history data (not the metadata).
// See CaptureConsoleHistory
// for details about using the console history operations.
func (client ComputeClient) GetConsoleHistoryContent(ctx context.Context, request GetConsoleHistoryContentRequest) (response GetConsoleHistoryContentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getConsoleHistoryContent, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetConsoleHistoryContentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetConsoleHistoryContentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetConsoleHistoryContentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetConsoleHistoryContentResponse")
	}
	return
}

// getConsoleHistoryContent implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getConsoleHistoryContent(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instanceConsoleHistories/{instanceConsoleHistoryId}/data")
	if err != nil {
		return nil, err
	}

	var response GetConsoleHistoryContentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetDedicatedVmHost Gets information about the specified dedicated virtual machine host.
func (client ComputeClient) GetDedicatedVmHost(ctx context.Context, request GetDedicatedVmHostRequest) (response GetDedicatedVmHostResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getDedicatedVmHost, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetDedicatedVmHostResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetDedicatedVmHostResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetDedicatedVmHostResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetDedicatedVmHostResponse")
	}
	return
}

// getDedicatedVmHost implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getDedicatedVmHost(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/dedicatedVmHosts/{dedicatedVmHostId}")
	if err != nil {
		return nil, err
	}

	var response GetDedicatedVmHostResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetImage Gets the specified image.
func (client ComputeClient) GetImage(ctx context.Context, request GetImageRequest) (response GetImageResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getImage, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetImageResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetImageResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetImageResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetImageResponse")
	}
	return
}

// getImage implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getImage(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/images/{imageId}")
	if err != nil {
		return nil, err
	}

	var response GetImageResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetImageShapeCompatibilityEntry Retrieves an image shape compatibility entry.
func (client ComputeClient) GetImageShapeCompatibilityEntry(ctx context.Context, request GetImageShapeCompatibilityEntryRequest) (response GetImageShapeCompatibilityEntryResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getImageShapeCompatibilityEntry, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetImageShapeCompatibilityEntryResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetImageShapeCompatibilityEntryResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetImageShapeCompatibilityEntryResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetImageShapeCompatibilityEntryResponse")
	}
	return
}

// getImageShapeCompatibilityEntry implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getImageShapeCompatibilityEntry(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/images/{imageId}/shapes/{shapeName}")
	if err != nil {
		return nil, err
	}

	var response GetImageShapeCompatibilityEntryResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetInstance Gets information about the specified instance.
func (client ComputeClient) GetInstance(ctx context.Context, request GetInstanceRequest) (response GetInstanceResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getInstance, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetInstanceResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetInstanceResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetInstanceResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetInstanceResponse")
	}
	return
}

// getInstance implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getInstance(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instances/{instanceId}")
	if err != nil {
		return nil, err
	}

	var response GetInstanceResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetInstanceConsoleConnection Gets the specified instance console connection's information.
func (client ComputeClient) GetInstanceConsoleConnection(ctx context.Context, request GetInstanceConsoleConnectionRequest) (response GetInstanceConsoleConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getInstanceConsoleConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetInstanceConsoleConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetInstanceConsoleConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetInstanceConsoleConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetInstanceConsoleConnectionResponse")
	}
	return
}

// getInstanceConsoleConnection implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getInstanceConsoleConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instanceConsoleConnections/{instanceConsoleConnectionId}")
	if err != nil {
		return nil, err
	}

	var response GetInstanceConsoleConnectionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetVnicAttachment Gets the information for the specified VNIC attachment.
func (client ComputeClient) GetVnicAttachment(ctx context.Context, request GetVnicAttachmentRequest) (response GetVnicAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVnicAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVnicAttachmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVnicAttachmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVnicAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVnicAttachmentResponse")
	}
	return
}

// getVnicAttachment implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getVnicAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vnicAttachments/{vnicAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response GetVnicAttachmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetVolumeAttachment Gets information about the specified volume attachment.
func (client ComputeClient) GetVolumeAttachment(ctx context.Context, request GetVolumeAttachmentRequest) (response GetVolumeAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVolumeAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVolumeAttachmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVolumeAttachmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVolumeAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVolumeAttachmentResponse")
	}
	return
}

// getVolumeAttachment implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getVolumeAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/volumeAttachments/{volumeAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response GetVolumeAttachmentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponseWithPolymorphicBody(httpResponse, &response, &volumeattachment{})
	return response, err
}

// GetWindowsInstanceInitialCredentials Gets the generated credentials for the instance. Only works for instances that require a password to log in, such as Windows.
// For certain operating systems, users will be forced to change the initial credentials.
func (client ComputeClient) GetWindowsInstanceInitialCredentials(ctx context.Context, request GetWindowsInstanceInitialCredentialsRequest) (response GetWindowsInstanceInitialCredentialsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getWindowsInstanceInitialCredentials, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetWindowsInstanceInitialCredentialsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetWindowsInstanceInitialCredentialsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetWindowsInstanceInitialCredentialsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetWindowsInstanceInitialCredentialsResponse")
	}
	return
}

// getWindowsInstanceInitialCredentials implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) getWindowsInstanceInitialCredentials(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instances/{instanceId}/initialCredentials")
	if err != nil {
		return nil, err
	}

	var response GetWindowsInstanceInitialCredentialsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// InstanceAction Performs one of the following power actions on the specified instance:
// - **START** - Powers on the instance.
// - **STOP** - Powers off the instance.
// - **RESET** - Powers off the instance and then powers it back on.
// - **SOFTSTOP** - Gracefully shuts down the instance by sending a shutdown command to the operating system.
// If the applications that run on the instance take a long time to shut down, they could be improperly stopped, resulting
// in data corruption. To avoid this, shut down the instance using the commands available in the OS before you softstop the
// instance.
// - **SOFTRESET** - Gracefully reboots the instance by sending a shutdown command to the operating system, and
// then powers the instance back on.
// For more information, see Stopping and Starting an Instance (https://docs.cloud.oracle.com/Content/Compute/Tasks/restartinginstance.htm).
func (client ComputeClient) InstanceAction(ctx context.Context, request InstanceActionRequest) (response InstanceActionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.instanceAction, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = InstanceActionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = InstanceActionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(InstanceActionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into InstanceActionResponse")
	}
	return
}

// instanceAction implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) instanceAction(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/instances/{instanceId}")
	if err != nil {
		return nil, err
	}

	var response InstanceActionResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// LaunchInstance Creates a new instance in the specified compartment and the specified availability domain.
// For general information about instances, see
// Overview of the Compute Service (https://docs.cloud.oracle.com/Content/Compute/Concepts/computeoverview.htm).
// For information about access control and compartments, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/Content/Identity/Concepts/overview.htm).
// For information about availability domains, see
// Regions and Availability Domains (https://docs.cloud.oracle.com/Content/General/Concepts/regions.htm).
// To get a list of availability domains, use the `ListAvailabilityDomains` operation
// in the Identity and Access Management Service API.
// All Oracle Cloud Infrastructure resources, including instances, get an Oracle-assigned,
// unique ID called an Oracle Cloud Identifier (OCID).
// When you create a resource, you can find its OCID in the response. You can
// also retrieve a resource's OCID by using a List API operation
// on that resource type, or by viewing the resource in the Console.
// To launch an instance using an image or a boot volume use the `sourceDetails` parameter in LaunchInstanceDetails.
// When you launch an instance, it is automatically attached to a virtual
// network interface card (VNIC), called the *primary VNIC*. The VNIC
// has a private IP address from the subnet's CIDR. You can either assign a
// private IP address of your choice or let Oracle automatically assign one.
// You can choose whether the instance has a public IP address. To retrieve the
// addresses, use the ListVnicAttachments
// operation to get the VNIC ID for the instance, and then call
// GetVnic with the VNIC ID.
// You can later add secondary VNICs to an instance. For more information, see
// Virtual Network Interface Cards (VNICs) (https://docs.cloud.oracle.com/Content/Network/Tasks/managingVNICs.htm).
// To launch an instance from a Marketplace image listing, you must provide the image ID of the
// listing resource version that you want, but you also must subscribe to the listing before you try
// to launch the instance. To subscribe to the listing, use the GetAppCatalogListingAgreements
// operation to get the signature for the terms of use agreement for the desired listing resource version.
// Then, call CreateAppCatalogSubscription
// with the signature. To get the image ID for the LaunchInstance operation, call
// GetAppCatalogListingResourceVersion.
func (client ComputeClient) LaunchInstance(ctx context.Context, request LaunchInstanceRequest) (response LaunchInstanceResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.launchInstance, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = LaunchInstanceResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = LaunchInstanceResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(LaunchInstanceResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into LaunchInstanceResponse")
	}
	return
}

// launchInstance implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) launchInstance(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/instances")
	if err != nil {
		return nil, err
	}

	var response LaunchInstanceResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListAppCatalogListingResourceVersions Gets all resource versions for a particular listing.
func (client ComputeClient) ListAppCatalogListingResourceVersions(ctx context.Context, request ListAppCatalogListingResourceVersionsRequest) (response ListAppCatalogListingResourceVersionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listAppCatalogListingResourceVersions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListAppCatalogListingResourceVersionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListAppCatalogListingResourceVersionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListAppCatalogListingResourceVersionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListAppCatalogListingResourceVersionsResponse")
	}
	return
}

// listAppCatalogListingResourceVersions implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listAppCatalogListingResourceVersions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/appCatalogListings/{listingId}/resourceVersions")
	if err != nil {
		return nil, err
	}

	var response ListAppCatalogListingResourceVersionsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListAppCatalogListings Lists the published listings.
func (client ComputeClient) ListAppCatalogListings(ctx context.Context, request ListAppCatalogListingsRequest) (response ListAppCatalogListingsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listAppCatalogListings, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListAppCatalogListingsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListAppCatalogListingsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListAppCatalogListingsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListAppCatalogListingsResponse")
	}
	return
}

// listAppCatalogListings implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listAppCatalogListings(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/appCatalogListings")
	if err != nil {
		return nil, err
	}

	var response ListAppCatalogListingsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListAppCatalogSubscriptions Lists subscriptions for a compartment.
func (client ComputeClient) ListAppCatalogSubscriptions(ctx context.Context, request ListAppCatalogSubscriptionsRequest) (response ListAppCatalogSubscriptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listAppCatalogSubscriptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListAppCatalogSubscriptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListAppCatalogSubscriptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListAppCatalogSubscriptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListAppCatalogSubscriptionsResponse")
	}
	return
}

// listAppCatalogSubscriptions implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listAppCatalogSubscriptions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/appCatalogSubscriptions")
	if err != nil {
		return nil, err
	}

	var response ListAppCatalogSubscriptionsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListBootVolumeAttachments Lists the boot volume attachments in the specified compartment. You can filter the
// list by specifying an instance OCID, boot volume OCID, or both.
func (client ComputeClient) ListBootVolumeAttachments(ctx context.Context, request ListBootVolumeAttachmentsRequest) (response ListBootVolumeAttachmentsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listBootVolumeAttachments, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListBootVolumeAttachmentsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListBootVolumeAttachmentsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListBootVolumeAttachmentsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListBootVolumeAttachmentsResponse")
	}
	return
}

// listBootVolumeAttachments implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listBootVolumeAttachments(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/bootVolumeAttachments")
	if err != nil {
		return nil, err
	}

	var response ListBootVolumeAttachmentsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListComputeGlobalImageCapabilitySchemaVersions Lists Compute Global Image Capability Schema versions in the specified compartment.
func (client ComputeClient) ListComputeGlobalImageCapabilitySchemaVersions(ctx context.Context, request ListComputeGlobalImageCapabilitySchemaVersionsRequest) (response ListComputeGlobalImageCapabilitySchemaVersionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listComputeGlobalImageCapabilitySchemaVersions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListComputeGlobalImageCapabilitySchemaVersionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListComputeGlobalImageCapabilitySchemaVersionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListComputeGlobalImageCapabilitySchemaVersionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListComputeGlobalImageCapabilitySchemaVersionsResponse")
	}
	return
}

// listComputeGlobalImageCapabilitySchemaVersions implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listComputeGlobalImageCapabilitySchemaVersions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/computeGlobalImageCapabilitySchemas/{computeGlobalImageCapabilitySchemaId}/versions")
	if err != nil {
		return nil, err
	}

	var response ListComputeGlobalImageCapabilitySchemaVersionsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListComputeGlobalImageCapabilitySchemas Lists Compute Global Image Capability Schema in the specified compartment.
func (client ComputeClient) ListComputeGlobalImageCapabilitySchemas(ctx context.Context, request ListComputeGlobalImageCapabilitySchemasRequest) (response ListComputeGlobalImageCapabilitySchemasResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listComputeGlobalImageCapabilitySchemas, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListComputeGlobalImageCapabilitySchemasResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListComputeGlobalImageCapabilitySchemasResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListComputeGlobalImageCapabilitySchemasResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListComputeGlobalImageCapabilitySchemasResponse")
	}
	return
}

// listComputeGlobalImageCapabilitySchemas implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listComputeGlobalImageCapabilitySchemas(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/computeGlobalImageCapabilitySchemas")
	if err != nil {
		return nil, err
	}

	var response ListComputeGlobalImageCapabilitySchemasResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListComputeImageCapabilitySchemas Lists Compute Image Capability Schema in the specified compartment. You can also query by a specific imageId.
func (client ComputeClient) ListComputeImageCapabilitySchemas(ctx context.Context, request ListComputeImageCapabilitySchemasRequest) (response ListComputeImageCapabilitySchemasResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listComputeImageCapabilitySchemas, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListComputeImageCapabilitySchemasResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListComputeImageCapabilitySchemasResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListComputeImageCapabilitySchemasResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListComputeImageCapabilitySchemasResponse")
	}
	return
}

// listComputeImageCapabilitySchemas implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listComputeImageCapabilitySchemas(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/computeImageCapabilitySchemas")
	if err != nil {
		return nil, err
	}

	var response ListComputeImageCapabilitySchemasResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListConsoleHistories Lists the console history metadata for the specified compartment or instance.
func (client ComputeClient) ListConsoleHistories(ctx context.Context, request ListConsoleHistoriesRequest) (response ListConsoleHistoriesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listConsoleHistories, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListConsoleHistoriesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListConsoleHistoriesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListConsoleHistoriesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListConsoleHistoriesResponse")
	}
	return
}

// listConsoleHistories implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listConsoleHistories(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instanceConsoleHistories")
	if err != nil {
		return nil, err
	}

	var response ListConsoleHistoriesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListDedicatedVmHostInstanceShapes Lists the shapes that can be used to launch a virtual machine instance on a dedicated virtual machine host within the specified compartment.
// You can filter the list by compatibility with a specific dedicated virtual machine host shape.
func (client ComputeClient) ListDedicatedVmHostInstanceShapes(ctx context.Context, request ListDedicatedVmHostInstanceShapesRequest) (response ListDedicatedVmHostInstanceShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listDedicatedVmHostInstanceShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListDedicatedVmHostInstanceShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListDedicatedVmHostInstanceShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListDedicatedVmHostInstanceShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListDedicatedVmHostInstanceShapesResponse")
	}
	return
}

// listDedicatedVmHostInstanceShapes implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listDedicatedVmHostInstanceShapes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/dedicatedVmHostInstanceShapes")
	if err != nil {
		return nil, err
	}

	var response ListDedicatedVmHostInstanceShapesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListDedicatedVmHostInstances Returns the list of instances on the dedicated virtual machine hosts that match the specified criteria.
func (client ComputeClient) ListDedicatedVmHostInstances(ctx context.Context, request ListDedicatedVmHostInstancesRequest) (response ListDedicatedVmHostInstancesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listDedicatedVmHostInstances, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListDedicatedVmHostInstancesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListDedicatedVmHostInstancesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListDedicatedVmHostInstancesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListDedicatedVmHostInstancesResponse")
	}
	return
}

// listDedicatedVmHostInstances implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listDedicatedVmHostInstances(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/dedicatedVmHosts/{dedicatedVmHostId}/instances")
	if err != nil {
		return nil, err
	}

	var response ListDedicatedVmHostInstancesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListDedicatedVmHostShapes Lists the shapes that can be used to launch a dedicated virtual machine host within the specified compartment.
func (client ComputeClient) ListDedicatedVmHostShapes(ctx context.Context, request ListDedicatedVmHostShapesRequest) (response ListDedicatedVmHostShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listDedicatedVmHostShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListDedicatedVmHostShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListDedicatedVmHostShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListDedicatedVmHostShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListDedicatedVmHostShapesResponse")
	}
	return
}

// listDedicatedVmHostShapes implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listDedicatedVmHostShapes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/dedicatedVmHostShapes")
	if err != nil {
		return nil, err
	}

	var response ListDedicatedVmHostShapesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListDedicatedVmHosts Returns the list of dedicated virtual machine hosts that match the specified criteria in the specified compartment.
// You can limit the list by specifying a dedicated virtual machine host display name. The list will include all the identically-named
// dedicated virtual machine hosts in the compartment.
func (client ComputeClient) ListDedicatedVmHosts(ctx context.Context, request ListDedicatedVmHostsRequest) (response ListDedicatedVmHostsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listDedicatedVmHosts, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListDedicatedVmHostsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListDedicatedVmHostsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListDedicatedVmHostsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListDedicatedVmHostsResponse")
	}
	return
}

// listDedicatedVmHosts implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listDedicatedVmHosts(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/dedicatedVmHosts")
	if err != nil {
		return nil, err
	}

	var response ListDedicatedVmHostsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListImageShapeCompatibilityEntries Lists the compatible shapes for the specified image.
func (client ComputeClient) ListImageShapeCompatibilityEntries(ctx context.Context, request ListImageShapeCompatibilityEntriesRequest) (response ListImageShapeCompatibilityEntriesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listImageShapeCompatibilityEntries, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListImageShapeCompatibilityEntriesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListImageShapeCompatibilityEntriesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListImageShapeCompatibilityEntriesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListImageShapeCompatibilityEntriesResponse")
	}
	return
}

// listImageShapeCompatibilityEntries implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listImageShapeCompatibilityEntries(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/images/{imageId}/shapes")
	if err != nil {
		return nil, err
	}

	var response ListImageShapeCompatibilityEntriesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListImages Lists the available images in the specified compartment, including both
// Oracle-provided images (https://docs.cloud.oracle.com/Content/Compute/References/images.htm) and
// custom images (https://docs.cloud.oracle.com/Content/Compute/Tasks/managingcustomimages.htm) that have
// been created. The list of images returned is ordered to first show all
// Oracle-provided images, then all custom images.
// The order of images returned may change when new images are released.
func (client ComputeClient) ListImages(ctx context.Context, request ListImagesRequest) (response ListImagesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listImages, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListImagesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListImagesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListImagesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListImagesResponse")
	}
	return
}

// listImages implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listImages(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/images")
	if err != nil {
		return nil, err
	}

	var response ListImagesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListInstanceConsoleConnections Lists the console connections for the specified compartment or instance.
// For more information about console access, see Accessing the Console (https://docs.cloud.oracle.com/Content/Compute/References/serialconsole.htm).
func (client ComputeClient) ListInstanceConsoleConnections(ctx context.Context, request ListInstanceConsoleConnectionsRequest) (response ListInstanceConsoleConnectionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listInstanceConsoleConnections, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListInstanceConsoleConnectionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListInstanceConsoleConnectionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListInstanceConsoleConnectionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListInstanceConsoleConnectionsResponse")
	}
	return
}

// listInstanceConsoleConnections implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listInstanceConsoleConnections(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instanceConsoleConnections")
	if err != nil {
		return nil, err
	}

	var response ListInstanceConsoleConnectionsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListInstanceDevices Gets a list of all the devices for given instance. You can optionally filter results by device availability.
func (client ComputeClient) ListInstanceDevices(ctx context.Context, request ListInstanceDevicesRequest) (response ListInstanceDevicesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listInstanceDevices, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListInstanceDevicesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListInstanceDevicesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListInstanceDevicesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListInstanceDevicesResponse")
	}
	return
}

// listInstanceDevices implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listInstanceDevices(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instances/{instanceId}/devices")
	if err != nil {
		return nil, err
	}

	var response ListInstanceDevicesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListInstances Lists the instances in the specified compartment and the specified availability domain.
// You can filter the results by specifying an instance name (the list will include all the identically-named
// instances in the compartment).
func (client ComputeClient) ListInstances(ctx context.Context, request ListInstancesRequest) (response ListInstancesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listInstances, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListInstancesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListInstancesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListInstancesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListInstancesResponse")
	}
	return
}

// listInstances implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listInstances(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/instances")
	if err != nil {
		return nil, err
	}

	var response ListInstancesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListShapes Lists the shapes that can be used to launch an instance within the specified compartment. You can
// filter the list by compatibility with a specific image.
func (client ComputeClient) ListShapes(ctx context.Context, request ListShapesRequest) (response ListShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListShapesResponse")
	}
	return
}

// listShapes implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listShapes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/shapes")
	if err != nil {
		return nil, err
	}

	var response ListShapesResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// ListVnicAttachments Lists the VNIC attachments in the specified compartment. A VNIC attachment
// resides in the same compartment as the attached instance. The list can be
// filtered by instance, VNIC, or availability domain.
func (client ComputeClient) ListVnicAttachments(ctx context.Context, request ListVnicAttachmentsRequest) (response ListVnicAttachmentsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVnicAttachments, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVnicAttachmentsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVnicAttachmentsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVnicAttachmentsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVnicAttachmentsResponse")
	}
	return
}

// listVnicAttachments implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listVnicAttachments(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vnicAttachments")
	if err != nil {
		return nil, err
	}

	var response ListVnicAttachmentsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

//listvolumeattachment allows to unmarshal list of polymorphic VolumeAttachment
type listvolumeattachment []volumeattachment

//UnmarshalPolymorphicJSON unmarshals polymorphic json list of items
func (m *listvolumeattachment) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {
	res := make([]VolumeAttachment, len(*m))
	for i, v := range *m {
		nn, err := v.UnmarshalPolymorphicJSON(v.JsonData)
		if err != nil {
			return nil, err
		}
		res[i] = nn.(VolumeAttachment)
	}
	return res, nil
}

// ListVolumeAttachments Lists the volume attachments in the specified compartment. You can filter the
// list by specifying an instance OCID, volume OCID, or both.
// Currently, the only supported volume attachment type are IScsiVolumeAttachment and
// ParavirtualizedVolumeAttachment.
func (client ComputeClient) ListVolumeAttachments(ctx context.Context, request ListVolumeAttachmentsRequest) (response ListVolumeAttachmentsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVolumeAttachments, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVolumeAttachmentsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVolumeAttachmentsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVolumeAttachmentsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVolumeAttachmentsResponse")
	}
	return
}

// listVolumeAttachments implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) listVolumeAttachments(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/volumeAttachments")
	if err != nil {
		return nil, err
	}

	var response ListVolumeAttachmentsResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponseWithPolymorphicBody(httpResponse, &response, &listvolumeattachment{})
	return response, err
}

// RemoveImageShapeCompatibilityEntry Removes a shape from the compatible shapes list for the image.
func (client ComputeClient) RemoveImageShapeCompatibilityEntry(ctx context.Context, request RemoveImageShapeCompatibilityEntryRequest) (response RemoveImageShapeCompatibilityEntryResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.removeImageShapeCompatibilityEntry, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = RemoveImageShapeCompatibilityEntryResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = RemoveImageShapeCompatibilityEntryResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(RemoveImageShapeCompatibilityEntryResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into RemoveImageShapeCompatibilityEntryResponse")
	}
	return
}

// removeImageShapeCompatibilityEntry implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) removeImageShapeCompatibilityEntry(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/images/{imageId}/shapes/{shapeName}")
	if err != nil {
		return nil, err
	}

	var response RemoveImageShapeCompatibilityEntryResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// TerminateInstance Terminates the specified instance. Any attached VNICs and volumes are automatically detached
// when the instance terminates.
// To preserve the boot volume associated with the instance, specify `true` for `PreserveBootVolumeQueryParam`.
// To delete the boot volume when the instance is deleted, specify `false` or do not specify a value for `PreserveBootVolumeQueryParam`.
// This is an asynchronous operation. The instance's `lifecycleState` will change to TERMINATING temporarily
// until the instance is completely removed.
func (client ComputeClient) TerminateInstance(ctx context.Context, request TerminateInstanceRequest) (response TerminateInstanceResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.terminateInstance, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = TerminateInstanceResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = TerminateInstanceResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(TerminateInstanceResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into TerminateInstanceResponse")
	}
	return
}

// terminateInstance implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) terminateInstance(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/instances/{instanceId}")
	if err != nil {
		return nil, err
	}

	var response TerminateInstanceResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateComputeImageCapabilitySchema Updates the specified Compute Image Capability Schema
func (client ComputeClient) UpdateComputeImageCapabilitySchema(ctx context.Context, request UpdateComputeImageCapabilitySchemaRequest) (response UpdateComputeImageCapabilitySchemaResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateComputeImageCapabilitySchema, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateComputeImageCapabilitySchemaResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateComputeImageCapabilitySchemaResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateComputeImageCapabilitySchemaResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateComputeImageCapabilitySchemaResponse")
	}
	return
}

// updateComputeImageCapabilitySchema implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) updateComputeImageCapabilitySchema(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/computeImageCapabilitySchemas/{computeImageCapabilitySchemaId}")
	if err != nil {
		return nil, err
	}

	var response UpdateComputeImageCapabilitySchemaResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateConsoleHistory Updates the specified console history metadata.
func (client ComputeClient) UpdateConsoleHistory(ctx context.Context, request UpdateConsoleHistoryRequest) (response UpdateConsoleHistoryResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateConsoleHistory, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateConsoleHistoryResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateConsoleHistoryResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateConsoleHistoryResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateConsoleHistoryResponse")
	}
	return
}

// updateConsoleHistory implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) updateConsoleHistory(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/instanceConsoleHistories/{instanceConsoleHistoryId}")
	if err != nil {
		return nil, err
	}

	var response UpdateConsoleHistoryResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateDedicatedVmHost Updates the displayName, freeformTags, and definedTags attributes for the specified dedicated virtual machine host.
// If an attribute value is not included, it will not be updated.
func (client ComputeClient) UpdateDedicatedVmHost(ctx context.Context, request UpdateDedicatedVmHostRequest) (response UpdateDedicatedVmHostResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.updateDedicatedVmHost, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateDedicatedVmHostResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateDedicatedVmHostResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateDedicatedVmHostResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateDedicatedVmHostResponse")
	}
	return
}

// updateDedicatedVmHost implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) updateDedicatedVmHost(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/dedicatedVmHosts/{dedicatedVmHostId}")
	if err != nil {
		return nil, err
	}

	var response UpdateDedicatedVmHostResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateImage Updates the display name of the image. Avoid entering confidential information.
func (client ComputeClient) UpdateImage(ctx context.Context, request UpdateImageRequest) (response UpdateImageResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.updateImage, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateImageResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateImageResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateImageResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateImageResponse")
	}
	return
}

// updateImage implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) updateImage(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/images/{imageId}")
	if err != nil {
		return nil, err
	}

	var response UpdateImageResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// UpdateInstance Updates certain fields on the specified instance. Fields that are not provided in the
// request will not be updated. Avoid entering confidential information.
// Changes to metadata fields will be reflected in the instance metadata service (this may take
// up to a minute).
// The OCID of the instance remains the same.
func (client ComputeClient) UpdateInstance(ctx context.Context, request UpdateInstanceRequest) (response UpdateInstanceResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.updateInstance, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateInstanceResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateInstanceResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateInstanceResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateInstanceResponse")
	}
	return
}

// updateInstance implements the OCIOperation interface (enables retrying operations)
func (client ComputeClient) updateInstance(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/instances/{instanceId}")
	if err != nil {
		return nil, err
	}

	var response UpdateInstanceResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	defer common.CloseBodyIfValid(httpResponse)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

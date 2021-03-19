// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
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
	"github.com/oracle/oci-go-sdk/v36/common"
	"github.com/oracle/oci-go-sdk/v36/common/auth"
	"net/http"
)

//VirtualNetworkClient a client for VirtualNetwork
type VirtualNetworkClient struct {
	common.BaseClient
	config *common.ConfigurationProvider
}

// NewVirtualNetworkClientWithConfigurationProvider Creates a new default VirtualNetwork client with the given configuration provider.
// the configuration provider will be used for the default signer as well as reading the region
func NewVirtualNetworkClientWithConfigurationProvider(configProvider common.ConfigurationProvider) (client VirtualNetworkClient, err error) {
	provider, err := auth.GetGenericConfigurationProvider(configProvider)
	if err != nil {
		return client, err
	}
	baseClient, e := common.NewClientWithConfig(provider)
	if e != nil {
		return client, e
	}
	return newVirtualNetworkClientFromBaseClient(baseClient, provider)
}

// NewVirtualNetworkClientWithOboToken Creates a new default VirtualNetwork client with the given configuration provider.
// The obotoken will be added to default headers and signed; the configuration provider will be used for the signer
//  as well as reading the region
func NewVirtualNetworkClientWithOboToken(configProvider common.ConfigurationProvider, oboToken string) (client VirtualNetworkClient, err error) {
	baseClient, err := common.NewClientWithOboToken(configProvider, oboToken)
	if err != nil {
		return client, err
	}

	return newVirtualNetworkClientFromBaseClient(baseClient, configProvider)
}

func newVirtualNetworkClientFromBaseClient(baseClient common.BaseClient, configProvider common.ConfigurationProvider) (client VirtualNetworkClient, err error) {
	client = VirtualNetworkClient{BaseClient: baseClient}
	client.BasePath = "20160918"
	err = client.setConfigurationProvider(configProvider)
	return
}

// SetRegion overrides the region of this client.
func (client *VirtualNetworkClient) SetRegion(region string) {
	client.Host = common.StringToRegion(region).EndpointForTemplate("iaas", "https://iaas.{region}.{secondLevelDomain}")
}

// SetConfigurationProvider sets the configuration provider including the region, returns an error if is not valid
func (client *VirtualNetworkClient) setConfigurationProvider(configProvider common.ConfigurationProvider) error {
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
func (client *VirtualNetworkClient) ConfigurationProvider() *common.ConfigurationProvider {
	return client.config
}

// AddNetworkSecurityGroupSecurityRules Adds one or more security rules to the specified network security group.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/AddNetworkSecurityGroupSecurityRules.go.html to see an example of how to use AddNetworkSecurityGroupSecurityRules API.
func (client VirtualNetworkClient) AddNetworkSecurityGroupSecurityRules(ctx context.Context, request AddNetworkSecurityGroupSecurityRulesRequest) (response AddNetworkSecurityGroupSecurityRulesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.addNetworkSecurityGroupSecurityRules, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AddNetworkSecurityGroupSecurityRulesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AddNetworkSecurityGroupSecurityRulesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AddNetworkSecurityGroupSecurityRulesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AddNetworkSecurityGroupSecurityRulesResponse")
	}
	return
}

// addNetworkSecurityGroupSecurityRules implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) addNetworkSecurityGroupSecurityRules(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/networkSecurityGroups/{networkSecurityGroupId}/actions/addSecurityRules")
	if err != nil {
		return nil, err
	}

	var response AddNetworkSecurityGroupSecurityRulesResponse
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

// AddPublicIpPoolCapacity Adds some or all of a CIDR block to a public IP pool.
// The CIDR block (or subrange) must not overlap with any other CIDR block already added to this or any other public IP pool.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/AddPublicIpPoolCapacity.go.html to see an example of how to use AddPublicIpPoolCapacity API.
func (client VirtualNetworkClient) AddPublicIpPoolCapacity(ctx context.Context, request AddPublicIpPoolCapacityRequest) (response AddPublicIpPoolCapacityResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.addPublicIpPoolCapacity, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AddPublicIpPoolCapacityResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AddPublicIpPoolCapacityResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AddPublicIpPoolCapacityResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AddPublicIpPoolCapacityResponse")
	}
	return
}

// addPublicIpPoolCapacity implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) addPublicIpPoolCapacity(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIpPools/{publicIpPoolId}/actions/addCapacity")
	if err != nil {
		return nil, err
	}

	var response AddPublicIpPoolCapacityResponse
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

// AddVcnCidr Adds a CIDR block to a VCN. The CIDR block you add:
// - Must be valid.
// - Must not overlap with another CIDR block in the VCN, a CIDR block of a peered VCN, or the on-premises network CIDR block.
// - Must not exceed the limit of CIDR blocks allowed per VCN.
// **Note:** Adding a CIDR block places your VCN in an updating state until the changes are complete. You cannot create or update the VCN's subnets, VLANs, LPGs, or route tables during this operation. The time to completion can take a few minutes. You can use the `GetWorkRequest` operation to check the status of the update.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/AddVcnCidr.go.html to see an example of how to use AddVcnCidr API.
func (client VirtualNetworkClient) AddVcnCidr(ctx context.Context, request AddVcnCidrRequest) (response AddVcnCidrResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.addVcnCidr, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AddVcnCidrResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AddVcnCidrResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AddVcnCidrResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AddVcnCidrResponse")
	}
	return
}

// addVcnCidr implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) addVcnCidr(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vcns/{vcnId}/actions/addCidr")
	if err != nil {
		return nil, err
	}

	var response AddVcnCidrResponse
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

// AdvertiseByoipRange Begins BGP route advertisements for the BYOIP CIDR block you imported to the Oracle Cloud.
// The `ByoipRange` resource must be in the PROVISIONED state before the BYOIP CIDR block routes can be advertised with BGP.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/AdvertiseByoipRange.go.html to see an example of how to use AdvertiseByoipRange API.
func (client VirtualNetworkClient) AdvertiseByoipRange(ctx context.Context, request AdvertiseByoipRangeRequest) (response AdvertiseByoipRangeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.advertiseByoipRange, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AdvertiseByoipRangeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AdvertiseByoipRangeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AdvertiseByoipRangeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AdvertiseByoipRangeResponse")
	}
	return
}

// advertiseByoipRange implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) advertiseByoipRange(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/byoipRanges/{byoipRangeId}/actions/advertise")
	if err != nil {
		return nil, err
	}

	var response AdvertiseByoipRangeResponse
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

// AttachServiceId Adds the specified Service to the list of enabled
// `Service` objects for the specified gateway. You must also set up a route rule with the
// `cidrBlock` of the `Service` as the rule's destination and the service gateway as the rule's
// target. See RouteTable.
// **Note:** The `AttachServiceId` operation is an easy way to add an individual `Service` to
// the service gateway. Compare it with
// UpdateServiceGateway, which replaces
// the entire existing list of enabled `Service` objects with the list that you provide in the
// `Update` call.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/AttachServiceId.go.html to see an example of how to use AttachServiceId API.
func (client VirtualNetworkClient) AttachServiceId(ctx context.Context, request AttachServiceIdRequest) (response AttachServiceIdResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.attachServiceId, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = AttachServiceIdResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = AttachServiceIdResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(AttachServiceIdResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into AttachServiceIdResponse")
	}
	return
}

// attachServiceId implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) attachServiceId(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/serviceGateways/{serviceGatewayId}/actions/attachService")
	if err != nil {
		return nil, err
	}

	var response AttachServiceIdResponse
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

// BulkAddVirtualCircuitPublicPrefixes Adds one or more customer public IP prefixes to the specified public virtual circuit.
// Use this operation (and not UpdateVirtualCircuit)
// to add prefixes to the virtual circuit. Oracle must verify the customer's ownership
// of each prefix before traffic for that prefix will flow across the virtual circuit.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/BulkAddVirtualCircuitPublicPrefixes.go.html to see an example of how to use BulkAddVirtualCircuitPublicPrefixes API.
func (client VirtualNetworkClient) BulkAddVirtualCircuitPublicPrefixes(ctx context.Context, request BulkAddVirtualCircuitPublicPrefixesRequest) (response BulkAddVirtualCircuitPublicPrefixesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.bulkAddVirtualCircuitPublicPrefixes, policy)
	if err != nil {
		if ociResponse != nil {
			response = BulkAddVirtualCircuitPublicPrefixesResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(BulkAddVirtualCircuitPublicPrefixesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into BulkAddVirtualCircuitPublicPrefixesResponse")
	}
	return
}

// bulkAddVirtualCircuitPublicPrefixes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) bulkAddVirtualCircuitPublicPrefixes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/virtualCircuits/{virtualCircuitId}/actions/bulkAddPublicPrefixes")
	if err != nil {
		return nil, err
	}

	var response BulkAddVirtualCircuitPublicPrefixesResponse
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

// BulkDeleteVirtualCircuitPublicPrefixes Removes one or more customer public IP prefixes from the specified public virtual circuit.
// Use this operation (and not UpdateVirtualCircuit)
// to remove prefixes from the virtual circuit. When the virtual circuit's state switches
// back to PROVISIONED, Oracle stops advertising the specified prefixes across the connection.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/BulkDeleteVirtualCircuitPublicPrefixes.go.html to see an example of how to use BulkDeleteVirtualCircuitPublicPrefixes API.
func (client VirtualNetworkClient) BulkDeleteVirtualCircuitPublicPrefixes(ctx context.Context, request BulkDeleteVirtualCircuitPublicPrefixesRequest) (response BulkDeleteVirtualCircuitPublicPrefixesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.bulkDeleteVirtualCircuitPublicPrefixes, policy)
	if err != nil {
		if ociResponse != nil {
			response = BulkDeleteVirtualCircuitPublicPrefixesResponse{RawResponse: ociResponse.HTTPResponse()}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(BulkDeleteVirtualCircuitPublicPrefixesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into BulkDeleteVirtualCircuitPublicPrefixesResponse")
	}
	return
}

// bulkDeleteVirtualCircuitPublicPrefixes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) bulkDeleteVirtualCircuitPublicPrefixes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/virtualCircuits/{virtualCircuitId}/actions/bulkDeletePublicPrefixes")
	if err != nil {
		return nil, err
	}

	var response BulkDeleteVirtualCircuitPublicPrefixesResponse
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

// ChangeByoipRangeCompartment Moves a BYOIP CIDR block to a different compartment. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeByoipRangeCompartment.go.html to see an example of how to use ChangeByoipRangeCompartment API.
func (client VirtualNetworkClient) ChangeByoipRangeCompartment(ctx context.Context, request ChangeByoipRangeCompartmentRequest) (response ChangeByoipRangeCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeByoipRangeCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeByoipRangeCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeByoipRangeCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeByoipRangeCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeByoipRangeCompartmentResponse")
	}
	return
}

// changeByoipRangeCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeByoipRangeCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/byoipRanges/{byoipRangeId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeByoipRangeCompartmentResponse
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

// ChangeCpeCompartment Moves a CPE object into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeCpeCompartment.go.html to see an example of how to use ChangeCpeCompartment API.
func (client VirtualNetworkClient) ChangeCpeCompartment(ctx context.Context, request ChangeCpeCompartmentRequest) (response ChangeCpeCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeCpeCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeCpeCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeCpeCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeCpeCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeCpeCompartmentResponse")
	}
	return
}

// changeCpeCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeCpeCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/cpes/{cpeId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeCpeCompartmentResponse
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

// ChangeCrossConnectCompartment Moves a cross-connect into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeCrossConnectCompartment.go.html to see an example of how to use ChangeCrossConnectCompartment API.
func (client VirtualNetworkClient) ChangeCrossConnectCompartment(ctx context.Context, request ChangeCrossConnectCompartmentRequest) (response ChangeCrossConnectCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeCrossConnectCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeCrossConnectCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeCrossConnectCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeCrossConnectCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeCrossConnectCompartmentResponse")
	}
	return
}

// changeCrossConnectCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeCrossConnectCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/crossConnects/{crossConnectId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeCrossConnectCompartmentResponse
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

// ChangeCrossConnectGroupCompartment Moves a cross-connect group into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeCrossConnectGroupCompartment.go.html to see an example of how to use ChangeCrossConnectGroupCompartment API.
func (client VirtualNetworkClient) ChangeCrossConnectGroupCompartment(ctx context.Context, request ChangeCrossConnectGroupCompartmentRequest) (response ChangeCrossConnectGroupCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeCrossConnectGroupCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeCrossConnectGroupCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeCrossConnectGroupCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeCrossConnectGroupCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeCrossConnectGroupCompartmentResponse")
	}
	return
}

// changeCrossConnectGroupCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeCrossConnectGroupCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/crossConnectGroups/{crossConnectGroupId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeCrossConnectGroupCompartmentResponse
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

// ChangeDhcpOptionsCompartment Moves a set of DHCP options into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeDhcpOptionsCompartment.go.html to see an example of how to use ChangeDhcpOptionsCompartment API.
func (client VirtualNetworkClient) ChangeDhcpOptionsCompartment(ctx context.Context, request ChangeDhcpOptionsCompartmentRequest) (response ChangeDhcpOptionsCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeDhcpOptionsCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeDhcpOptionsCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeDhcpOptionsCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeDhcpOptionsCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeDhcpOptionsCompartmentResponse")
	}
	return
}

// changeDhcpOptionsCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeDhcpOptionsCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/dhcps/{dhcpId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeDhcpOptionsCompartmentResponse
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

// ChangeDrgCompartment Moves a DRG into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeDrgCompartment.go.html to see an example of how to use ChangeDrgCompartment API.
func (client VirtualNetworkClient) ChangeDrgCompartment(ctx context.Context, request ChangeDrgCompartmentRequest) (response ChangeDrgCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeDrgCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeDrgCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeDrgCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeDrgCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeDrgCompartmentResponse")
	}
	return
}

// changeDrgCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeDrgCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/drgs/{drgId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeDrgCompartmentResponse
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

// ChangeIPSecConnectionCompartment Moves an IPSec connection into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeIPSecConnectionCompartment.go.html to see an example of how to use ChangeIPSecConnectionCompartment API.
func (client VirtualNetworkClient) ChangeIPSecConnectionCompartment(ctx context.Context, request ChangeIPSecConnectionCompartmentRequest) (response ChangeIPSecConnectionCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeIPSecConnectionCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeIPSecConnectionCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeIPSecConnectionCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeIPSecConnectionCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeIPSecConnectionCompartmentResponse")
	}
	return
}

// changeIPSecConnectionCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeIPSecConnectionCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/ipsecConnections/{ipscId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeIPSecConnectionCompartmentResponse
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

// ChangeInternetGatewayCompartment Moves an internet gateway into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeInternetGatewayCompartment.go.html to see an example of how to use ChangeInternetGatewayCompartment API.
func (client VirtualNetworkClient) ChangeInternetGatewayCompartment(ctx context.Context, request ChangeInternetGatewayCompartmentRequest) (response ChangeInternetGatewayCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeInternetGatewayCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeInternetGatewayCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeInternetGatewayCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeInternetGatewayCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeInternetGatewayCompartmentResponse")
	}
	return
}

// changeInternetGatewayCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeInternetGatewayCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/internetGateways/{igId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeInternetGatewayCompartmentResponse
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

// ChangeLocalPeeringGatewayCompartment Moves a local peering gateway into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeLocalPeeringGatewayCompartment.go.html to see an example of how to use ChangeLocalPeeringGatewayCompartment API.
func (client VirtualNetworkClient) ChangeLocalPeeringGatewayCompartment(ctx context.Context, request ChangeLocalPeeringGatewayCompartmentRequest) (response ChangeLocalPeeringGatewayCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeLocalPeeringGatewayCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeLocalPeeringGatewayCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeLocalPeeringGatewayCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeLocalPeeringGatewayCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeLocalPeeringGatewayCompartmentResponse")
	}
	return
}

// changeLocalPeeringGatewayCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeLocalPeeringGatewayCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/localPeeringGateways/{localPeeringGatewayId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeLocalPeeringGatewayCompartmentResponse
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

// ChangeNatGatewayCompartment Moves a NAT gateway into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeNatGatewayCompartment.go.html to see an example of how to use ChangeNatGatewayCompartment API.
func (client VirtualNetworkClient) ChangeNatGatewayCompartment(ctx context.Context, request ChangeNatGatewayCompartmentRequest) (response ChangeNatGatewayCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeNatGatewayCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeNatGatewayCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeNatGatewayCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeNatGatewayCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeNatGatewayCompartmentResponse")
	}
	return
}

// changeNatGatewayCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeNatGatewayCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/natGateways/{natGatewayId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeNatGatewayCompartmentResponse
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

// ChangeNetworkSecurityGroupCompartment Moves a network security group into a different compartment within the same tenancy. For
// information about moving resources between compartments, see Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeNetworkSecurityGroupCompartment.go.html to see an example of how to use ChangeNetworkSecurityGroupCompartment API.
func (client VirtualNetworkClient) ChangeNetworkSecurityGroupCompartment(ctx context.Context, request ChangeNetworkSecurityGroupCompartmentRequest) (response ChangeNetworkSecurityGroupCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeNetworkSecurityGroupCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeNetworkSecurityGroupCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeNetworkSecurityGroupCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeNetworkSecurityGroupCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeNetworkSecurityGroupCompartmentResponse")
	}
	return
}

// changeNetworkSecurityGroupCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeNetworkSecurityGroupCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/networkSecurityGroups/{networkSecurityGroupId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeNetworkSecurityGroupCompartmentResponse
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

// ChangePublicIpCompartment Moves a public IP into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
// This operation applies only to reserved public IPs. Ephemeral public IPs always belong to the
// same compartment as their VNIC and move accordingly.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangePublicIpCompartment.go.html to see an example of how to use ChangePublicIpCompartment API.
func (client VirtualNetworkClient) ChangePublicIpCompartment(ctx context.Context, request ChangePublicIpCompartmentRequest) (response ChangePublicIpCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changePublicIpCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangePublicIpCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangePublicIpCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangePublicIpCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangePublicIpCompartmentResponse")
	}
	return
}

// changePublicIpCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changePublicIpCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIps/{publicIpId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangePublicIpCompartmentResponse
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

// ChangePublicIpPoolCompartment Moves a public IP pool to a different compartment. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangePublicIpPoolCompartment.go.html to see an example of how to use ChangePublicIpPoolCompartment API.
func (client VirtualNetworkClient) ChangePublicIpPoolCompartment(ctx context.Context, request ChangePublicIpPoolCompartmentRequest) (response ChangePublicIpPoolCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changePublicIpPoolCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangePublicIpPoolCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangePublicIpPoolCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangePublicIpPoolCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangePublicIpPoolCompartmentResponse")
	}
	return
}

// changePublicIpPoolCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changePublicIpPoolCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIpPools/{publicIpPoolId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangePublicIpPoolCompartmentResponse
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

// ChangeRemotePeeringConnectionCompartment Moves a remote peering connection (RPC) into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeRemotePeeringConnectionCompartment.go.html to see an example of how to use ChangeRemotePeeringConnectionCompartment API.
func (client VirtualNetworkClient) ChangeRemotePeeringConnectionCompartment(ctx context.Context, request ChangeRemotePeeringConnectionCompartmentRequest) (response ChangeRemotePeeringConnectionCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeRemotePeeringConnectionCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeRemotePeeringConnectionCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeRemotePeeringConnectionCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeRemotePeeringConnectionCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeRemotePeeringConnectionCompartmentResponse")
	}
	return
}

// changeRemotePeeringConnectionCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeRemotePeeringConnectionCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/remotePeeringConnections/{remotePeeringConnectionId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeRemotePeeringConnectionCompartmentResponse
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

// ChangeRouteTableCompartment Moves a route table into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeRouteTableCompartment.go.html to see an example of how to use ChangeRouteTableCompartment API.
func (client VirtualNetworkClient) ChangeRouteTableCompartment(ctx context.Context, request ChangeRouteTableCompartmentRequest) (response ChangeRouteTableCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeRouteTableCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeRouteTableCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeRouteTableCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeRouteTableCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeRouteTableCompartmentResponse")
	}
	return
}

// changeRouteTableCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeRouteTableCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/routeTables/{rtId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeRouteTableCompartmentResponse
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

// ChangeSecurityListCompartment Moves a security list into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeSecurityListCompartment.go.html to see an example of how to use ChangeSecurityListCompartment API.
func (client VirtualNetworkClient) ChangeSecurityListCompartment(ctx context.Context, request ChangeSecurityListCompartmentRequest) (response ChangeSecurityListCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeSecurityListCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeSecurityListCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeSecurityListCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeSecurityListCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeSecurityListCompartmentResponse")
	}
	return
}

// changeSecurityListCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeSecurityListCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/securityLists/{securityListId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeSecurityListCompartmentResponse
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

// ChangeServiceGatewayCompartment Moves a service gateway into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeServiceGatewayCompartment.go.html to see an example of how to use ChangeServiceGatewayCompartment API.
func (client VirtualNetworkClient) ChangeServiceGatewayCompartment(ctx context.Context, request ChangeServiceGatewayCompartmentRequest) (response ChangeServiceGatewayCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeServiceGatewayCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeServiceGatewayCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeServiceGatewayCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeServiceGatewayCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeServiceGatewayCompartmentResponse")
	}
	return
}

// changeServiceGatewayCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeServiceGatewayCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/serviceGateways/{serviceGatewayId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeServiceGatewayCompartmentResponse
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

// ChangeSubnetCompartment Moves a subnet into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeSubnetCompartment.go.html to see an example of how to use ChangeSubnetCompartment API.
func (client VirtualNetworkClient) ChangeSubnetCompartment(ctx context.Context, request ChangeSubnetCompartmentRequest) (response ChangeSubnetCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeSubnetCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeSubnetCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeSubnetCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeSubnetCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeSubnetCompartmentResponse")
	}
	return
}

// changeSubnetCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeSubnetCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/subnets/{subnetId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeSubnetCompartmentResponse
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

// ChangeVcnCompartment Moves a VCN into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeVcnCompartment.go.html to see an example of how to use ChangeVcnCompartment API.
func (client VirtualNetworkClient) ChangeVcnCompartment(ctx context.Context, request ChangeVcnCompartmentRequest) (response ChangeVcnCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeVcnCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeVcnCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeVcnCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeVcnCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeVcnCompartmentResponse")
	}
	return
}

// changeVcnCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeVcnCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vcns/{vcnId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeVcnCompartmentResponse
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

// ChangeVirtualCircuitCompartment Moves a virtual circuit into a different compartment within the same tenancy. For information
// about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeVirtualCircuitCompartment.go.html to see an example of how to use ChangeVirtualCircuitCompartment API.
func (client VirtualNetworkClient) ChangeVirtualCircuitCompartment(ctx context.Context, request ChangeVirtualCircuitCompartmentRequest) (response ChangeVirtualCircuitCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeVirtualCircuitCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeVirtualCircuitCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeVirtualCircuitCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeVirtualCircuitCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeVirtualCircuitCompartmentResponse")
	}
	return
}

// changeVirtualCircuitCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeVirtualCircuitCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/virtualCircuits/{virtualCircuitId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeVirtualCircuitCompartmentResponse
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

// ChangeVlanCompartment Moves a VLAN into a different compartment within the same tenancy.
// For information about moving resources between compartments, see
// Moving Resources to a Different Compartment (https://docs.cloud.oracle.com/iaas/Content/Identity/Tasks/managingcompartments.htm#moveRes).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ChangeVlanCompartment.go.html to see an example of how to use ChangeVlanCompartment API.
func (client VirtualNetworkClient) ChangeVlanCompartment(ctx context.Context, request ChangeVlanCompartmentRequest) (response ChangeVlanCompartmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.changeVlanCompartment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ChangeVlanCompartmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ChangeVlanCompartmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ChangeVlanCompartmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ChangeVlanCompartmentResponse")
	}
	return
}

// changeVlanCompartment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) changeVlanCompartment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vlans/{vlanId}/actions/changeCompartment")
	if err != nil {
		return nil, err
	}

	var response ChangeVlanCompartmentResponse
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

// ConnectLocalPeeringGateways Connects this local peering gateway (LPG) to another one in the same region.
// This operation must be called by the VCN administrator who is designated as
// the *requestor* in the peering relationship. The *acceptor* must implement
// an Identity and Access Management (IAM) policy that gives the requestor permission
// to connect to LPGs in the acceptor's compartment. Without that permission, this
// operation will fail. For more information, see
// VCN Peering (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/VCNpeering.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ConnectLocalPeeringGateways.go.html to see an example of how to use ConnectLocalPeeringGateways API.
func (client VirtualNetworkClient) ConnectLocalPeeringGateways(ctx context.Context, request ConnectLocalPeeringGatewaysRequest) (response ConnectLocalPeeringGatewaysResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.connectLocalPeeringGateways, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ConnectLocalPeeringGatewaysResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ConnectLocalPeeringGatewaysResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ConnectLocalPeeringGatewaysResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ConnectLocalPeeringGatewaysResponse")
	}
	return
}

// connectLocalPeeringGateways implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) connectLocalPeeringGateways(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/localPeeringGateways/{localPeeringGatewayId}/actions/connect")
	if err != nil {
		return nil, err
	}

	var response ConnectLocalPeeringGatewaysResponse
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

// ConnectRemotePeeringConnections Connects this RPC to another one in a different region.
// This operation must be called by the VCN administrator who is designated as
// the *requestor* in the peering relationship. The *acceptor* must implement
// an Identity and Access Management (IAM) policy that gives the requestor permission
// to connect to RPCs in the acceptor's compartment. Without that permission, this
// operation will fail. For more information, see
// VCN Peering (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/VCNpeering.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ConnectRemotePeeringConnections.go.html to see an example of how to use ConnectRemotePeeringConnections API.
func (client VirtualNetworkClient) ConnectRemotePeeringConnections(ctx context.Context, request ConnectRemotePeeringConnectionsRequest) (response ConnectRemotePeeringConnectionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.connectRemotePeeringConnections, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ConnectRemotePeeringConnectionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ConnectRemotePeeringConnectionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ConnectRemotePeeringConnectionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ConnectRemotePeeringConnectionsResponse")
	}
	return
}

// connectRemotePeeringConnections implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) connectRemotePeeringConnections(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/remotePeeringConnections/{remotePeeringConnectionId}/actions/connect")
	if err != nil {
		return nil, err
	}

	var response ConnectRemotePeeringConnectionsResponse
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

// CreateByoipRange Creates a subrange of the BYOIP CIDR block.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateByoipRange.go.html to see an example of how to use CreateByoipRange API.
func (client VirtualNetworkClient) CreateByoipRange(ctx context.Context, request CreateByoipRangeRequest) (response CreateByoipRangeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createByoipRange, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateByoipRangeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateByoipRangeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateByoipRangeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateByoipRangeResponse")
	}
	return
}

// createByoipRange implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createByoipRange(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/byoipRanges")
	if err != nil {
		return nil, err
	}

	var response CreateByoipRangeResponse
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

// CreateCpe Creates a new virtual customer-premises equipment (CPE) object in the specified compartment. For
// more information, see IPSec VPNs (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingIPsec.htm).
// For the purposes of access control, you must provide the OCID of the compartment where you want
// the CPE to reside. Notice that the CPE doesn't have to be in the same compartment as the IPSec
// connection or other Networking Service components. If you're not sure which compartment to
// use, put the CPE in the same compartment as the DRG. For more information about
// compartments and access control, see Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// For information about OCIDs, see Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You must provide the public IP address of your on-premises router. See
// Configuring Your On-Premises Router for an IPSec VPN (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/configuringCPE.htm).
// You may optionally specify a *display name* for the CPE, otherwise a default is provided. It does not have to
// be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateCpe.go.html to see an example of how to use CreateCpe API.
func (client VirtualNetworkClient) CreateCpe(ctx context.Context, request CreateCpeRequest) (response CreateCpeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createCpe, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateCpeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateCpeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateCpeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateCpeResponse")
	}
	return
}

// createCpe implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createCpe(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/cpes")
	if err != nil {
		return nil, err
	}

	var response CreateCpeResponse
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

// CreateCrossConnect Creates a new cross-connect. Oracle recommends you create each cross-connect in a
// CrossConnectGroup so you can use link aggregation
// with the connection.
// After creating the `CrossConnect` object, you need to go the FastConnect location
// and request to have the physical cable installed. For more information, see
// FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
// For the purposes of access control, you must provide the OCID of the
// compartment where you want the cross-connect to reside. If you're
// not sure which compartment to use, put the cross-connect in the
// same compartment with your VCN. For more information about
// compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the cross-connect.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateCrossConnect.go.html to see an example of how to use CreateCrossConnect API.
func (client VirtualNetworkClient) CreateCrossConnect(ctx context.Context, request CreateCrossConnectRequest) (response CreateCrossConnectResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createCrossConnect, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateCrossConnectResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateCrossConnectResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateCrossConnectResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateCrossConnectResponse")
	}
	return
}

// createCrossConnect implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createCrossConnect(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/crossConnects")
	if err != nil {
		return nil, err
	}

	var response CreateCrossConnectResponse
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

// CreateCrossConnectGroup Creates a new cross-connect group to use with Oracle Cloud Infrastructure
// FastConnect. For more information, see
// FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
// For the purposes of access control, you must provide the OCID of the
// compartment where you want the cross-connect group to reside. If you're
// not sure which compartment to use, put the cross-connect group in the
// same compartment with your VCN. For more information about
// compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the cross-connect group.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateCrossConnectGroup.go.html to see an example of how to use CreateCrossConnectGroup API.
func (client VirtualNetworkClient) CreateCrossConnectGroup(ctx context.Context, request CreateCrossConnectGroupRequest) (response CreateCrossConnectGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createCrossConnectGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateCrossConnectGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateCrossConnectGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateCrossConnectGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateCrossConnectGroupResponse")
	}
	return
}

// createCrossConnectGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createCrossConnectGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/crossConnectGroups")
	if err != nil {
		return nil, err
	}

	var response CreateCrossConnectGroupResponse
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

// CreateDhcpOptions Creates a new set of DHCP options for the specified VCN. For more information, see
// DhcpOptions.
// For the purposes of access control, you must provide the OCID of the compartment where you want the set of
// DHCP options to reside. Notice that the set of options doesn't have to be in the same compartment as the VCN,
// subnets, or other Networking Service components. If you're not sure which compartment to use, put the set
// of DHCP options in the same compartment as the VCN. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm). For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the set of DHCP options, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateDhcpOptions.go.html to see an example of how to use CreateDhcpOptions API.
func (client VirtualNetworkClient) CreateDhcpOptions(ctx context.Context, request CreateDhcpOptionsRequest) (response CreateDhcpOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createDhcpOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateDhcpOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateDhcpOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateDhcpOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateDhcpOptionsResponse")
	}
	return
}

// createDhcpOptions implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createDhcpOptions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/dhcps")
	if err != nil {
		return nil, err
	}

	var response CreateDhcpOptionsResponse
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

// CreateDrg Creates a new dynamic routing gateway (DRG) in the specified compartment. For more information,
// see Dynamic Routing Gateways (DRGs) (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingDRGs.htm).
// For the purposes of access control, you must provide the OCID of the compartment where you want
// the DRG to reside. Notice that the DRG doesn't have to be in the same compartment as the VCN,
// the DRG attachment, or other Networking Service components. If you're not sure which compartment
// to use, put the DRG in the same compartment as the VCN. For more information about compartments
// and access control, see Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// For information about OCIDs, see Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the DRG, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateDrg.go.html to see an example of how to use CreateDrg API.
func (client VirtualNetworkClient) CreateDrg(ctx context.Context, request CreateDrgRequest) (response CreateDrgResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createDrg, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateDrgResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateDrgResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateDrgResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateDrgResponse")
	}
	return
}

// createDrg implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createDrg(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/drgs")
	if err != nil {
		return nil, err
	}

	var response CreateDrgResponse
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

// CreateDrgAttachment Attaches the specified DRG to the specified VCN. A VCN can be attached to only one DRG at a time,
// and vice versa. The response includes a `DrgAttachment` object with its own OCID. For more
// information about DRGs, see
// Dynamic Routing Gateways (DRGs) (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingDRGs.htm).
// You may optionally specify a *display name* for the attachment, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
// For the purposes of access control, the DRG attachment is automatically placed into the same compartment
// as the VCN. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateDrgAttachment.go.html to see an example of how to use CreateDrgAttachment API.
func (client VirtualNetworkClient) CreateDrgAttachment(ctx context.Context, request CreateDrgAttachmentRequest) (response CreateDrgAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createDrgAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateDrgAttachmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateDrgAttachmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateDrgAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateDrgAttachmentResponse")
	}
	return
}

// createDrgAttachment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createDrgAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/drgAttachments")
	if err != nil {
		return nil, err
	}

	var response CreateDrgAttachmentResponse
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

// CreateIPSecConnection Creates a new IPSec connection between the specified DRG and CPE. For more information, see
// IPSec VPNs (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingIPsec.htm).
// If you configure at least one tunnel to use static routing, then in the request you must provide
// at least one valid static route (you're allowed a maximum of 10). For example: 10.0.0.0/16.
// If you configure both tunnels to use BGP dynamic routing, you can provide an empty list for
// the static routes. For more information, see the important note in
// IPSecConnection.
// For the purposes of access control, you must provide the OCID of the compartment where you want the
// IPSec connection to reside. Notice that the IPSec connection doesn't have to be in the same compartment
// as the DRG, CPE, or other Networking Service components. If you're not sure which compartment to
// use, put the IPSec connection in the same compartment as the DRG. For more information about
// compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// For information about OCIDs, see Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the IPSec connection, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
// After creating the IPSec connection, you need to configure your on-premises router
// with tunnel-specific information. For tunnel status and the required configuration information, see:
//   * IPSecConnectionTunnel
//   * IPSecConnectionTunnelSharedSecret
// For each tunnel, you need the IP address of Oracle's VPN headend and the shared secret
// (that is, the pre-shared key). For more information, see
// Configuring Your On-Premises Router for an IPSec VPN (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/configuringCPE.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateIPSecConnection.go.html to see an example of how to use CreateIPSecConnection API.
func (client VirtualNetworkClient) CreateIPSecConnection(ctx context.Context, request CreateIPSecConnectionRequest) (response CreateIPSecConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createIPSecConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateIPSecConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateIPSecConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateIPSecConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateIPSecConnectionResponse")
	}
	return
}

// createIPSecConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createIPSecConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/ipsecConnections")
	if err != nil {
		return nil, err
	}

	var response CreateIPSecConnectionResponse
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

// CreateInternetGateway Creates a new internet gateway for the specified VCN. For more information, see
// Access to the Internet (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingIGs.htm).
// For the purposes of access control, you must provide the OCID of the compartment where you want the Internet
// Gateway to reside. Notice that the internet gateway doesn't have to be in the same compartment as the VCN or
// other Networking Service components. If you're not sure which compartment to use, put the Internet
// Gateway in the same compartment with the VCN. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm). For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the internet gateway, otherwise a default is provided. It
// does not have to be unique, and you can change it. Avoid entering confidential information.
// For traffic to flow between a subnet and an internet gateway, you must create a route rule accordingly in
// the subnet's route table (for example, 0.0.0.0/0 > internet gateway). See
// UpdateRouteTable.
// You must specify whether the internet gateway is enabled when you create it. If it's disabled, that means no
// traffic will flow to/from the internet even if there's a route rule that enables that traffic. You can later
// use UpdateInternetGateway to easily disable/enable
// the gateway without changing the route rule.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateInternetGateway.go.html to see an example of how to use CreateInternetGateway API.
func (client VirtualNetworkClient) CreateInternetGateway(ctx context.Context, request CreateInternetGatewayRequest) (response CreateInternetGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createInternetGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateInternetGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateInternetGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateInternetGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateInternetGatewayResponse")
	}
	return
}

// createInternetGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createInternetGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/internetGateways")
	if err != nil {
		return nil, err
	}

	var response CreateInternetGatewayResponse
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

// CreateIpv6 Creates an IPv6 for the specified VNIC.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateIpv6.go.html to see an example of how to use CreateIpv6 API.
func (client VirtualNetworkClient) CreateIpv6(ctx context.Context, request CreateIpv6Request) (response CreateIpv6Response, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createIpv6, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateIpv6Response{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateIpv6Response{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateIpv6Response); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateIpv6Response")
	}
	return
}

// createIpv6 implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createIpv6(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/ipv6")
	if err != nil {
		return nil, err
	}

	var response CreateIpv6Response
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

// CreateLocalPeeringGateway Creates a new local peering gateway (LPG) for the specified VCN.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateLocalPeeringGateway.go.html to see an example of how to use CreateLocalPeeringGateway API.
func (client VirtualNetworkClient) CreateLocalPeeringGateway(ctx context.Context, request CreateLocalPeeringGatewayRequest) (response CreateLocalPeeringGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createLocalPeeringGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateLocalPeeringGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateLocalPeeringGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateLocalPeeringGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateLocalPeeringGatewayResponse")
	}
	return
}

// createLocalPeeringGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createLocalPeeringGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/localPeeringGateways")
	if err != nil {
		return nil, err
	}

	var response CreateLocalPeeringGatewayResponse
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

// CreateNatGateway Creates a new NAT gateway for the specified VCN. You must also set up a route rule with the
// NAT gateway as the rule's target. See RouteTable.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateNatGateway.go.html to see an example of how to use CreateNatGateway API.
func (client VirtualNetworkClient) CreateNatGateway(ctx context.Context, request CreateNatGatewayRequest) (response CreateNatGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createNatGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateNatGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateNatGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateNatGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateNatGatewayResponse")
	}
	return
}

// createNatGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createNatGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/natGateways")
	if err != nil {
		return nil, err
	}

	var response CreateNatGatewayResponse
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

// CreateNetworkSecurityGroup Creates a new network security group for the specified VCN.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateNetworkSecurityGroup.go.html to see an example of how to use CreateNetworkSecurityGroup API.
func (client VirtualNetworkClient) CreateNetworkSecurityGroup(ctx context.Context, request CreateNetworkSecurityGroupRequest) (response CreateNetworkSecurityGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createNetworkSecurityGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateNetworkSecurityGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateNetworkSecurityGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateNetworkSecurityGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateNetworkSecurityGroupResponse")
	}
	return
}

// createNetworkSecurityGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createNetworkSecurityGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/networkSecurityGroups")
	if err != nil {
		return nil, err
	}

	var response CreateNetworkSecurityGroupResponse
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

// CreatePrivateIp Creates a secondary private IP for the specified VNIC.
// For more information about secondary private IPs, see
// IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingIPaddresses.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreatePrivateIp.go.html to see an example of how to use CreatePrivateIp API.
func (client VirtualNetworkClient) CreatePrivateIp(ctx context.Context, request CreatePrivateIpRequest) (response CreatePrivateIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createPrivateIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreatePrivateIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreatePrivateIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreatePrivateIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreatePrivateIpResponse")
	}
	return
}

// createPrivateIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createPrivateIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/privateIps")
	if err != nil {
		return nil, err
	}

	var response CreatePrivateIpResponse
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

// CreatePublicIp Creates a public IP. Use the `lifetime` property to specify whether it's an ephemeral or
// reserved public IP. For information about limits on how many you can create, see
// Public IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingpublicIPs.htm).
// * **For an ephemeral public IP assigned to a private IP:** You must also specify a `privateIpId`
// with the OCID of the primary private IP you want to assign the public IP to. The public IP is
// created in the same availability domain as the private IP. An ephemeral public IP must always be
// assigned to a private IP, and only to the *primary* private IP on a VNIC, not a secondary
// private IP. Exception: If you create a NatGateway, Oracle
// automatically assigns the NAT gateway a regional ephemeral public IP that you cannot remove.
// * **For a reserved public IP:** You may also optionally assign the public IP to a private
// IP by specifying `privateIpId`. Or you can later assign the public IP with
// UpdatePublicIp.
// **Note:** When assigning a public IP to a private IP, the private IP must not already have
// a public IP with `lifecycleState` = ASSIGNING or ASSIGNED. If it does, an error is returned.
// Also, for reserved public IPs, the optional assignment part of this operation is
// asynchronous. Poll the public IP's `lifecycleState` to determine if the assignment
// succeeded.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreatePublicIp.go.html to see an example of how to use CreatePublicIp API.
func (client VirtualNetworkClient) CreatePublicIp(ctx context.Context, request CreatePublicIpRequest) (response CreatePublicIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createPublicIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreatePublicIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreatePublicIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreatePublicIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreatePublicIpResponse")
	}
	return
}

// createPublicIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createPublicIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIps")
	if err != nil {
		return nil, err
	}

	var response CreatePublicIpResponse
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

// CreatePublicIpPool Creates a public IP pool.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreatePublicIpPool.go.html to see an example of how to use CreatePublicIpPool API.
func (client VirtualNetworkClient) CreatePublicIpPool(ctx context.Context, request CreatePublicIpPoolRequest) (response CreatePublicIpPoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createPublicIpPool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreatePublicIpPoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreatePublicIpPoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreatePublicIpPoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreatePublicIpPoolResponse")
	}
	return
}

// createPublicIpPool implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createPublicIpPool(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIpPools")
	if err != nil {
		return nil, err
	}

	var response CreatePublicIpPoolResponse
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

// CreateRemotePeeringConnection Creates a new remote peering connection (RPC) for the specified DRG.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateRemotePeeringConnection.go.html to see an example of how to use CreateRemotePeeringConnection API.
func (client VirtualNetworkClient) CreateRemotePeeringConnection(ctx context.Context, request CreateRemotePeeringConnectionRequest) (response CreateRemotePeeringConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createRemotePeeringConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateRemotePeeringConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateRemotePeeringConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateRemotePeeringConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateRemotePeeringConnectionResponse")
	}
	return
}

// createRemotePeeringConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createRemotePeeringConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/remotePeeringConnections")
	if err != nil {
		return nil, err
	}

	var response CreateRemotePeeringConnectionResponse
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

// CreateRouteTable Creates a new route table for the specified VCN. In the request you must also include at least one route
// rule for the new route table. For information on the number of rules you can have in a route table, see
// Service Limits (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/servicelimits.htm). For general information about route
// tables in your VCN and the types of targets you can use in route rules,
// see Route Tables (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm).
// For the purposes of access control, you must provide the OCID of the compartment where you want the route
// table to reside. Notice that the route table doesn't have to be in the same compartment as the VCN, subnets,
// or other Networking Service components. If you're not sure which compartment to use, put the route
// table in the same compartment as the VCN. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm). For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the route table, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateRouteTable.go.html to see an example of how to use CreateRouteTable API.
func (client VirtualNetworkClient) CreateRouteTable(ctx context.Context, request CreateRouteTableRequest) (response CreateRouteTableResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createRouteTable, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateRouteTableResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateRouteTableResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateRouteTableResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateRouteTableResponse")
	}
	return
}

// createRouteTable implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createRouteTable(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/routeTables")
	if err != nil {
		return nil, err
	}

	var response CreateRouteTableResponse
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

// CreateSecurityList Creates a new security list for the specified VCN. For more information
// about security lists, see Security Lists (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/securitylists.htm).
// For information on the number of rules you can have in a security list, see
// Service Limits (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/servicelimits.htm).
// For the purposes of access control, you must provide the OCID of the compartment where you want the security
// list to reside. Notice that the security list doesn't have to be in the same compartment as the VCN, subnets,
// or other Networking Service components. If you're not sure which compartment to use, put the security
// list in the same compartment as the VCN. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm). For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the security list, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateSecurityList.go.html to see an example of how to use CreateSecurityList API.
func (client VirtualNetworkClient) CreateSecurityList(ctx context.Context, request CreateSecurityListRequest) (response CreateSecurityListResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createSecurityList, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateSecurityListResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateSecurityListResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateSecurityListResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateSecurityListResponse")
	}
	return
}

// createSecurityList implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createSecurityList(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/securityLists")
	if err != nil {
		return nil, err
	}

	var response CreateSecurityListResponse
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

// CreateServiceGateway Creates a new service gateway in the specified compartment.
// For the purposes of access control, you must provide the OCID of the compartment where you want
// the service gateway to reside. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// For information about OCIDs, see Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the service gateway, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateServiceGateway.go.html to see an example of how to use CreateServiceGateway API.
func (client VirtualNetworkClient) CreateServiceGateway(ctx context.Context, request CreateServiceGatewayRequest) (response CreateServiceGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createServiceGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateServiceGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateServiceGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateServiceGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateServiceGatewayResponse")
	}
	return
}

// createServiceGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createServiceGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/serviceGateways")
	if err != nil {
		return nil, err
	}

	var response CreateServiceGatewayResponse
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

// CreateSubnet Creates a new subnet in the specified VCN. You can't change the size of the subnet after creation,
// so it's important to think about the size of subnets you need before creating them.
// For more information, see VCNs and Subnets (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingVCNs.htm).
// For information on the number of subnets you can have in a VCN, see
// Service Limits (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/servicelimits.htm).
// For the purposes of access control, you must provide the OCID of the compartment where you want the subnet
// to reside. Notice that the subnet doesn't have to be in the same compartment as the VCN, route tables, or
// other Networking Service components. If you're not sure which compartment to use, put the subnet in
// the same compartment as the VCN. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm). For information about OCIDs,
// see Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally associate a route table with the subnet. If you don't, the subnet will use the
// VCN's default route table. For more information about route tables, see
// Route Tables (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm).
// You may optionally associate a security list with the subnet. If you don't, the subnet will use the
// VCN's default security list. For more information about security lists, see
// Security Lists (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/securitylists.htm).
// You may optionally associate a set of DHCP options with the subnet. If you don't, the subnet will use the
// VCN's default set. For more information about DHCP options, see
// DHCP Options (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingDHCP.htm).
// You may optionally specify a *display name* for the subnet, otherwise a default is provided.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
// You can also add a DNS label for the subnet, which is required if you want the Internet and
// VCN Resolver to resolve hostnames for instances in the subnet. For more information, see
// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateSubnet.go.html to see an example of how to use CreateSubnet API.
func (client VirtualNetworkClient) CreateSubnet(ctx context.Context, request CreateSubnetRequest) (response CreateSubnetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createSubnet, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateSubnetResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateSubnetResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateSubnetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateSubnetResponse")
	}
	return
}

// createSubnet implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createSubnet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/subnets")
	if err != nil {
		return nil, err
	}

	var response CreateSubnetResponse
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

// CreateVcn Creates a new virtual cloud network (VCN). For more information, see
// VCNs and Subnets (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingVCNs.htm).
// For the VCN, you specify a list of one or more IPv4 CIDR blocks that meet the following criteria:
// - The CIDR blocks must be valid.
// - They must not overlap with each other or with the on-premises network CIDR block.
// - The number of CIDR blocks does not exceed the limit of CIDR blocks allowed per VCN.
// For a CIDR block, Oracle recommends that you use one of the private IP address ranges specified in RFC 1918 (https://tools.ietf.org/html/rfc1918) (10.0.0.0/8, 172.16/12, and 192.168/16). Example:
// 172.16.0.0/16. The CIDR blocks can range from /16 to /30.
// For the purposes of access control, you must provide the OCID of the compartment where you want the VCN to
// reside. Consult an Oracle Cloud Infrastructure administrator in your organization if you're not sure which
// compartment to use. Notice that the VCN doesn't have to be in the same compartment as the subnets or other
// Networking Service components. For more information about compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm). For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the VCN, otherwise a default is provided. It does not have to
// be unique, and you can change it. Avoid entering confidential information.
// You can also add a DNS label for the VCN, which is required if you want the instances to use the
// Interent and VCN Resolver option for DNS in the VCN. For more information, see
// DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm).
// The VCN automatically comes with a default route table, default security list, and default set of DHCP options.
// The OCID for each is returned in the response. You can't delete these default objects, but you can change their
// contents (that is, change the route rules, security list rules, and so on).
// The VCN and subnets you create are not accessible until you attach an internet gateway or set up an IPSec VPN
// or FastConnect. For more information, see
// Overview of the Networking Service (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateVcn.go.html to see an example of how to use CreateVcn API.
func (client VirtualNetworkClient) CreateVcn(ctx context.Context, request CreateVcnRequest) (response CreateVcnResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createVcn, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateVcnResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateVcnResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateVcnResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateVcnResponse")
	}
	return
}

// createVcn implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createVcn(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vcns")
	if err != nil {
		return nil, err
	}

	var response CreateVcnResponse
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

// CreateVirtualCircuit Creates a new virtual circuit to use with Oracle Cloud
// Infrastructure FastConnect. For more information, see
// FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
// For the purposes of access control, you must provide the OCID of the
// compartment where you want the virtual circuit to reside. If you're
// not sure which compartment to use, put the virtual circuit in the
// same compartment with the DRG it's using. For more information about
// compartments and access control, see
// Overview of the IAM Service (https://docs.cloud.oracle.com/iaas/Content/Identity/Concepts/overview.htm).
// For information about OCIDs, see
// Resource Identifiers (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// You may optionally specify a *display name* for the virtual circuit.
// It does not have to be unique, and you can change it. Avoid entering confidential information.
// **Important:** When creating a virtual circuit, you specify a DRG for
// the traffic to flow through. Make sure you attach the DRG to your
// VCN and confirm the VCN's routing sends traffic to the DRG. Otherwise
// traffic will not flow. For more information, see
// Route Tables (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateVirtualCircuit.go.html to see an example of how to use CreateVirtualCircuit API.
func (client VirtualNetworkClient) CreateVirtualCircuit(ctx context.Context, request CreateVirtualCircuitRequest) (response CreateVirtualCircuitResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createVirtualCircuit, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateVirtualCircuitResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateVirtualCircuitResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateVirtualCircuitResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateVirtualCircuitResponse")
	}
	return
}

// createVirtualCircuit implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createVirtualCircuit(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/virtualCircuits")
	if err != nil {
		return nil, err
	}

	var response CreateVirtualCircuitResponse
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

// CreateVlan Creates a VLAN in the specified VCN and the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/CreateVlan.go.html to see an example of how to use CreateVlan API.
func (client VirtualNetworkClient) CreateVlan(ctx context.Context, request CreateVlanRequest) (response CreateVlanResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.createVlan, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = CreateVlanResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = CreateVlanResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(CreateVlanResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into CreateVlanResponse")
	}
	return
}

// createVlan implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) createVlan(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vlans")
	if err != nil {
		return nil, err
	}

	var response CreateVlanResponse
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

// DeleteByoipRange Deletes the specified `ByoipRange` resource.
// The resource must be in one of the following states: CREATING, PROVISIONED, ACTIVE, or FAILED.
// It must not have any subranges currently allocated to a PublicIpPool object or the deletion will fail.
// You must specify the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
// If the `ByoipRange` resource is currently in the PROVISIONED or ACTIVE state, it will be de-provisioned and then deleted.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteByoipRange.go.html to see an example of how to use DeleteByoipRange API.
func (client VirtualNetworkClient) DeleteByoipRange(ctx context.Context, request DeleteByoipRangeRequest) (response DeleteByoipRangeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteByoipRange, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteByoipRangeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteByoipRangeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteByoipRangeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteByoipRangeResponse")
	}
	return
}

// deleteByoipRange implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteByoipRange(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/byoipRanges/{byoipRangeId}")
	if err != nil {
		return nil, err
	}

	var response DeleteByoipRangeResponse
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

// DeleteCpe Deletes the specified CPE object. The CPE must not be connected to a DRG. This is an asynchronous
// operation. The CPE's `lifecycleState` will change to TERMINATING temporarily until the CPE is completely
// removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteCpe.go.html to see an example of how to use DeleteCpe API.
func (client VirtualNetworkClient) DeleteCpe(ctx context.Context, request DeleteCpeRequest) (response DeleteCpeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteCpe, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteCpeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteCpeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteCpeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteCpeResponse")
	}
	return
}

// deleteCpe implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteCpe(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/cpes/{cpeId}")
	if err != nil {
		return nil, err
	}

	var response DeleteCpeResponse
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

// DeleteCrossConnect Deletes the specified cross-connect. It must not be mapped to a
// VirtualCircuit.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteCrossConnect.go.html to see an example of how to use DeleteCrossConnect API.
func (client VirtualNetworkClient) DeleteCrossConnect(ctx context.Context, request DeleteCrossConnectRequest) (response DeleteCrossConnectResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteCrossConnect, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteCrossConnectResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteCrossConnectResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteCrossConnectResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteCrossConnectResponse")
	}
	return
}

// deleteCrossConnect implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteCrossConnect(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/crossConnects/{crossConnectId}")
	if err != nil {
		return nil, err
	}

	var response DeleteCrossConnectResponse
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

// DeleteCrossConnectGroup Deletes the specified cross-connect group. It must not contain any
// cross-connects, and it cannot be mapped to a
// VirtualCircuit.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteCrossConnectGroup.go.html to see an example of how to use DeleteCrossConnectGroup API.
func (client VirtualNetworkClient) DeleteCrossConnectGroup(ctx context.Context, request DeleteCrossConnectGroupRequest) (response DeleteCrossConnectGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteCrossConnectGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteCrossConnectGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteCrossConnectGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteCrossConnectGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteCrossConnectGroupResponse")
	}
	return
}

// deleteCrossConnectGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteCrossConnectGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/crossConnectGroups/{crossConnectGroupId}")
	if err != nil {
		return nil, err
	}

	var response DeleteCrossConnectGroupResponse
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

// DeleteDhcpOptions Deletes the specified set of DHCP options, but only if it's not associated with a subnet. You can't delete a
// VCN's default set of DHCP options.
// This is an asynchronous operation. The state of the set of options will switch to TERMINATING temporarily
// until the set is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteDhcpOptions.go.html to see an example of how to use DeleteDhcpOptions API.
func (client VirtualNetworkClient) DeleteDhcpOptions(ctx context.Context, request DeleteDhcpOptionsRequest) (response DeleteDhcpOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteDhcpOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteDhcpOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteDhcpOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteDhcpOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteDhcpOptionsResponse")
	}
	return
}

// deleteDhcpOptions implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteDhcpOptions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/dhcps/{dhcpId}")
	if err != nil {
		return nil, err
	}

	var response DeleteDhcpOptionsResponse
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

// DeleteDrg Deletes the specified DRG. The DRG must not be attached to a VCN or be connected to your on-premise
// network. Also, there must not be a route table that lists the DRG as a target. This is an asynchronous
// operation. The DRG's `lifecycleState` will change to TERMINATING temporarily until the DRG is completely
// removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteDrg.go.html to see an example of how to use DeleteDrg API.
func (client VirtualNetworkClient) DeleteDrg(ctx context.Context, request DeleteDrgRequest) (response DeleteDrgResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteDrg, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteDrgResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteDrgResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteDrgResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteDrgResponse")
	}
	return
}

// deleteDrg implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteDrg(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/drgs/{drgId}")
	if err != nil {
		return nil, err
	}

	var response DeleteDrgResponse
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

// DeleteDrgAttachment Detaches a DRG from a VCN by deleting the corresponding `DrgAttachment`. This is an asynchronous
// operation. The attachment's `lifecycleState` will change to DETACHING temporarily until the attachment
// is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteDrgAttachment.go.html to see an example of how to use DeleteDrgAttachment API.
func (client VirtualNetworkClient) DeleteDrgAttachment(ctx context.Context, request DeleteDrgAttachmentRequest) (response DeleteDrgAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteDrgAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteDrgAttachmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteDrgAttachmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteDrgAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteDrgAttachmentResponse")
	}
	return
}

// deleteDrgAttachment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteDrgAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/drgAttachments/{drgAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response DeleteDrgAttachmentResponse
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

// DeleteIPSecConnection Deletes the specified IPSec connection. If your goal is to disable the IPSec VPN between your VCN and
// on-premises network, it's easiest to simply detach the DRG but keep all the IPSec VPN components intact.
// If you were to delete all the components and then later need to create an IPSec VPN again, you would
// need to configure your on-premises router again with the new information returned from
// CreateIPSecConnection.
// This is an asynchronous operation. The connection's `lifecycleState` will change to TERMINATING temporarily
// until the connection is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteIPSecConnection.go.html to see an example of how to use DeleteIPSecConnection API.
func (client VirtualNetworkClient) DeleteIPSecConnection(ctx context.Context, request DeleteIPSecConnectionRequest) (response DeleteIPSecConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteIPSecConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteIPSecConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteIPSecConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteIPSecConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteIPSecConnectionResponse")
	}
	return
}

// deleteIPSecConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteIPSecConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/ipsecConnections/{ipscId}")
	if err != nil {
		return nil, err
	}

	var response DeleteIPSecConnectionResponse
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

// DeleteInternetGateway Deletes the specified internet gateway. The internet gateway does not have to be disabled, but
// there must not be a route table that lists it as a target.
// This is an asynchronous operation. The gateway's `lifecycleState` will change to TERMINATING temporarily
// until the gateway is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteInternetGateway.go.html to see an example of how to use DeleteInternetGateway API.
func (client VirtualNetworkClient) DeleteInternetGateway(ctx context.Context, request DeleteInternetGatewayRequest) (response DeleteInternetGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteInternetGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteInternetGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteInternetGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteInternetGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteInternetGatewayResponse")
	}
	return
}

// deleteInternetGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteInternetGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/internetGateways/{igId}")
	if err != nil {
		return nil, err
	}

	var response DeleteInternetGatewayResponse
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

// DeleteIpv6 Unassigns and deletes the specified IPv6. You must specify the object's OCID.
// The IPv6 address is returned to the subnet's pool of available addresses.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteIpv6.go.html to see an example of how to use DeleteIpv6 API.
func (client VirtualNetworkClient) DeleteIpv6(ctx context.Context, request DeleteIpv6Request) (response DeleteIpv6Response, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteIpv6, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteIpv6Response{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteIpv6Response{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteIpv6Response); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteIpv6Response")
	}
	return
}

// deleteIpv6 implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteIpv6(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/ipv6/{ipv6Id}")
	if err != nil {
		return nil, err
	}

	var response DeleteIpv6Response
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

// DeleteLocalPeeringGateway Deletes the specified local peering gateway (LPG).
// This is an asynchronous operation; the local peering gateway's `lifecycleState` changes to TERMINATING temporarily
// until the local peering gateway is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteLocalPeeringGateway.go.html to see an example of how to use DeleteLocalPeeringGateway API.
func (client VirtualNetworkClient) DeleteLocalPeeringGateway(ctx context.Context, request DeleteLocalPeeringGatewayRequest) (response DeleteLocalPeeringGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteLocalPeeringGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteLocalPeeringGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteLocalPeeringGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteLocalPeeringGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteLocalPeeringGatewayResponse")
	}
	return
}

// deleteLocalPeeringGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteLocalPeeringGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/localPeeringGateways/{localPeeringGatewayId}")
	if err != nil {
		return nil, err
	}

	var response DeleteLocalPeeringGatewayResponse
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

// DeleteNatGateway Deletes the specified NAT gateway. The NAT gateway does not have to be disabled, but there
// must not be a route rule that lists the NAT gateway as a target.
// This is an asynchronous operation. The NAT gateway's `lifecycleState` will change to
// TERMINATING temporarily until the NAT gateway is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteNatGateway.go.html to see an example of how to use DeleteNatGateway API.
func (client VirtualNetworkClient) DeleteNatGateway(ctx context.Context, request DeleteNatGatewayRequest) (response DeleteNatGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteNatGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteNatGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteNatGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteNatGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteNatGatewayResponse")
	}
	return
}

// deleteNatGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteNatGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/natGateways/{natGatewayId}")
	if err != nil {
		return nil, err
	}

	var response DeleteNatGatewayResponse
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

// DeleteNetworkSecurityGroup Deletes the specified network security group. The group must not contain any VNICs.
// To get a list of the VNICs in a network security group, use
// ListNetworkSecurityGroupVnics.
// Each returned NetworkSecurityGroupVnic object
// contains both the OCID of the VNIC and the OCID of the VNIC's parent resource (for example,
// the Compute instance that the VNIC is attached to).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteNetworkSecurityGroup.go.html to see an example of how to use DeleteNetworkSecurityGroup API.
func (client VirtualNetworkClient) DeleteNetworkSecurityGroup(ctx context.Context, request DeleteNetworkSecurityGroupRequest) (response DeleteNetworkSecurityGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteNetworkSecurityGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteNetworkSecurityGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteNetworkSecurityGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteNetworkSecurityGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteNetworkSecurityGroupResponse")
	}
	return
}

// deleteNetworkSecurityGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteNetworkSecurityGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/networkSecurityGroups/{networkSecurityGroupId}")
	if err != nil {
		return nil, err
	}

	var response DeleteNetworkSecurityGroupResponse
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

// DeletePrivateIp Unassigns and deletes the specified private IP. You must
// specify the object's OCID. The private IP address is returned to
// the subnet's pool of available addresses.
// This operation cannot be used with primary private IPs, which are
// automatically unassigned and deleted when the VNIC is terminated.
// **Important:** If a secondary private IP is the
// target of a route rule (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingroutetables.htm#privateip),
// unassigning it from the VNIC causes that route rule to blackhole and the traffic
// will be dropped.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeletePrivateIp.go.html to see an example of how to use DeletePrivateIp API.
func (client VirtualNetworkClient) DeletePrivateIp(ctx context.Context, request DeletePrivateIpRequest) (response DeletePrivateIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deletePrivateIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeletePrivateIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeletePrivateIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeletePrivateIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeletePrivateIpResponse")
	}
	return
}

// deletePrivateIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deletePrivateIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/privateIps/{privateIpId}")
	if err != nil {
		return nil, err
	}

	var response DeletePrivateIpResponse
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

// DeletePublicIp Unassigns and deletes the specified public IP (either ephemeral or reserved).
// You must specify the object's OCID. The public IP address is returned to the
// Oracle Cloud Infrastructure public IP pool.
// **Note:** You cannot update, unassign, or delete the public IP that Oracle automatically
// assigned to an entity for you (such as a load balancer or NAT gateway). The public IP is
// automatically deleted if the assigned entity is terminated.
// For an assigned reserved public IP, the initial unassignment portion of this operation
// is asynchronous. Poll the public IP's `lifecycleState` to determine
// if the operation succeeded.
// If you want to simply unassign a reserved public IP and return it to your pool
// of reserved public IPs, instead use
// UpdatePublicIp.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeletePublicIp.go.html to see an example of how to use DeletePublicIp API.
func (client VirtualNetworkClient) DeletePublicIp(ctx context.Context, request DeletePublicIpRequest) (response DeletePublicIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deletePublicIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeletePublicIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeletePublicIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeletePublicIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeletePublicIpResponse")
	}
	return
}

// deletePublicIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deletePublicIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/publicIps/{publicIpId}")
	if err != nil {
		return nil, err
	}

	var response DeletePublicIpResponse
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

// DeletePublicIpPool Deletes the specified public IP pool.
// To delete a public IP pool it must not have any active IP address allocations.
// You must specify the object's OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm) when deleting an IP pool.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeletePublicIpPool.go.html to see an example of how to use DeletePublicIpPool API.
func (client VirtualNetworkClient) DeletePublicIpPool(ctx context.Context, request DeletePublicIpPoolRequest) (response DeletePublicIpPoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deletePublicIpPool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeletePublicIpPoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeletePublicIpPoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeletePublicIpPoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeletePublicIpPoolResponse")
	}
	return
}

// deletePublicIpPool implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deletePublicIpPool(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/publicIpPools/{publicIpPoolId}")
	if err != nil {
		return nil, err
	}

	var response DeletePublicIpPoolResponse
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

// DeleteRemotePeeringConnection Deletes the remote peering connection (RPC).
// This is an asynchronous operation; the RPC's `lifecycleState` changes to TERMINATING temporarily
// until the RPC is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteRemotePeeringConnection.go.html to see an example of how to use DeleteRemotePeeringConnection API.
func (client VirtualNetworkClient) DeleteRemotePeeringConnection(ctx context.Context, request DeleteRemotePeeringConnectionRequest) (response DeleteRemotePeeringConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteRemotePeeringConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteRemotePeeringConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteRemotePeeringConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteRemotePeeringConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteRemotePeeringConnectionResponse")
	}
	return
}

// deleteRemotePeeringConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteRemotePeeringConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/remotePeeringConnections/{remotePeeringConnectionId}")
	if err != nil {
		return nil, err
	}

	var response DeleteRemotePeeringConnectionResponse
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

// DeleteRouteTable Deletes the specified route table, but only if it's not associated with a subnet. You can't delete a
// VCN's default route table.
// This is an asynchronous operation. The route table's `lifecycleState` will change to TERMINATING temporarily
// until the route table is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteRouteTable.go.html to see an example of how to use DeleteRouteTable API.
func (client VirtualNetworkClient) DeleteRouteTable(ctx context.Context, request DeleteRouteTableRequest) (response DeleteRouteTableResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteRouteTable, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteRouteTableResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteRouteTableResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteRouteTableResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteRouteTableResponse")
	}
	return
}

// deleteRouteTable implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteRouteTable(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/routeTables/{rtId}")
	if err != nil {
		return nil, err
	}

	var response DeleteRouteTableResponse
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

// DeleteSecurityList Deletes the specified security list, but only if it's not associated with a subnet. You can't delete
// a VCN's default security list.
// This is an asynchronous operation. The security list's `lifecycleState` will change to TERMINATING temporarily
// until the security list is completely removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteSecurityList.go.html to see an example of how to use DeleteSecurityList API.
func (client VirtualNetworkClient) DeleteSecurityList(ctx context.Context, request DeleteSecurityListRequest) (response DeleteSecurityListResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteSecurityList, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteSecurityListResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteSecurityListResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteSecurityListResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteSecurityListResponse")
	}
	return
}

// deleteSecurityList implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteSecurityList(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/securityLists/{securityListId}")
	if err != nil {
		return nil, err
	}

	var response DeleteSecurityListResponse
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

// DeleteServiceGateway Deletes the specified service gateway. There must not be a route table that lists the service
// gateway as a target.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteServiceGateway.go.html to see an example of how to use DeleteServiceGateway API.
func (client VirtualNetworkClient) DeleteServiceGateway(ctx context.Context, request DeleteServiceGatewayRequest) (response DeleteServiceGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteServiceGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteServiceGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteServiceGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteServiceGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteServiceGatewayResponse")
	}
	return
}

// deleteServiceGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteServiceGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/serviceGateways/{serviceGatewayId}")
	if err != nil {
		return nil, err
	}

	var response DeleteServiceGatewayResponse
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

// DeleteSubnet Deletes the specified subnet, but only if there are no instances in the subnet. This is an asynchronous
// operation. The subnet's `lifecycleState` will change to TERMINATING temporarily. If there are any
// instances in the subnet, the state will instead change back to AVAILABLE.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteSubnet.go.html to see an example of how to use DeleteSubnet API.
func (client VirtualNetworkClient) DeleteSubnet(ctx context.Context, request DeleteSubnetRequest) (response DeleteSubnetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteSubnet, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteSubnetResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteSubnetResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteSubnetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteSubnetResponse")
	}
	return
}

// deleteSubnet implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteSubnet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/subnets/{subnetId}")
	if err != nil {
		return nil, err
	}

	var response DeleteSubnetResponse
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

// DeleteVcn Deletes the specified VCN. The VCN must be empty and have no attached gateways. This is an asynchronous
// operation. The VCN's `lifecycleState` will change to TERMINATING temporarily until the VCN is completely
// removed.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteVcn.go.html to see an example of how to use DeleteVcn API.
func (client VirtualNetworkClient) DeleteVcn(ctx context.Context, request DeleteVcnRequest) (response DeleteVcnResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteVcn, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteVcnResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteVcnResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteVcnResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteVcnResponse")
	}
	return
}

// deleteVcn implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteVcn(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/vcns/{vcnId}")
	if err != nil {
		return nil, err
	}

	var response DeleteVcnResponse
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

// DeleteVirtualCircuit Deletes the specified virtual circuit.
// **Important:** If you're using FastConnect via a provider,
// make sure to also terminate the connection with
// the provider, or else the provider may continue to bill you.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteVirtualCircuit.go.html to see an example of how to use DeleteVirtualCircuit API.
func (client VirtualNetworkClient) DeleteVirtualCircuit(ctx context.Context, request DeleteVirtualCircuitRequest) (response DeleteVirtualCircuitResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteVirtualCircuit, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteVirtualCircuitResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteVirtualCircuitResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteVirtualCircuitResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteVirtualCircuitResponse")
	}
	return
}

// deleteVirtualCircuit implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteVirtualCircuit(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/virtualCircuits/{virtualCircuitId}")
	if err != nil {
		return nil, err
	}

	var response DeleteVirtualCircuitResponse
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

// DeleteVlan Deletes the specified VLAN, but only if there are no VNICs in the VLAN.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DeleteVlan.go.html to see an example of how to use DeleteVlan API.
func (client VirtualNetworkClient) DeleteVlan(ctx context.Context, request DeleteVlanRequest) (response DeleteVlanResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.deleteVlan, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DeleteVlanResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DeleteVlanResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DeleteVlanResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DeleteVlanResponse")
	}
	return
}

// deleteVlan implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) deleteVlan(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodDelete, "/vlans/{vlanId}")
	if err != nil {
		return nil, err
	}

	var response DeleteVlanResponse
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

// DetachServiceId Removes the specified Service from the list of enabled
// `Service` objects for the specified gateway. You do not need to remove any route
// rules that specify this `Service` object's `cidrBlock` as the destination CIDR. However, consider
// removing the rules if your intent is to permanently disable use of the `Service` through this
// service gateway.
// **Note:** The `DetachServiceId` operation is an easy way to remove an individual `Service` from
// the service gateway. Compare it with
// UpdateServiceGateway, which replaces
// the entire existing list of enabled `Service` objects with the list that you provide in the
// `Update` call. `UpdateServiceGateway` also lets you block all traffic through the service
// gateway without having to remove each of the individual `Service` objects.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/DetachServiceId.go.html to see an example of how to use DetachServiceId API.
func (client VirtualNetworkClient) DetachServiceId(ctx context.Context, request DetachServiceIdRequest) (response DetachServiceIdResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.detachServiceId, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = DetachServiceIdResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = DetachServiceIdResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(DetachServiceIdResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into DetachServiceIdResponse")
	}
	return
}

// detachServiceId implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) detachServiceId(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/serviceGateways/{serviceGatewayId}/actions/detachService")
	if err != nil {
		return nil, err
	}

	var response DetachServiceIdResponse
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

// GetByoipRange Gets the `ByoipRange` resource. You must specify the OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetByoipRange.go.html to see an example of how to use GetByoipRange API.
func (client VirtualNetworkClient) GetByoipRange(ctx context.Context, request GetByoipRangeRequest) (response GetByoipRangeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getByoipRange, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetByoipRangeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetByoipRangeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetByoipRangeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetByoipRangeResponse")
	}
	return
}

// getByoipRange implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getByoipRange(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/byoipRanges/{byoipRangeId}")
	if err != nil {
		return nil, err
	}

	var response GetByoipRangeResponse
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

// GetCpe Gets the specified CPE's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetCpe.go.html to see an example of how to use GetCpe API.
func (client VirtualNetworkClient) GetCpe(ctx context.Context, request GetCpeRequest) (response GetCpeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCpe, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCpeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCpeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCpeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCpeResponse")
	}
	return
}

// getCpe implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getCpe(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/cpes/{cpeId}")
	if err != nil {
		return nil, err
	}

	var response GetCpeResponse
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

// GetCpeDeviceConfigContent Renders a set of CPE configuration content that can help a network engineer configure the actual
// CPE device (for example, a hardware router) represented by the specified Cpe
// object.
// The rendered content is specific to the type of CPE device (for example, Cisco ASA). Therefore the
// Cpe must have the CPE's device type specified by the `cpeDeviceShapeId`
// attribute. The content optionally includes answers that the customer provides (see
// UpdateTunnelCpeDeviceConfig),
// merged with a template of other information specific to the CPE device type.
// The operation returns configuration information for *all* of the
// IPSecConnection objects that use the specified CPE.
// Here are similar operations:
//   * GetIpsecCpeDeviceConfigContent
//   returns CPE configuration content for all tunnels in a single IPSec connection.
//   * GetTunnelCpeDeviceConfigContent
//   returns CPE configuration content for a specific tunnel within an IPSec connection.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetCpeDeviceConfigContent.go.html to see an example of how to use GetCpeDeviceConfigContent API.
func (client VirtualNetworkClient) GetCpeDeviceConfigContent(ctx context.Context, request GetCpeDeviceConfigContentRequest) (response GetCpeDeviceConfigContentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCpeDeviceConfigContent, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCpeDeviceConfigContentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCpeDeviceConfigContentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCpeDeviceConfigContentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCpeDeviceConfigContentResponse")
	}
	return
}

// getCpeDeviceConfigContent implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getCpeDeviceConfigContent(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/cpes/{cpeId}/cpeConfigContent")
	if err != nil {
		return nil, err
	}

	var response GetCpeDeviceConfigContentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetCpeDeviceShape Gets the detailed information about the specified CPE device type. This might include a set of questions
// that are specific to the particular CPE device type. The customer must supply answers to those questions
// (see UpdateTunnelCpeDeviceConfig).
// The service merges the answers with a template of other information for the CPE device type. The following
// operations return the merged content:
//   * GetCpeDeviceConfigContent
//   * GetIpsecCpeDeviceConfigContent
//   * GetTunnelCpeDeviceConfigContent
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetCpeDeviceShape.go.html to see an example of how to use GetCpeDeviceShape API.
func (client VirtualNetworkClient) GetCpeDeviceShape(ctx context.Context, request GetCpeDeviceShapeRequest) (response GetCpeDeviceShapeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCpeDeviceShape, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCpeDeviceShapeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCpeDeviceShapeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCpeDeviceShapeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCpeDeviceShapeResponse")
	}
	return
}

// getCpeDeviceShape implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getCpeDeviceShape(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/cpeDeviceShapes/{cpeDeviceShapeId}")
	if err != nil {
		return nil, err
	}

	var response GetCpeDeviceShapeResponse
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

// GetCrossConnect Gets the specified cross-connect's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetCrossConnect.go.html to see an example of how to use GetCrossConnect API.
func (client VirtualNetworkClient) GetCrossConnect(ctx context.Context, request GetCrossConnectRequest) (response GetCrossConnectResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCrossConnect, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCrossConnectResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCrossConnectResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCrossConnectResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCrossConnectResponse")
	}
	return
}

// getCrossConnect implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getCrossConnect(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnects/{crossConnectId}")
	if err != nil {
		return nil, err
	}

	var response GetCrossConnectResponse
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

// GetCrossConnectGroup Gets the specified cross-connect group's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetCrossConnectGroup.go.html to see an example of how to use GetCrossConnectGroup API.
func (client VirtualNetworkClient) GetCrossConnectGroup(ctx context.Context, request GetCrossConnectGroupRequest) (response GetCrossConnectGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCrossConnectGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCrossConnectGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCrossConnectGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCrossConnectGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCrossConnectGroupResponse")
	}
	return
}

// getCrossConnectGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getCrossConnectGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnectGroups/{crossConnectGroupId}")
	if err != nil {
		return nil, err
	}

	var response GetCrossConnectGroupResponse
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

// GetCrossConnectLetterOfAuthority Gets the Letter of Authority for the specified cross-connect.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetCrossConnectLetterOfAuthority.go.html to see an example of how to use GetCrossConnectLetterOfAuthority API.
func (client VirtualNetworkClient) GetCrossConnectLetterOfAuthority(ctx context.Context, request GetCrossConnectLetterOfAuthorityRequest) (response GetCrossConnectLetterOfAuthorityResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCrossConnectLetterOfAuthority, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCrossConnectLetterOfAuthorityResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCrossConnectLetterOfAuthorityResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCrossConnectLetterOfAuthorityResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCrossConnectLetterOfAuthorityResponse")
	}
	return
}

// getCrossConnectLetterOfAuthority implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getCrossConnectLetterOfAuthority(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnects/{crossConnectId}/letterOfAuthority")
	if err != nil {
		return nil, err
	}

	var response GetCrossConnectLetterOfAuthorityResponse
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

// GetCrossConnectStatus Gets the status of the specified cross-connect.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetCrossConnectStatus.go.html to see an example of how to use GetCrossConnectStatus API.
func (client VirtualNetworkClient) GetCrossConnectStatus(ctx context.Context, request GetCrossConnectStatusRequest) (response GetCrossConnectStatusResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getCrossConnectStatus, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetCrossConnectStatusResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetCrossConnectStatusResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetCrossConnectStatusResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetCrossConnectStatusResponse")
	}
	return
}

// getCrossConnectStatus implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getCrossConnectStatus(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnects/{crossConnectId}/status")
	if err != nil {
		return nil, err
	}

	var response GetCrossConnectStatusResponse
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

// GetDhcpOptions Gets the specified set of DHCP options.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetDhcpOptions.go.html to see an example of how to use GetDhcpOptions API.
func (client VirtualNetworkClient) GetDhcpOptions(ctx context.Context, request GetDhcpOptionsRequest) (response GetDhcpOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getDhcpOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetDhcpOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetDhcpOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetDhcpOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetDhcpOptionsResponse")
	}
	return
}

// getDhcpOptions implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getDhcpOptions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/dhcps/{dhcpId}")
	if err != nil {
		return nil, err
	}

	var response GetDhcpOptionsResponse
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

// GetDrg Gets the specified DRG's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetDrg.go.html to see an example of how to use GetDrg API.
func (client VirtualNetworkClient) GetDrg(ctx context.Context, request GetDrgRequest) (response GetDrgResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getDrg, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetDrgResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetDrgResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetDrgResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetDrgResponse")
	}
	return
}

// getDrg implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getDrg(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/drgs/{drgId}")
	if err != nil {
		return nil, err
	}

	var response GetDrgResponse
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

// GetDrgAttachment Gets the information for the specified `DrgAttachment`.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetDrgAttachment.go.html to see an example of how to use GetDrgAttachment API.
func (client VirtualNetworkClient) GetDrgAttachment(ctx context.Context, request GetDrgAttachmentRequest) (response GetDrgAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getDrgAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetDrgAttachmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetDrgAttachmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetDrgAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetDrgAttachmentResponse")
	}
	return
}

// getDrgAttachment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getDrgAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/drgAttachments/{drgAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response GetDrgAttachmentResponse
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

// GetDrgRedundancyStatus Gets the redundancy status for the specified DRG. For more information, see
// Redundancy Remedies (https://docs.cloud.oracle.com/iaas/Content/Network/Troubleshoot/drgredundancy.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetDrgRedundancyStatus.go.html to see an example of how to use GetDrgRedundancyStatus API.
func (client VirtualNetworkClient) GetDrgRedundancyStatus(ctx context.Context, request GetDrgRedundancyStatusRequest) (response GetDrgRedundancyStatusResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getDrgRedundancyStatus, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetDrgRedundancyStatusResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetDrgRedundancyStatusResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetDrgRedundancyStatusResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetDrgRedundancyStatusResponse")
	}
	return
}

// getDrgRedundancyStatus implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getDrgRedundancyStatus(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/drgs/{drgId}/redundancyStatus")
	if err != nil {
		return nil, err
	}

	var response GetDrgRedundancyStatusResponse
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

// GetFastConnectProviderService Gets the specified provider service.
// For more information, see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetFastConnectProviderService.go.html to see an example of how to use GetFastConnectProviderService API.
func (client VirtualNetworkClient) GetFastConnectProviderService(ctx context.Context, request GetFastConnectProviderServiceRequest) (response GetFastConnectProviderServiceResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getFastConnectProviderService, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetFastConnectProviderServiceResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetFastConnectProviderServiceResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetFastConnectProviderServiceResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetFastConnectProviderServiceResponse")
	}
	return
}

// getFastConnectProviderService implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getFastConnectProviderService(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/fastConnectProviderServices/{providerServiceId}")
	if err != nil {
		return nil, err
	}

	var response GetFastConnectProviderServiceResponse
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

// GetFastConnectProviderServiceKey Gets the specified provider service key's information. Use this operation to validate a
// provider service key. An invalid key returns a 404 error.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetFastConnectProviderServiceKey.go.html to see an example of how to use GetFastConnectProviderServiceKey API.
func (client VirtualNetworkClient) GetFastConnectProviderServiceKey(ctx context.Context, request GetFastConnectProviderServiceKeyRequest) (response GetFastConnectProviderServiceKeyResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getFastConnectProviderServiceKey, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetFastConnectProviderServiceKeyResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetFastConnectProviderServiceKeyResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetFastConnectProviderServiceKeyResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetFastConnectProviderServiceKeyResponse")
	}
	return
}

// getFastConnectProviderServiceKey implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getFastConnectProviderServiceKey(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/fastConnectProviderServices/{providerServiceId}/providerServiceKeys/{providerServiceKeyName}")
	if err != nil {
		return nil, err
	}

	var response GetFastConnectProviderServiceKeyResponse
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

// GetIPSecConnection Gets the specified IPSec connection's basic information, including the static routes for the
// on-premises router. If you want the status of the connection (whether it's up or down), use
// GetIPSecConnectionTunnel.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetIPSecConnection.go.html to see an example of how to use GetIPSecConnection API.
func (client VirtualNetworkClient) GetIPSecConnection(ctx context.Context, request GetIPSecConnectionRequest) (response GetIPSecConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getIPSecConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetIPSecConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetIPSecConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetIPSecConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetIPSecConnectionResponse")
	}
	return
}

// getIPSecConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getIPSecConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}")
	if err != nil {
		return nil, err
	}

	var response GetIPSecConnectionResponse
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

// GetIPSecConnectionDeviceConfig Deprecated. To get tunnel information, instead use:
// * GetIPSecConnectionTunnel
// * GetIPSecConnectionTunnelSharedSecret
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetIPSecConnectionDeviceConfig.go.html to see an example of how to use GetIPSecConnectionDeviceConfig API.
func (client VirtualNetworkClient) GetIPSecConnectionDeviceConfig(ctx context.Context, request GetIPSecConnectionDeviceConfigRequest) (response GetIPSecConnectionDeviceConfigResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getIPSecConnectionDeviceConfig, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetIPSecConnectionDeviceConfigResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetIPSecConnectionDeviceConfigResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetIPSecConnectionDeviceConfigResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetIPSecConnectionDeviceConfigResponse")
	}
	return
}

// getIPSecConnectionDeviceConfig implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getIPSecConnectionDeviceConfig(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/deviceConfig")
	if err != nil {
		return nil, err
	}

	var response GetIPSecConnectionDeviceConfigResponse
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

// GetIPSecConnectionDeviceStatus Deprecated. To get the tunnel status, instead use
// GetIPSecConnectionTunnel.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetIPSecConnectionDeviceStatus.go.html to see an example of how to use GetIPSecConnectionDeviceStatus API.
func (client VirtualNetworkClient) GetIPSecConnectionDeviceStatus(ctx context.Context, request GetIPSecConnectionDeviceStatusRequest) (response GetIPSecConnectionDeviceStatusResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getIPSecConnectionDeviceStatus, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetIPSecConnectionDeviceStatusResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetIPSecConnectionDeviceStatusResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetIPSecConnectionDeviceStatusResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetIPSecConnectionDeviceStatusResponse")
	}
	return
}

// getIPSecConnectionDeviceStatus implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getIPSecConnectionDeviceStatus(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/deviceStatus")
	if err != nil {
		return nil, err
	}

	var response GetIPSecConnectionDeviceStatusResponse
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

// GetIPSecConnectionTunnel Gets the specified tunnel's information. The resulting object does not include the tunnel's
// shared secret (pre-shared key). To retrieve that, use
// GetIPSecConnectionTunnelSharedSecret.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetIPSecConnectionTunnel.go.html to see an example of how to use GetIPSecConnectionTunnel API.
func (client VirtualNetworkClient) GetIPSecConnectionTunnel(ctx context.Context, request GetIPSecConnectionTunnelRequest) (response GetIPSecConnectionTunnelResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getIPSecConnectionTunnel, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetIPSecConnectionTunnelResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetIPSecConnectionTunnelResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetIPSecConnectionTunnelResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetIPSecConnectionTunnelResponse")
	}
	return
}

// getIPSecConnectionTunnel implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getIPSecConnectionTunnel(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/tunnels/{tunnelId}")
	if err != nil {
		return nil, err
	}

	var response GetIPSecConnectionTunnelResponse
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

// GetIPSecConnectionTunnelSharedSecret Gets the specified tunnel's shared secret (pre-shared key). To get other information
// about the tunnel, use GetIPSecConnectionTunnel.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetIPSecConnectionTunnelSharedSecret.go.html to see an example of how to use GetIPSecConnectionTunnelSharedSecret API.
func (client VirtualNetworkClient) GetIPSecConnectionTunnelSharedSecret(ctx context.Context, request GetIPSecConnectionTunnelSharedSecretRequest) (response GetIPSecConnectionTunnelSharedSecretResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getIPSecConnectionTunnelSharedSecret, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetIPSecConnectionTunnelSharedSecretResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetIPSecConnectionTunnelSharedSecretResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetIPSecConnectionTunnelSharedSecretResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetIPSecConnectionTunnelSharedSecretResponse")
	}
	return
}

// getIPSecConnectionTunnelSharedSecret implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getIPSecConnectionTunnelSharedSecret(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/tunnels/{tunnelId}/sharedSecret")
	if err != nil {
		return nil, err
	}

	var response GetIPSecConnectionTunnelSharedSecretResponse
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

// GetInternetGateway Gets the specified internet gateway's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetInternetGateway.go.html to see an example of how to use GetInternetGateway API.
func (client VirtualNetworkClient) GetInternetGateway(ctx context.Context, request GetInternetGatewayRequest) (response GetInternetGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getInternetGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetInternetGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetInternetGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetInternetGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetInternetGatewayResponse")
	}
	return
}

// getInternetGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getInternetGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/internetGateways/{igId}")
	if err != nil {
		return nil, err
	}

	var response GetInternetGatewayResponse
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

// GetIpsecCpeDeviceConfigContent Renders a set of CPE configuration content for the specified IPSec connection (for all the
// tunnels in the connection). The content helps a network engineer configure the actual CPE
// device (for example, a hardware router) that the specified IPSec connection terminates on.
// The rendered content is specific to the type of CPE device (for example, Cisco ASA). Therefore the
// Cpe used by the specified IPSecConnection
// must have the CPE's device type specified by the `cpeDeviceShapeId` attribute. The content
// optionally includes answers that the customer provides (see
// UpdateTunnelCpeDeviceConfig),
// merged with a template of other information specific to the CPE device type.
// The operation returns configuration information for all tunnels in the single specified
// IPSecConnection object. Here are other similar
// operations:
//   * GetTunnelCpeDeviceConfigContent
//   returns CPE configuration content for a specific tunnel within an IPSec connection.
//   * GetCpeDeviceConfigContent
//   returns CPE configuration content for *all* IPSec connections that use a specific CPE.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetIpsecCpeDeviceConfigContent.go.html to see an example of how to use GetIpsecCpeDeviceConfigContent API.
func (client VirtualNetworkClient) GetIpsecCpeDeviceConfigContent(ctx context.Context, request GetIpsecCpeDeviceConfigContentRequest) (response GetIpsecCpeDeviceConfigContentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getIpsecCpeDeviceConfigContent, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetIpsecCpeDeviceConfigContentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetIpsecCpeDeviceConfigContentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetIpsecCpeDeviceConfigContentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetIpsecCpeDeviceConfigContentResponse")
	}
	return
}

// getIpsecCpeDeviceConfigContent implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getIpsecCpeDeviceConfigContent(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/cpeConfigContent")
	if err != nil {
		return nil, err
	}

	var response GetIpsecCpeDeviceConfigContentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetIpv6 Gets the specified IPv6. You must specify the object's OCID.
// Alternatively, you can get the object by using
// ListIpv6s
// with the IPv6 address (for example, 2001:0db8:0123:1111:98fe:dcba:9876:4321) and subnet OCID.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetIpv6.go.html to see an example of how to use GetIpv6 API.
func (client VirtualNetworkClient) GetIpv6(ctx context.Context, request GetIpv6Request) (response GetIpv6Response, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getIpv6, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetIpv6Response{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetIpv6Response{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetIpv6Response); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetIpv6Response")
	}
	return
}

// getIpv6 implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getIpv6(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipv6/{ipv6Id}")
	if err != nil {
		return nil, err
	}

	var response GetIpv6Response
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

// GetLocalPeeringGateway Gets the specified local peering gateway's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetLocalPeeringGateway.go.html to see an example of how to use GetLocalPeeringGateway API.
func (client VirtualNetworkClient) GetLocalPeeringGateway(ctx context.Context, request GetLocalPeeringGatewayRequest) (response GetLocalPeeringGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getLocalPeeringGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetLocalPeeringGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetLocalPeeringGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetLocalPeeringGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetLocalPeeringGatewayResponse")
	}
	return
}

// getLocalPeeringGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getLocalPeeringGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/localPeeringGateways/{localPeeringGatewayId}")
	if err != nil {
		return nil, err
	}

	var response GetLocalPeeringGatewayResponse
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

// GetNatGateway Gets the specified NAT gateway's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetNatGateway.go.html to see an example of how to use GetNatGateway API.
func (client VirtualNetworkClient) GetNatGateway(ctx context.Context, request GetNatGatewayRequest) (response GetNatGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getNatGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetNatGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetNatGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetNatGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetNatGatewayResponse")
	}
	return
}

// getNatGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getNatGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/natGateways/{natGatewayId}")
	if err != nil {
		return nil, err
	}

	var response GetNatGatewayResponse
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

// GetNetworkSecurityGroup Gets the specified network security group's information.
// To list the VNICs in an NSG, see
// ListNetworkSecurityGroupVnics.
// To list the security rules in an NSG, see
// ListNetworkSecurityGroupSecurityRules.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetNetworkSecurityGroup.go.html to see an example of how to use GetNetworkSecurityGroup API.
func (client VirtualNetworkClient) GetNetworkSecurityGroup(ctx context.Context, request GetNetworkSecurityGroupRequest) (response GetNetworkSecurityGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getNetworkSecurityGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetNetworkSecurityGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetNetworkSecurityGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetNetworkSecurityGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetNetworkSecurityGroupResponse")
	}
	return
}

// getNetworkSecurityGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getNetworkSecurityGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/networkSecurityGroups/{networkSecurityGroupId}")
	if err != nil {
		return nil, err
	}

	var response GetNetworkSecurityGroupResponse
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

// GetPrivateIp Gets the specified private IP. You must specify the object's OCID.
// Alternatively, you can get the object by using
// ListPrivateIps
// with the private IP address (for example, 10.0.3.3) and subnet OCID.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetPrivateIp.go.html to see an example of how to use GetPrivateIp API.
func (client VirtualNetworkClient) GetPrivateIp(ctx context.Context, request GetPrivateIpRequest) (response GetPrivateIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getPrivateIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetPrivateIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetPrivateIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetPrivateIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetPrivateIpResponse")
	}
	return
}

// getPrivateIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getPrivateIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/privateIps/{privateIpId}")
	if err != nil {
		return nil, err
	}

	var response GetPrivateIpResponse
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

// GetPublicIp Gets the specified public IP. You must specify the object's OCID.
// Alternatively, you can get the object by using GetPublicIpByIpAddress
// with the public IP address (for example, 203.0.113.2).
// Or you can use GetPublicIpByPrivateIpId
// with the OCID of the private IP that the public IP is assigned to.
// **Note:** If you're fetching a reserved public IP that is in the process of being
// moved to a different private IP, the service returns the public IP object with
// `lifecycleState` = ASSIGNING and `assignedEntityId` = OCID of the target private IP.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetPublicIp.go.html to see an example of how to use GetPublicIp API.
func (client VirtualNetworkClient) GetPublicIp(ctx context.Context, request GetPublicIpRequest) (response GetPublicIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getPublicIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetPublicIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetPublicIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetPublicIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetPublicIpResponse")
	}
	return
}

// getPublicIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getPublicIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/publicIps/{publicIpId}")
	if err != nil {
		return nil, err
	}

	var response GetPublicIpResponse
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

// GetPublicIpByIpAddress Gets the public IP based on the public IP address (for example, 203.0.113.2).
// **Note:** If you're fetching a reserved public IP that is in the process of being
// moved to a different private IP, the service returns the public IP object with
// `lifecycleState` = ASSIGNING and `assignedEntityId` = OCID of the target private IP.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetPublicIpByIpAddress.go.html to see an example of how to use GetPublicIpByIpAddress API.
func (client VirtualNetworkClient) GetPublicIpByIpAddress(ctx context.Context, request GetPublicIpByIpAddressRequest) (response GetPublicIpByIpAddressResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getPublicIpByIpAddress, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetPublicIpByIpAddressResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetPublicIpByIpAddressResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetPublicIpByIpAddressResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetPublicIpByIpAddressResponse")
	}
	return
}

// getPublicIpByIpAddress implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getPublicIpByIpAddress(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIps/actions/getByIpAddress")
	if err != nil {
		return nil, err
	}

	var response GetPublicIpByIpAddressResponse
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

// GetPublicIpByPrivateIpId Gets the public IP assigned to the specified private IP. You must specify the OCID
// of the private IP. If no public IP is assigned, a 404 is returned.
// **Note:** If you're fetching a reserved public IP that is in the process of being
// moved to a different private IP, and you provide the OCID of the original private
// IP, this operation returns a 404. If you instead provide the OCID of the target
// private IP, or if you instead call
// GetPublicIp or
// GetPublicIpByIpAddress, the
// service returns the public IP object with `lifecycleState` = ASSIGNING and
// `assignedEntityId` = OCID of the target private IP.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetPublicIpByPrivateIpId.go.html to see an example of how to use GetPublicIpByPrivateIpId API.
func (client VirtualNetworkClient) GetPublicIpByPrivateIpId(ctx context.Context, request GetPublicIpByPrivateIpIdRequest) (response GetPublicIpByPrivateIpIdResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getPublicIpByPrivateIpId, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetPublicIpByPrivateIpIdResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetPublicIpByPrivateIpIdResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetPublicIpByPrivateIpIdResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetPublicIpByPrivateIpIdResponse")
	}
	return
}

// getPublicIpByPrivateIpId implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getPublicIpByPrivateIpId(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIps/actions/getByPrivateIpId")
	if err != nil {
		return nil, err
	}

	var response GetPublicIpByPrivateIpIdResponse
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

// GetPublicIpPool Gets the specified `PublicIpPool` object. You must specify the object's OCID (https://docs.cloud.oracle.com/iaas/Content/General/Concepts/identifiers.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetPublicIpPool.go.html to see an example of how to use GetPublicIpPool API.
func (client VirtualNetworkClient) GetPublicIpPool(ctx context.Context, request GetPublicIpPoolRequest) (response GetPublicIpPoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getPublicIpPool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetPublicIpPoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetPublicIpPoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetPublicIpPoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetPublicIpPoolResponse")
	}
	return
}

// getPublicIpPool implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getPublicIpPool(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/publicIpPools/{publicIpPoolId}")
	if err != nil {
		return nil, err
	}

	var response GetPublicIpPoolResponse
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

// GetRemotePeeringConnection Get the specified remote peering connection's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetRemotePeeringConnection.go.html to see an example of how to use GetRemotePeeringConnection API.
func (client VirtualNetworkClient) GetRemotePeeringConnection(ctx context.Context, request GetRemotePeeringConnectionRequest) (response GetRemotePeeringConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getRemotePeeringConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetRemotePeeringConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetRemotePeeringConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetRemotePeeringConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetRemotePeeringConnectionResponse")
	}
	return
}

// getRemotePeeringConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getRemotePeeringConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/remotePeeringConnections/{remotePeeringConnectionId}")
	if err != nil {
		return nil, err
	}

	var response GetRemotePeeringConnectionResponse
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

// GetRouteTable Gets the specified route table's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetRouteTable.go.html to see an example of how to use GetRouteTable API.
func (client VirtualNetworkClient) GetRouteTable(ctx context.Context, request GetRouteTableRequest) (response GetRouteTableResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getRouteTable, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetRouteTableResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetRouteTableResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetRouteTableResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetRouteTableResponse")
	}
	return
}

// getRouteTable implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getRouteTable(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/routeTables/{rtId}")
	if err != nil {
		return nil, err
	}

	var response GetRouteTableResponse
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

// GetSecurityList Gets the specified security list's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetSecurityList.go.html to see an example of how to use GetSecurityList API.
func (client VirtualNetworkClient) GetSecurityList(ctx context.Context, request GetSecurityListRequest) (response GetSecurityListResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getSecurityList, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetSecurityListResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetSecurityListResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetSecurityListResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetSecurityListResponse")
	}
	return
}

// getSecurityList implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getSecurityList(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/securityLists/{securityListId}")
	if err != nil {
		return nil, err
	}

	var response GetSecurityListResponse
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

// GetService Gets the specified Service object.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetService.go.html to see an example of how to use GetService API.
func (client VirtualNetworkClient) GetService(ctx context.Context, request GetServiceRequest) (response GetServiceResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getService, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetServiceResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetServiceResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetServiceResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetServiceResponse")
	}
	return
}

// getService implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getService(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/services/{serviceId}")
	if err != nil {
		return nil, err
	}

	var response GetServiceResponse
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

// GetServiceGateway Gets the specified service gateway's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetServiceGateway.go.html to see an example of how to use GetServiceGateway API.
func (client VirtualNetworkClient) GetServiceGateway(ctx context.Context, request GetServiceGatewayRequest) (response GetServiceGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getServiceGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetServiceGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetServiceGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetServiceGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetServiceGatewayResponse")
	}
	return
}

// getServiceGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getServiceGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/serviceGateways/{serviceGatewayId}")
	if err != nil {
		return nil, err
	}

	var response GetServiceGatewayResponse
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

// GetSubnet Gets the specified subnet's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetSubnet.go.html to see an example of how to use GetSubnet API.
func (client VirtualNetworkClient) GetSubnet(ctx context.Context, request GetSubnetRequest) (response GetSubnetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getSubnet, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetSubnetResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetSubnetResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetSubnetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetSubnetResponse")
	}
	return
}

// getSubnet implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getSubnet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/subnets/{subnetId}")
	if err != nil {
		return nil, err
	}

	var response GetSubnetResponse
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

// GetTunnelCpeDeviceConfig Gets the set of CPE configuration answers for the tunnel, which the customer provided in
// UpdateTunnelCpeDeviceConfig.
// To get the full set of content for the tunnel (any answers merged with the template of other
// information specific to the CPE device type), use
// GetTunnelCpeDeviceConfigContent.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetTunnelCpeDeviceConfig.go.html to see an example of how to use GetTunnelCpeDeviceConfig API.
func (client VirtualNetworkClient) GetTunnelCpeDeviceConfig(ctx context.Context, request GetTunnelCpeDeviceConfigRequest) (response GetTunnelCpeDeviceConfigResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getTunnelCpeDeviceConfig, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetTunnelCpeDeviceConfigResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetTunnelCpeDeviceConfigResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetTunnelCpeDeviceConfigResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetTunnelCpeDeviceConfigResponse")
	}
	return
}

// getTunnelCpeDeviceConfig implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getTunnelCpeDeviceConfig(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/tunnels/{tunnelId}/tunnelDeviceConfig")
	if err != nil {
		return nil, err
	}

	var response GetTunnelCpeDeviceConfigResponse
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

// GetTunnelCpeDeviceConfigContent Renders a set of CPE configuration content for the specified IPSec tunnel. The content helps a
// network engineer configure the actual CPE device (for example, a hardware router) that the specified
// IPSec tunnel terminates on.
// The rendered content is specific to the type of CPE device (for example, Cisco ASA). Therefore the
// Cpe used by the specified IPSecConnection
// must have the CPE's device type specified by the `cpeDeviceShapeId` attribute. The content
// optionally includes answers that the customer provides (see
// UpdateTunnelCpeDeviceConfig),
// merged with a template of other information specific to the CPE device type.
// The operation returns configuration information for only the specified IPSec tunnel.
// Here are other similar operations:
//   * GetIpsecCpeDeviceConfigContent
//   returns CPE configuration content for all tunnels in a single IPSec connection.
//   * GetCpeDeviceConfigContent
//   returns CPE configuration content for *all* IPSec connections that use a specific CPE.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetTunnelCpeDeviceConfigContent.go.html to see an example of how to use GetTunnelCpeDeviceConfigContent API.
func (client VirtualNetworkClient) GetTunnelCpeDeviceConfigContent(ctx context.Context, request GetTunnelCpeDeviceConfigContentRequest) (response GetTunnelCpeDeviceConfigContentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getTunnelCpeDeviceConfigContent, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetTunnelCpeDeviceConfigContentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetTunnelCpeDeviceConfigContentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetTunnelCpeDeviceConfigContentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetTunnelCpeDeviceConfigContentResponse")
	}
	return
}

// getTunnelCpeDeviceConfigContent implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getTunnelCpeDeviceConfigContent(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/tunnels/{tunnelId}/tunnelDeviceConfig/content")
	if err != nil {
		return nil, err
	}

	var response GetTunnelCpeDeviceConfigContentResponse
	var httpResponse *http.Response
	httpResponse, err = client.Call(ctx, &httpRequest)
	response.RawResponse = httpResponse
	if err != nil {
		return response, err
	}

	err = common.UnmarshalResponse(httpResponse, &response)
	return response, err
}

// GetVcn Gets the specified VCN's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetVcn.go.html to see an example of how to use GetVcn API.
func (client VirtualNetworkClient) GetVcn(ctx context.Context, request GetVcnRequest) (response GetVcnResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVcn, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVcnResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVcnResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVcnResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVcnResponse")
	}
	return
}

// getVcn implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getVcn(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vcns/{vcnId}")
	if err != nil {
		return nil, err
	}

	var response GetVcnResponse
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

// GetVcnDnsResolverAssociation Get the associated DNS resolver information with a vcn
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetVcnDnsResolverAssociation.go.html to see an example of how to use GetVcnDnsResolverAssociation API.
func (client VirtualNetworkClient) GetVcnDnsResolverAssociation(ctx context.Context, request GetVcnDnsResolverAssociationRequest) (response GetVcnDnsResolverAssociationResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVcnDnsResolverAssociation, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVcnDnsResolverAssociationResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVcnDnsResolverAssociationResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVcnDnsResolverAssociationResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVcnDnsResolverAssociationResponse")
	}
	return
}

// getVcnDnsResolverAssociation implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getVcnDnsResolverAssociation(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vcns/{vcnId}/dnsResolverAssociation")
	if err != nil {
		return nil, err
	}

	var response GetVcnDnsResolverAssociationResponse
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

// GetVirtualCircuit Gets the specified virtual circuit's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetVirtualCircuit.go.html to see an example of how to use GetVirtualCircuit API.
func (client VirtualNetworkClient) GetVirtualCircuit(ctx context.Context, request GetVirtualCircuitRequest) (response GetVirtualCircuitResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVirtualCircuit, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVirtualCircuitResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVirtualCircuitResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVirtualCircuitResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVirtualCircuitResponse")
	}
	return
}

// getVirtualCircuit implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getVirtualCircuit(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualCircuits/{virtualCircuitId}")
	if err != nil {
		return nil, err
	}

	var response GetVirtualCircuitResponse
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

// GetVlan Gets the specified VLAN's information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetVlan.go.html to see an example of how to use GetVlan API.
func (client VirtualNetworkClient) GetVlan(ctx context.Context, request GetVlanRequest) (response GetVlanResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVlan, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVlanResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVlanResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVlanResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVlanResponse")
	}
	return
}

// getVlan implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getVlan(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vlans/{vlanId}")
	if err != nil {
		return nil, err
	}

	var response GetVlanResponse
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

// GetVnic Gets the information for the specified virtual network interface card (VNIC).
// You can get the VNIC OCID from the
// ListVnicAttachments
// operation.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/GetVnic.go.html to see an example of how to use GetVnic API.
func (client VirtualNetworkClient) GetVnic(ctx context.Context, request GetVnicRequest) (response GetVnicResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.getVnic, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = GetVnicResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = GetVnicResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(GetVnicResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into GetVnicResponse")
	}
	return
}

// getVnic implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) getVnic(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vnics/{vnicId}")
	if err != nil {
		return nil, err
	}

	var response GetVnicResponse
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

// ListAllowedPeerRegionsForRemotePeering Lists the regions that support remote VCN peering (which is peering across regions).
// For more information, see VCN Peering (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/VCNpeering.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListAllowedPeerRegionsForRemotePeering.go.html to see an example of how to use ListAllowedPeerRegionsForRemotePeering API.
func (client VirtualNetworkClient) ListAllowedPeerRegionsForRemotePeering(ctx context.Context, request ListAllowedPeerRegionsForRemotePeeringRequest) (response ListAllowedPeerRegionsForRemotePeeringResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listAllowedPeerRegionsForRemotePeering, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListAllowedPeerRegionsForRemotePeeringResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListAllowedPeerRegionsForRemotePeeringResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListAllowedPeerRegionsForRemotePeeringResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListAllowedPeerRegionsForRemotePeeringResponse")
	}
	return
}

// listAllowedPeerRegionsForRemotePeering implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listAllowedPeerRegionsForRemotePeering(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/allowedPeerRegionsForRemotePeering")
	if err != nil {
		return nil, err
	}

	var response ListAllowedPeerRegionsForRemotePeeringResponse
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

// ListByoipAllocatedRanges Lists the subranges of a BYOIP CIDR block currently allocated to an IP pool.
// Each `ByoipAllocatedRange` object also lists the IP pool where it is allocated.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListByoipAllocatedRanges.go.html to see an example of how to use ListByoipAllocatedRanges API.
func (client VirtualNetworkClient) ListByoipAllocatedRanges(ctx context.Context, request ListByoipAllocatedRangesRequest) (response ListByoipAllocatedRangesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listByoipAllocatedRanges, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListByoipAllocatedRangesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListByoipAllocatedRangesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListByoipAllocatedRangesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListByoipAllocatedRangesResponse")
	}
	return
}

// listByoipAllocatedRanges implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listByoipAllocatedRanges(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/byoipRanges/{byoipRangeId}/byoipAllocatedRanges")
	if err != nil {
		return nil, err
	}

	var response ListByoipAllocatedRangesResponse
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

// ListByoipRanges Lists the `ByoipRange` resources in the specified compartment.
// You can filter the list using query parameters.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListByoipRanges.go.html to see an example of how to use ListByoipRanges API.
func (client VirtualNetworkClient) ListByoipRanges(ctx context.Context, request ListByoipRangesRequest) (response ListByoipRangesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listByoipRanges, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListByoipRangesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListByoipRangesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListByoipRangesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListByoipRangesResponse")
	}
	return
}

// listByoipRanges implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listByoipRanges(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/byoipRanges")
	if err != nil {
		return nil, err
	}

	var response ListByoipRangesResponse
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

// ListCpeDeviceShapes Lists the CPE device types that the Networking service provides CPE configuration
// content for (example: Cisco ASA). The content helps a network engineer configure
// the actual CPE device represented by a Cpe object.
// If you want to generate CPE configuration content for one of the returned CPE device types,
// ensure that the Cpe object's `cpeDeviceShapeId` attribute is set
// to the CPE device type's OCID (returned by this operation).
// For information about generating CPE configuration content, see these operations:
//   * GetCpeDeviceConfigContent
//   * GetIpsecCpeDeviceConfigContent
//   * GetTunnelCpeDeviceConfigContent
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCpeDeviceShapes.go.html to see an example of how to use ListCpeDeviceShapes API.
func (client VirtualNetworkClient) ListCpeDeviceShapes(ctx context.Context, request ListCpeDeviceShapesRequest) (response ListCpeDeviceShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listCpeDeviceShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListCpeDeviceShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListCpeDeviceShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListCpeDeviceShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListCpeDeviceShapesResponse")
	}
	return
}

// listCpeDeviceShapes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listCpeDeviceShapes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/cpeDeviceShapes")
	if err != nil {
		return nil, err
	}

	var response ListCpeDeviceShapesResponse
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

// ListCpes Lists the customer-premises equipment objects (CPEs) in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCpes.go.html to see an example of how to use ListCpes API.
func (client VirtualNetworkClient) ListCpes(ctx context.Context, request ListCpesRequest) (response ListCpesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listCpes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListCpesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListCpesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListCpesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListCpesResponse")
	}
	return
}

// listCpes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listCpes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/cpes")
	if err != nil {
		return nil, err
	}

	var response ListCpesResponse
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

// ListCrossConnectGroups Lists the cross-connect groups in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCrossConnectGroups.go.html to see an example of how to use ListCrossConnectGroups API.
func (client VirtualNetworkClient) ListCrossConnectGroups(ctx context.Context, request ListCrossConnectGroupsRequest) (response ListCrossConnectGroupsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listCrossConnectGroups, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListCrossConnectGroupsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListCrossConnectGroupsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListCrossConnectGroupsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListCrossConnectGroupsResponse")
	}
	return
}

// listCrossConnectGroups implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listCrossConnectGroups(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnectGroups")
	if err != nil {
		return nil, err
	}

	var response ListCrossConnectGroupsResponse
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

// ListCrossConnectLocations Lists the available FastConnect locations for cross-connect installation. You need
// this information so you can specify your desired location when you create a cross-connect.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCrossConnectLocations.go.html to see an example of how to use ListCrossConnectLocations API.
func (client VirtualNetworkClient) ListCrossConnectLocations(ctx context.Context, request ListCrossConnectLocationsRequest) (response ListCrossConnectLocationsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listCrossConnectLocations, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListCrossConnectLocationsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListCrossConnectLocationsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListCrossConnectLocationsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListCrossConnectLocationsResponse")
	}
	return
}

// listCrossConnectLocations implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listCrossConnectLocations(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnectLocations")
	if err != nil {
		return nil, err
	}

	var response ListCrossConnectLocationsResponse
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

// ListCrossConnects Lists the cross-connects in the specified compartment. You can filter the list
// by specifying the OCID of a cross-connect group.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCrossConnects.go.html to see an example of how to use ListCrossConnects API.
func (client VirtualNetworkClient) ListCrossConnects(ctx context.Context, request ListCrossConnectsRequest) (response ListCrossConnectsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listCrossConnects, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListCrossConnectsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListCrossConnectsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListCrossConnectsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListCrossConnectsResponse")
	}
	return
}

// listCrossConnects implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listCrossConnects(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnects")
	if err != nil {
		return nil, err
	}

	var response ListCrossConnectsResponse
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

// ListCrossconnectPortSpeedShapes Lists the available port speeds for cross-connects. You need this information
// so you can specify your desired port speed (that is, shape) when you create a
// cross-connect.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListCrossconnectPortSpeedShapes.go.html to see an example of how to use ListCrossconnectPortSpeedShapes API.
func (client VirtualNetworkClient) ListCrossconnectPortSpeedShapes(ctx context.Context, request ListCrossconnectPortSpeedShapesRequest) (response ListCrossconnectPortSpeedShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listCrossconnectPortSpeedShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListCrossconnectPortSpeedShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListCrossconnectPortSpeedShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListCrossconnectPortSpeedShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListCrossconnectPortSpeedShapesResponse")
	}
	return
}

// listCrossconnectPortSpeedShapes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listCrossconnectPortSpeedShapes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/crossConnectPortSpeedShapes")
	if err != nil {
		return nil, err
	}

	var response ListCrossconnectPortSpeedShapesResponse
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

// ListDhcpOptions Lists the sets of DHCP options in the specified VCN and specified compartment.
// If the VCN ID is not provided, then the list includes the sets of DHCP options from all VCNs in the specified compartment.
// The response includes the default set of options that automatically comes with each VCN,
// plus any other sets you've created.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListDhcpOptions.go.html to see an example of how to use ListDhcpOptions API.
func (client VirtualNetworkClient) ListDhcpOptions(ctx context.Context, request ListDhcpOptionsRequest) (response ListDhcpOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listDhcpOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListDhcpOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListDhcpOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListDhcpOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListDhcpOptionsResponse")
	}
	return
}

// listDhcpOptions implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listDhcpOptions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/dhcps")
	if err != nil {
		return nil, err
	}

	var response ListDhcpOptionsResponse
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

// ListDrgAttachments Lists the `DrgAttachment` objects for the specified compartment. You can filter the
// results by VCN or DRG.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListDrgAttachments.go.html to see an example of how to use ListDrgAttachments API.
func (client VirtualNetworkClient) ListDrgAttachments(ctx context.Context, request ListDrgAttachmentsRequest) (response ListDrgAttachmentsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listDrgAttachments, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListDrgAttachmentsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListDrgAttachmentsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListDrgAttachmentsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListDrgAttachmentsResponse")
	}
	return
}

// listDrgAttachments implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listDrgAttachments(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/drgAttachments")
	if err != nil {
		return nil, err
	}

	var response ListDrgAttachmentsResponse
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

// ListDrgs Lists the DRGs in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListDrgs.go.html to see an example of how to use ListDrgs API.
func (client VirtualNetworkClient) ListDrgs(ctx context.Context, request ListDrgsRequest) (response ListDrgsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listDrgs, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListDrgsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListDrgsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListDrgsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListDrgsResponse")
	}
	return
}

// listDrgs implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listDrgs(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/drgs")
	if err != nil {
		return nil, err
	}

	var response ListDrgsResponse
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

// ListFastConnectProviderServices Lists the service offerings from supported providers. You need this
// information so you can specify your desired provider and service
// offering when you create a virtual circuit.
// For the compartment ID, provide the OCID of your tenancy (the root compartment).
// For more information, see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListFastConnectProviderServices.go.html to see an example of how to use ListFastConnectProviderServices API.
func (client VirtualNetworkClient) ListFastConnectProviderServices(ctx context.Context, request ListFastConnectProviderServicesRequest) (response ListFastConnectProviderServicesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listFastConnectProviderServices, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListFastConnectProviderServicesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListFastConnectProviderServicesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListFastConnectProviderServicesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListFastConnectProviderServicesResponse")
	}
	return
}

// listFastConnectProviderServices implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listFastConnectProviderServices(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/fastConnectProviderServices")
	if err != nil {
		return nil, err
	}

	var response ListFastConnectProviderServicesResponse
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

// ListFastConnectProviderVirtualCircuitBandwidthShapes Gets the list of available virtual circuit bandwidth levels for a provider.
// You need this information so you can specify your desired bandwidth level (shape) when you create a virtual circuit.
// For more information about virtual circuits, see FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListFastConnectProviderVirtualCircuitBandwidthShapes.go.html to see an example of how to use ListFastConnectProviderVirtualCircuitBandwidthShapes API.
func (client VirtualNetworkClient) ListFastConnectProviderVirtualCircuitBandwidthShapes(ctx context.Context, request ListFastConnectProviderVirtualCircuitBandwidthShapesRequest) (response ListFastConnectProviderVirtualCircuitBandwidthShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listFastConnectProviderVirtualCircuitBandwidthShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListFastConnectProviderVirtualCircuitBandwidthShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListFastConnectProviderVirtualCircuitBandwidthShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListFastConnectProviderVirtualCircuitBandwidthShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListFastConnectProviderVirtualCircuitBandwidthShapesResponse")
	}
	return
}

// listFastConnectProviderVirtualCircuitBandwidthShapes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listFastConnectProviderVirtualCircuitBandwidthShapes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/fastConnectProviderServices/{providerServiceId}/virtualCircuitBandwidthShapes")
	if err != nil {
		return nil, err
	}

	var response ListFastConnectProviderVirtualCircuitBandwidthShapesResponse
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

// ListIPSecConnectionTunnels Lists the tunnel information for the specified IPSec connection.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListIPSecConnectionTunnels.go.html to see an example of how to use ListIPSecConnectionTunnels API.
func (client VirtualNetworkClient) ListIPSecConnectionTunnels(ctx context.Context, request ListIPSecConnectionTunnelsRequest) (response ListIPSecConnectionTunnelsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listIPSecConnectionTunnels, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListIPSecConnectionTunnelsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListIPSecConnectionTunnelsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListIPSecConnectionTunnelsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListIPSecConnectionTunnelsResponse")
	}
	return
}

// listIPSecConnectionTunnels implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listIPSecConnectionTunnels(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections/{ipscId}/tunnels")
	if err != nil {
		return nil, err
	}

	var response ListIPSecConnectionTunnelsResponse
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

// ListIPSecConnections Lists the IPSec connections for the specified compartment. You can filter the
// results by DRG or CPE.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListIPSecConnections.go.html to see an example of how to use ListIPSecConnections API.
func (client VirtualNetworkClient) ListIPSecConnections(ctx context.Context, request ListIPSecConnectionsRequest) (response ListIPSecConnectionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listIPSecConnections, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListIPSecConnectionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListIPSecConnectionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListIPSecConnectionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListIPSecConnectionsResponse")
	}
	return
}

// listIPSecConnections implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listIPSecConnections(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipsecConnections")
	if err != nil {
		return nil, err
	}

	var response ListIPSecConnectionsResponse
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

// ListInternetGateways Lists the internet gateways in the specified VCN and the specified compartment.
// If the VCN ID is not provided, then the list includes the internet gateways from all VCNs in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListInternetGateways.go.html to see an example of how to use ListInternetGateways API.
func (client VirtualNetworkClient) ListInternetGateways(ctx context.Context, request ListInternetGatewaysRequest) (response ListInternetGatewaysResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listInternetGateways, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListInternetGatewaysResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListInternetGatewaysResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListInternetGatewaysResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListInternetGatewaysResponse")
	}
	return
}

// listInternetGateways implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listInternetGateways(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/internetGateways")
	if err != nil {
		return nil, err
	}

	var response ListInternetGatewaysResponse
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

// ListIpv6s Lists the Ipv6 objects based
// on one of these filters:
//   * Subnet OCID.
//   * VNIC OCID.
//   * Both IPv6 address and subnet OCID: This lets you get an `Ipv6` object based on its private
//   IPv6 address (for example, 2001:0db8:0123:1111:abcd:ef01:2345:6789) and not its OCID. For comparison,
//   GetIpv6 requires the OCID.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListIpv6s.go.html to see an example of how to use ListIpv6s API.
func (client VirtualNetworkClient) ListIpv6s(ctx context.Context, request ListIpv6sRequest) (response ListIpv6sResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listIpv6s, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListIpv6sResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListIpv6sResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListIpv6sResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListIpv6sResponse")
	}
	return
}

// listIpv6s implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listIpv6s(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/ipv6")
	if err != nil {
		return nil, err
	}

	var response ListIpv6sResponse
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

// ListLocalPeeringGateways Lists the local peering gateways (LPGs) for the specified VCN and specified compartment.
// If the VCN ID is not provided, then the list includes the LPGs from all VCNs in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListLocalPeeringGateways.go.html to see an example of how to use ListLocalPeeringGateways API.
func (client VirtualNetworkClient) ListLocalPeeringGateways(ctx context.Context, request ListLocalPeeringGatewaysRequest) (response ListLocalPeeringGatewaysResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listLocalPeeringGateways, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListLocalPeeringGatewaysResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListLocalPeeringGatewaysResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListLocalPeeringGatewaysResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListLocalPeeringGatewaysResponse")
	}
	return
}

// listLocalPeeringGateways implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listLocalPeeringGateways(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/localPeeringGateways")
	if err != nil {
		return nil, err
	}

	var response ListLocalPeeringGatewaysResponse
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

// ListNatGateways Lists the NAT gateways in the specified compartment. You may optionally specify a VCN OCID
// to filter the results by VCN.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListNatGateways.go.html to see an example of how to use ListNatGateways API.
func (client VirtualNetworkClient) ListNatGateways(ctx context.Context, request ListNatGatewaysRequest) (response ListNatGatewaysResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listNatGateways, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListNatGatewaysResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListNatGatewaysResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListNatGatewaysResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListNatGatewaysResponse")
	}
	return
}

// listNatGateways implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listNatGateways(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/natGateways")
	if err != nil {
		return nil, err
	}

	var response ListNatGatewaysResponse
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

// ListNetworkSecurityGroupSecurityRules Lists the security rules in the specified network security group.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListNetworkSecurityGroupSecurityRules.go.html to see an example of how to use ListNetworkSecurityGroupSecurityRules API.
func (client VirtualNetworkClient) ListNetworkSecurityGroupSecurityRules(ctx context.Context, request ListNetworkSecurityGroupSecurityRulesRequest) (response ListNetworkSecurityGroupSecurityRulesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listNetworkSecurityGroupSecurityRules, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListNetworkSecurityGroupSecurityRulesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListNetworkSecurityGroupSecurityRulesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListNetworkSecurityGroupSecurityRulesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListNetworkSecurityGroupSecurityRulesResponse")
	}
	return
}

// listNetworkSecurityGroupSecurityRules implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listNetworkSecurityGroupSecurityRules(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/networkSecurityGroups/{networkSecurityGroupId}/securityRules")
	if err != nil {
		return nil, err
	}

	var response ListNetworkSecurityGroupSecurityRulesResponse
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

// ListNetworkSecurityGroupVnics Lists the VNICs in the specified network security group.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListNetworkSecurityGroupVnics.go.html to see an example of how to use ListNetworkSecurityGroupVnics API.
func (client VirtualNetworkClient) ListNetworkSecurityGroupVnics(ctx context.Context, request ListNetworkSecurityGroupVnicsRequest) (response ListNetworkSecurityGroupVnicsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listNetworkSecurityGroupVnics, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListNetworkSecurityGroupVnicsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListNetworkSecurityGroupVnicsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListNetworkSecurityGroupVnicsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListNetworkSecurityGroupVnicsResponse")
	}
	return
}

// listNetworkSecurityGroupVnics implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listNetworkSecurityGroupVnics(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/networkSecurityGroups/{networkSecurityGroupId}/vnics")
	if err != nil {
		return nil, err
	}

	var response ListNetworkSecurityGroupVnicsResponse
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

// ListNetworkSecurityGroups Lists the network security groups in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListNetworkSecurityGroups.go.html to see an example of how to use ListNetworkSecurityGroups API.
func (client VirtualNetworkClient) ListNetworkSecurityGroups(ctx context.Context, request ListNetworkSecurityGroupsRequest) (response ListNetworkSecurityGroupsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listNetworkSecurityGroups, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListNetworkSecurityGroupsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListNetworkSecurityGroupsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListNetworkSecurityGroupsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListNetworkSecurityGroupsResponse")
	}
	return
}

// listNetworkSecurityGroups implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listNetworkSecurityGroups(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/networkSecurityGroups")
	if err != nil {
		return nil, err
	}

	var response ListNetworkSecurityGroupsResponse
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

// ListPrivateIps Lists the PrivateIp objects based
// on one of these filters:
//   - Subnet OCID.
//   - VNIC OCID.
//   - Both private IP address and subnet OCID: This lets
//   you get a `privateIP` object based on its private IP
//   address (for example, 10.0.3.3) and not its OCID. For comparison,
//   GetPrivateIp
//   requires the OCID.
// If you're listing all the private IPs associated with a given subnet
// or VNIC, the response includes both primary and secondary private IPs.
// If you are an Oracle Cloud VMware Solution customer and have VLANs
// in your VCN, you can filter the list by VLAN OCID. See Vlan.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListPrivateIps.go.html to see an example of how to use ListPrivateIps API.
func (client VirtualNetworkClient) ListPrivateIps(ctx context.Context, request ListPrivateIpsRequest) (response ListPrivateIpsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listPrivateIps, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListPrivateIpsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListPrivateIpsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListPrivateIpsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListPrivateIpsResponse")
	}
	return
}

// listPrivateIps implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listPrivateIps(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/privateIps")
	if err != nil {
		return nil, err
	}

	var response ListPrivateIpsResponse
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

// ListPublicIpPools Lists the public IP pools in the specified compartment.
// You can filter the list using query parameters.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListPublicIpPools.go.html to see an example of how to use ListPublicIpPools API.
func (client VirtualNetworkClient) ListPublicIpPools(ctx context.Context, request ListPublicIpPoolsRequest) (response ListPublicIpPoolsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listPublicIpPools, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListPublicIpPoolsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListPublicIpPoolsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListPublicIpPoolsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListPublicIpPoolsResponse")
	}
	return
}

// listPublicIpPools implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listPublicIpPools(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/publicIpPools")
	if err != nil {
		return nil, err
	}

	var response ListPublicIpPoolsResponse
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

// ListPublicIps Lists the PublicIp objects
// in the specified compartment. You can filter the list by using query parameters.
// To list your reserved public IPs:
//   * Set `scope` = `REGION`  (required)
//   * Leave the `availabilityDomain` parameter empty
//   * Set `lifetime` = `RESERVED`
// To list the ephemeral public IPs assigned to a regional entity such as a NAT gateway:
//   * Set `scope` = `REGION`  (required)
//   * Leave the `availabilityDomain` parameter empty
//   * Set `lifetime` = `EPHEMERAL`
// To list the ephemeral public IPs assigned to private IPs:
//   * Set `scope` = `AVAILABILITY_DOMAIN` (required)
//   * Set the `availabilityDomain` parameter to the desired availability domain (required)
//   * Set `lifetime` = `EPHEMERAL`
// **Note:** An ephemeral public IP assigned to a private IP
// is always in the same availability domain and compartment as the private IP.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListPublicIps.go.html to see an example of how to use ListPublicIps API.
func (client VirtualNetworkClient) ListPublicIps(ctx context.Context, request ListPublicIpsRequest) (response ListPublicIpsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listPublicIps, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListPublicIpsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListPublicIpsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListPublicIpsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListPublicIpsResponse")
	}
	return
}

// listPublicIps implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listPublicIps(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/publicIps")
	if err != nil {
		return nil, err
	}

	var response ListPublicIpsResponse
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

// ListRemotePeeringConnections Lists the remote peering connections (RPCs) for the specified DRG and compartment
// (the RPC's compartment).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListRemotePeeringConnections.go.html to see an example of how to use ListRemotePeeringConnections API.
func (client VirtualNetworkClient) ListRemotePeeringConnections(ctx context.Context, request ListRemotePeeringConnectionsRequest) (response ListRemotePeeringConnectionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listRemotePeeringConnections, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListRemotePeeringConnectionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListRemotePeeringConnectionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListRemotePeeringConnectionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListRemotePeeringConnectionsResponse")
	}
	return
}

// listRemotePeeringConnections implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listRemotePeeringConnections(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/remotePeeringConnections")
	if err != nil {
		return nil, err
	}

	var response ListRemotePeeringConnectionsResponse
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

// ListRouteTables Lists the route tables in the specified VCN and specified compartment.
// If the VCN ID is not provided, then the list includes the route tables from all VCNs in the specified compartment.
// The response includes the default route table that automatically comes with
// each VCN in the specified compartment, plus any route tables you've created.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListRouteTables.go.html to see an example of how to use ListRouteTables API.
func (client VirtualNetworkClient) ListRouteTables(ctx context.Context, request ListRouteTablesRequest) (response ListRouteTablesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listRouteTables, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListRouteTablesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListRouteTablesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListRouteTablesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListRouteTablesResponse")
	}
	return
}

// listRouteTables implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listRouteTables(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/routeTables")
	if err != nil {
		return nil, err
	}

	var response ListRouteTablesResponse
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

// ListSecurityLists Lists the security lists in the specified VCN and compartment.
// If the VCN ID is not provided, then the list includes the security lists from all VCNs in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListSecurityLists.go.html to see an example of how to use ListSecurityLists API.
func (client VirtualNetworkClient) ListSecurityLists(ctx context.Context, request ListSecurityListsRequest) (response ListSecurityListsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listSecurityLists, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListSecurityListsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListSecurityListsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListSecurityListsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListSecurityListsResponse")
	}
	return
}

// listSecurityLists implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listSecurityLists(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/securityLists")
	if err != nil {
		return nil, err
	}

	var response ListSecurityListsResponse
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

// ListServiceGateways Lists the service gateways in the specified compartment. You may optionally specify a VCN OCID
// to filter the results by VCN.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListServiceGateways.go.html to see an example of how to use ListServiceGateways API.
func (client VirtualNetworkClient) ListServiceGateways(ctx context.Context, request ListServiceGatewaysRequest) (response ListServiceGatewaysResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listServiceGateways, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListServiceGatewaysResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListServiceGatewaysResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListServiceGatewaysResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListServiceGatewaysResponse")
	}
	return
}

// listServiceGateways implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listServiceGateways(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/serviceGateways")
	if err != nil {
		return nil, err
	}

	var response ListServiceGatewaysResponse
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

// ListServices Lists the available Service objects that you can enable for a
// service gateway in this region.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListServices.go.html to see an example of how to use ListServices API.
func (client VirtualNetworkClient) ListServices(ctx context.Context, request ListServicesRequest) (response ListServicesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listServices, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListServicesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListServicesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListServicesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListServicesResponse")
	}
	return
}

// listServices implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listServices(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/services")
	if err != nil {
		return nil, err
	}

	var response ListServicesResponse
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

// ListSubnets Lists the subnets in the specified VCN and the specified compartment.
// If the VCN ID is not provided, then the list includes the subnets from all VCNs in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListSubnets.go.html to see an example of how to use ListSubnets API.
func (client VirtualNetworkClient) ListSubnets(ctx context.Context, request ListSubnetsRequest) (response ListSubnetsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listSubnets, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListSubnetsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListSubnetsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListSubnetsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListSubnetsResponse")
	}
	return
}

// listSubnets implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listSubnets(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/subnets")
	if err != nil {
		return nil, err
	}

	var response ListSubnetsResponse
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

// ListVcns Lists the virtual cloud networks (VCNs) in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVcns.go.html to see an example of how to use ListVcns API.
func (client VirtualNetworkClient) ListVcns(ctx context.Context, request ListVcnsRequest) (response ListVcnsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVcns, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVcnsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVcnsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVcnsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVcnsResponse")
	}
	return
}

// listVcns implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listVcns(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vcns")
	if err != nil {
		return nil, err
	}

	var response ListVcnsResponse
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

// ListVirtualCircuitBandwidthShapes The deprecated operation lists available bandwidth levels for virtual circuits. For the compartment ID, provide the OCID of your tenancy (the root compartment).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVirtualCircuitBandwidthShapes.go.html to see an example of how to use ListVirtualCircuitBandwidthShapes API.
func (client VirtualNetworkClient) ListVirtualCircuitBandwidthShapes(ctx context.Context, request ListVirtualCircuitBandwidthShapesRequest) (response ListVirtualCircuitBandwidthShapesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVirtualCircuitBandwidthShapes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVirtualCircuitBandwidthShapesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVirtualCircuitBandwidthShapesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVirtualCircuitBandwidthShapesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVirtualCircuitBandwidthShapesResponse")
	}
	return
}

// listVirtualCircuitBandwidthShapes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listVirtualCircuitBandwidthShapes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualCircuitBandwidthShapes")
	if err != nil {
		return nil, err
	}

	var response ListVirtualCircuitBandwidthShapesResponse
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

// ListVirtualCircuitPublicPrefixes Lists the public IP prefixes and their details for the specified
// public virtual circuit.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVirtualCircuitPublicPrefixes.go.html to see an example of how to use ListVirtualCircuitPublicPrefixes API.
func (client VirtualNetworkClient) ListVirtualCircuitPublicPrefixes(ctx context.Context, request ListVirtualCircuitPublicPrefixesRequest) (response ListVirtualCircuitPublicPrefixesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVirtualCircuitPublicPrefixes, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVirtualCircuitPublicPrefixesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVirtualCircuitPublicPrefixesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVirtualCircuitPublicPrefixesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVirtualCircuitPublicPrefixesResponse")
	}
	return
}

// listVirtualCircuitPublicPrefixes implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listVirtualCircuitPublicPrefixes(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualCircuits/{virtualCircuitId}/publicPrefixes")
	if err != nil {
		return nil, err
	}

	var response ListVirtualCircuitPublicPrefixesResponse
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

// ListVirtualCircuits Lists the virtual circuits in the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVirtualCircuits.go.html to see an example of how to use ListVirtualCircuits API.
func (client VirtualNetworkClient) ListVirtualCircuits(ctx context.Context, request ListVirtualCircuitsRequest) (response ListVirtualCircuitsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVirtualCircuits, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVirtualCircuitsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVirtualCircuitsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVirtualCircuitsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVirtualCircuitsResponse")
	}
	return
}

// listVirtualCircuits implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listVirtualCircuits(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/virtualCircuits")
	if err != nil {
		return nil, err
	}

	var response ListVirtualCircuitsResponse
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

// ListVlans Lists the VLANs in the specified VCN and the specified compartment.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ListVlans.go.html to see an example of how to use ListVlans API.
func (client VirtualNetworkClient) ListVlans(ctx context.Context, request ListVlansRequest) (response ListVlansResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.listVlans, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ListVlansResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ListVlansResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ListVlansResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ListVlansResponse")
	}
	return
}

// listVlans implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) listVlans(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodGet, "/vlans")
	if err != nil {
		return nil, err
	}

	var response ListVlansResponse
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

// ModifyVcnCidr Updates the specified CIDR block of a VCN. The new CIDR IP range must meet the following criteria:
// - Must be valid.
// - Must not overlap with another CIDR block in the VCN, a CIDR block of a peered VCN, or the on-premises network CIDR block.
// - Must not exceed the limit of CIDR blocks allowed per VCN.
// - Must include IP addresses from the original CIDR block that are used in the VCN's existing route rules.
// - No IP address in an existing subnet should be outside of the new CIDR block range.
// **Note:** Modifying a CIDR block places your VCN in an updating state until the changes are complete. You cannot create or update the VCN's subnets, VLANs, LPGs, or route tables during this operation. The time to completion can vary depending on the size of your network. Updating a small network could take about a minute, and updating a large network could take up to an hour. You can use the `GetWorkRequest` operation to check the status of the update.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ModifyVcnCidr.go.html to see an example of how to use ModifyVcnCidr API.
func (client VirtualNetworkClient) ModifyVcnCidr(ctx context.Context, request ModifyVcnCidrRequest) (response ModifyVcnCidrResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.modifyVcnCidr, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ModifyVcnCidrResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ModifyVcnCidrResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ModifyVcnCidrResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ModifyVcnCidrResponse")
	}
	return
}

// modifyVcnCidr implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) modifyVcnCidr(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vcns/{vcnId}/actions/modifyCidr")
	if err != nil {
		return nil, err
	}

	var response ModifyVcnCidrResponse
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

// RemoveNetworkSecurityGroupSecurityRules Removes one or more security rules from the specified network security group.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/RemoveNetworkSecurityGroupSecurityRules.go.html to see an example of how to use RemoveNetworkSecurityGroupSecurityRules API.
func (client VirtualNetworkClient) RemoveNetworkSecurityGroupSecurityRules(ctx context.Context, request RemoveNetworkSecurityGroupSecurityRulesRequest) (response RemoveNetworkSecurityGroupSecurityRulesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.removeNetworkSecurityGroupSecurityRules, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = RemoveNetworkSecurityGroupSecurityRulesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = RemoveNetworkSecurityGroupSecurityRulesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(RemoveNetworkSecurityGroupSecurityRulesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into RemoveNetworkSecurityGroupSecurityRulesResponse")
	}
	return
}

// removeNetworkSecurityGroupSecurityRules implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) removeNetworkSecurityGroupSecurityRules(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/networkSecurityGroups/{networkSecurityGroupId}/actions/removeSecurityRules")
	if err != nil {
		return nil, err
	}

	var response RemoveNetworkSecurityGroupSecurityRulesResponse
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

// RemovePublicIpPoolCapacity Removes a CIDR block from the referenced public IP pool.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/RemovePublicIpPoolCapacity.go.html to see an example of how to use RemovePublicIpPoolCapacity API.
func (client VirtualNetworkClient) RemovePublicIpPoolCapacity(ctx context.Context, request RemovePublicIpPoolCapacityRequest) (response RemovePublicIpPoolCapacityResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.removePublicIpPoolCapacity, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = RemovePublicIpPoolCapacityResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = RemovePublicIpPoolCapacityResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(RemovePublicIpPoolCapacityResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into RemovePublicIpPoolCapacityResponse")
	}
	return
}

// removePublicIpPoolCapacity implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) removePublicIpPoolCapacity(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/publicIpPools/{publicIpPoolId}/actions/removeCapacity")
	if err != nil {
		return nil, err
	}

	var response RemovePublicIpPoolCapacityResponse
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

// RemoveVcnCidr Removes a specified CIDR block from a VCN.
// **Notes:**
// - You cannot remove a CIDR block if an IP address in its range is in use.
// - Removing a CIDR block places your VCN in an updating state until the changes are complete. You cannot create or update the VCN's subnets, VLANs, LPGs, or route tables during this operation. The time to completion can take a few minutes. You can use the `GetWorkRequest` operation to check the status of the update.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/RemoveVcnCidr.go.html to see an example of how to use RemoveVcnCidr API.
func (client VirtualNetworkClient) RemoveVcnCidr(ctx context.Context, request RemoveVcnCidrRequest) (response RemoveVcnCidrResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.removeVcnCidr, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = RemoveVcnCidrResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = RemoveVcnCidrResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(RemoveVcnCidrResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into RemoveVcnCidrResponse")
	}
	return
}

// removeVcnCidr implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) removeVcnCidr(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/vcns/{vcnId}/actions/removeCidr")
	if err != nil {
		return nil, err
	}

	var response RemoveVcnCidrResponse
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

// UpdateByoipRange Updates the tags or display name associated to the specified BYOIP CIDR block.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateByoipRange.go.html to see an example of how to use UpdateByoipRange API.
func (client VirtualNetworkClient) UpdateByoipRange(ctx context.Context, request UpdateByoipRangeRequest) (response UpdateByoipRangeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateByoipRange, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateByoipRangeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateByoipRangeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateByoipRangeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateByoipRangeResponse")
	}
	return
}

// updateByoipRange implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateByoipRange(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/byoipRanges/{byoipRangeId}")
	if err != nil {
		return nil, err
	}

	var response UpdateByoipRangeResponse
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

// UpdateCpe Updates the specified CPE's display name or tags.
// Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateCpe.go.html to see an example of how to use UpdateCpe API.
func (client VirtualNetworkClient) UpdateCpe(ctx context.Context, request UpdateCpeRequest) (response UpdateCpeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateCpe, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateCpeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateCpeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateCpeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateCpeResponse")
	}
	return
}

// updateCpe implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateCpe(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/cpes/{cpeId}")
	if err != nil {
		return nil, err
	}

	var response UpdateCpeResponse
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

// UpdateCrossConnect Updates the specified cross-connect.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateCrossConnect.go.html to see an example of how to use UpdateCrossConnect API.
func (client VirtualNetworkClient) UpdateCrossConnect(ctx context.Context, request UpdateCrossConnectRequest) (response UpdateCrossConnectResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateCrossConnect, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateCrossConnectResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateCrossConnectResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateCrossConnectResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateCrossConnectResponse")
	}
	return
}

// updateCrossConnect implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateCrossConnect(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/crossConnects/{crossConnectId}")
	if err != nil {
		return nil, err
	}

	var response UpdateCrossConnectResponse
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

// UpdateCrossConnectGroup Updates the specified cross-connect group's display name.
// Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateCrossConnectGroup.go.html to see an example of how to use UpdateCrossConnectGroup API.
func (client VirtualNetworkClient) UpdateCrossConnectGroup(ctx context.Context, request UpdateCrossConnectGroupRequest) (response UpdateCrossConnectGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateCrossConnectGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateCrossConnectGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateCrossConnectGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateCrossConnectGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateCrossConnectGroupResponse")
	}
	return
}

// updateCrossConnectGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateCrossConnectGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/crossConnectGroups/{crossConnectGroupId}")
	if err != nil {
		return nil, err
	}

	var response UpdateCrossConnectGroupResponse
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

// UpdateDhcpOptions Updates the specified set of DHCP options. You can update the display name or the options
// themselves. Avoid entering confidential information.
// Note that the `options` object you provide replaces the entire existing set of options.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateDhcpOptions.go.html to see an example of how to use UpdateDhcpOptions API.
func (client VirtualNetworkClient) UpdateDhcpOptions(ctx context.Context, request UpdateDhcpOptionsRequest) (response UpdateDhcpOptionsResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateDhcpOptions, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateDhcpOptionsResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateDhcpOptionsResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateDhcpOptionsResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateDhcpOptionsResponse")
	}
	return
}

// updateDhcpOptions implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateDhcpOptions(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/dhcps/{dhcpId}")
	if err != nil {
		return nil, err
	}

	var response UpdateDhcpOptionsResponse
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

// UpdateDrg Updates the specified DRG's display name or tags. Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateDrg.go.html to see an example of how to use UpdateDrg API.
func (client VirtualNetworkClient) UpdateDrg(ctx context.Context, request UpdateDrgRequest) (response UpdateDrgResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateDrg, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateDrgResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateDrgResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateDrgResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateDrgResponse")
	}
	return
}

// updateDrg implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateDrg(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/drgs/{drgId}")
	if err != nil {
		return nil, err
	}

	var response UpdateDrgResponse
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

// UpdateDrgAttachment Updates the display name for the specified `DrgAttachment`.
// Avoid entering confidential information.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateDrgAttachment.go.html to see an example of how to use UpdateDrgAttachment API.
func (client VirtualNetworkClient) UpdateDrgAttachment(ctx context.Context, request UpdateDrgAttachmentRequest) (response UpdateDrgAttachmentResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateDrgAttachment, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateDrgAttachmentResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateDrgAttachmentResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateDrgAttachmentResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateDrgAttachmentResponse")
	}
	return
}

// updateDrgAttachment implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateDrgAttachment(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/drgAttachments/{drgAttachmentId}")
	if err != nil {
		return nil, err
	}

	var response UpdateDrgAttachmentResponse
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

// UpdateIPSecConnection Updates the specified IPSec connection.
// To update an individual IPSec tunnel's attributes, use
// UpdateIPSecConnectionTunnel.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateIPSecConnection.go.html to see an example of how to use UpdateIPSecConnection API.
func (client VirtualNetworkClient) UpdateIPSecConnection(ctx context.Context, request UpdateIPSecConnectionRequest) (response UpdateIPSecConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateIPSecConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateIPSecConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateIPSecConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateIPSecConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateIPSecConnectionResponse")
	}
	return
}

// updateIPSecConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateIPSecConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/ipsecConnections/{ipscId}")
	if err != nil {
		return nil, err
	}

	var response UpdateIPSecConnectionResponse
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

// UpdateIPSecConnectionTunnel Updates the specified tunnel. This operation lets you change tunnel attributes such as the
// routing type (BGP dynamic routing or static routing). Here are some important notes:
//   * If you change the tunnel's routing type or BGP session configuration, the tunnel will go
//     down while it's reprovisioned.
//   * If you want to switch the tunnel's `routing` from `STATIC` to `BGP`, make sure the tunnel's
//     BGP session configuration attributes have been set (BgpSessionInfo).
//   * If you want to switch the tunnel's `routing` from `BGP` to `STATIC`, make sure the
//     IPSecConnection already has at least one valid CIDR
//     static route.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateIPSecConnectionTunnel.go.html to see an example of how to use UpdateIPSecConnectionTunnel API.
func (client VirtualNetworkClient) UpdateIPSecConnectionTunnel(ctx context.Context, request UpdateIPSecConnectionTunnelRequest) (response UpdateIPSecConnectionTunnelResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateIPSecConnectionTunnel, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateIPSecConnectionTunnelResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateIPSecConnectionTunnelResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateIPSecConnectionTunnelResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateIPSecConnectionTunnelResponse")
	}
	return
}

// updateIPSecConnectionTunnel implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateIPSecConnectionTunnel(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/ipsecConnections/{ipscId}/tunnels/{tunnelId}")
	if err != nil {
		return nil, err
	}

	var response UpdateIPSecConnectionTunnelResponse
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

// UpdateIPSecConnectionTunnelSharedSecret Updates the shared secret (pre-shared key) for the specified tunnel.
// **Important:** If you change the shared secret, the tunnel will go down while it's reprovisioned.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateIPSecConnectionTunnelSharedSecret.go.html to see an example of how to use UpdateIPSecConnectionTunnelSharedSecret API.
func (client VirtualNetworkClient) UpdateIPSecConnectionTunnelSharedSecret(ctx context.Context, request UpdateIPSecConnectionTunnelSharedSecretRequest) (response UpdateIPSecConnectionTunnelSharedSecretResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateIPSecConnectionTunnelSharedSecret, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateIPSecConnectionTunnelSharedSecretResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateIPSecConnectionTunnelSharedSecretResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateIPSecConnectionTunnelSharedSecretResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateIPSecConnectionTunnelSharedSecretResponse")
	}
	return
}

// updateIPSecConnectionTunnelSharedSecret implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateIPSecConnectionTunnelSharedSecret(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/ipsecConnections/{ipscId}/tunnels/{tunnelId}/sharedSecret")
	if err != nil {
		return nil, err
	}

	var response UpdateIPSecConnectionTunnelSharedSecretResponse
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

// UpdateInternetGateway Updates the specified internet gateway. You can disable/enable it, or change its display name
// or tags. Avoid entering confidential information.
// If the gateway is disabled, that means no traffic will flow to/from the internet even if there's
// a route rule that enables that traffic.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateInternetGateway.go.html to see an example of how to use UpdateInternetGateway API.
func (client VirtualNetworkClient) UpdateInternetGateway(ctx context.Context, request UpdateInternetGatewayRequest) (response UpdateInternetGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateInternetGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateInternetGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateInternetGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateInternetGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateInternetGatewayResponse")
	}
	return
}

// updateInternetGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateInternetGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/internetGateways/{igId}")
	if err != nil {
		return nil, err
	}

	var response UpdateInternetGatewayResponse
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

// UpdateIpv6 Updates the specified IPv6. You must specify the object's OCID.
// Use this operation if you want to:
//   * Move an IPv6 to a different VNIC in the same subnet.
//   * Enable/disable internet access for an IPv6.
//   * Change the display name for an IPv6.
//   * Update resource tags for an IPv6.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateIpv6.go.html to see an example of how to use UpdateIpv6 API.
func (client VirtualNetworkClient) UpdateIpv6(ctx context.Context, request UpdateIpv6Request) (response UpdateIpv6Response, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateIpv6, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateIpv6Response{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateIpv6Response{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateIpv6Response); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateIpv6Response")
	}
	return
}

// updateIpv6 implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateIpv6(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/ipv6/{ipv6Id}")
	if err != nil {
		return nil, err
	}

	var response UpdateIpv6Response
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

// UpdateLocalPeeringGateway Updates the specified local peering gateway (LPG).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateLocalPeeringGateway.go.html to see an example of how to use UpdateLocalPeeringGateway API.
func (client VirtualNetworkClient) UpdateLocalPeeringGateway(ctx context.Context, request UpdateLocalPeeringGatewayRequest) (response UpdateLocalPeeringGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateLocalPeeringGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateLocalPeeringGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateLocalPeeringGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateLocalPeeringGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateLocalPeeringGatewayResponse")
	}
	return
}

// updateLocalPeeringGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateLocalPeeringGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/localPeeringGateways/{localPeeringGatewayId}")
	if err != nil {
		return nil, err
	}

	var response UpdateLocalPeeringGatewayResponse
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

// UpdateNatGateway Updates the specified NAT gateway.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateNatGateway.go.html to see an example of how to use UpdateNatGateway API.
func (client VirtualNetworkClient) UpdateNatGateway(ctx context.Context, request UpdateNatGatewayRequest) (response UpdateNatGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateNatGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateNatGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateNatGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateNatGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateNatGatewayResponse")
	}
	return
}

// updateNatGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateNatGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/natGateways/{natGatewayId}")
	if err != nil {
		return nil, err
	}

	var response UpdateNatGatewayResponse
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

// UpdateNetworkSecurityGroup Updates the specified network security group.
// To add or remove an existing VNIC from the group, use
// UpdateVnic.
// To add a VNIC to the group *when you create the VNIC*, specify the NSG's OCID during creation.
// For example, see the `nsgIds` attribute in CreateVnicDetails.
// To add or remove security rules from the group, use
// AddNetworkSecurityGroupSecurityRules
// or
// RemoveNetworkSecurityGroupSecurityRules.
// To edit the contents of existing security rules in the group, use
// UpdateNetworkSecurityGroupSecurityRules.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateNetworkSecurityGroup.go.html to see an example of how to use UpdateNetworkSecurityGroup API.
func (client VirtualNetworkClient) UpdateNetworkSecurityGroup(ctx context.Context, request UpdateNetworkSecurityGroupRequest) (response UpdateNetworkSecurityGroupResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateNetworkSecurityGroup, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateNetworkSecurityGroupResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateNetworkSecurityGroupResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateNetworkSecurityGroupResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateNetworkSecurityGroupResponse")
	}
	return
}

// updateNetworkSecurityGroup implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateNetworkSecurityGroup(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/networkSecurityGroups/{networkSecurityGroupId}")
	if err != nil {
		return nil, err
	}

	var response UpdateNetworkSecurityGroupResponse
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

// UpdateNetworkSecurityGroupSecurityRules Updates one or more security rules in the specified network security group.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateNetworkSecurityGroupSecurityRules.go.html to see an example of how to use UpdateNetworkSecurityGroupSecurityRules API.
func (client VirtualNetworkClient) UpdateNetworkSecurityGroupSecurityRules(ctx context.Context, request UpdateNetworkSecurityGroupSecurityRulesRequest) (response UpdateNetworkSecurityGroupSecurityRulesResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateNetworkSecurityGroupSecurityRules, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateNetworkSecurityGroupSecurityRulesResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateNetworkSecurityGroupSecurityRulesResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateNetworkSecurityGroupSecurityRulesResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateNetworkSecurityGroupSecurityRulesResponse")
	}
	return
}

// updateNetworkSecurityGroupSecurityRules implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateNetworkSecurityGroupSecurityRules(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/networkSecurityGroups/{networkSecurityGroupId}/actions/updateSecurityRules")
	if err != nil {
		return nil, err
	}

	var response UpdateNetworkSecurityGroupSecurityRulesResponse
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

// UpdatePrivateIp Updates the specified private IP. You must specify the object's OCID.
// Use this operation if you want to:
//   - Move a secondary private IP to a different VNIC in the same subnet.
//   - Change the display name for a secondary private IP.
//   - Change the hostname for a secondary private IP.
// This operation cannot be used with primary private IPs.
// To update the hostname for the primary IP on a VNIC, use
// UpdateVnic.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdatePrivateIp.go.html to see an example of how to use UpdatePrivateIp API.
func (client VirtualNetworkClient) UpdatePrivateIp(ctx context.Context, request UpdatePrivateIpRequest) (response UpdatePrivateIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updatePrivateIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdatePrivateIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdatePrivateIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdatePrivateIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdatePrivateIpResponse")
	}
	return
}

// updatePrivateIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updatePrivateIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/privateIps/{privateIpId}")
	if err != nil {
		return nil, err
	}

	var response UpdatePrivateIpResponse
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

// UpdatePublicIp Updates the specified public IP. You must specify the object's OCID. Use this operation if you want to:
// * Assign a reserved public IP in your pool to a private IP.
// * Move a reserved public IP to a different private IP.
// * Unassign a reserved public IP from a private IP (which returns it to your pool
// of reserved public IPs).
// * Change the display name or tags for a public IP.
// Assigning, moving, and unassigning a reserved public IP are asynchronous
// operations. Poll the public IP's `lifecycleState` to determine if the operation
// succeeded.
// **Note:** When moving a reserved public IP, the target private IP
// must not already have a public IP with `lifecycleState` = ASSIGNING or ASSIGNED. If it
// does, an error is returned. Also, the initial unassignment from the original
// private IP always succeeds, but the assignment to the target private IP is asynchronous and
// could fail silently (for example, if the target private IP is deleted or has a different public IP
// assigned to it in the interim). If that occurs, the public IP remains unassigned and its
// `lifecycleState` switches to AVAILABLE (it is not reassigned to its original private IP).
// You must poll the public IP's `lifecycleState` to determine if the move succeeded.
// Regarding ephemeral public IPs:
// * If you want to assign an ephemeral public IP to a primary private IP, use
// CreatePublicIp.
// * You can't move an ephemeral public IP to a different private IP.
// * If you want to unassign an ephemeral public IP from its private IP, use
// DeletePublicIp, which
// unassigns and deletes the ephemeral public IP.
// **Note:** If a public IP is assigned to a secondary private
// IP (see PrivateIp), and you move that secondary
// private IP to another VNIC, the public IP moves with it.
// **Note:** There's a limit to the number of PublicIp
// a VNIC or instance can have. If you try to move a reserved public IP
// to a VNIC or instance that has already reached its public IP limit, an error is
// returned. For information about the public IP limits, see
// Public IP Addresses (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingpublicIPs.htm).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdatePublicIp.go.html to see an example of how to use UpdatePublicIp API.
func (client VirtualNetworkClient) UpdatePublicIp(ctx context.Context, request UpdatePublicIpRequest) (response UpdatePublicIpResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updatePublicIp, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdatePublicIpResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdatePublicIpResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdatePublicIpResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdatePublicIpResponse")
	}
	return
}

// updatePublicIp implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updatePublicIp(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/publicIps/{publicIpId}")
	if err != nil {
		return nil, err
	}

	var response UpdatePublicIpResponse
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

// UpdatePublicIpPool Updates the specified public IP pool.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdatePublicIpPool.go.html to see an example of how to use UpdatePublicIpPool API.
func (client VirtualNetworkClient) UpdatePublicIpPool(ctx context.Context, request UpdatePublicIpPoolRequest) (response UpdatePublicIpPoolResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updatePublicIpPool, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdatePublicIpPoolResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdatePublicIpPoolResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdatePublicIpPoolResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdatePublicIpPoolResponse")
	}
	return
}

// updatePublicIpPool implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updatePublicIpPool(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/publicIpPools/{publicIpPoolId}")
	if err != nil {
		return nil, err
	}

	var response UpdatePublicIpPoolResponse
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

// UpdateRemotePeeringConnection Updates the specified remote peering connection (RPC).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateRemotePeeringConnection.go.html to see an example of how to use UpdateRemotePeeringConnection API.
func (client VirtualNetworkClient) UpdateRemotePeeringConnection(ctx context.Context, request UpdateRemotePeeringConnectionRequest) (response UpdateRemotePeeringConnectionResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateRemotePeeringConnection, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateRemotePeeringConnectionResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateRemotePeeringConnectionResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateRemotePeeringConnectionResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateRemotePeeringConnectionResponse")
	}
	return
}

// updateRemotePeeringConnection implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateRemotePeeringConnection(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/remotePeeringConnections/{remotePeeringConnectionId}")
	if err != nil {
		return nil, err
	}

	var response UpdateRemotePeeringConnectionResponse
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

// UpdateRouteTable Updates the specified route table's display name or route rules.
// Avoid entering confidential information.
// Note that the `routeRules` object you provide replaces the entire existing set of rules.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateRouteTable.go.html to see an example of how to use UpdateRouteTable API.
func (client VirtualNetworkClient) UpdateRouteTable(ctx context.Context, request UpdateRouteTableRequest) (response UpdateRouteTableResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateRouteTable, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateRouteTableResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateRouteTableResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateRouteTableResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateRouteTableResponse")
	}
	return
}

// updateRouteTable implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateRouteTable(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/routeTables/{rtId}")
	if err != nil {
		return nil, err
	}

	var response UpdateRouteTableResponse
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

// UpdateSecurityList Updates the specified security list's display name or rules.
// Avoid entering confidential information.
// Note that the `egressSecurityRules` or `ingressSecurityRules` objects you provide replace the entire
// existing objects.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateSecurityList.go.html to see an example of how to use UpdateSecurityList API.
func (client VirtualNetworkClient) UpdateSecurityList(ctx context.Context, request UpdateSecurityListRequest) (response UpdateSecurityListResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateSecurityList, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateSecurityListResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateSecurityListResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateSecurityListResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateSecurityListResponse")
	}
	return
}

// updateSecurityList implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateSecurityList(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/securityLists/{securityListId}")
	if err != nil {
		return nil, err
	}

	var response UpdateSecurityListResponse
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

// UpdateServiceGateway Updates the specified service gateway. The information you provide overwrites the existing
// attributes of the gateway.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateServiceGateway.go.html to see an example of how to use UpdateServiceGateway API.
func (client VirtualNetworkClient) UpdateServiceGateway(ctx context.Context, request UpdateServiceGatewayRequest) (response UpdateServiceGatewayResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateServiceGateway, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateServiceGatewayResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateServiceGatewayResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateServiceGatewayResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateServiceGatewayResponse")
	}
	return
}

// updateServiceGateway implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateServiceGateway(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/serviceGateways/{serviceGatewayId}")
	if err != nil {
		return nil, err
	}

	var response UpdateServiceGatewayResponse
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

// UpdateSubnet Updates the specified subnet.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateSubnet.go.html to see an example of how to use UpdateSubnet API.
func (client VirtualNetworkClient) UpdateSubnet(ctx context.Context, request UpdateSubnetRequest) (response UpdateSubnetResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateSubnet, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateSubnetResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateSubnetResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateSubnetResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateSubnetResponse")
	}
	return
}

// updateSubnet implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateSubnet(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/subnets/{subnetId}")
	if err != nil {
		return nil, err
	}

	var response UpdateSubnetResponse
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

// UpdateTunnelCpeDeviceConfig Creates or updates the set of CPE configuration answers for the specified tunnel.
// The answers correlate to the questions that are specific to the CPE device type (see the
// `parameters` attribute of CpeDeviceShapeDetail).
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateTunnelCpeDeviceConfig.go.html to see an example of how to use UpdateTunnelCpeDeviceConfig API.
func (client VirtualNetworkClient) UpdateTunnelCpeDeviceConfig(ctx context.Context, request UpdateTunnelCpeDeviceConfigRequest) (response UpdateTunnelCpeDeviceConfigResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}

	if !(request.OpcRetryToken != nil && *request.OpcRetryToken != "") {
		request.OpcRetryToken = common.String(common.RetryToken())
	}

	ociResponse, err = common.Retry(ctx, request, client.updateTunnelCpeDeviceConfig, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateTunnelCpeDeviceConfigResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateTunnelCpeDeviceConfigResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateTunnelCpeDeviceConfigResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateTunnelCpeDeviceConfigResponse")
	}
	return
}

// updateTunnelCpeDeviceConfig implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateTunnelCpeDeviceConfig(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/ipsecConnections/{ipscId}/tunnels/{tunnelId}/tunnelDeviceConfig")
	if err != nil {
		return nil, err
	}

	var response UpdateTunnelCpeDeviceConfigResponse
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

// UpdateVcn Updates the specified VCN.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateVcn.go.html to see an example of how to use UpdateVcn API.
func (client VirtualNetworkClient) UpdateVcn(ctx context.Context, request UpdateVcnRequest) (response UpdateVcnResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateVcn, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateVcnResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateVcnResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateVcnResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateVcnResponse")
	}
	return
}

// updateVcn implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateVcn(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/vcns/{vcnId}")
	if err != nil {
		return nil, err
	}

	var response UpdateVcnResponse
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

// UpdateVirtualCircuit Updates the specified virtual circuit. This can be called by
// either the customer who owns the virtual circuit, or the
// provider (when provisioning or de-provisioning the virtual
// circuit from their end). The documentation for
// UpdateVirtualCircuitDetails
// indicates who can update each property of the virtual circuit.
// **Important:** If the virtual circuit is working and in the
// PROVISIONED state, updating any of the network-related properties
// (such as the DRG being used, the BGP ASN, and so on) will cause the virtual
// circuit's state to switch to PROVISIONING and the related BGP
// session to go down. After Oracle re-provisions the virtual circuit,
// its state will return to PROVISIONED. Make sure you confirm that
// the associated BGP session is back up. For more information
// about the various states and how to test connectivity, see
// FastConnect Overview (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/fastconnect.htm).
// To change the list of public IP prefixes for a public virtual circuit,
// use BulkAddVirtualCircuitPublicPrefixes
// and
// BulkDeleteVirtualCircuitPublicPrefixes.
// Updating the list of prefixes does NOT cause the BGP session to go down. However,
// Oracle must verify the customer's ownership of each added prefix before
// traffic for that prefix will flow across the virtual circuit.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateVirtualCircuit.go.html to see an example of how to use UpdateVirtualCircuit API.
func (client VirtualNetworkClient) UpdateVirtualCircuit(ctx context.Context, request UpdateVirtualCircuitRequest) (response UpdateVirtualCircuitResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateVirtualCircuit, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateVirtualCircuitResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateVirtualCircuitResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateVirtualCircuitResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateVirtualCircuitResponse")
	}
	return
}

// updateVirtualCircuit implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateVirtualCircuit(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/virtualCircuits/{virtualCircuitId}")
	if err != nil {
		return nil, err
	}

	var response UpdateVirtualCircuitResponse
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

// UpdateVlan Updates the specified VLAN. Note that this operation might require changes to all
// the VNICs in the VLAN, which can take a while. The VLAN will be in the UPDATING state until the changes are complete.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateVlan.go.html to see an example of how to use UpdateVlan API.
func (client VirtualNetworkClient) UpdateVlan(ctx context.Context, request UpdateVlanRequest) (response UpdateVlanResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateVlan, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateVlanResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateVlanResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateVlanResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateVlanResponse")
	}
	return
}

// updateVlan implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateVlan(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/vlans/{vlanId}")
	if err != nil {
		return nil, err
	}

	var response UpdateVlanResponse
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

// UpdateVnic Updates the specified VNIC.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/UpdateVnic.go.html to see an example of how to use UpdateVnic API.
func (client VirtualNetworkClient) UpdateVnic(ctx context.Context, request UpdateVnicRequest) (response UpdateVnicResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.updateVnic, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = UpdateVnicResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = UpdateVnicResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(UpdateVnicResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into UpdateVnicResponse")
	}
	return
}

// updateVnic implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) updateVnic(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPut, "/vnics/{vnicId}")
	if err != nil {
		return nil, err
	}

	var response UpdateVnicResponse
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

// ValidateByoipRange Submits the BYOIP CIDR block you are importing for validation. Do not submit to Oracle for validation if you have not already
// modified the information for the BYOIP CIDR block with your Regional Internet Registry. See To import a CIDR block (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/BYOIP.htm#import_cidr) for details.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/ValidateByoipRange.go.html to see an example of how to use ValidateByoipRange API.
func (client VirtualNetworkClient) ValidateByoipRange(ctx context.Context, request ValidateByoipRangeRequest) (response ValidateByoipRangeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.validateByoipRange, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = ValidateByoipRangeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = ValidateByoipRangeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(ValidateByoipRangeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into ValidateByoipRangeResponse")
	}
	return
}

// validateByoipRange implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) validateByoipRange(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/byoipRanges/{byoipRangeId}/actions/validate")
	if err != nil {
		return nil, err
	}

	var response ValidateByoipRangeResponse
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

// WithdrawByoipRange Withdraws BGP route advertisement for the BYOIP CIDR block.
//
// See also
//
// Click https://docs.cloud.oracle.com/en-us/iaas/tools/go-sdk-examples/latest/core/WithdrawByoipRange.go.html to see an example of how to use WithdrawByoipRange API.
func (client VirtualNetworkClient) WithdrawByoipRange(ctx context.Context, request WithdrawByoipRangeRequest) (response WithdrawByoipRangeResponse, err error) {
	var ociResponse common.OCIResponse
	policy := common.NoRetryPolicy()
	if client.RetryPolicy() != nil {
		policy = *client.RetryPolicy()
	}
	if request.RetryPolicy() != nil {
		policy = *request.RetryPolicy()
	}
	ociResponse, err = common.Retry(ctx, request, client.withdrawByoipRange, policy)
	if err != nil {
		if ociResponse != nil {
			if httpResponse := ociResponse.HTTPResponse(); httpResponse != nil {
				opcRequestId := httpResponse.Header.Get("opc-request-id")
				response = WithdrawByoipRangeResponse{RawResponse: httpResponse, OpcRequestId: &opcRequestId}
			} else {
				response = WithdrawByoipRangeResponse{}
			}
		}
		return
	}
	if convertedResponse, ok := ociResponse.(WithdrawByoipRangeResponse); ok {
		response = convertedResponse
	} else {
		err = fmt.Errorf("failed to convert OCIResponse into WithdrawByoipRangeResponse")
	}
	return
}

// withdrawByoipRange implements the OCIOperation interface (enables retrying operations)
func (client VirtualNetworkClient) withdrawByoipRange(ctx context.Context, request common.OCIRequest) (common.OCIResponse, error) {
	httpRequest, err := request.HTTPRequest(http.MethodPost, "/byoipRanges/{byoipRangeId}/actions/withdraw")
	if err != nil {
		return nil, err
	}

	var response WithdrawByoipRangeResponse
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

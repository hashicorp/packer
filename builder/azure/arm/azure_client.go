// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	armStorage "github.com/Azure/azure-sdk-for-go/arm/storage"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/version"
)

const (
	EnvPackerLogAzureMaxLen = "PACKER_LOG_AZURE_MAXLEN"
)

var (
	packerUserAgent = fmt.Sprintf(";packer/%s", version.FormattedVersion())
)

type AzureClient struct {
	storage.BlobStorageClient
	resources.DeploymentsClient
	resources.GroupsClient
	network.PublicIPAddressesClient
	network.InterfacesClient
	network.SubnetsClient
	network.VirtualNetworksClient
	compute.ImagesClient
	compute.VirtualMachinesClient
	common.VaultClient
	armStorage.AccountsClient

	InspectorMaxLength int
	Template           *CaptureTemplate
	LastError          azureErrorResponse
}

func getCaptureResponse(body string) *CaptureTemplate {
	var operation CaptureOperation
	err := json.Unmarshal([]byte(body), &operation)
	if err != nil {
		return nil
	}

	if operation.Properties != nil && operation.Properties.Output != nil {
		return operation.Properties.Output
	}

	return nil
}

// HACK(chrboum): This method is a hack.  It was written to work around this issue
// (https://github.com/Azure/azure-sdk-for-go/issues/307) and to an extent this
// issue (https://github.com/Azure/azure-rest-api-specs/issues/188).
//
// Capturing a VM is a long running operation that requires polling.  There are
// couple different forms of polling, and the end result of a poll operation is
// discarded by the SDK.  It is expected that any discarded data can be re-fetched,
// so discarding it has minimal impact.  Unfortunately, there is no way to re-fetch
// the template returned by a capture call that I am aware of.
//
// If the second issue were fixed the VM ID would be included when GET'ing a VM.  The
// VM ID could be used to locate the captured VHD, and captured template.
// Unfortunately, the VM ID is not included so this method cannot be used either.
//
// This code captures the template and saves it to the client (the AzureClient type).
// It expects that the capture API is called only once, or rather you only care that the
// last call's value is important because subsequent requests are not persisted.  There
// is no care given to multiple threads writing this value because for our use case
// it does not matter.
func templateCapture(client *AzureClient) autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(resp *http.Response) error {
			body, bodyString := handleBody(resp.Body, math.MaxInt64)
			resp.Body = body

			captureTemplate := getCaptureResponse(bodyString)
			if captureTemplate != nil {
				client.Template = captureTemplate
			}

			return r.Respond(resp)
		})
	}
}

func errorCapture(client *AzureClient) autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(resp *http.Response) error {
			body, bodyString := handleBody(resp.Body, math.MaxInt64)
			resp.Body = body

			errorResponse := newAzureErrorResponse(bodyString)
			if errorResponse != nil {
				client.LastError = *errorResponse
			}

			return r.Respond(resp)
		})
	}
}

// WAITING(chrboum): I have logged https://github.com/Azure/azure-sdk-for-go/issues/311 to get this
// method included in the SDK.  It has been accepted, and I'll cut over to the official way
// once it ships.
func byConcatDecorators(decorators ...autorest.RespondDecorator) autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.DecorateResponder(r, decorators...)
	}
}

func NewAzureClient(subscriptionID, resourceGroupName, storageAccountName string,
	cloud *azure.Environment,
	servicePrincipalToken, servicePrincipalTokenVault *adal.ServicePrincipalToken) (*AzureClient, error) {

	var azureClient = &AzureClient{}

	maxlen := getInspectorMaxLength()

	azureClient.DeploymentsClient = resources.NewDeploymentsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DeploymentsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DeploymentsClient.RequestInspector = withInspection(maxlen)
	azureClient.DeploymentsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.DeploymentsClient.UserAgent += packerUserAgent

	azureClient.GroupsClient = resources.NewGroupsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.GroupsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.GroupsClient.RequestInspector = withInspection(maxlen)
	azureClient.GroupsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GroupsClient.UserAgent += packerUserAgent

	azureClient.ImagesClient = compute.NewImagesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.ImagesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.ImagesClient.RequestInspector = withInspection(maxlen)
	azureClient.ImagesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.ImagesClient.UserAgent += packerUserAgent

	azureClient.InterfacesClient = network.NewInterfacesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.InterfacesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.InterfacesClient.RequestInspector = withInspection(maxlen)
	azureClient.InterfacesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.InterfacesClient.UserAgent += packerUserAgent

	azureClient.SubnetsClient = network.NewSubnetsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.SubnetsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.SubnetsClient.RequestInspector = withInspection(maxlen)
	azureClient.SubnetsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.SubnetsClient.UserAgent += packerUserAgent

	azureClient.VirtualNetworksClient = network.NewVirtualNetworksClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.VirtualNetworksClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.VirtualNetworksClient.RequestInspector = withInspection(maxlen)
	azureClient.VirtualNetworksClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.VirtualNetworksClient.UserAgent += packerUserAgent

	azureClient.PublicIPAddressesClient = network.NewPublicIPAddressesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.PublicIPAddressesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.PublicIPAddressesClient.RequestInspector = withInspection(maxlen)
	azureClient.PublicIPAddressesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.PublicIPAddressesClient.UserAgent += packerUserAgent

	azureClient.VirtualMachinesClient = compute.NewVirtualMachinesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.VirtualMachinesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.VirtualMachinesClient.RequestInspector = withInspection(maxlen)
	azureClient.VirtualMachinesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), templateCapture(azureClient), errorCapture(azureClient))
	azureClient.VirtualMachinesClient.UserAgent += packerUserAgent

	azureClient.AccountsClient = armStorage.NewAccountsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.AccountsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.AccountsClient.RequestInspector = withInspection(maxlen)
	azureClient.AccountsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.AccountsClient.UserAgent += packerUserAgent

	keyVaultURL, err := url.Parse(cloud.KeyVaultEndpoint)
	if err != nil {
		return nil, err
	}

	azureClient.VaultClient = common.NewVaultClient(*keyVaultURL)
	azureClient.VaultClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalTokenVault)
	azureClient.VaultClient.RequestInspector = withInspection(maxlen)
	azureClient.VaultClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.VaultClient.UserAgent += packerUserAgent

	// If this is a managed disk build, this should be ignored.
	if resourceGroupName != "" && storageAccountName != "" {
		accountKeys, err := azureClient.AccountsClient.ListKeys(resourceGroupName, storageAccountName)
		if err != nil {
			return nil, err
		}

		storageClient, err := storage.NewClient(
			storageAccountName,
			*(*accountKeys.Keys)[0].Value,
			cloud.StorageEndpointSuffix,
			storage.DefaultAPIVersion,
			true /*useHttps*/)

		if err != nil {
			return nil, err
		}

		azureClient.BlobStorageClient = storageClient.GetBlobService()
	}

	return azureClient, nil
}

func getInspectorMaxLength() int64 {
	value, ok := os.LookupEnv(EnvPackerLogAzureMaxLen)
	if !ok {
		return math.MaxInt64
	}

	i, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	if i < 0 {
		return 0
	}

	return i
}

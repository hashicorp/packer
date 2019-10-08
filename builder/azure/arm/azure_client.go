package arm

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	newCompute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-01-01/network"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	armStorage "github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-10-01/storage"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/helper/useragent"
)

const (
	EnvPackerLogAzureMaxLen = "PACKER_LOG_AZURE_MAXLEN"
)

type AzureClient struct {
	storage.BlobStorageClient
	resources.DeploymentsClient
	resources.DeploymentOperationsClient
	resources.GroupsClient
	network.PublicIPAddressesClient
	network.InterfacesClient
	network.SubnetsClient
	network.VirtualNetworksClient
	network.SecurityGroupsClient
	compute.ImagesClient
	compute.VirtualMachinesClient
	common.VaultClient
	armStorage.AccountsClient
	compute.DisksClient
	compute.SnapshotsClient
	newCompute.GalleryImageVersionsClient
	newCompute.GalleryImagesClient

	InspectorMaxLength int
	Template           *CaptureTemplate
	LastError          azureErrorResponse
	VaultClientDelete  common.VaultClient
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
	cloud *azure.Environment, SharedGalleryTimeout time.Duration,
	servicePrincipalToken, servicePrincipalTokenVault *adal.ServicePrincipalToken) (*AzureClient, error) {

	var azureClient = &AzureClient{}

	maxlen := getInspectorMaxLength()

	azureClient.DeploymentsClient = resources.NewDeploymentsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DeploymentsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DeploymentsClient.RequestInspector = withInspection(maxlen)
	azureClient.DeploymentsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.DeploymentsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DeploymentsClient.UserAgent)

	azureClient.DeploymentOperationsClient = resources.NewDeploymentOperationsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DeploymentOperationsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DeploymentOperationsClient.RequestInspector = withInspection(maxlen)
	azureClient.DeploymentOperationsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.DeploymentOperationsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DeploymentOperationsClient.UserAgent)

	azureClient.DisksClient = compute.NewDisksClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DisksClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DisksClient.RequestInspector = withInspection(maxlen)
	azureClient.DisksClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.DisksClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DisksClient.UserAgent)

	azureClient.GroupsClient = resources.NewGroupsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.GroupsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.GroupsClient.RequestInspector = withInspection(maxlen)
	azureClient.GroupsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GroupsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.GroupsClient.UserAgent)

	azureClient.ImagesClient = compute.NewImagesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.ImagesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.ImagesClient.RequestInspector = withInspection(maxlen)
	azureClient.ImagesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.ImagesClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.ImagesClient.UserAgent)

	azureClient.InterfacesClient = network.NewInterfacesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.InterfacesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.InterfacesClient.RequestInspector = withInspection(maxlen)
	azureClient.InterfacesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.InterfacesClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.InterfacesClient.UserAgent)

	azureClient.SubnetsClient = network.NewSubnetsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.SubnetsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.SubnetsClient.RequestInspector = withInspection(maxlen)
	azureClient.SubnetsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.SubnetsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.SubnetsClient.UserAgent)

	azureClient.VirtualNetworksClient = network.NewVirtualNetworksClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.VirtualNetworksClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.VirtualNetworksClient.RequestInspector = withInspection(maxlen)
	azureClient.VirtualNetworksClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.VirtualNetworksClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.VirtualNetworksClient.UserAgent)

	azureClient.SecurityGroupsClient = network.NewSecurityGroupsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.SecurityGroupsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.SecurityGroupsClient.RequestInspector = withInspection(maxlen)
	azureClient.SecurityGroupsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.SecurityGroupsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.SecurityGroupsClient.UserAgent)

	azureClient.PublicIPAddressesClient = network.NewPublicIPAddressesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.PublicIPAddressesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.PublicIPAddressesClient.RequestInspector = withInspection(maxlen)
	azureClient.PublicIPAddressesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.PublicIPAddressesClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.PublicIPAddressesClient.UserAgent)

	azureClient.VirtualMachinesClient = compute.NewVirtualMachinesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.VirtualMachinesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.VirtualMachinesClient.RequestInspector = withInspection(maxlen)
	azureClient.VirtualMachinesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), templateCapture(azureClient), errorCapture(azureClient))
	azureClient.VirtualMachinesClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.VirtualMachinesClient.UserAgent)

	azureClient.SnapshotsClient = compute.NewSnapshotsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.SnapshotsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.SnapshotsClient.RequestInspector = withInspection(maxlen)
	azureClient.SnapshotsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.SnapshotsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.SnapshotsClient.UserAgent)

	azureClient.AccountsClient = armStorage.NewAccountsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.AccountsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.AccountsClient.RequestInspector = withInspection(maxlen)
	azureClient.AccountsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.AccountsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.AccountsClient.UserAgent)

	azureClient.GalleryImageVersionsClient = newCompute.NewGalleryImageVersionsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.GalleryImageVersionsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.GalleryImageVersionsClient.RequestInspector = withInspection(maxlen)
	azureClient.GalleryImageVersionsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GalleryImageVersionsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.GalleryImageVersionsClient.UserAgent)
	azureClient.GalleryImageVersionsClient.Client.PollingDuration = SharedGalleryTimeout

	azureClient.GalleryImagesClient = newCompute.NewGalleryImagesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.GalleryImagesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.GalleryImagesClient.RequestInspector = withInspection(maxlen)
	azureClient.GalleryImagesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GalleryImagesClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.GalleryImagesClient.UserAgent)

	keyVaultURL, err := url.Parse(cloud.KeyVaultEndpoint)
	if err != nil {
		return nil, err
	}

	azureClient.VaultClient = common.NewVaultClient(*keyVaultURL)
	azureClient.VaultClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalTokenVault)
	azureClient.VaultClient.RequestInspector = withInspection(maxlen)
	azureClient.VaultClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.VaultClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.VaultClient.UserAgent)

	// TODO(boumenot) - SDK still does not have a full KeyVault client.
	// There are two ways that KeyVault has to be accessed, and each one has their own SPN.  An authenticated SPN
	// is tied to the URL, and the URL associated with getting the secret is different than the URL
	// associated with deleting the KeyVault.  As a result, I need to have *two* different clients to
	// access KeyVault.  I did not want to split it into two separate files, so I am starting with this.
	//
	// I do not like this implementation.  It is getting long in the tooth, and should be re-examined now
	// that we have a "working" solution.
	azureClient.VaultClientDelete = common.NewVaultClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.VaultClientDelete.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.VaultClientDelete.RequestInspector = withInspection(maxlen)
	azureClient.VaultClientDelete.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.VaultClientDelete.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.VaultClientDelete.UserAgent)

	// If this is a managed disk build, this should be ignored.
	if resourceGroupName != "" && storageAccountName != "" {
		accountKeys, err := azureClient.AccountsClient.ListKeys(context.TODO(), resourceGroupName, storageAccountName)
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

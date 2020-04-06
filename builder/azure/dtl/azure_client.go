package dtl

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	newCompute "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-03-01/compute"
	dtl "github.com/Azure/azure-sdk-for-go/services/devtestlabs/mgmt/2018-09-15/dtl"
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
	compute.ImagesClient
	compute.VirtualMachinesClient
	common.VaultClient
	armStorage.AccountsClient
	compute.DisksClient
	compute.SnapshotsClient
	newCompute.GalleryImageVersionsClient
	newCompute.GalleryImagesClient

	InspectorMaxLength       int
	Template                 *CaptureTemplate
	LastError                azureErrorResponse
	VaultClientDelete        common.VaultClient
	DtlLabsClient            dtl.LabsClient
	DtlVirtualMachineClient  dtl.VirtualMachinesClient
	DtlEnvironmentsClient    dtl.EnvironmentsClient
	DtlCustomImageClient     dtl.CustomImagesClient
	DtlVirtualNetworksClient dtl.VirtualNetworksClient
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

func NewAzureClient(subscriptionID, resourceGroupName string,
	cloud *azure.Environment, SharedGalleryTimeout time.Duration, PollingDuration time.Duration,
	servicePrincipalToken *adal.ServicePrincipalToken) (*AzureClient, error) {

	var azureClient = &AzureClient{}

	maxlen := getInspectorMaxLength()

	azureClient.DtlVirtualMachineClient = dtl.NewVirtualMachinesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DtlVirtualMachineClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DtlVirtualMachineClient.RequestInspector = withInspection(maxlen)
	azureClient.DtlVirtualMachineClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), templateCapture(azureClient), errorCapture(azureClient))
	azureClient.DtlVirtualMachineClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DtlVirtualMachineClient.UserAgent)
	azureClient.DtlVirtualMachineClient.Client.PollingDuration = PollingDuration

	azureClient.DtlEnvironmentsClient = dtl.NewEnvironmentsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DtlEnvironmentsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DtlEnvironmentsClient.RequestInspector = withInspection(maxlen)
	azureClient.DtlEnvironmentsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), templateCapture(azureClient), errorCapture(azureClient))
	azureClient.DtlEnvironmentsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DtlEnvironmentsClient.UserAgent)
	azureClient.DtlEnvironmentsClient.Client.PollingDuration = PollingDuration

	azureClient.DtlLabsClient = dtl.NewLabsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DtlLabsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DtlLabsClient.RequestInspector = withInspection(maxlen)
	azureClient.DtlLabsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), templateCapture(azureClient), errorCapture(azureClient))
	azureClient.DtlLabsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DtlLabsClient.UserAgent)
	azureClient.DtlLabsClient.Client.PollingDuration = PollingDuration

	azureClient.DtlCustomImageClient = dtl.NewCustomImagesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DtlCustomImageClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DtlCustomImageClient.RequestInspector = withInspection(maxlen)
	azureClient.DtlCustomImageClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), templateCapture(azureClient), errorCapture(azureClient))
	azureClient.DtlCustomImageClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DtlCustomImageClient.UserAgent)
	azureClient.DtlCustomImageClient.PollingDuration = autorest.DefaultPollingDuration
	azureClient.DtlCustomImageClient.Client.PollingDuration = PollingDuration

	azureClient.DtlVirtualNetworksClient = dtl.NewVirtualNetworksClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.DtlVirtualNetworksClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.DtlVirtualNetworksClient.RequestInspector = withInspection(maxlen)
	azureClient.DtlVirtualNetworksClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), templateCapture(azureClient), errorCapture(azureClient))
	azureClient.DtlVirtualNetworksClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.DtlVirtualNetworksClient.UserAgent)
	azureClient.DtlVirtualNetworksClient.Client.PollingDuration = PollingDuration

	azureClient.GalleryImageVersionsClient = newCompute.NewGalleryImageVersionsClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.GalleryImageVersionsClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.GalleryImageVersionsClient.RequestInspector = withInspection(maxlen)
	azureClient.GalleryImageVersionsClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GalleryImageVersionsClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.GalleryImageVersionsClient.UserAgent)
	azureClient.GalleryImageVersionsClient.Client.PollingDuration = SharedGalleryTimeout
	azureClient.GalleryImageVersionsClient.Client.PollingDuration = PollingDuration

	azureClient.GalleryImagesClient = newCompute.NewGalleryImagesClientWithBaseURI(cloud.ResourceManagerEndpoint, subscriptionID)
	azureClient.GalleryImagesClient.Authorizer = autorest.NewBearerAuthorizer(servicePrincipalToken)
	azureClient.GalleryImagesClient.RequestInspector = withInspection(maxlen)
	azureClient.GalleryImagesClient.ResponseInspector = byConcatDecorators(byInspecting(maxlen), errorCapture(azureClient))
	azureClient.GalleryImagesClient.UserAgent = fmt.Sprintf("%s %s", useragent.String(), azureClient.GalleryImagesClient.UserAgent)
	azureClient.GalleryImagesClient.Client.PollingDuration = PollingDuration

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

package yandex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/useragent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/hashicorp/packer/builder/yandex/version"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/endpoint"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yandex-cloud/go-sdk/pkg/requestid"
	"github.com/yandex-cloud/go-sdk/pkg/retry"
	"github.com/yandex-cloud/go-sdk/sdkresolvers"
)

const (
	defaultExponentialBackoffBase = 50 * time.Millisecond
	defaultExponentialBackoffCap  = 1 * time.Minute
)

type driverYC struct {
	sdk *ycsdk.SDK
	ui  packersdk.Ui
}

func NewDriverYC(ui packersdk.Ui, ac *AccessConfig) (Driver, error) {
	log.Printf("[INFO] Initialize Yandex.Cloud client...")

	sdkConfig := ycsdk.Config{}

	if ac.Endpoint != "" {
		sdkConfig.Endpoint = ac.Endpoint
	}

	switch {
	case ac.Token == "" && ac.ServiceAccountKeyFile == "":
		log.Printf("[INFO] Use Instance Service Account for authentication")
		sdkConfig.Credentials = ycsdk.InstanceServiceAccount()

	case ac.Token != "":
		if strings.HasPrefix(ac.Token, "t1.") && strings.Count(ac.Token, ".") == 2 {
			log.Printf("[INFO] Use IAM token for authentication")
			sdkConfig.Credentials = ycsdk.NewIAMTokenCredentials(ac.Token)
		} else {
			log.Printf("[INFO] Use OAuth token for authentication")
			sdkConfig.Credentials = ycsdk.OAuthToken(ac.Token)
		}
	case ac.ServiceAccountKeyFile != "":
		log.Printf("[INFO] Use Service Account key file %q for authentication", ac.ServiceAccountKeyFile)
		key, err := iamkey.ReadFromJSONFile(ac.ServiceAccountKeyFile)
		if err != nil {
			return nil, err
		}

		credentials, err := ycsdk.ServiceAccountKey(key)
		if err != nil {
			return nil, err
		}

		sdkConfig.Credentials = credentials
	}

	requestIDInterceptor := requestid.Interceptor()

	retryInterceptor := retry.Interceptor(
		retry.WithMax(ac.MaxRetries),
		retry.WithCodes(codes.Unavailable),
		retry.WithAttemptHeader(true),
		retry.WithBackoff(retry.BackoffExponentialWithJitter(defaultExponentialBackoffBase, defaultExponentialBackoffCap)))

	// Make sure retry interceptor is above id interceptor.
	// Now we will have new request id for every retry attempt.
	interceptorChain := grpc_middleware.ChainUnaryClient(retryInterceptor, requestIDInterceptor)

	userAgentMD := metadata.Pairs("user-agent", useragent.String(version.YandexPluginVersion.FormattedVersion()))

	sdk, err := ycsdk.Build(context.Background(), sdkConfig,
		grpc.WithDefaultCallOptions(grpc.Header(&userAgentMD)),
		grpc.WithUnaryInterceptor(interceptorChain))

	if err != nil {
		return nil, err
	}

	if _, err = sdk.ApiEndpoint().ApiEndpoint().List(context.Background(), &endpoint.ListApiEndpointsRequest{}); err != nil {
		return nil, err
	}

	return &driverYC{
		sdk: sdk,
		ui:  ui,
	}, nil

}

func (d *driverYC) GetImage(imageID string) (*Image, error) {
	image, err := d.sdk.Compute().Image().Get(context.Background(), &compute.GetImageRequest{
		ImageId: imageID,
	})
	if err != nil {
		return nil, err
	}

	return &Image{
		ID:            image.Id,
		Labels:        image.Labels,
		Licenses:      image.ProductIds,
		Name:          image.Name,
		Description:   image.Description,
		FolderID:      image.FolderId,
		MinDiskSizeGb: toGigabytes(image.MinDiskSize),
		SizeGb:        toGigabytes(image.StorageSize),
	}, nil
}

func (d *driverYC) GetImageFromFolder(ctx context.Context, folderID string, family string) (*Image, error) {
	image, err := d.sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
		FolderId: folderID,
		Family:   family,
	})
	if err != nil {
		return nil, err
	}

	return &Image{
		ID:            image.Id,
		Labels:        image.Labels,
		Licenses:      image.ProductIds,
		Name:          image.Name,
		Description:   image.Description,
		FolderID:      image.FolderId,
		Family:        image.Family,
		MinDiskSizeGb: toGigabytes(image.MinDiskSize),
		SizeGb:        toGigabytes(image.StorageSize),
	}, nil
}

func (d *driverYC) GetImageFromFolderByName(ctx context.Context, folderID string, imageName string) (*Image, error) {
	imageResolver := sdkresolvers.ImageResolver(imageName, sdkresolvers.FolderID(folderID))

	if err := d.sdk.Resolve(ctx, imageResolver); err != nil {
		return nil, fmt.Errorf("failed to resolve image name: %s", err)
	}

	return d.GetImage(imageResolver.ID())
}

func (d *driverYC) DeleteImage(ID string) error {
	ctx := context.TODO()
	op, err := d.sdk.WrapOperation(d.sdk.Compute().Image().Delete(ctx, &compute.DeleteImageRequest{
		ImageId: ID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err
}

func (d *driverYC) SDK() *ycsdk.SDK {
	return d.sdk
}

func (d *driverYC) DeleteInstance(ctx context.Context, instanceID string) error {
	op, err := d.sdk.WrapOperation(d.sdk.Compute().Instance().Delete(ctx, &compute.DeleteInstanceRequest{
		InstanceId: instanceID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err
}

func (d *driverYC) DeleteSubnet(ctx context.Context, subnetID string) error {
	op, err := d.sdk.WrapOperation(d.sdk.VPC().Subnet().Delete(ctx, &vpc.DeleteSubnetRequest{
		SubnetId: subnetID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err
}

func (d *driverYC) DeleteNetwork(ctx context.Context, networkID string) error {
	op, err := d.sdk.WrapOperation(d.sdk.VPC().Network().Delete(ctx, &vpc.DeleteNetworkRequest{
		NetworkId: networkID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err

}

func (d *driverYC) DeleteDisk(ctx context.Context, diskID string) error {
	op, err := d.sdk.WrapOperation(d.sdk.Compute().Disk().Delete(ctx, &compute.DeleteDiskRequest{
		DiskId: diskID,
	}))
	if err != nil {
		return err
	}

	err = op.Wait(ctx)
	if err != nil {
		return err
	}

	_, err = op.Response()
	return err
}

func (d *driverYC) GetInstanceMetadata(ctx context.Context, instanceID string, key string) (string, error) {
	instance, err := d.sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: instanceID,
		View:       compute.InstanceView_FULL,
	})
	if err != nil {
		return "", err
	}

	for k, v := range instance.GetMetadata() {
		if k == key {
			return v, nil
		}
	}

	return "", fmt.Errorf("Instance metadata key, %s, not found.", key)
}

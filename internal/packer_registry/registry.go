package packer_registry

import (
	"context"
	"fmt"

	"github.com/go-openapi/runtime"
	packerSvc "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	"google.golang.org/grpc/codes"
)

// CreateBucket creates a bucket on a Packer Artifact Registry.
func CreateBucket(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBucketRequest) error {

	// Create bucket if exist we continue as is, eventually we want to treat this like an upsert

	params := packerSvc.NewCreateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.Body = input

	_, err := client.Packer.CreateBucket(params, nil, func(*runtime.ClientOperation) {})
	if err != nil {
		return err
	}

	return nil
}

// UpsertBucket tries to update a bucket on a Packer Artifact Registry. If the bucket doesn't exist it creates it.
func UpsertBucket(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBucketRequest) error {

	err := CreateBucket(ctx, client, input)
	if err != nil && !checkErrorCode(err, codes.AlreadyExists) {
		return err
	}

	if err == nil {
		return nil
	}

	params := packerSvc.NewUpdateBucketParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.BucketSlug = input.BucketSlug
	_, err = client.Packer.UpdateBucket(params, nil, func(*runtime.ClientOperation) {})

	return err
}

/* CreateIteration creates an Iteration for some Bucket on a Packer Artifact Registry for the given input
and returns the ID associated with the persisted Bucket iteration.
input: *models.HashicorpCloudPackerCreateIterationRequest{BucketSlug: "bucket name"
*/
func CreateIteration(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateIterationRequest) (string, error) {
	// Create/find iteration
	params := packerSvc.NewCreateIterationParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.BucketSlug = input.BucketSlug
	params.Body = input

	it, err := client.Packer.CreateIteration(params, nil, func(*runtime.ClientOperation) {})
	if err != nil {
		return "", err
	}

	return it.Payload.Iteration.ID, nil
}

func UpsertBuild(ctx context.Context, client *Client, input *models.HashicorpCloudPackerCreateBuildRequest) error {
	params := packerSvc.NewCreateBuildParamsWithContext(ctx)
	params.LocationOrganizationID = client.Config.OrganizationID
	params.LocationProjectID = client.Config.ProjectID
	params.BucketSlug = input.BucketSlug
	params.BuildIterationID = input.Build.IterationID
	params.Body = input

	b, err := client.Packer.CreateBuild(params, nil, func(*runtime.ClientOperation) {})
	fmt.Printf("%#v", b)
	return err

}

package paracctest

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"

	structenv "github.com/caarlos0/env/v6"
	cloud_packer "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/hcp-sdk-go/httpclient"
	"github.com/hashicorp/packer/acctest"
)

type hcpConf struct {
	OrgId        string `env:"HCP_ORG_ID"`
	ProjectId    string `env:"HCP_PROJECT_ID"`
	ApiHost      string `env:"HCP_API_HOST"`
	AuthUrl      string `env:"HCP_AUTH_URL"`
	ClientId     string `env:"HCP_CLIENT_ID"`
	ClientSecret string `env:"HCP_CLIENT_SECRET"`
	UserAgent    string `env:"HCP_CLIENT_USER_AGENT" envDefault:"packer-par-acc-test"`
}

type Config struct {
	Client packer_service.ClientService
	Loc    *sharedmodels.HashicorpCloudLocationLocation
	T      *testing.T
}

func NewParConfig(t *testing.T) (*Config, error) {
	checkEnvVars(t)
	cfg := hcpConf{}
	if err := structenv.Parse(&cfg); err != nil {
		t.Errorf("%+v\n", err)
	}

	httpClient, err := httpclient.New(httpclient.Config{
		ClientID:      cfg.ClientId,
		ClientSecret:  cfg.ClientSecret,
		SourceChannel: cfg.UserAgent,
	})
	if err != nil {
		return nil, err
	}

	return &Config{
		Client: cloud_packer.New(httpClient, nil).PackerService,
		Loc: &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: cfg.OrgId,
			ProjectID:      cfg.ProjectId,
		},
		T: t,
	}, nil
}

func checkEnvVars(t *testing.T) {
	t.Helper()
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skip(fmt.Sprintf(
			"Acceptance tests skipped unless env '%s' set",
			acctest.TestEnvVar))
		return
	}
	if os.Getenv("HCP_CLIENT_ID") == "" {
		t.Fatal("HCP_CLIENT_ID must be set for acceptance tests")
	}
	if os.Getenv("HCP_CLIENT_SECRET") == "" {
		t.Fatal("HCP_CLIENT_SECRET must be set for acceptance tests")
	}
	if os.Getenv("HCP_ORG_ID") == "" {
		t.Fatal("HCP_ORG_ID must be set for acceptance tests")
	}
	if os.Getenv("HCP_PROJECT_ID") == "" {
		t.Fatal("HCP_PROJECT_ID must be set for acceptance tests")
	}
}

// UpsertBucket creates a new bucket if it does not already exists.
func (cfg *Config) CreateBucket(
	bucketSlug string,
) (*packer_service.CreateBucketOK, error) {
	cfg.T.Helper()

	createBktParams := packer_service.NewCreateBucketParams()
	createBktParams.LocationOrganizationID = cfg.Loc.OrganizationID
	createBktParams.LocationProjectID = cfg.Loc.ProjectID
	createBktParams.Body = &models.HashicorpCloudPackerCreateBucketRequest{
		BucketSlug: bucketSlug,
		Location:   cfg.Loc,
	}
	return cfg.Client.CreateBucket(createBktParams, nil)
}

func (cfg *Config) UpsertBucket(
	bucketSlug string,
) {
	cfg.T.Helper()
	_, err := cfg.CreateBucket(bucketSlug)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.CreateBucketDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return
		}
	}

	cfg.T.Errorf("unexpected CreateBucket error, expected nil or 409. Got %v", err)
}

// UpsertIteration creates a new iteration if it does not already exists.
func (cfg *Config) GetIterationByID(
	bucketSlug,
	id string,
) string {
	cfg.T.Helper()

	getItParams := packer_service.NewGetIterationParams()
	getItParams.LocationOrganizationID = cfg.Loc.OrganizationID
	getItParams.LocationProjectID = cfg.Loc.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.IterationID = &id

	ok, err := cfg.Client.GetIteration(getItParams, nil)
	if err != nil {
		cfg.T.Fatal(err)
	}
	return ok.Payload.Iteration.ID
}

// UpsertIteration creates a new iteration if it does not already exists.
func (cfg *Config) UpsertIteration(
	bucketSlug,
	fingerprint string,
) string {
	cfg.T.Helper()

	createItParams := packer_service.NewCreateIterationParams()
	createItParams.LocationOrganizationID = cfg.Loc.OrganizationID
	createItParams.LocationProjectID = cfg.Loc.ProjectID
	createItParams.BucketSlug = bucketSlug

	createItParams.Body = &models.HashicorpCloudPackerCreateIterationRequest{
		BucketSlug:  bucketSlug,
		Fingerprint: fingerprint,
		Location:    cfg.Loc,
	}
	_, err := cfg.Client.CreateIteration(createItParams, nil)
	if err == nil {
		return cfg.GetIterationIDFromFingerPrint(bucketSlug, fingerprint)
	}
	if err, ok := err.(*packer_service.CreateIterationDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return cfg.GetIterationIDFromFingerPrint(bucketSlug, fingerprint)

		}
	}

	cfg.T.Fatalf("unexpected CreateIteration error, expected nil or 409. Got %v", err)
	return ""
}

// GetIterationIDFromFingerPrint returns an iteration ID given its unique
// fingerprincfg.t.
func (cfg *Config) GetIterationIDFromFingerPrint(
	bucketSlug,
	fingerprint string,
) string {
	cfg.T.Helper()

	getItParams := packer_service.NewGetIterationParams()
	getItParams.LocationOrganizationID = cfg.Loc.OrganizationID
	getItParams.LocationProjectID = cfg.Loc.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.Fingerprint = &fingerprint

	ok, err := cfg.Client.GetIteration(getItParams, nil)
	if err != nil {
		cfg.T.Fatal(err)
	}
	return ok.Payload.Iteration.ID
}

// UpsertBuild creates a new build for iteration if it does not already exists.
func (cfg *Config) UpsertBuild(
	bucketSlug,
	iterationFingerprint,
	iterationID,
	cloudProvider,
	region string,
	imageIDs []string,
) {

	createBuildParams := packer_service.NewCreateBuildParams()
	createBuildParams.LocationOrganizationID = cfg.Loc.OrganizationID
	createBuildParams.LocationProjectID = cfg.Loc.ProjectID
	createBuildParams.BucketSlug = bucketSlug
	createBuildParams.BuildIterationID = iterationID

	createBuildParams.Body = &models.HashicorpCloudPackerCreateBuildRequest{
		Fingerprint: iterationFingerprint,
		BucketSlug:  bucketSlug,
		Location:    cfg.Loc,
	}
	createBuildParams.Body.Build = &models.HashicorpCloudPackerBuild{
		PackerRunUUID: uuid.New().String(),
		CloudProvider: cloudProvider,
		ComponentType: "acceptance.test",
		IterationID:   iterationID,
		Status:        models.HashicorpCloudPackerBuildStatusRUNNING,
	}

	build, err := cfg.Client.CreateBuild(createBuildParams, nil)
	if err, ok := err.(*packer_service.CreateBuildDefault); ok {
		switch err.Code() {
		case int(codes.Aborted), http.StatusConflict:
			// all good here !
			return
		}
	}

	if build == nil {
		cfg.T.Errorf("unexpected CreateBuild error, expected non nil build response. Got %v", err)
		return
	}

	// Iterations are currently only assigned an incremental version when publishing image metadata on update.
	// Incremental versions are a requirement for assigning the channel.
	updateBuildParams := packer_service.NewUpdateBuildParams()
	updateBuildParams.LocationOrganizationID = cfg.Loc.OrganizationID
	updateBuildParams.LocationProjectID = cfg.Loc.ProjectID
	updateBuildParams.BuildID = build.Payload.Build.ID
	updateBuildParams.Body = &models.HashicorpCloudPackerUpdateBuildRequest{
		Updates: &models.HashicorpCloudPackerBuildUpdates{
			CloudProvider: cloudProvider,
			Status:        models.HashicorpCloudPackerBuildStatusDONE,
		},
	}
	for _, imageID := range imageIDs {
		updateBuildParams.Body.Updates.Images = append(updateBuildParams.Body.Updates.Images, &models.HashicorpCloudPackerImage{
			ImageID: imageID,
			Region:  region,
		})
	}
	_, err = cfg.Client.UpdateBuild(updateBuildParams, nil)
	if err, ok := err.(*packer_service.UpdateBuildDefault); ok {
		cfg.T.Errorf("unexpected UpdateBuild error, expected nil. Got %v", err)
	}
}

func (cfg *Config) UpsertChannel(
	bucketSlug,
	channelSlug,
	iterationID string,
) {
	cfg.T.Helper()

	createChParams := packer_service.NewCreateChannelParams()
	createChParams.LocationOrganizationID = cfg.Loc.OrganizationID
	createChParams.LocationProjectID = cfg.Loc.ProjectID
	createChParams.BucketSlug = bucketSlug
	createChParams.Body = &models.HashicorpCloudPackerCreateChannelRequest{
		Slug:               channelSlug,
		IncrementalVersion: 1,
		IterationID:        iterationID,
	}

	_, err := cfg.Client.CreateChannel(createChParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.CreateChannelDefault); ok {
		switch err.Code() {
		case int(codes.Aborted), http.StatusConflict:
			// all good here !
			cfg.UpdateChannel(bucketSlug, channelSlug)
			return
		}
	}
	cfg.T.Errorf("unexpected CreateChannel error, expected nil. Got %v", err)
}

func (cfg *Config) UpdateChannel(
	bucketSlug,
	channelSlug string,
) {
	cfg.T.Helper()

	updateChParams := packer_service.NewUpdateChannelParams()
	updateChParams.LocationOrganizationID = cfg.Loc.OrganizationID
	updateChParams.LocationProjectID = cfg.Loc.ProjectID
	updateChParams.BucketSlug = bucketSlug
	updateChParams.Slug = channelSlug
	updateChParams.Body = &models.HashicorpCloudPackerUpdateChannelRequest{
		IncrementalVersion: 1,
	}

	_, err := cfg.Client.UpdateChannel(updateChParams, nil)
	if err == nil {
		return
	}
	cfg.T.Errorf("unexpected UpdateChannel error, expected nil. Got %v", err)
}

func (cfg *Config) DeleteBucket(
	bucketSlug string,
) {
	cfg.T.Helper()

	deleteBktParams := packer_service.NewDeleteBucketParams()
	deleteBktParams.LocationOrganizationID = cfg.Loc.OrganizationID
	deleteBktParams.LocationProjectID = cfg.Loc.ProjectID
	deleteBktParams.BucketSlug = bucketSlug

	_, err := cfg.Client.DeleteBucket(deleteBktParams, nil)
	if err == nil {
		return
	}
	cfg.T.Errorf("unexpected DeleteBucket error, expected nil. Got %v", err)
}

func (cfg *Config) DeleteIteration(
	bucketSlug,
	iterationID string,
) {
	cfg.T.Helper()

	deleteItParams := packer_service.NewDeleteIterationParams()
	deleteItParams.LocationOrganizationID = cfg.Loc.OrganizationID
	deleteItParams.LocationProjectID = cfg.Loc.ProjectID
	deleteItParams.BucketSlug = &bucketSlug
	deleteItParams.IterationID = iterationID

	_, err := cfg.Client.DeleteIteration(deleteItParams, nil)
	if err == nil {
		return
	}
	cfg.T.Errorf("unexpected DeleteIteration error, expected nil. Got %v", err)
}

func (cfg *Config) DeleteChannel(
	bucketSlug,
	channelSlug string,
) {
	cfg.T.Helper()

	deleteChParams := packer_service.NewDeleteChannelParams()
	deleteChParams.LocationOrganizationID = cfg.Loc.OrganizationID
	deleteChParams.LocationProjectID = cfg.Loc.ProjectID
	deleteChParams.BucketSlug = bucketSlug
	deleteChParams.Slug = channelSlug

	_, err := cfg.Client.DeleteChannel(deleteChParams, nil)
	if err == nil {
		return
	}
	cfg.T.Errorf("unexpected DeleteChannel error, expected nil. Got %v", err)
}

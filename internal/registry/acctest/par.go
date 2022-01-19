package acctest

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	structenv "github.com/caarlos0/env/v6"
	packer_service "github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/client/packer_service"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-packer-service/preview/2021-04-30/models"
	sharedmodels "github.com/hashicorp/hcp-sdk-go/clients/cloud-shared/v1/models"
	"github.com/hashicorp/packer/acctest"
	"github.com/hashicorp/packer/internal/registry"
	"google.golang.org/grpc/codes"
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
	*registry.Client
	Loc *sharedmodels.HashicorpCloudLocationLocation
	T   *testing.T
}

func NewTestConfig(t *testing.T) (*Config, error) {
	checkEnvVars(t)
	cfg := hcpConf{}
	if err := structenv.Parse(&cfg); err != nil {
		t.Errorf("%+v\n", err)
	}
	cli, err := registry.NewClient()
	if err != nil {
		return nil, err
	}

	return &Config{
		Client: cli,
		Loc: &sharedmodels.HashicorpCloudLocationLocation{
			OrganizationID: cli.OrganizationID,
			ProjectID:      cli.ProjectID,
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
}

// GetIterationByID is a helper to validate that the id for a given bucketSlug returns a valid Iteration ID.
func (cfg *Config) GetIterationByID(
	bucketSlug,
	id string,
) string {
	cfg.T.Helper()

	getItParams := packer_service.NewPackerServiceGetIterationParams()
	getItParams.LocationOrganizationID = cfg.Loc.OrganizationID
	getItParams.LocationProjectID = cfg.Loc.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.IterationID = &id

	ok, err := cfg.Packer.PackerServiceGetIteration(getItParams, nil)
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

	_, err := cfg.CreateIteration(context.Background(), bucketSlug, fingerprint)
	if err == nil {
		return cfg.GetIterationIDFromFingerPrint(bucketSlug, fingerprint)
	}
	if err, ok := err.(*packer_service.PackerServiceCreateIterationDefault); ok {
		switch err.Code() {
		case int(codes.AlreadyExists), http.StatusConflict:
			// all good here !
			return cfg.GetIterationIDFromFingerPrint(bucketSlug, fingerprint)
		}
	}

	cfg.T.Fatalf("unexpected CreateIteration error, expected nil or 409. Got %v", err)
	return ""
}

// MarkIterationAsDone sends an update request to mark an existing iteration as complete.
func (cfg *Config) MarkIterationAsDone(
	bucketSlug,
	iterID string,
) {
	cfg.T.Helper()

	updateItParams := packer_service.NewPackerServiceUpdateIterationParams()
	updateItParams.Body = &models.HashicorpCloudPackerUpdateIterationRequest{
		BucketSlug:  bucketSlug,
		IterationID: iterID,
		Complete:    true,
	}
	updateItParams.IterationID = iterID
	updateItParams.LocationOrganizationID = cfg.OrganizationID
	updateItParams.LocationProjectID = cfg.ProjectID

	_, err := cfg.Packer.PackerServiceUpdateIteration(updateItParams, nil)
	if err == nil {
		return
	}

	cfg.T.Errorf("unexpected UpdateIteration error: %v", err)
}

// GetIterationIDFromFingerPrint returns an iteration ID given its unique
// fingerprincfg.t.
func (cfg *Config) GetIterationIDFromFingerPrint(
	bucketSlug,
	fingerprint string,
) string {
	cfg.T.Helper()

	getItParams := packer_service.NewPackerServiceGetIterationParams()
	getItParams.LocationOrganizationID = cfg.Loc.OrganizationID
	getItParams.LocationProjectID = cfg.Loc.ProjectID
	getItParams.BucketSlug = bucketSlug
	getItParams.Fingerprint = &fingerprint

	ok, err := cfg.Packer.PackerServiceGetIteration(getItParams, nil)
	if err != nil {
		cfg.T.Fatal(err)
	}
	return ok.Payload.Iteration.ID
}

// UpsertBuild creates a new build for iteration if it does not already exists.
func (cfg *Config) UpsertBuild(
	bucketSlug,
	iterationFingerprint,
	runUUID,
	iterationID,
	cloudProvider,
	componentType,
	region string,
	imageIDs []string,
) {

	build, err := cfg.CreateBuild(context.Background(),
		bucketSlug,
		runUUID,
		iterationID,
		iterationFingerprint,
		componentType,
		models.HashicorpCloudPackerBuildStatusUNSET,
	)
	if err, ok := err.(*packer_service.PackerServiceCreateBuildDefault); ok {
		switch err.Code() {
		case int(codes.Aborted), http.StatusConflict:
			// all good here !
			return
		default:
			cfg.T.Fatalf("couldn't create build: %v", err)
		}
	}

	if build == nil {
		cfg.T.Errorf("unexpected CreateBuild error, expected non nil build response. Got %v", err)
		return
	}

	// Iterations are currently only assigned an incremental version when publishing image metadata on update.
	// Incremental versions are a requirement for assigning the channel.
	updateBuildParams := packer_service.NewPackerServiceUpdateBuildParams()
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
		updateBuildParams.Body.Updates.Images = append(updateBuildParams.Body.Updates.Images, &models.HashicorpCloudPackerImageCreateBody{
			ImageID: imageID,
			Region:  region,
		})
	}
	_, err = cfg.Packer.PackerServiceUpdateBuild(updateBuildParams, nil)
	if err, ok := err.(*packer_service.PackerServiceUpdateBuildDefault); ok {
		cfg.T.Errorf("unexpected UpdateBuild error, expected nil. Got %v", err)
	}
}

func (cfg *Config) UpsertChannel(
	bucketSlug,
	channelSlug,
	iterationID string,
) {
	cfg.T.Helper()

	createChParams := packer_service.NewPackerServiceCreateChannelParams()
	createChParams.LocationOrganizationID = cfg.Loc.OrganizationID
	createChParams.LocationProjectID = cfg.Loc.ProjectID
	createChParams.BucketSlug = bucketSlug
	createChParams.Body = &models.HashicorpCloudPackerCreateChannelRequest{
		Slug:               channelSlug,
		IncrementalVersion: 1,
		IterationID:        iterationID,
	}

	_, err := cfg.Packer.PackerServiceCreateChannel(createChParams, nil)
	if err == nil {
		return
	}
	if err, ok := err.(*packer_service.PackerServiceCreateChannelDefault); ok {
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

	updateChParams := packer_service.NewPackerServiceUpdateChannelParams()
	updateChParams.LocationOrganizationID = cfg.Loc.OrganizationID
	updateChParams.LocationProjectID = cfg.Loc.ProjectID
	updateChParams.BucketSlug = bucketSlug
	updateChParams.Slug = channelSlug
	updateChParams.Body = &models.HashicorpCloudPackerUpdateChannelRequest{
		IncrementalVersion: 1,
	}

	_, err := cfg.Packer.PackerServiceUpdateChannel(updateChParams, nil)
	if err == nil {
		return
	}
	cfg.T.Errorf("unexpected UpdateChannel error, expected nil. Got %v", err)
}

func (cfg *Config) DeleteIteration(
	bucketSlug,
	iterationID string,
) {
	cfg.T.Helper()

	deleteItParams := packer_service.NewPackerServiceDeleteIterationParams()
	deleteItParams.LocationOrganizationID = cfg.Loc.OrganizationID
	deleteItParams.LocationProjectID = cfg.Loc.ProjectID
	deleteItParams.BucketSlug = &bucketSlug
	deleteItParams.IterationID = iterationID

	_, err := cfg.Packer.PackerServiceDeleteIteration(deleteItParams, nil)
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

	deleteChParams := packer_service.NewPackerServiceDeleteChannelParams()
	deleteChParams.LocationOrganizationID = cfg.Loc.OrganizationID
	deleteChParams.LocationProjectID = cfg.Loc.ProjectID
	deleteChParams.BucketSlug = bucketSlug
	deleteChParams.Slug = channelSlug

	_, err := cfg.Packer.PackerServiceDeleteChannel(deleteChParams, nil)
	if err == nil {
		return
	}
	cfg.T.Errorf("unexpected DeleteChannel error, expected nil. Got %v", err)
}

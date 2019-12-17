//go:generate mapstructure-to-hcl2 -type Config

package googlecomputeimport

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/storage/v1"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/googlecompute"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/compress"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	AccountFile string `mapstructure:"account_file"`
	ProjectId   string `mapstructure:"project_id"`

	Bucket               string            `mapstructure:"bucket"`
	GCSObjectName        string            `mapstructure:"gcs_object_name"`
	ImageDescription     string            `mapstructure:"image_description"`
	ImageFamily          string            `mapstructure:"image_family"`
	ImageGuestOsFeatures []string          `mapstructure:"image_guest_os_features"`
	ImageLabels          map[string]string `mapstructure:"image_labels"`
	ImageName            string            `mapstructure:"image_name"`
	SkipClean            bool              `mapstructure:"skip_clean"`
	VaultGCPOauthEngine  string            `mapstructure:"vault_gcp_oauth_engine"`

	account *jwt.Config
	ctx     interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"gcs_object_name",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	errs := new(packer.MultiError)

	// Set defaults
	if p.config.GCSObjectName == "" {
		p.config.GCSObjectName = "packer-import-{{timestamp}}.tar.gz"
	}

	// Check and render gcs_object_name
	if err = interpolate.Validate(p.config.GCSObjectName, &p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing gcs_object_name template: %s", err))
	}

	if p.config.AccountFile != "" {
		cfg, err := googlecompute.ProcessAccountFile(p.config.AccountFile)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
		p.config.account = cfg
	}

	if p.config.AccountFile != "" && p.config.VaultGCPOauthEngine != "" {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("May set either account_file or "+
				"vault_gcp_oauth_engine, but not both."))
	}

	templates := map[string]*string{
		"bucket":     &p.config.Bucket,
		"image_name": &p.config.ImageName,
		"project_id": &p.config.ProjectId,
	}
	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	client, err := googlecompute.NewClientGCE(p.config.account, p.config.VaultGCPOauthEngine)
	if err != nil {
		return nil, false, false, err
	}

	if artifact.BuilderId() != compress.BuilderId {
		err = fmt.Errorf(
			"incompatible artifact type: %s\nCan only import from Compress post-processor artifacts",
			artifact.BuilderId())
		return nil, false, false, err
	}

	p.config.GCSObjectName, err = interpolate.Render(p.config.GCSObjectName, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("Error rendering gcs_object_name template: %s", err)
	}

	rawImageGcsPath, err := UploadToBucket(client, ui, artifact, p.config.Bucket, p.config.GCSObjectName)
	if err != nil {
		return nil, false, false, err
	}

	gceImageArtifact, err := CreateGceImage(client, ui, p.config.ProjectId, rawImageGcsPath, p.config.ImageName, p.config.ImageDescription, p.config.ImageFamily, p.config.ImageLabels, p.config.ImageGuestOsFeatures)
	if err != nil {
		return nil, false, false, err
	}

	if !p.config.SkipClean {
		err = DeleteFromBucket(client, ui, p.config.Bucket, p.config.GCSObjectName)
		if err != nil {
			return nil, false, false, err
		}
	}

	return gceImageArtifact, false, false, nil
}

func UploadToBucket(client *http.Client, ui packer.Ui, artifact packer.Artifact, bucket string, gcsObjectName string) (string, error) {
	service, err := storage.New(client)
	if err != nil {
		return "", err
	}

	ui.Say("Looking for tar.gz file in list of artifacts...")
	source := ""
	for _, path := range artifact.Files() {
		ui.Say(fmt.Sprintf("Found artifact %v...", path))
		if strings.HasSuffix(path, ".tar.gz") {
			source = path
			break
		}
	}

	if source == "" {
		return "", fmt.Errorf("No tar.gz file found in list of artifacts")
	}

	artifactFile, err := os.Open(source)
	if err != nil {
		err := fmt.Errorf("error opening %v", source)
		return "", err
	}

	ui.Say(fmt.Sprintf("Uploading file %v to GCS bucket %v/%v...", source, bucket, gcsObjectName))
	storageObject, err := service.Objects.Insert(bucket, &storage.Object{Name: gcsObjectName}).Media(artifactFile).Do()
	if err != nil {
		ui.Say(fmt.Sprintf("Failed to upload: %v", storageObject))
		return "", err
	}

	return storageObject.SelfLink, nil
}

func CreateGceImage(client *http.Client, ui packer.Ui, project string, rawImageURL string, imageName string, imageDescription string, imageFamily string, imageLabels map[string]string, imageGuestOsFeatures []string) (packer.Artifact, error) {
	service, err := compute.New(client)
	if err != nil {
		return nil, err
	}

	// Build up the imageFeatures
	imageFeatures := make([]*compute.GuestOsFeature, len(imageGuestOsFeatures))
	for _, v := range imageGuestOsFeatures {
		imageFeatures = append(imageFeatures, &compute.GuestOsFeature{
			Type: v,
		})
	}

	gceImage := &compute.Image{
		Description:     imageDescription,
		Family:          imageFamily,
		GuestOsFeatures: imageFeatures,
		Labels:          imageLabels,
		Name:            imageName,
		RawDisk:         &compute.ImageRawDisk{Source: rawImageURL},
		SourceType:      "RAW",
	}

	ui.Say(fmt.Sprintf("Creating GCE image %v...", imageName))
	op, err := service.Images.Insert(project, gceImage).Do()
	if err != nil {
		ui.Say("Error creating GCE image")
		return nil, err
	}

	ui.Say("Waiting for GCE image creation operation to complete...")
	for op.Status != "DONE" {
		op, err = service.GlobalOperations.Get(project, op.Name).Do()
		if err != nil {
			return nil, err
		}

		time.Sleep(5 * time.Second)
	}

	// fail if image creation operation has an error
	if op.Error != nil {
		var imageError string
		for _, error := range op.Error.Errors {
			imageError += error.Message
		}
		err = fmt.Errorf("failed to create GCE image %s: %s", imageName, imageError)
		return nil, err
	}

	return &Artifact{paths: []string{op.TargetLink}}, nil
}

func DeleteFromBucket(client *http.Client, ui packer.Ui, bucket string, gcsObjectName string) error {
	service, err := storage.New(client)
	if err != nil {
		return err
	}

	ui.Say(fmt.Sprintf("Deleting import source from GCS %s/%s...", bucket, gcsObjectName))
	err = service.Objects.Delete(bucket, gcsObjectName).Do()
	if err != nil {
		ui.Say(fmt.Sprintf("Failed to delete: %v/%v", bucket, gcsObjectName))
		return err
	}

	return nil
}

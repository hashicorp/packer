//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package googlecomputeimport

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/googlecompute"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/artifice"
	"github.com/hashicorp/packer/post-processor/compress"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	//The JSON file containing your account credentials.
	//If specified, the account file will take precedence over any `googlecompute` builder authentication method.
	AccountFile string `mapstructure:"account_file" required:"true"`
	// This allows service account impersonation as per the [docs](https://cloud.google.com/iam/docs/impersonating-service-accounts).
	ImpersonateServiceAccount string `mapstructure:"impersonate_service_account" required:"false"`
	//The project ID where the GCS bucket exists and where the GCE image is stored.
	ProjectId string `mapstructure:"project_id" required:"true"`
	IAP       bool   `mapstructure-to-hcl:",skip"`
	//The name of the GCS bucket where the raw disk image will be uploaded.
	Bucket string `mapstructure:"bucket" required:"true"`
	//The name of the GCS object in `bucket` where
	//the RAW disk image will be copied for import. This is treated as a
	//[template engine](/docs/templates/engine). Therefore, you
	//may use user variables and template functions in this field. Defaults to
	//`packer-import-{{timestamp}}.tar.gz`.
	GCSObjectName string `mapstructure:"gcs_object_name"`
	//The description of the resulting image.
	ImageDescription string `mapstructure:"image_description"`
	//The name of the image family to which the resulting image belongs.
	ImageFamily string `mapstructure:"image_family"`
	//A list of features to enable on the guest operating system. Applicable only for bootable images. Valid
	//values are `MULTI_IP_SUBNET`, `SECURE_BOOT`, `UEFI_COMPATIBLE`,
	//`VIRTIO_SCSI_MULTIQUEUE` and `WINDOWS` currently.
	ImageGuestOsFeatures []string `mapstructure:"image_guest_os_features"`
	//Key/value pair labels to apply to the created image.
	ImageLabels map[string]string `mapstructure:"image_labels"`
	//The unique name of the resulting image.
	ImageName string `mapstructure:"image_name" required:"true"`
	//Skip removing the TAR file uploaded to the GCS
	//bucket after the import process has completed. "true" means that we should
	//leave it in the GCS bucket, "false" means to clean it out. Defaults to
	//`false`.
	SkipClean           bool   `mapstructure:"skip_clean"`
	VaultGCPOauthEngine string `mapstructure:"vault_gcp_oauth_engine"`

	account *googlecompute.ServiceAccount
	ctx     interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec { return p.config.FlatMapstructure().HCL2Spec() }

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		PluginType:         BuilderId,
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
		if p.config.VaultGCPOauthEngine != "" && p.config.ImpersonateServiceAccount != "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("You cannot "+
				"specify impersonate_service_account, account_file and vault_gcp_oauth_engine at the same time"))
		}
		cfg, err := googlecompute.ProcessAccountFile(p.config.AccountFile)
		if err != nil {
			errs = packer.MultiErrorAppend(errs, err)
		}
		p.config.account = cfg
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
	generatedData := artifact.State("generated_data")
	if generatedData == nil {
		// Make sure it's not a nil map so we can assign to it later.
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData
	var err error
	var opts option.ClientOption
	opts, err = googlecompute.NewClientOptionGoogle(p.config.account, p.config.VaultGCPOauthEngine, p.config.ImpersonateServiceAccount)
	if err != nil {
		return nil, false, false, err
	}

	switch artifact.BuilderId() {
	case compress.BuilderId, artifice.BuilderId:
		break
	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from Compress post-processor and Artifice post-processor artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	p.config.GCSObjectName, err = interpolate.Render(p.config.GCSObjectName, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("Error rendering gcs_object_name template: %s", err)
	}

	rawImageGcsPath, err := UploadToBucket(opts, ui, artifact, p.config.Bucket, p.config.GCSObjectName)
	if err != nil {
		return nil, false, false, err
	}

	gceImageArtifact, err := CreateGceImage(opts, ui, p.config.ProjectId, rawImageGcsPath, p.config.ImageName, p.config.ImageDescription, p.config.ImageFamily, p.config.ImageLabels, p.config.ImageGuestOsFeatures)
	if err != nil {
		return nil, false, false, err
	}

	if !p.config.SkipClean {
		err = DeleteFromBucket(opts, ui, p.config.Bucket, p.config.GCSObjectName)
		if err != nil {
			return nil, false, false, err
		}
	}

	return gceImageArtifact, false, false, nil
}

func UploadToBucket(opts option.ClientOption, ui packer.Ui, artifact packer.Artifact, bucket string, gcsObjectName string) (string, error) {
	service, err := storage.NewService(context.TODO(), opts)
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

func CreateGceImage(opts option.ClientOption, ui packer.Ui, project string, rawImageURL string, imageName string, imageDescription string, imageFamily string, imageLabels map[string]string, imageGuestOsFeatures []string) (packer.Artifact, error) {
	service, err := compute.NewService(context.TODO(), opts)

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

func DeleteFromBucket(opts option.ClientOption, ui packer.Ui, bucket string, gcsObjectName string) error {
	service, err := storage.NewService(context.TODO(), opts)

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

//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package googlecomputeimport

import (
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/builder/googlecompute"
	"github.com/hashicorp/packer/post-processor/artifice"
	"github.com/hashicorp/packer/post-processor/compress"
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
	//[template engine](/docs/templates/legacy_json_templates/engine). Therefore, you
	//may use user variables and template functions in this field. Defaults to
	//`packer-import-{{timestamp}}.tar.gz`.
	GCSObjectName string `mapstructure:"gcs_object_name"`
	//The description of the resulting image.
	ImageDescription string `mapstructure:"image_description"`
	//The name of the image family to which the resulting image belongs.
	ImageFamily string `mapstructure:"image_family"`
	//A list of features to enable on the guest operating system. Applicable only for bootable images. Valid
	//values are `MULTI_IP_SUBNET`, `UEFI_COMPATIBLE`,
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
	//A key used to establish the trust relationship between the platform owner and the firmware. You may only specify one platform key, and it must be a valid X.509 certificate.
	ImagePlatformKey string `mapstructure:"image_platform_key"`
	//A key used to establish a trust relationship between the firmware and the OS. You may specify multiple comma-separated keys for this value.
	ImageKeyExchangeKey []string `mapstructure:"image_key_exchange_key"`
	//A database of certificates that have been revoked and will cause the system to stop booting if a boot file is signed with one of them. You may specify single or multiple comma-separated values for this value.
	ImageSignaturesDB []string `mapstructure:"image_signatures_db"`
	//A database of certificates that are trusted and can be used to sign boot files. You may specify single or multiple comma-separated values for this value.
	ImageForbiddenSignaturesDB []string `mapstructure:"image_forbidden_signatures_db"`

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

	errs := new(packersdk.MultiError)

	// Set defaults
	if p.config.GCSObjectName == "" {
		p.config.GCSObjectName = "packer-import-{{timestamp}}.tar.gz"
	}

	// Check and render gcs_object_name
	if err = interpolate.Validate(p.config.GCSObjectName, &p.config.ctx); err != nil {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("Error parsing gcs_object_name template: %s", err))
	}

	if p.config.AccountFile != "" {
		if p.config.VaultGCPOauthEngine != "" && p.config.ImpersonateServiceAccount != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("You cannot "+
				"specify impersonate_service_account, account_file and vault_gcp_oauth_engine at the same time"))
		}
		cfg, err := googlecompute.ProcessAccountFile(p.config.AccountFile)
		if err != nil {
			errs = packersdk.MultiErrorAppend(errs, err)
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
			errs = packersdk.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set", key))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
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

	shieldedVMStateConfig, err := CreateShieldedVMStateConfig(p.config.ImageGuestOsFeatures, p.config.ImagePlatformKey, p.config.ImageKeyExchangeKey, p.config.ImageSignaturesDB, p.config.ImageForbiddenSignaturesDB)
	if err != nil {
		return nil, false, false, err
	}

	gceImageArtifact, err := CreateGceImage(opts, ui, p.config.ProjectId, rawImageGcsPath, p.config.ImageName, p.config.ImageDescription, p.config.ImageFamily, p.config.ImageLabels, p.config.ImageGuestOsFeatures, shieldedVMStateConfig)
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

func FillFileContentBuffer(certOrKeyFile string) (*compute.FileContentBuffer, error) {
	data, err := ioutil.ReadFile(certOrKeyFile)
	if err != nil {
		err := fmt.Errorf("Unable to read Certificate or Key file %s", certOrKeyFile)
		return nil, err
	}
	shield := &compute.FileContentBuffer{
		Content:  base64.StdEncoding.EncodeToString(data),
		FileType: "X509",
	}
	block, _ := pem.Decode(data)

	if block == nil || block.Type != "CERTIFICATE" {
		_, err = x509.ParseCertificate(data)
	} else {
		_, err = x509.ParseCertificate(block.Bytes)
	}
	if err != nil {
		shield.FileType = "BIN"
	}
	return shield, nil

}

func CreateShieldedVMStateConfig(imageGuestOsFeatures []string, imagePlatformKey string, imageKeyExchangeKey []string, imageSignaturesDB []string, imageForbiddenSignaturesDB []string) (*compute.InitialStateConfig, error) {
	shieldedVMStateConfig := &compute.InitialStateConfig{}
	for _, v := range imageGuestOsFeatures {
		if v == "UEFI_COMPATIBLE" {
			if imagePlatformKey != "" {
				shieldedData, err := FillFileContentBuffer(imagePlatformKey)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Pk = shieldedData
			}
			for _, v := range imageKeyExchangeKey {
				shieldedData, err := FillFileContentBuffer(v)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Keks = append(shieldedVMStateConfig.Keks, shieldedData)
			}
			for _, v := range imageSignaturesDB {
				shieldedData, err := FillFileContentBuffer(v)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Dbs = append(shieldedVMStateConfig.Dbs, shieldedData)
			}
			for _, v := range imageForbiddenSignaturesDB {
				shieldedData, err := FillFileContentBuffer(v)
				if err != nil {
					return nil, err
				}
				shieldedVMStateConfig.Dbxs = append(shieldedVMStateConfig.Dbxs, shieldedData)
			}

		}
	}
	return shieldedVMStateConfig, nil
}

func UploadToBucket(opts option.ClientOption, ui packersdk.Ui, artifact packersdk.Artifact, bucket string, gcsObjectName string) (string, error) {
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

func CreateGceImage(opts option.ClientOption, ui packersdk.Ui, project string, rawImageURL string, imageName string, imageDescription string, imageFamily string, imageLabels map[string]string, imageGuestOsFeatures []string, shieldedVMStateConfig *compute.InitialStateConfig) (packersdk.Artifact, error) {
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
		Description:                  imageDescription,
		Family:                       imageFamily,
		GuestOsFeatures:              imageFeatures,
		Labels:                       imageLabels,
		Name:                         imageName,
		RawDisk:                      &compute.ImageRawDisk{Source: rawImageURL},
		SourceType:                   "RAW",
		ShieldedInstanceInitialState: shieldedVMStateConfig,
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

func DeleteFromBucket(opts option.ClientOption, ui packersdk.Ui, bucket string, gcsObjectName string) error {
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

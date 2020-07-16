//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package yandeximport

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
	"github.com/yandex-cloud/go-sdk/iamkey"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/post-processor/artifice"
	"github.com/hashicorp/packer/post-processor/compress"
	yandexexport "github.com/hashicorp/packer/post-processor/yandex-export"
	"github.com/hashicorp/packer/template/interpolate"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// The folder ID that will be used to store imported Image.
	FolderID string `mapstructure:"folder_id" required:"true"`
	// Service Account ID with proper permission to use Storage service
	// for operations 'upload' and 'delete' object to `bucket`
	ServiceAccountID string `mapstructure:"service_account_id" required:"true"`

	// OAuth token to use to authenticate to Yandex.Cloud.
	Token string `mapstructure:"token" required:"false"`
	// Path to file with Service Account key in json format. This
	// is an alternative method to authenticate to Yandex.Cloud.
	ServiceAccountKeyFile string `mapstructure:"service_account_key_file" required:"false"`

	// The name of the bucket where the qcow2 file will be uploaded to for import.
	// This bucket must exist when the post-processor is run.
	//
	// If import occurred after Yandex-Export post-processor, artifact already
	// in storage service and first paths (URL) is used to, so no need to set this param.
	Bucket string `mapstructure:"bucket" required:"false"`
	// The name of the object key in `bucket` where the qcow2 file will be copied to import.
	// This is a [template engine](/docs/templates/engine).
	// Therefore, you may use user variables and template functions in this field.
	ObjectName string `mapstructure:"object_name" required:"false"`
	// Whether skip removing the qcow2 file uploaded to Storage
	// after the import process has completed. Possible values are: `true` to
	// leave it in the bucket, `false` to remove it. (Default: `false`).
	SkipClean bool `mapstructure:"skip_clean" required:"false"`

	// The name of the image, which contains 1-63 characters and only
	// supports lowercase English characters, numbers and hyphen.
	ImageName string `mapstructure:"image_name" required:"false"`
	// The description of the image.
	ImageDescription string `mapstructure:"image_description" required:"false"`
	// The family name of the imported image.
	ImageFamily string `mapstructure:"image_family" required:"false"`
	// Key/value pair labels to apply to the imported image.
	ImageLabels map[string]string `mapstructure:"image_labels" required:"false"`

	ctx interpolate.Context
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
				"object_name",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	errs := new(packer.MultiError)

	// provision config by OS environment variables
	if p.config.Token == "" {
		p.config.Token = os.Getenv("YC_TOKEN")
	}

	if p.config.ServiceAccountKeyFile == "" {
		p.config.ServiceAccountKeyFile = os.Getenv("YC_SERVICE_ACCOUNT_KEY_FILE")
	}

	if p.config.Token != "" {
		packer.LogSecretFilter.Set(p.config.Token)
	}

	if p.config.ServiceAccountKeyFile != "" {
		if _, err := iamkey.ReadFromJSONFile(p.config.ServiceAccountKeyFile); err != nil {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("fail to read service account key file: %s", err))
		}
	}

	if p.config.FolderID == "" {
		p.config.FolderID = os.Getenv("YC_FOLDER_ID")
	}

	// Set defaults
	if p.config.ObjectName == "" {
		p.config.ObjectName = "packer-import-{{timestamp}}.qcow2"
	}

	// Check and render object_name
	if err = interpolate.Validate(p.config.ObjectName, &p.config.ctx); err != nil {
		errs = packer.MultiErrorAppend(
			errs, fmt.Errorf("error parsing object_name template: %s", err))
	}

	// TODO: make common code to check and prepare Yandex.Cloud auth configuration data

	templates := map[string]*string{
		"object_name": &p.config.ObjectName,
		"folder_id":   &p.config.FolderID,
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

	cfg := &yandex.Config{
		Token:                 p.config.Token,
		ServiceAccountKeyFile: p.config.ServiceAccountKeyFile,
	}

	client, err := yandex.NewDriverYC(ui, cfg)
	if err != nil {
		return nil, false, false, err
	}

	p.config.ObjectName, err = interpolate.Render(p.config.ObjectName, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("error rendering object_name template: %s", err)
	}

	var url string

	// Create temporary storage Access Key
	respWithKey, err := client.SDK().IAM().AWSCompatibility().AccessKey().Create(ctx, &awscompatibility.CreateAccessKeyRequest{
		ServiceAccountId: p.config.ServiceAccountID,
		Description:      "this temporary key is for upload image to storage; created by Packer",
	})
	if err != nil {
		return nil, false, false, err
	}

	storageClient, err := newYCStorageClient("", respWithKey.GetAccessKey().GetKeyId(), respWithKey.GetSecret())
	if err != nil {
		return nil, false, false, fmt.Errorf("error create object storage client: %s", err)
	}

	switch artifact.BuilderId() {
	case compress.BuilderId, artifice.BuilderId, file.BuilderId:
		// Artifact as a file, need to be uploaded to storage before create Compute Image

		// As `bucket` option validate input here
		if p.config.Bucket == "" {
			return nil, false, false, fmt.Errorf("To upload artfact you need to specify `bucket` value")
		}

		url, err = uploadToBucket(storageClient, ui, artifact, p.config.Bucket, p.config.ObjectName)
		if err != nil {
			return nil, false, false, err
		}

	case yandexexport.BuilderId:
		// Artifact already in storage, just get URL
		url = artifact.Id()

	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from Yandex-Export, Compress, Artifice and File post-processor artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	presignedUrl, err := presignUrl(storageClient, ui, url)
	if err != nil {
		return nil, false, false, err
	}

	ycImage, err := createYCImage(ctx, client, ui, p.config.FolderID, presignedUrl, p.config.ImageName, p.config.ImageDescription, p.config.ImageFamily, p.config.ImageLabels)
	if err != nil {
		return nil, false, false, err
	}

	if !p.config.SkipClean {
		err = deleteFromBucket(storageClient, ui, url)
		if err != nil {
			return nil, false, false, err
		}
	}

	// Delete temporary storage Access Key
	_, err = client.SDK().IAM().AWSCompatibility().AccessKey().Delete(ctx, &awscompatibility.DeleteAccessKeyRequest{
		AccessKeyId: respWithKey.GetAccessKey().GetId(),
	})
	if err != nil {
		return nil, false, false, fmt.Errorf("error delete static access key: %s", err)
	}

	return ycImage, false, false, nil
}

func uploadToBucket(s3conn *s3.S3, ui packer.Ui, artifact packer.Artifact, bucket string, objectName string) (string, error) {
	ui.Say("Looking for qcow2 file in list of artifacts...")
	source := ""
	for _, path := range artifact.Files() {
		ui.Say(fmt.Sprintf("Found artifact %v...", path))
		if strings.HasSuffix(path, ".qcow2") {
			source = path
			break
		}
	}

	if source == "" {
		return "", fmt.Errorf("no qcow2 file found in list of artifacts")
	}

	artifactFile, err := os.Open(source)
	if err != nil {
		err := fmt.Errorf("error opening %v", source)
		return "", err
	}

	ui.Say(fmt.Sprintf("Uploading file %v to bucket %v/%v...", source, bucket, objectName))

	_, err = s3conn.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
		Body:   artifactFile,
	})

	if err != nil {
		ui.Say(fmt.Sprintf("Failed to upload: %v", objectName))
		return "", err
	}

	req, _ := s3conn.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
	})

	// Compute service allow only `https://storage.yandexcloud.net/...` URLs for Image create process
	req.Config.S3ForcePathStyle = aws.Bool(true)

	return req.HTTPRequest.URL.String(), nil
}

func createYCImage(ctx context.Context, driver yandex.Driver, ui packer.Ui, folderID string, rawImageURL string, imageName string, imageDescription string, imageFamily string, imageLabels map[string]string) (packer.Artifact, error) {
	op, err := driver.SDK().WrapOperation(driver.SDK().Compute().Image().Create(ctx, &compute.CreateImageRequest{
		FolderId:    folderID,
		Name:        imageName,
		Description: imageDescription,
		Labels:      imageLabels,
		Family:      imageFamily,
		Source:      &compute.CreateImageRequest_Uri{Uri: rawImageURL},
	}))
	if err != nil {
		ui.Say("Error creating Yandex Compute Image")
		return nil, err
	}

	ui.Say(fmt.Sprintf("Source url for Image creation: %v", rawImageURL))

	ui.Say(fmt.Sprintf("Creating Yandex Compute Image %v within operation %#v", imageName, op.Id()))

	ui.Say("Waiting for Yandex Compute Image creation operation to complete...")
	err = op.Wait(ctx)

	// fail if image creation operation has an error
	if err != nil {
		return nil, fmt.Errorf("failed to create Yandex Compute Image: %s", err)
	}

	protoMetadata, err := op.Metadata()
	if err != nil {
		return nil, fmt.Errorf("error while get image create operation metadata: %s", err)
	}

	md, ok := protoMetadata.(*compute.CreateImageMetadata)
	if !ok {
		return nil, fmt.Errorf("could not get Image ID from create operation metadata")
	}

	image, err := driver.SDK().Compute().Image().Get(ctx, &compute.GetImageRequest{
		ImageId: md.ImageId,
	})
	if err != nil {
		return nil, fmt.Errorf("error while image get request: %s", err)
	}

	return &yandex.Artifact{
		Image: image,
	}, nil
}

func deleteFromBucket(s3conn *s3.S3, ui packer.Ui, url string) error {
	bucket, objectName, err := s3URLToBucketKey(url)
	if err != nil {
		return err
	}

	ui.Say(fmt.Sprintf("Deleting import source from Object Storage %s/%s...", bucket, objectName))

	_, err = s3conn.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
	})
	if err != nil {
		ui.Say(fmt.Sprintf("Failed to delete: %v/%v", bucket, objectName))
		return fmt.Errorf("error deleting storage object %q in bucket %q: %s ", objectName, bucket, err)
	}

	return nil
}

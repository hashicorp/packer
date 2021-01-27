//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config

package yandeximport

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/hashicorp/packer/builder/file"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/hashicorp/packer/post-processor/artifice"
	"github.com/hashicorp/packer/post-processor/compress"
	yandexexport "github.com/hashicorp/packer/post-processor/yandex-export"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

type Config struct {
	common.PackerConfig         `mapstructure:",squash"`
	yandex.AccessConfig         `mapstructure:",squash"`
	yandex.CloudConfig          `mapstructure:",squash"`
	yandexexport.ExchangeConfig `mapstructure:",squash"`
	yandex.ImageConfig          `mapstructure:",squash"`

	// The name of the bucket where the qcow2 file will be uploaded to for import.
	// This bucket must exist when the post-processor is run.
	//
	// If import occurred after Yandex-Export post-processor, artifact already
	// in storage service and first paths (URL) is used to, so no need to set this param.
	Bucket string `mapstructure:"bucket" required:"false"`
	// The name of the object key in `bucket` where the qcow2 file will be copied to import.
	// This is a [template engine](/docs/templates/legacy_json_templates/engine).
	// Therefore, you may use user variables and template functions in this field.
	ObjectName string `mapstructure:"object_name" required:"false"`
	// Whether skip removing the qcow2 file uploaded to Storage
	// after the import process has completed. Possible values are: `true` to
	// leave it in the bucket, `false` to remove it. Default is `false`.
	SkipClean bool `mapstructure:"skip_clean" required:"false"`

	ctx interpolate.Context
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
				"object_name",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Accumulate any errors
	var errs *packersdk.MultiError

	errs = packersdk.MultiErrorAppend(errs, p.config.AccessConfig.Prepare(&p.config.ctx)...)
	errs = p.config.CloudConfig.Prepare(errs)
	errs = p.config.ImageConfig.Prepare(errs)
	errs = p.config.ExchangeConfig.Prepare(errs)

	// Set defaults
	if p.config.ObjectName == "" {
		p.config.ObjectName = "packer-import-{{timestamp}}.qcow2"
	}

	// Check and render object_name
	if err = interpolate.Validate(p.config.ObjectName, &p.config.ctx); err != nil {
		errs = packersdk.MultiErrorAppend(
			errs, fmt.Errorf("error parsing object_name template: %s", err))
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, artifact packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	var imageSrc cloudImageSource
	var fileSource bool
	var err error

	generatedData := artifact.State("generated_data")
	if generatedData == nil {
		// Make sure it's not a nil map so we can assign to it later.
		generatedData = make(map[string]interface{})
	}
	p.config.ctx.Data = generatedData

	p.config.ObjectName, err = interpolate.Render(p.config.ObjectName, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("error rendering object_name template: %s", err)
	}

	client, err := yandex.NewDriverYC(ui, &p.config.AccessConfig)

	if err != nil {
		return nil, false, false, err
	}

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
		fileSource = true

		// As `bucket` option validate input here
		if p.config.Bucket == "" {
			return nil, false, false, fmt.Errorf("To upload artfact you need to specify `bucket` value")
		}

		imageSrc, err = uploadToBucket(storageClient, ui, artifact, p.config.Bucket, p.config.ObjectName)
		if err != nil {
			return nil, false, false, err
		}

	case yandexexport.BuilderId:
		// Artifact already in storage, just get URL
		imageSrc, err = presignUrl(storageClient, ui, artifact.Id())
		if err != nil {
			return nil, false, false, err
		}

	case yandex.BuilderID:
		// Artifact is plain Yandex Compute Image, just create new one based on provided
		imageSrc = &imageSource{
			imageID: artifact.Id(),
		}
	case BuilderId:
		// Artifact from prev yandex-import PP, reuse URL or Cloud Image ID
		imageSrc, err = chooseSource(artifact)
		if err != nil {
			return nil, false, false, err
		}

	default:
		err := fmt.Errorf(
			"Unknown artifact type: %s\nCan only import from Yandex-Export, Yandex-Import, Compress, Artifice and File post-processor artifacts.",
			artifact.BuilderId())
		return nil, false, false, err
	}

	ycImage, err := createYCImage(ctx, client, ui, imageSrc, &p.config)
	if err != nil {
		return nil, false, false, err
	}

	if fileSource && !p.config.SkipClean {
		err = deleteFromBucket(storageClient, ui, imageSrc)
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

	return &Artifact{
		imageID: ycImage.GetId(),
		StateData: map[string]interface{}{
			"source_type": imageSrc.GetSourceType(),
			"source_id":   imageSrc.GetSourceID(),
		},
	}, false, false, nil
}

func chooseSource(a packersdk.Artifact) (cloudImageSource, error) {
	st := a.State("source_type").(string)
	if st == "" {
		return nil, fmt.Errorf("could not determine source type of yandex-import artifact: %v", a)
	}
	switch st {
	case sourceType_IMAGE:
		return &imageSource{
			imageID: a.State("source_id").(string),
		}, nil

	case sourceType_OBJECT:
		return &objectSource{
			url: a.State("source_id").(string),
		}, nil
	}
	return nil, fmt.Errorf("unknow source type of yandex-import artifact: %s", st)
}

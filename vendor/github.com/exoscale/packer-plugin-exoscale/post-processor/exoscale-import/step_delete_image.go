package exoscaleimport

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepDeleteImage struct{}

func (s *stepDeleteImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		ui       = state.Get("ui").(packer.Ui)
		config   = state.Get("config").(*Config)
		sos      = state.Get("sos").(*s3.Client)
		artifact = state.Get("artifact").(packer.Artifact)

		imageFile  = artifact.Files()[0]
		bucketFile = filepath.Base(imageFile)
	)

	if config.SkipClean {
		return multistep.ActionContinue
	}

	ui.Say("Deleting uploaded template image")

	if _, err := sos.DeleteObject(ctx,
		&s3.DeleteObjectInput{
			Bucket: aws.String(config.ImageBucket),
			Key:    aws.String(bucketFile),
		}); err != nil {
		ui.Error(fmt.Sprintf("unable to delete template image: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDeleteImage) Cleanup(state multistep.StateBag) {}

package exoscaleimport

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepDeleteImage struct{}

func (s *stepDeleteImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		ui       = state.Get("ui").(packer.Ui)
		config   = state.Get("config").(*Config)
		artifact = state.Get("artifact").(packer.Artifact)

		imageFile  = artifact.Files()[0]
		bucketFile = filepath.Base(imageFile)
	)

	if config.SkipClean {
		return multistep.ActionContinue
	}

	ui.Say("Deleting uploaded template image")

	sess, err := session.NewSessionWithOptions(session.Options{Config: aws.Config{
		Region:      aws.String(config.TemplateZone),
		Endpoint:    aws.String(config.SOSEndpoint),
		Credentials: credentials.NewStaticCredentials(config.APIKey, config.APISecret, "")}})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to initialize session: %v", err))
		return multistep.ActionHalt
	}

	svc := s3.New(sess)
	if _, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.ImageBucket),
		Key:    aws.String(bucketFile),
	}); err != nil {
		ui.Error(fmt.Sprintf("unable to delete template image: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *stepDeleteImage) Cleanup(state multistep.StateBag) {}

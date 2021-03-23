package exoscaleimport

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

type stepUploadImage struct{}

func (s *stepUploadImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	var (
		ui       = state.Get("ui").(packer.Ui)
		config   = state.Get("config").(*Config)
		artifact = state.Get("artifact").(packer.Artifact)

		imageFile  = artifact.Files()[0]
		bucketFile = filepath.Base(imageFile)
	)

	ui.Say("Uploading template image")

	f, err := os.Open(imageFile)
	if err != nil {
		ui.Error(fmt.Sprint(err))
		return multistep.ActionHalt
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		ui.Error(fmt.Sprint(err))
		return multistep.ActionHalt
	}

	// For tracking image file upload progress
	pf := ui.TrackProgress(imageFile, 0, fileInfo.Size(), f)
	defer pf.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, f); err != nil {
		ui.Error(fmt.Sprintf("unable to compute template file checksum: %v", err))
		return multistep.ActionHalt
	}
	if _, err := f.Seek(0, 0); err != nil {
		ui.Error(fmt.Sprintf("unable to compute template file checksum: %v", err))
		return multistep.ActionHalt
	}

	sess, err := session.NewSessionWithOptions(session.Options{Config: aws.Config{
		Region:      aws.String(config.TemplateZone),
		Endpoint:    aws.String(config.SOSEndpoint),
		Credentials: credentials.NewStaticCredentials(config.APIKey, config.APISecret, "")}})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to initialize session: %v", err))
		return multistep.ActionHalt
	}

	uploader := s3manager.NewUploader(sess)
	output, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Body:       pf,
		Bucket:     aws.String(config.ImageBucket),
		Key:        aws.String(bucketFile),
		ContentMD5: aws.String(base64.StdEncoding.EncodeToString(hash.Sum(nil))),
		ACL:        aws.String("public-read"),
	})
	if err != nil {
		ui.Error(fmt.Sprintf("unable to upload template image: %v", err))
		return multistep.ActionHalt
	}

	state.Put("image_url", output.Location)
	state.Put("image_checksum", fmt.Sprintf("%x", hash.Sum(nil)))

	return multistep.ActionContinue
}

func (s *stepUploadImage) Cleanup(state multistep.StateBag) {}

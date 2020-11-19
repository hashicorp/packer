package yandeximport

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/hashicorp/packer/packer"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
)

func uploadToBucket(s3conn *s3.S3, ui packersdk.Ui, artifact packer.Artifact, bucket string, objectName string) (cloudImageSource, error) {
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
		return nil, fmt.Errorf("no qcow2 file found in list of artifacts")
	}

	artifactFile, err := os.Open(source)
	if err != nil {
		err := fmt.Errorf("error opening %v", source)
		return nil, err
	}

	ui.Say(fmt.Sprintf("Uploading file %v to bucket %v/%v...", source, bucket, objectName))

	_, err = s3conn.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
		Body:   artifactFile,
	})

	if err != nil {
		ui.Say(fmt.Sprintf("Failed to upload: %v", objectName))
		return nil, err
	}

	req, _ := s3conn.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectName),
	})

	// Compute service allow only `https://storage.yandexcloud.net/...` URLs for Image create process
	req.Config.S3ForcePathStyle = aws.Bool(true)

	err = req.Build()
	if err != nil {
		ui.Say(fmt.Sprintf("Failed to build S3 request: %v", err))
		return nil, err
	}

	return &objectSource{
		url: req.HTTPRequest.URL.String(),
	}, nil
}

func createYCImage(ctx context.Context, driver yandex.Driver, ui packersdk.Ui, folderID string, imageSrc cloudImageSource, imageName string, imageDescription string, imageFamily string, imageLabels map[string]string) (*compute.Image, error) {
	req := &compute.CreateImageRequest{
		FolderId:    folderID,
		Name:        imageName,
		Description: imageDescription,
		Labels:      imageLabels,
		Family:      imageFamily,
	}

	// switch on cloudImageSource type: cloud image id or storage URL
	switch v := imageSrc.(type) {
	case *imageSource:
		req.Source = &compute.CreateImageRequest_ImageId{ImageId: v.imageID}
	case *objectSource:
		req.Source = &compute.CreateImageRequest_Uri{Uri: v.url}
	}

	op, err := driver.SDK().WrapOperation(driver.SDK().Compute().Image().Create(ctx, req))
	if err != nil {
		ui.Say("Error creating Yandex Compute Image")
		return nil, err
	}

	ui.Say(fmt.Sprintf("Source of Image creation: %s", imageSrc.Description()))

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

	return image, nil

}

func deleteFromBucket(s3conn *s3.S3, ui packersdk.Ui, imageSrc cloudImageSource) error {
	var url string
	// switch on cloudImageSource type: cloud image id or storage URL
	switch v := imageSrc.(type) {
	case *objectSource:
		url = v.GetSourceID()
	case *imageSource:
		return fmt.Errorf("invalid argument for `deleteFromBucket` method: %v", v)
	}

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

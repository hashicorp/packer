package yandexexport

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/builder/yandex"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1/awscompatibility"
)

type StepCreateS3Keys struct {
	ServiceAccountID string
	Paths            []string
}

func (c *StepCreateS3Keys) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(yandex.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	ui.Say("Create temporary storage Access Key")
	// Create temporary storage Access Key
	respWithKey, err := driver.SDK().IAM().AWSCompatibility().AccessKey().Create(ctx, &awscompatibility.CreateAccessKeyRequest{
		ServiceAccountId: c.ServiceAccountID,
		Description:      "this temporary key is for upload image to storage; created by Packer",
	})
	if err != nil {
		err := fmt.Errorf("Error waiting for cloud-init script to finish: %s", err)
		return yandex.StepHaltWithError(state, err)
	}
	state.Put("s3_secret", respWithKey)

	ui.Say("Verify access to paths")
	if err := verfiyAccess(respWithKey.GetAccessKey().GetKeyId(), respWithKey.Secret, c.Paths); err != nil {
		return yandex.StepHaltWithError(state, err)
	}
	return multistep.ActionContinue
}

func (s *StepCreateS3Keys) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(yandex.Driver)
	ui := state.Get("ui").(packersdk.Ui)

	if val, ok := state.GetOk("s3_secret"); ok {
		ui.Say("S3 secrets have been found")
		s3Secret := val.(*awscompatibility.CreateAccessKeyResponse)

		ui.Message("Cleanup empty objects...")
		cleanUpEmptyObjects(s3Secret.GetAccessKey().GetKeyId(), s3Secret.GetSecret(), s.Paths)

		ui.Say("Delete S3 secrets...")
		_, err := driver.SDK().IAM().AWSCompatibility().AccessKey().Delete(context.Background(), &awscompatibility.DeleteAccessKeyRequest{
			AccessKeyId: s3Secret.GetAccessKey().GetId(),
		})

		if err != nil {
			ui.Error(err.Error())
		}
	}
}

func verfiyAccess(keyID, secret string, paths []string) error {
	newSession, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(defaultStorageEndpoint),
		Region:   aws.String(defaultStorageRegion),
		Credentials: credentials.NewStaticCredentials(
			keyID, secret, "",
		),
	})
	if err != nil {
		return err
	}

	s3Conn := s3.New(newSession)

	for _, path := range paths {
		u, err := url.Parse(path)
		if err != nil {
			return err
		}
		key := u.Path
		if strings.HasSuffix(key, "/") {
			key = filepath.Join(key, "disk.qcow2")
		}
		_, err = s3Conn.PutObject(&s3.PutObjectInput{
			Body:   aws.ReadSeekCloser(strings.NewReader("")),
			Bucket: aws.String(u.Host),
			Key:    aws.String(key),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanUpEmptyObjects(keyID, secret string, paths []string) {
	newSession, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(defaultStorageEndpoint),
		Region:   aws.String(defaultStorageRegion),
		Credentials: credentials.NewStaticCredentials(
			keyID, secret, "",
		),
	})
	if err != nil {
		log.Printf("[WARN] %s", err)
		return
	}
	s3Conn := s3.New(newSession)

	for _, path := range paths {
		u, err := url.Parse(path)
		if err != nil {
			log.Printf("[WARN] %s", err)
			continue
		}
		key := u.Path
		if strings.HasSuffix(key, "/") {
			key = filepath.Join(key, "disk.qcow2")
		}

		log.Printf("Check object: '%s'", path)
		respHead, err := s3Conn.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(u.Host),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("[WARN] %s", err)
			continue
		}
		if *respHead.ContentLength > 0 {
			continue
		}

		log.Printf("[DEBUG] Delete object: '%s'", path)
		_, err = s3Conn.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(u.Host),
			Key:    aws.String(key),
		})
		if err != nil {
			log.Printf("[WARN] %s", err)
		}
	}
}

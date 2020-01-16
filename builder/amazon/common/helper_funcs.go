package common

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/common/retry"
)

// DestroyAMIs deregisters the AWS machine images in imageids from an active AWS account
func DestroyAMIs(imageids []*string, ec2conn *ec2.EC2) error {
	resp, err := ec2conn.DescribeImages(&ec2.DescribeImagesInput{
		ImageIds: imageids,
	})

	if err != nil {
		err := fmt.Errorf("Error describing AMI: %s", err)
		return err
	}

	// Deregister image by name.
	for _, i := range resp.Images {

		ctx := context.TODO()
		err = retry.Config{
			Tries: 11,
			ShouldRetry: func(err error) bool {
				return isAWSErr(err, "UnauthorizedOperation", "")
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			_, err := ec2conn.DeregisterImage(&ec2.DeregisterImageInput{
				ImageId: i.ImageId,
			})
			return err
		})

		if err != nil {
			err := fmt.Errorf("Error deregistering existing AMI: %s", err)
			return err
		}
		log.Printf("Deregistered AMI id: %s", *i.ImageId)

		// Delete snapshot(s) by image
		for _, b := range i.BlockDeviceMappings {
			if b.Ebs != nil && aws.StringValue(b.Ebs.SnapshotId) != "" {

				err = retry.Config{
					Tries: 11,
					ShouldRetry: func(err error) bool {
						return isAWSErr(err, "UnauthorizedOperation", "")
					},
					RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
				}.Run(ctx, func(ctx context.Context) error {
					_, err := ec2conn.DeleteSnapshot(&ec2.DeleteSnapshotInput{
						SnapshotId: b.Ebs.SnapshotId,
					})
					return err
				})

				if err != nil {
					err := fmt.Errorf("Error deleting existing snapshot: %s", err)
					return err
				}
				log.Printf("Deleted snapshot: %s", *b.Ebs.SnapshotId)
			}
		}
	}
	return nil
}

// Returns true if the error matches all these conditions:
//  * err is of type awserr.Error
//  * Error.Code() matches code
//  * Error.Message() contains message
func isAWSErr(err error, code string, message string) bool {
	if err, ok := err.(awserr.Error); ok {
		return err.Code() == code && strings.Contains(err.Message(), message)
	}
	return false
}

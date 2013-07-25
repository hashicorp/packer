package common

import (
	"github.com/mitchellh/goamz/ec2"
	"log"
	"time"
)

// WaitForAMI waits for the given AMI ID to become ready.
func WaitForAMI(c *ec2.EC2, imageId string) error {
	for {
		imageResp, err := c.Images([]string{imageId}, ec2.NewFilter())
		if err != nil {
			return err
		}

		if imageResp.Images[0].State == "available" {
			return nil
		}

		log.Printf("Image in state %s, sleeping 2s before checking again",
			imageResp.Images[0].State)
		time.Sleep(2 * time.Second)
	}
}

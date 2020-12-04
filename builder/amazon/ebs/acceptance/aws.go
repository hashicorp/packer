package amazon_acc

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awscommon "github.com/hashicorp/packer/builder/amazon/common"
)

type AWSHelper struct {
	Region  string
	AMIName string
}

func (a *AWSHelper) CleanUpAmi() error {
	accessConfig := &awscommon.AccessConfig{}
	session, err := accessConfig.Session()
	if err != nil {
		return fmt.Errorf("AWSAMICleanUp: Unable to create aws session %s", err.Error())
	}

	regionconn := ec2.New(session.Copy(&aws.Config{
		Region: aws.String(a.Region),
	}))

	resp, err := regionconn.DescribeImages(&ec2.DescribeImagesInput{
		Owners: aws.StringSlice([]string{"self"}),
		Filters: []*ec2.Filter{{
			Name:   aws.String("name"),
			Values: aws.StringSlice([]string{a.AMIName}),
		}}})
	if err != nil {
		return fmt.Errorf("AWSAMICleanUp: Unable to find Image %s: %s", a.AMIName, err.Error())
	}

	_, err = regionconn.DeregisterImage(&ec2.DeregisterImageInput{
		ImageId: resp.Images[0].ImageId,
	})
	if err != nil {
		return fmt.Errorf("AWSAMICleanUp: Unable to Deregister Image %s", err.Error())
	}
	return nil
}

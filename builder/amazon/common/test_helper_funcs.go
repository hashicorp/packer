package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

type mockEC2Client struct {
	ec2iface.EC2API
}

func FakeAccessConfig() *AccessConfig {
	accessConfig := AccessConfig{
		getEC2Connection: func() ec2iface.EC2API {
			return &mockEC2Client{}
		},
		PollingConfig: new(AWSPollingConfig),
	}
	accessConfig.session = session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-1"),
	}))

	return &accessConfig
}

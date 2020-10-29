package common

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

const (
	sessionManagerPluginName string = "session-manager-plugin"

	//sessionCommand is the AWS-SDK equivalent to the command you would specify to `aws ssm ...`
	sessionCommand string = "StartSession"
)

type SSMDriverConfig struct {
	SvcClient   ssmiface.SSMAPI
	Region      string
	ProfileName string
	SvcEndpoint string
}

type SSMDriver struct {
	SSMDriverConfig
	session       *ssm.StartSessionOutput
	sessionParams ssm.StartSessionInput
	pluginCmdFunc func(context.Context) error
}

func NewSSMDriver(config SSMDriverConfig) *SSMDriver {
	d := SSMDriver{SSMDriverConfig: config}
	return &d
}

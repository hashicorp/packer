package common

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

type MockSSMSvc struct {
	ssmiface.SSMAPI
	StartSessionError      error
	TerminateSessionError  error
	StartSessionCalled     bool
	TerminateSessionCalled bool
}

func (svc *MockSSMSvc) StartSessionWithContext(ctx aws.Context, input *ssm.StartSessionInput, options ...request.Option) (*ssm.StartSessionOutput, error) {
	svc.StartSessionCalled = true
	return MockStartSessionOutput(), svc.StartSessionError
}

func (svc *MockSSMSvc) TerminateSession(input *ssm.TerminateSessionInput) (*ssm.TerminateSessionOutput, error) {
	svc.TerminateSessionCalled = true
	return new(ssm.TerminateSessionOutput), svc.TerminateSessionError
}

func MockPluginCmdFunc(ctx context.Context) error {
	return nil
}

func MockStartSessionOutput() *ssm.StartSessionOutput {
	id, url, token := "packerid", "http://packer.io", "packer-token"
	output := ssm.StartSessionOutput{
		SessionId:  &id,
		StreamUrl:  &url,
		TokenValue: &token,
	}
	return &output
}

func MockStartSessionInput(instance string) ssm.StartSessionInput {
	params := map[string][]*string{
		"portNumber":      []*string{aws.String("22")},
		"localPortNumber": []*string{aws.String("8001")},
	}

	input := ssm.StartSessionInput{
		DocumentName: aws.String("AWS-StartPortForwardingSession"),
		Parameters:   params,
		Target:       aws.String(instance),
	}

	return input
}

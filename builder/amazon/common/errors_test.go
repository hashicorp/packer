package common

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sts"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

type mockSTS struct {
}

func (m *mockSTS) DecodeAuthorizationMessage(input *sts.DecodeAuthorizationMessageInput) (*sts.DecodeAuthorizationMessageOutput, error) {
	return &sts.DecodeAuthorizationMessageOutput{
		DecodedMessage: aws.String(`{
			"allowed": false,
			"explicitDeny": true,
			"matchedStatements": {}
		}`),
	}, nil
}

func TestErrorsParsing_RequestFailure(t *testing.T) {

	ae := awserr.New("UnauthorizedOperation",
		`You are not authorized to perform this operation. Encoded authorization failure message: D9Q7oicjOMr9l2CC-NPP1FiZXK9Ijia1k-3l0siBFCcrK3oSuMFMkBIO5TNj0HdXE-WfwnAcdycFOohfKroNO6toPJEns8RFVfy_M_IjNGmrEFJ6E62pnmBW0OLrMsXxR9FQE4gB4gJzSM0AD6cV6S3FOfqYzWBRX-sQdOT4HryGkFNRoFBr9Xbp-tRwiadwkbdHdfnV9fbRkXmnwCdULml16NBSofC4ZPepLMKmIB5rKjwk-m179UUh2XA-J5no0si6XcRo5GbHQB5QfCIwSHL4vsro2wLZUd16-8OWKyr3tVlTbQe0ERZskqRqRQ5E28QuiBCVV6XstUyo-T4lBSr75Fgnyr3wCO-dS3b_5Ns3WzA2JD4E2AJOAStXIU8IH5YuKkAg7C-dJMuBMPpmKCBEXhNoHDwCyOo5PsV3xMlc0jSb0qYGpfst_TDDtejcZfn7NssUjxVq9qkdH-OXz2gPoQB-hX8ycmZCL5UZwKc3TCLUr7TGnudHjmnMrE9cUo-yTCWfyHPLprhiYhTCKW18EikJ0O1EKI3FJ_b4F19_jFBPARjSwQc7Ut6MNCVzrPdZGYSF6acj5gPaxdy9uSkVQwWXK7Pd5MFP7EBDE1_DgYbzodgwDO2PXeVFUbSLBHKWo_ebZS9ZX2nYPcGss_sYaly0ZVSIJXp7G58B5BoFVhvVH6jYnF9XiAOjMltuP_ycu1pQP1lki500RY3baLvfeYeAsB38XZHKEgWZzq7Fei-uh89q0cjJTmlVyrfRU3q6`,
		fmt.Errorf("You can't do it!!"))
	rf := awserr.NewRequestFailure(ae, 400, "abc-def-123-456")

	result := decodeAWSError(&mockSTS{}, rf)
	if result == nil {
		t.Error("Expected resulting error")
	}
	if !strings.Contains(result.Error(), "Authorization failure message:") {
		t.Error("Expected authorization failure message")
	}
}

func TestErrorsParsing_NonAuthorizationFailure(t *testing.T) {

	ae := awserr.New("BadRequest",
		`You did something wrong. Try again`,
		fmt.Errorf("Request was no good."))
	rf := awserr.NewRequestFailure(ae, 400, "abc-def-123-456")

	result := decodeAWSError(&mockSTS{}, rf)
	if result == nil {
		t.Error("Expected resulting error")
	}
	if result != rf {
		t.Error("Expected original error to be returned unchanged")
	}
}

func TestErrorsParsing_NonAWSError(t *testing.T) {

	err := fmt.Errorf("Random error occurred")

	result := decodeAWSError(&mockSTS{}, err)
	if result == nil {
		t.Error("Expected resulting error")
	}
	if result != err {
		t.Error("Expected original error to be returned unchanged")
	}
}

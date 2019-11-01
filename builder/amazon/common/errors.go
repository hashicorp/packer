package common

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

var encodedFailureMessagePattern = regexp.MustCompile(`(?i)(.*) Encoded authorization failure message: ([\w-]+) ?( .*)?`)

type stsDecoder interface {
	DecodeAuthorizationMessage(input *sts.DecodeAuthorizationMessageInput) (*sts.DecodeAuthorizationMessageOutput, error)
}

// decodeError replaces encoded authorization messages with the
// decoded results
func decodeAWSError(decoder stsDecoder, err error) error {

	groups := encodedFailureMessagePattern.FindStringSubmatch(err.Error())
	if len(groups) > 1 {
		result, decodeErr := decoder.DecodeAuthorizationMessage(&sts.DecodeAuthorizationMessageInput{
			EncodedMessage: aws.String(groups[2]),
		})
		if decodeErr == nil {
			msg := aws.StringValue(result.DecodedMessage)
			return fmt.Errorf("%s Authorization failure message: '%s'%s", groups[1], msg, groups[3])
		}
		log.Printf("[WARN] Attempted to decode authorization message, but received: %v", decodeErr)
	}
	return err
}

// DecodeAuthZMessages enables automatic decoding of any
// encoded authorization messages
func DecodeAuthZMessages(sess *session.Session) {
	azd := &authZMessageDecoder{
		Decoder: sts.New(sess),
	}
	sess.Handlers.UnmarshalError.AfterEachFn = azd.afterEachFn
}

type authZMessageDecoder struct {
	Decoder stsDecoder
}

func (a *authZMessageDecoder) afterEachFn(item request.HandlerListRunItem) bool {
	if err, ok := item.Request.Error.(awserr.Error); ok && err.Code() == "UnauthorizedOperation" {
		item.Request.Error = decodeAWSError(a.Decoder, err)
	}
	return true
}

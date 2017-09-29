package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/mitchellh/multistep"
)

var encodedFailureMessagePattern = regexp.MustCompile(`(?i).* Encoded authorization failure message: ([\w-]+)`)

// DecodeError replaces encoded authorization messages with the
// decoded results
func DecodeError(state multistep.StateBag, err error) error {

	if ae, ok := err.(awserr.Error); ok && ae.Code() == "UnauthorizedOperation" {
		parts := encodedFailureMessagePattern.FindStringSubmatch(ae.Message())
		if parts != nil && len(parts) > 1 {
			stsConn := state.Get("sts").(*sts.STS)
			result, decodeErr := stsConn.DecodeAuthorizationMessage(&sts.DecodeAuthorizationMessageInput{
				EncodedMessage: aws.String(parts[1]),
			})
			if decodeErr == nil {
				msg, ppErr := prettyPrint(aws.StringValue(result.DecodedMessage))
				if ppErr != nil {
					log.Printf("[WARN] Attempted to pretty print authorization message: %v", ppErr)
					msg = aws.StringValue(result.DecodedMessage)
				}

				return fmt.Errorf("You are not authorized to perform this operation. Authorization failure message: \n%s",
					msg)

			}
			log.Printf("[WARN] Attempted to decode authorization message, but received: %v", decodeErr)
		}
	}
	return err
}

func prettyPrint(str string) (string, error) {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(str), "", "  ")
	return out.String(), err
}

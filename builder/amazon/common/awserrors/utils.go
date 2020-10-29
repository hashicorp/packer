package awserrors

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// Returns true if the err matches all these conditions:
//  * err is of type awserr.Error
//  * Error.Code() matches code
//  * Error.Message() contains message
func Matches(err error, code string, message string) bool {
	if err, ok := err.(awserr.Error); ok {
		return err.Code() == code && strings.Contains(err.Message(), message)
	}
	return false
}

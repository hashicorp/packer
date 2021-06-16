package packer_registry

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
)

// Do we need a models.Error for the client to properly handle errors?
func checkErrorCode(err error, code codes.Code) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), fmt.Sprintf("Code:%d", code))

}

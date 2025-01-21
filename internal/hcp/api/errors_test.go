package api

import (
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
)

func TestCheckErrorCode(t *testing.T) {
	tests := []struct {
		name          string
		codeString    string
		expectCode    codes.Code
		expectSuccess bool
	}{
		{
			"old format, code matches what is looked for",
			`{Code:5,"details":[],"message":"Error: The bucket etc."}`,
			codes.Code(5),
			true,
		},
		{
			"old format, code doesn't match what is looked for",
			`{Code:55,"details":[],"message":"Error: The bucket etc."}`,
			codes.Code(5),
			false,
		},
		{
			"new format, code matches what is looked for",
			`{"code":5,"details":[],"message":"Error: The bucket etc."}`,
			codes.Code(5),
			true,
		},
		{
			"new format, code doesn't match what is looked for",
			`{"code":55,"details":[],"message":"Error: The bucket etc."}`,
			codes.Code(5),
			false,
		},
		{
			"bad format, should always be false",
			`"ceod":55`,
			codes.Code(5),
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found := CheckErrorCode(fmt.Errorf(tt.codeString), tt.expectCode)
			if found != tt.expectSuccess {
				t.Errorf("check error code returned %t, expected %t", found, tt.expectSuccess)
			}
		})
	}
}

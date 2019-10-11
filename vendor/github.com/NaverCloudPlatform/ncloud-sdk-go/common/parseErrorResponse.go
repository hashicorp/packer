package common

import (
	"encoding/xml"
	"strings"
)

func ParseErrorResponse(bytes []byte) (*ResponseError, error) {
	responseError := ResponseError{}

	if err := xml.Unmarshal([]byte(bytes), &responseError); err != nil {
		return nil, err
	}

	if responseError.ReturnMessage != "" {
		responseError.ReturnMessage = strings.TrimSpace(responseError.ReturnMessage)
	}

	return &responseError, nil
}

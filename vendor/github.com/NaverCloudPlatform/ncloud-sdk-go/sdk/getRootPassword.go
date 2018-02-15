package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processGetRootPasswordParams(reqParams *RequestGetRootPassword) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams.ServerInstanceNo == "" {
		return params, errors.New("Required field is not specified. location : serverInstanceNo")
	}

	if reqParams.PrivateKey == "" {
		return params, errors.New("Required field is not specified. location : privateKey")
	}

	params["serverInstanceNo"] = reqParams.ServerInstanceNo
	params["privateKey"] = reqParams.PrivateKey

	return params, nil
}

// GetRootPassword get root password from server instance
func (s *Conn) GetRootPassword(reqParams *RequestGetRootPassword) (*RootPassword, error) {
	params, err := processGetRootPasswordParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getRootPassword"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := RootPassword{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	responseGetRootPassword := RootPassword{}
	if err := xml.Unmarshal([]byte(bytes), &responseGetRootPassword); err != nil {
		return nil, err
	}

	if responseGetRootPassword.RootPassword != "" {
		responseGetRootPassword.RootPassword = strings.TrimSpace(responseGetRootPassword.RootPassword)
	}

	return &responseGetRootPassword, nil
}

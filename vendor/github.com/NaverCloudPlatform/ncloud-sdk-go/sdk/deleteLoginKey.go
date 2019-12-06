package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processDeleteLoginKeyParams(keyName string) error {
	if keyName == "" {
		return errors.New("KeyName is required field")
	}

	if len := len(keyName); len < 3 || len > 30 {
		return errors.New("Length of KeyName should be min 3 or max 30")
	}

	return nil
}

// DeleteLoginKey delete login key with keyName
func (s *Conn) DeleteLoginKey(keyName string) (*common.CommonResponse, error) {
	if err := processDeleteLoginKeyParams(keyName); err != nil {
		return nil, err
	}

	params := make(map[string]string)

	params["keyName"] = keyName
	params["action"] = "deleteLoginKey"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := common.CommonResponse{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var responseDeleteLoginKey = common.CommonResponse{}
	if err := xml.Unmarshal([]byte(bytes), &responseDeleteLoginKey); err != nil {
		return nil, err
	}

	return &responseDeleteLoginKey, nil
}

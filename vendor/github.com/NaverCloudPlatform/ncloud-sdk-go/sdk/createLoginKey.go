package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strings"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processCreateLoginKeyParams(keyName string) error {
	if keyName == "" {
		return errors.New("KeyName is required field")
	}

	if len := len(keyName); len < 3 || len > 30 {
		return errors.New("Length of KeyName should be min 3 or max 30")
	}

	return nil
}

// CreateLoginKey create loginkey with keyName
func (s *Conn) CreateLoginKey(keyName string) (*PrivateKey, error) {
	if err := processCreateLoginKeyParams(keyName); err != nil {
		return nil, err
	}

	params := make(map[string]string)

	params["keyName"] = keyName
	params["action"] = "createLoginKey"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := PrivateKey{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	privateKey := PrivateKey{}
	if err := xml.Unmarshal([]byte(bytes), &privateKey); err != nil {
		return nil, err
	}

	if privateKey.PrivateKey != "" {
		privateKey.PrivateKey = strings.TrimSpace(privateKey.PrivateKey)
	}

	return &privateKey, nil
}

package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processGetLoginKeyListParams(reqParams *RequestGetLoginKeyList) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if reqParams.KeyName != "" {
		if len := len(reqParams.KeyName); len < 3 || len > 20 {
			return nil, errors.New("Length of KeyName should be min 3 or max 30")
		}
		params["keyName"] = reqParams.KeyName
	}

	if reqParams.PageNo > 0 {
		if reqParams.PageNo > 2147483647 {
			return nil, errors.New("PageNo should be min 0 or max 2147483647")
		}
		params["pageNo"] = strconv.Itoa(reqParams.PageNo)
	}

	if reqParams.PageSize > 0 {
		if reqParams.PageSize > 2147483647 {
			return nil, errors.New("PageSize should be min 0 or max 2147483647")
		}
		params["pageSize"] = strconv.Itoa(reqParams.PageSize)
	}

	return params, nil
}

// GetLoginKeyList get login key list
func (s *Conn) GetLoginKeyList(reqParams *RequestGetLoginKeyList) (*LoginKeyList, error) {
	params, err := processGetLoginKeyListParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getLoginKeyList"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := LoginKeyList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	loginKeyList := LoginKeyList{}
	if err := xml.Unmarshal([]byte(bytes), &loginKeyList); err != nil {
		return nil, err
	}

	return &loginKeyList, nil
}

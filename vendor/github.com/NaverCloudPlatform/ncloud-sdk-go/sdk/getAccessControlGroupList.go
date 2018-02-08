package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processGetAccessControlGroupListParams(reqParams *RequestAccessControlGroupList) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if len(reqParams.AccessControlGroupConfigurationNoList) > 0 {
		for k, v := range reqParams.AccessControlGroupConfigurationNoList {
			params[fmt.Sprintf("accessControlGroupConfigurationNoList.%d", k+1)] = v
		}
	}

	if reqParams.IsDefault {
		params["isDefault"] = "true"
	}

	if reqParams.AccessControlGroupName != "" {
		if len(reqParams.AccessControlGroupName) < 3 || len(reqParams.AccessControlGroupName) > 30 {
			return nil, errors.New("AccessControlGroupName must be between 3 and 30 characters in length")
		}
		params["accessControlGroupName"] = reqParams.AccessControlGroupName
	}

	if reqParams.PageNo != 0 {
		if reqParams.PageNo > 2147483647 {
			return nil, errors.New("PageNo should be up to 2147483647")
		}

		params["pageNo"] = strconv.Itoa(reqParams.PageNo)
	}

	if reqParams.PageSize != 0 {
		if reqParams.PageSize > 2147483647 {
			return nil, errors.New("PageSize should be up to 2147483647")
		}

		params["pageSize"] = strconv.Itoa(reqParams.PageSize)
	}

	return params, nil
}

// GetAccessControlGroupList get access control group list
func (s *Conn) GetAccessControlGroupList(reqParams *RequestAccessControlGroupList) (*AccessControlGroupList, error) {
	params, err := processGetAccessControlGroupListParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getAccessControlGroupList"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := AccessControlGroupList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var AccessControlGroupList = AccessControlGroupList{}
	if err := xml.Unmarshal([]byte(bytes), &AccessControlGroupList); err != nil {
		return nil, err
	}

	return &AccessControlGroupList, nil
}

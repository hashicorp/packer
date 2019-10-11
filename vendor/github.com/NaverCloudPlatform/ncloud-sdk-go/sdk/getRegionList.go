package sdk

import (
	"encoding/xml"
	"fmt"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

// GetRegionList gets region list
func (s *Conn) GetRegionList() (*RegionList, error) {
	params := make(map[string]string)
	params["action"] = "getRegionList"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := RegionList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	regionList := RegionList{}
	if err := xml.Unmarshal([]byte(bytes), &regionList); err != nil {
		return nil, err
	}

	return &regionList, nil
}

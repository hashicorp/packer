package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processStopServerInstancesParams(reqParams *RequestStopServerInstances) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil || len(reqParams.ServerInstanceNoList) == 0 {
		return params, errors.New("serverInstanceNoList is required")
	}

	if len(reqParams.ServerInstanceNoList) > 0 {
		for k, v := range reqParams.ServerInstanceNoList {
			params[fmt.Sprintf("serverInstanceNoList.%d", k+1)] = v
		}
	}

	return params, nil
}

// StopServerInstances stop server instances
func (s *Conn) StopServerInstances(reqParams *RequestStopServerInstances) (*ServerInstanceList, error) {
	params, err := processStopServerInstancesParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "stopServerInstances"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := ServerInstanceList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var serverInstanceList = ServerInstanceList{}
	if err := xml.Unmarshal([]byte(bytes), &serverInstanceList); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &serverInstanceList, nil
}

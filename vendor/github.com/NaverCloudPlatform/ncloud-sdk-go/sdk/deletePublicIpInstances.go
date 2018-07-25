package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processDeletePublicIPInstancesParams(reqParams *RequestDeletePublicIPInstances) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil || len(reqParams.PublicIPInstanceNoList) == 0 {
		return params, errors.New("Required field is not specified. location : publicIpInstanceNoList.N")
	}

	if len(reqParams.PublicIPInstanceNoList) > 0 {
		for k, v := range reqParams.PublicIPInstanceNoList {
			params[fmt.Sprintf("publicIpInstanceNoList.%d", k+1)] = v
		}
	}

	return params, nil
}

// DeletePublicIPInstances delete public ip instances
func (s *Conn) DeletePublicIPInstances(reqParams *RequestDeletePublicIPInstances) (*PublicIPInstanceList, error) {
	params, err := processDeletePublicIPInstancesParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "deletePublicIpInstances"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := PublicIPInstanceList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var publicIPInstanceList = PublicIPInstanceList{}
	if err := xml.Unmarshal([]byte(bytes), &publicIPInstanceList); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &publicIPInstanceList, nil
}

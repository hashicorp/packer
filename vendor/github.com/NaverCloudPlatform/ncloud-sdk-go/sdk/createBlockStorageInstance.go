package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processCreateBlockStorageInstanceParams(reqParams *RequestBlockStorageInstance) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams.BlockStorageName != "" {
		if len := len(reqParams.BlockStorageName); len < 3 || len > 30 {
			return nil, errors.New("Length of BlockStorageName should be min 3 or max 30")
		}
		params["blockStorageName"] = reqParams.BlockStorageName
	}

	if reqParams.BlockStorageSize < 10 || reqParams.BlockStorageSize > 2000 {
		return nil, errors.New("BlockStorageSize should be min 10 or max 2000")
	}

	if reqParams.BlockStorageDescription != "" {
		if len := len(reqParams.BlockStorageDescription); len > 1000 {
			return nil, errors.New("Length of BlockStorageDescription should be max 1000")
		}
		params["blockStorageDescription"] = reqParams.BlockStorageDescription
	}

	if int(reqParams.BlockStorageSize/10)*10 != reqParams.BlockStorageSize {
		return nil, errors.New("BlockStorageSize must be a multiple of 10 GB")
	}

	if reqParams.BlockStorageSize == 0 {
		return nil, errors.New("BlockStorageSize field is required")
	}

	params["blockStorageSize"] = strconv.Itoa(reqParams.BlockStorageSize)

	if reqParams.ServerInstanceNo == "" {
		return nil, errors.New("ServerInstanceNo field is required")
	}

	params["serverInstanceNo"] = reqParams.ServerInstanceNo

	return params, nil
}

// CreateBlockStorageInstance create block storage instance
func (s *Conn) CreateBlockStorageInstance(reqParams *RequestBlockStorageInstance) (*BlockStorageInstanceList, error) {

	params, err := processCreateBlockStorageInstanceParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "createBlockStorageInstance"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := BlockStorageInstanceList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var blockStorageInstanceList = BlockStorageInstanceList{}
	if err := xml.Unmarshal([]byte(bytes), &blockStorageInstanceList); err != nil {
		return nil, err
	}

	return &blockStorageInstanceList, nil
}

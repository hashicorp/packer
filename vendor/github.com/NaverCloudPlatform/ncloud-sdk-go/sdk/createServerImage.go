package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processCreateMemberServerImageParams(reqParams *RequestCreateServerImage) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil || reqParams.ServerInstanceNo == "" {
		return params, errors.New("ServerInstanceNo is required field")
	}

	if reqParams.MemberServerImageName != "" {
		if len := len(reqParams.MemberServerImageName); len < 3 || len > 30 {
			return nil, errors.New("Length of MemberServerImageName should be min 3 or max 30")
		}
		params["memberServerImageName"] = reqParams.MemberServerImageName
	}

	if reqParams.MemberServerImageDescription != "" {
		if len := len(reqParams.MemberServerImageDescription); len > 1000 {
			return nil, errors.New("Length of MemberServerImageDescription should be smaller than 1000")
		}
		params["memberServerImageDescription"] = reqParams.MemberServerImageDescription
	}

	params["serverInstanceNo"] = reqParams.ServerInstanceNo

	return params, nil
}

// CreateMemberServerImage create member server image and retrun member server image list
func (s *Conn) CreateMemberServerImage(reqParams *RequestCreateServerImage) (*MemberServerImageList, error) {
	params, err := processCreateMemberServerImageParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "createMemberServerImage"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := MemberServerImageList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var serverImageListsResp = MemberServerImageList{}
	if err := xml.Unmarshal([]byte(bytes), &serverImageListsResp); err != nil {
		return nil, err
	}

	return &serverImageListsResp, nil
}

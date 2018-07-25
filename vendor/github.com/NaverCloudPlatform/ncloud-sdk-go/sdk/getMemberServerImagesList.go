package sdk

import (
	"encoding/xml"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processGetMemberServerImageListParams(reqParams *RequestServerImageList) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if len(reqParams.MemberServerImageNoList) > 0 {
		for k, v := range reqParams.MemberServerImageNoList {
			params[fmt.Sprintf("memberServerImageNoList.%d", k+1)] = v
		}
	}

	if len(reqParams.PlatformTypeCodeList) > 0 {
		for k, v := range reqParams.PlatformTypeCodeList {
			params[fmt.Sprintf("platformTypeCodeList.%d", k+1)] = v
		}
	}

	if reqParams.PageNo != 0 {
		params["pageNo"] = strconv.Itoa(reqParams.PageNo)
	}

	if reqParams.PageSize != 0 {
		params["pageSize"] = strconv.Itoa(reqParams.PageSize)
	}

	if reqParams.RegionNo != "" {
		params["regionNo"] = reqParams.RegionNo
	}

	if reqParams.SortedBy != "" {
		params["sortedBy"] = reqParams.SortedBy
	}

	if reqParams.SortingOrder != "" {
		params["sortingOrder"] = reqParams.SortingOrder
	}

	return params, nil
}

// GetMemberServerImageList get member server image list
func (s *Conn) GetMemberServerImageList(reqParams *RequestServerImageList) (*MemberServerImageList, error) {
	params, err := processGetMemberServerImageListParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getMemberServerImageList"

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

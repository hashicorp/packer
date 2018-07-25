package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processGetPublicIPInstanceListParams(reqParams *RequestPublicIPInstanceList) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if reqParams.IsAssociated {
		params["isAssociated"] = "true"
	}

	if len(reqParams.PublicIPInstanceNoList) > 0 {
		for k, v := range reqParams.PublicIPInstanceNoList {
			params[fmt.Sprintf("publicIpInstanceNoList.%d", k+1)] = v
		}
	}

	if len(reqParams.PublicIPList) > 0 {
		for k, v := range reqParams.PublicIPList {
			if len(reqParams.PublicIPList) < 5 || len(reqParams.PublicIPList) > 15 {
				return nil, errors.New("PublicIPList must be between 5 and 15 characters in length")
			}
			params[fmt.Sprintf("publicIpList.%d", k+1)] = v
		}
	}

	if reqParams.SearchFilterName != "" {
		if reqParams.SearchFilterName != "publicIp" && reqParams.SearchFilterName != "associatedServerName" {
			return nil, errors.New("SearchFilterName must be publicIp or associatedServerName")
		}
		params["searchFilterName"] = reqParams.SearchFilterName
	}

	if reqParams.SearchFilterValue != "" {
		params["searchFilterValue"] = reqParams.SearchFilterValue
	}

	if reqParams.InternetLineTypeCode != "" {
		if reqParams.InternetLineTypeCode != "PUBLC" && reqParams.InternetLineTypeCode != "GLBL" {
			return params, errors.New("InternetLineTypeCode must be PUBLC or GLBL")
		}
		params["internetLineTypeCode"] = reqParams.InternetLineTypeCode
	}

	if reqParams.RegionNo != "" {
		params["regionNo"] = reqParams.RegionNo
	}

	if reqParams.PageNo != 0 {
		if reqParams.PageNo > 2147483647 {
			return nil, errors.New("PageNo must be up to 2147483647")
		}

		params["pageNo"] = strconv.Itoa(reqParams.PageNo)
	}

	if reqParams.PageSize != 0 {
		if reqParams.PageSize > 2147483647 {
			return nil, errors.New("PageSize must be up to 2147483647")
		}

		params["pageSize"] = strconv.Itoa(reqParams.PageSize)
	}

	if reqParams.SortedBy != "" {
		if strings.EqualFold(reqParams.SortedBy, "publicIp") || strings.EqualFold(reqParams.SortedBy, "publicIpInstanceNo") {
			params["sortedBy"] = reqParams.SortedBy
		} else {
			return nil, errors.New("SortedBy must be publicIp or publicIpInstanceNo")
		}
	}

	if reqParams.SortingOrder != "" {
		if strings.EqualFold(reqParams.SortingOrder, "ascending") || strings.EqualFold(reqParams.SortingOrder, "descending") {
			params["sortingOrder"] = reqParams.SortingOrder
		} else {
			return nil, errors.New("SortingOrder must be ascending or descending")
		}
	}

	return params, nil
}

// GetPublicIPInstanceList get public ip instance list
func (s *Conn) GetPublicIPInstanceList(reqParams *RequestPublicIPInstanceList) (*PublicIPInstanceList, error) {
	params, err := processGetPublicIPInstanceListParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getPublicIpInstanceList"

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
		return nil, err
	}

	return &publicIPInstanceList, nil
}

package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processGetServerInstanceListParams(reqParams *RequestGetServerInstanceList) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if len(reqParams.ServerInstanceNoList) > 0 {
		for k, v := range reqParams.ServerInstanceNoList {
			params[fmt.Sprintf("serverInstanceNoList.%d", k+1)] = v
		}
	}

	if reqParams.SearchFilterName != "" {
		params["searchFilterName"] = reqParams.SearchFilterName
	}

	if reqParams.SearchFilterValue != "" {
		params["searchFilterValue"] = reqParams.SearchFilterValue
	}

	if reqParams.PageNo > 0 {
		if reqParams.PageNo > 2147483647 {
			return nil, errors.New("PageNo should be less than 2147483647")

		}
		params["pageNo"] = strconv.Itoa(reqParams.PageNo)
	}

	if reqParams.PageSize > 0 {
		if reqParams.PageSize > 2147483647 {
			return nil, errors.New("PageSize should be less than 2147483647")

		}
		params["pageSize"] = strconv.Itoa(reqParams.PageSize)
	}

	if reqParams.ServerInstanceStatusCode != "" {
		if reqParams.ServerInstanceStatusCode != "RUN" && reqParams.ServerInstanceStatusCode != "NSTOP" && reqParams.ServerInstanceStatusCode != "ING" {
			return nil, errors.New("ServerInstanceStatusCode should be RUN, NSTOP or ING")
		}
		params["serverInstanceStatusCode"] = reqParams.ServerInstanceStatusCode
	}

	if reqParams.InternetLineTypeCode != "" {
		if reqParams.InternetLineTypeCode != "PUBLC" && reqParams.InternetLineTypeCode != "GLBL" {
			return nil, errors.New("InternetLineTypeCode should be PUBLC or GLBL")
		}
		params["internetLineTypeCode"] = reqParams.InternetLineTypeCode
	}

	if reqParams.RegionNo != "" {
		params["regionNo"] = reqParams.RegionNo
	}

	if reqParams.BaseBlockStorageDiskTypeCode != "" {
		if reqParams.BaseBlockStorageDiskTypeCode != "NET" && reqParams.BaseBlockStorageDiskTypeCode != "LOCAL" {
			return nil, errors.New("BaseBlockStorageDiskTypeCode should be NET or LOCAL")
		}
		params["baseBlockStorageDiskTypeCode"] = reqParams.BaseBlockStorageDiskTypeCode
	}

	if reqParams.BaseBlockStorageDiskDetailTypeCode != "" {
		if reqParams.BaseBlockStorageDiskDetailTypeCode != "HDD" && reqParams.BaseBlockStorageDiskDetailTypeCode != "SSD" {
			return nil, errors.New("BaseBlockStorageDiskDetailTypeCode should be HDD or SSD")
		}
		params["baseBlockStorageDiskDetailTypeCode"] = reqParams.BaseBlockStorageDiskDetailTypeCode
	}

	if reqParams.SortedBy != "" {
		if reqParams.SortedBy != "serverName" && reqParams.SortedBy != "serverInstanceNo" {
			return nil, errors.New("SortedBy should be serverName or serverInstanceNo")
		}
		params["sortedBy"] = reqParams.SortedBy
	}

	if reqParams.SortingOrder != "" {
		if reqParams.SortingOrder != "ascending" && reqParams.SortingOrder != "descending" {
			return nil, errors.New("SortingOrder should be ascending or descending")
		}
		params["sortingOrder"] = reqParams.SortingOrder
	}

	return params, nil
}

// GetServerInstanceList get server instance list
func (s *Conn) GetServerInstanceList(reqParams *RequestGetServerInstanceList) (*ServerInstanceList, error) {
	params, err := processGetServerInstanceListParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getServerInstanceList"

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

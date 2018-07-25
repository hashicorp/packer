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

func processGetBlockStorageInstanceListParams(reqParams *RequestBlockStorageInstanceList) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if reqParams.ServerInstanceNo != "" {
		params["serverInstanceNo"] = reqParams.ServerInstanceNo
	}

	if len(reqParams.BlockStorageInstanceNoList) > 0 {
		for k, v := range reqParams.BlockStorageInstanceNoList {
			params[fmt.Sprintf("blockStorageInstanceNoList.%d", k+1)] = v
		}
	}

	if reqParams.SearchFilterName != "" {
		if reqParams.SearchFilterName != "blockStorageName" && reqParams.SearchFilterName != "attachmentInformation" {
			return nil, errors.New("SearchFilterName should be blockStorageName or attachmentInformation")
		}
		params["searchFilterName"] = reqParams.SearchFilterName
	}

	if reqParams.SearchFilterValue == "" {
		params["searchFilterValue"] = reqParams.SearchFilterValue
	}

	if len(reqParams.BlockStorageTypeCodeList) > 0 {
		for k, v := range reqParams.BlockStorageTypeCodeList {
			if v != "BASIC" && v != "SVRBS" {
				return nil, errors.New("BlockStorageTypeCodeList value should be BASIC or SVRBS")
			}
			params[fmt.Sprintf("blockStorageTypeCodeList.%d", k+1)] = v
		}
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

	if reqParams.BlockStorageInstanceStatusCode != "" {
		if reqParams.BlockStorageInstanceStatusCode != "ATTAC" && reqParams.BlockStorageInstanceStatusCode != "CRAET" {
			return nil, errors.New("BlockStorageInstanceStatusCode should be ATTAC or CRAET")
		}
		params["blockStorageInstanceStatusCode"] = reqParams.BlockStorageInstanceStatusCode
	}

	if reqParams.DiskTypeCode != "" {
		if reqParams.DiskTypeCode != "NET" && reqParams.DiskTypeCode != "LOCAL" {
			return nil, errors.New("DiskTypeCode should be NET or LOCAL")
		}
		params["diskTypeCode"] = reqParams.DiskTypeCode
	}

	if reqParams.DiskDetailTypeCode != "" {
		if reqParams.DiskDetailTypeCode != "HDD" && reqParams.DiskDetailTypeCode != "SSD" {
			return nil, errors.New("DiskDetailTypeCode should be HDD or SSD")
		}
		params["diskDetailTypeCode"] = reqParams.DiskDetailTypeCode
	}

	if reqParams.RegionNo != "" {
		params["regionNo"] = reqParams.RegionNo
	}

	if reqParams.SortedBy != "" {
		if strings.EqualFold(reqParams.SortedBy, "blockStorageName") || strings.EqualFold(reqParams.SortedBy, "blockStorageInstanceNo") {
			params["sortedBy"] = reqParams.SortedBy
		} else {
			return nil, errors.New("SortedBy should be blockStorageName or blockStorageInstanceNo")
		}
	}

	if reqParams.SortingOrder != "" {
		if strings.EqualFold(reqParams.SortingOrder, "ascending") || strings.EqualFold(reqParams.SortingOrder, "descending") {
			params["sortingOrder"] = reqParams.SortingOrder
		} else {
			return nil, errors.New("SortingOrder should be ascending or descending")
		}
	}

	return params, nil
}

// GetBlockStorageInstance Get block storage instance list
func (s *Conn) GetBlockStorageInstance(reqParams *RequestBlockStorageInstanceList) (*BlockStorageInstanceList, error) {
	params, err := processGetBlockStorageInstanceListParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getBlockStorageInstanceList"

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

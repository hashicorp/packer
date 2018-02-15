package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processGetServerImageProductListParams(reqParams *RequestGetServerImageProductList) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if reqParams.ExclusionProductCode != "" {
		if len(reqParams.ExclusionProductCode) > 20 {
			return params, errors.New("Length of exclusionProductCode should be max 20")
		}
		params["exclusionProductCode"] = reqParams.ExclusionProductCode
	}

	if reqParams.ProductCode != "" {
		if len(reqParams.ProductCode) > 20 {
			return params, errors.New("Length of productCode should be max 20")
		}
		params["productCode"] = reqParams.ProductCode
	}

	if len(reqParams.PlatformTypeCodeList) > 0 {
		for k, v := range reqParams.PlatformTypeCodeList {
			params[fmt.Sprintf("platformTypeCodeList.%d", k+1)] = v
		}
	}

	if reqParams.BlockStorageSize > 0 {
		if reqParams.BlockStorageSize != 50 && reqParams.BlockStorageSize != 100 {
			return nil, errors.New("blockStorageSize should be null, 50 or 100")
		}
		params["blockStorageSize"] = strconv.Itoa(reqParams.BlockStorageSize)
	}

	if reqParams.RegionNo != "" {
		params["regionNo"] = reqParams.RegionNo
	}

	return params, nil
}

// GetServerImageProductList gets server image product list
func (s *Conn) GetServerImageProductList(reqParams *RequestGetServerImageProductList) (*ProductList, error) {
	params, err := processGetServerImageProductListParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "getServerImageProductList"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := ProductList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var productList = ProductList{}
	if err := xml.Unmarshal([]byte(bytes), &productList); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &productList, nil
}

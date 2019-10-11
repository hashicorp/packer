package sdk

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processCreateServerInstancesParams(reqParams *RequestCreateServerInstance) (map[string]string, error) {
	params := make(map[string]string)

	if reqParams == nil {
		return params, nil
	}

	if reqParams.ServerImageProductCode != "" {
		if len := len(reqParams.ServerImageProductCode); len > 20 {
			return nil, errors.New("Length of ServerImageProductCode should be min 1 or max 20")
		}
		params["serverImageProductCode"] = reqParams.ServerImageProductCode
	}

	if reqParams.ServerProductCode != "" {
		if len := len(reqParams.ServerProductCode); len > 20 {
			return nil, errors.New("Length of ServerProductCode should be min 1 or max 20")
		}
		params["serverProductCode"] = reqParams.ServerProductCode
	}

	if reqParams.MemberServerImageNo != "" {
		params["memberServerImageNo"] = reqParams.MemberServerImageNo
	}

	if reqParams.ServerName != "" {
		if len := len(reqParams.ServerName); len < 3 || len > 30 {
			return nil, errors.New("Length of ServerName should be min 3 or max 30")
		}
		params["serverName"] = reqParams.ServerName
	}

	if reqParams.ServerDescription != "" {
		if len := len(reqParams.ServerDescription); len > 1000 {
			return nil, errors.New("Length of ServerDescription should be min 1 or max 1000")
		}
		params["serverDescription"] = reqParams.ServerDescription
	}

	if reqParams.LoginKeyName != "" {
		if len := len(reqParams.LoginKeyName); len < 3 || len > 30 {
			return nil, errors.New("Length of LoginKeyName should be min 3 or max 30")
		}
		params["loginKeyName"] = reqParams.LoginKeyName
	}

	if reqParams.IsProtectServerTermination == true {
		params["isProtectServerTermination"] = "true"
	}

	if reqParams.ServerCreateCount > 0 {
		if reqParams.ServerCreateCount > 20 {
			return nil, errors.New("ServerCreateCount should be min 1 or max 20")

		}
		params["serverCreateCount"] = strconv.Itoa(reqParams.ServerCreateCount)
	}

	if reqParams.ServerCreateStartNo > 0 {
		if reqParams.ServerCreateCount+reqParams.ServerCreateStartNo > 1000 {
			return nil, errors.New("Sum of ServerCreateCount and ServerCreateStartNo should be less than 1000")

		}
		params["serverCreateStartNo"] = strconv.Itoa(reqParams.ServerCreateStartNo)
	}

	if reqParams.InternetLineTypeCode != "" {
		if reqParams.InternetLineTypeCode != "PUBLC" && reqParams.InternetLineTypeCode != "GLBL" {
			return nil, errors.New("InternetLineTypeCode should be PUBLC or GLBL")
		}
		params["internetLineTypeCode"] = reqParams.InternetLineTypeCode
	}

	if reqParams.FeeSystemTypeCode != "" {
		if reqParams.FeeSystemTypeCode != "FXSUM" && reqParams.FeeSystemTypeCode != "MTRAT" {
			return nil, errors.New("FeeSystemTypeCode should be FXSUM or MTRAT")
		}
		params["feeSystemTypeCode"] = reqParams.FeeSystemTypeCode
	}

	if reqParams.UserData != "" {
		if len := len(reqParams.UserData); len > 21847 {
			return nil, errors.New("Length of UserData should be min 1 or max 21847")
		}
		params["userData"] = base64.StdEncoding.EncodeToString([]byte(reqParams.UserData))
	}

	if reqParams.ZoneNo != "" {
		params["zoneNo"] = reqParams.ZoneNo
	}

	if len(reqParams.AccessControlGroupConfigurationNoList) > 0 {
		for k, v := range reqParams.AccessControlGroupConfigurationNoList {
			params[fmt.Sprintf("accessControlGroupConfigurationNoList.%d", k+1)] = v
		}
	}

	return params, nil
}

// CreateServerInstances create server instances
func (s *Conn) CreateServerInstances(reqParams *RequestCreateServerInstance) (*ServerInstanceList, error) {
	params, err := processCreateServerInstancesParams(reqParams)
	if err != nil {
		return nil, err
	}

	params["action"] = "createServerInstances"

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

	var responseCreateServerInstances = ServerInstanceList{}
	if err := xml.Unmarshal([]byte(bytes), &responseCreateServerInstances); err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &responseCreateServerInstances, nil
}

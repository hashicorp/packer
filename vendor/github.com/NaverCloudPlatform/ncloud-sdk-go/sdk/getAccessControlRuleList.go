package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func checkGetAccessControlRuleListParams(accessControlGroupConfigurationNo string) error {
	if accessControlGroupConfigurationNo == "" {
		return errors.New("accessControlGroupConfigurationNo is required")
	}

	if no, err := strconv.Atoi(accessControlGroupConfigurationNo); err != nil {
		return err
	} else if no < 0 || no > 2147483647 {
		return errors.New("accessControlGroupConfigurationNoeNo must be up to 2147483647")
	}

	return nil
}

// GetAccessControlRuleList get access control group list
func (s *Conn) GetAccessControlRuleList(accessControlGroupConfigurationNo string) (*AccessControlRuleList, error) {
	if err := checkGetAccessControlRuleListParams(accessControlGroupConfigurationNo); err != nil {
		return nil, err
	}

	params := make(map[string]string)
	params["accessControlGroupConfigurationNo"] = accessControlGroupConfigurationNo
	params["action"] = "getAccessControlRuleList"

	bytes, resp, err := request.NewRequest(s.accessKey, s.secretKey, "GET", s.apiURL+"server/", params)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		responseError, err := common.ParseErrorResponse(bytes)
		if err != nil {
			return nil, err
		}

		respError := AccessControlRuleList{}
		respError.ReturnCode = responseError.ReturnCode
		respError.ReturnMessage = responseError.ReturnMessage

		return &respError, fmt.Errorf("%s %s - error code: %d , error message: %s", resp.Status, string(bytes), responseError.ReturnCode, responseError.ReturnMessage)
	}

	var AccessControlRuleList = AccessControlRuleList{}
	if err := xml.Unmarshal([]byte(bytes), &AccessControlRuleList); err != nil {
		return nil, err
	}

	return &AccessControlRuleList, nil
}

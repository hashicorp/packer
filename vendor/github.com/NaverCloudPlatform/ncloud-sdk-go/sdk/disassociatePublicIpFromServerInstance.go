package sdk

import (
	"encoding/xml"
	"errors"
	"fmt"

	common "github.com/NaverCloudPlatform/ncloud-sdk-go/common"
	request "github.com/NaverCloudPlatform/ncloud-sdk-go/request"
)

func processDisassociatePublicIPParams(PublicIPInstanceNo string) error {
	if PublicIPInstanceNo == "" {
		return errors.New("Required field is not specified. location : publicIpInstanceNo")
	}

	return nil
}

// DisassociatePublicIP diassociate public ip from server instance
func (s *Conn) DisassociatePublicIP(PublicIPInstanceNo string) (*PublicIPInstanceList, error) {
	if err := processDisassociatePublicIPParams(PublicIPInstanceNo); err != nil {
		return nil, err
	}

	params := make(map[string]string)
	params["publicIpInstanceNo"] = PublicIPInstanceNo
	params["action"] = "disassociatePublicIpFromServerInstance"

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

	var responseDisassociatePublicIPInstances = PublicIPInstanceList{}
	if err := xml.Unmarshal([]byte(bytes), &responseDisassociatePublicIPInstances); err != nil {
		return nil, err
	}

	return &responseDisassociatePublicIPInstances, nil
}

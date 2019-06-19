package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
)

type Response interface {
	ParseErrorFromHTTPResponse(body []byte) error
}

type BaseResponse struct {
}

type ErrorResponse struct {
	Response struct {
		Error struct {
			Code    string `json:"Code"`
			Message string `json:"Message"`
		} `json:"Error" omitempty`
		RequestId string `json:"RequestId"`
	} `json:"Response"`
}

type DeprecatedAPIErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	CodeDesc string `json:"codeDesc"`
}

func (r *BaseResponse) ParseErrorFromHTTPResponse(body []byte) (err error) {
	resp := &ErrorResponse{}
	err = json.Unmarshal(body, resp)
	if err != nil {
		return
	}
	if resp.Response.Error.Code != "" {
		return errors.NewTencentCloudSDKError(resp.Response.Error.Code, resp.Response.Error.Message, resp.Response.RequestId)
	}

	deprecated := &DeprecatedAPIErrorResponse{}
	err = json.Unmarshal(body, deprecated)
	if err != nil {
		return
	}
	if deprecated.Code != 0 {
		return errors.NewTencentCloudSDKError(deprecated.CodeDesc, deprecated.Message, "")
	}
	return nil
}

func ParseFromHttpResponse(hr *http.Response, response Response) (err error) {
	defer hr.Body.Close()
	body, err := ioutil.ReadAll(hr.Body)
	if err != nil {
		return
	}
	if hr.StatusCode != 200 {
		return fmt.Errorf("Request fail with status: %s, with body: %s", hr.Status, body)
	}
	//log.Printf("[DEBUG] Response Body=%s", body)
	err = response.ParseErrorFromHTTPResponse(body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		log.Printf("Unexpected Error occurs when parsing API response\n%s\n", string(body[:]))
	}
	return
}

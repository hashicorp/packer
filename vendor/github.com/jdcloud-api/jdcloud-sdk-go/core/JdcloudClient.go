// Copyright 2018-2025 JDCLOUD.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"net/http"
	"strings"
	"time"
	"fmt"
	"encoding/json"
	"encoding/base64"
)

// JDCloudClient is the base struct of service clients
type JDCloudClient struct {
	Credential  Credential
	Config      Config
	ServiceName string
	Revision    string
	Logger      Logger
}

type SignFunc func(*http.Request) error

// Send send the request and return the response to the client.
// Parameter request accepts concrete request object which follow RequestInterface.
func (c JDCloudClient) Send(request RequestInterface, serviceName string) ([]byte, error) {
	method := request.GetMethod()
	builder := GetParameterBuilder(method, c.Logger)
	jsonReq, _ := json.Marshal(request)
	encodedUrl, err := builder.BuildURL(request.GetURL(), jsonReq)
	if err != nil {
		return nil, err
	}
	reqUrl := fmt.Sprintf("%s://%s/%s%s", c.Config.Scheme, c.Config.Endpoint, request.GetVersion(), encodedUrl)

	body, err := builder.BuildBody(jsonReq)
	if err != nil {
		return nil, err
	}

	sign := func(r *http.Request) error {
		regionId := request.GetRegionId()
		// some request has no region parameter, so give a value to it,
		// then API gateway can calculate sign successfully.
		if regionId == "" {
			regionId = "jdcloud-api"
		}

		signer := NewSigner(c.Credential, c.Logger)
		_, err := signer.Sign(r, strings.NewReader(body), serviceName, regionId, time.Now())
		return err
	}

	return c.doSend(method, reqUrl, body, request.GetHeaders(), c.Config.Timeout, sign)
}

func (c JDCloudClient) doSend(method, url, data string, header map[string]string, timeout time.Duration, sign SignFunc) ([]byte, error) {
	client := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if err != nil {
		c.Logger.Log(LogFatal, err.Error())
		return nil, err
	}

	c.setHeader(req, header)

	err = sign(req)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		c.Logger.Log(LogError, err.Error())
		return nil, err
	}

	processor := GetResponseProcessor(req.Method)
	result, err := processor.Process(resp)
	if err != nil {
		c.Logger.Log(LogError, err.Error())
		return nil, err
	}
	return result, nil
}

func (c JDCloudClient) setHeader(req *http.Request, header map[string]string) {

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", fmt.Sprintf("JdcloudSdkGo/%s %s/%s", Version, c.ServiceName, c.Revision))

	base64Headers := []string{HeaderJdcloudPrefix + "-pin", HeaderJdcloudPrefix + "-erp", HeaderJdcloudPrefix + "-security-token",
		HeaderJcloudPrefix + "-pin", HeaderJcloudPrefix + "-erp", HeaderJcloudPrefix + "-security-token"}

	for k, v := range header {
		if includes(base64Headers, strings.ToLower(k)) {
			v = base64.StdEncoding.EncodeToString([]byte(v))
		}

		req.Header.Set(k, v)
	}

	for k, v := range req.Header {
		c.Logger.Log(LogInfo, k, v)
	}
}

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

type RequestInterface interface {
	GetURL() string
	GetMethod() string
	GetVersion() string
	GetHeaders() map[string]string
	GetRegionId() string
}

// JDCloudRequest is the base struct of service requests
type JDCloudRequest struct {
	URL     string // resource url, i.e. /regions/${regionId}/elasticIps/${elasticIpId}
	Method  string
	Header  map[string]string
	Version string
}

func (r JDCloudRequest) GetURL() string {
	return r.URL
}

func (r JDCloudRequest) GetMethod() string {
	return r.Method
}

func (r JDCloudRequest) GetVersion() string {
	return r.Version
}

func (r JDCloudRequest) GetHeaders() map[string]string {
	return r.Header
}

// AddHeader only adds pin or erp, they will be encoded to base64 code
func (r *JDCloudRequest) AddHeader(key, value string) {
	if r.Header == nil {
		r.Header = make(map[string]string)
	}
	r.Header[key] = value
}

//
// Copyright 2018, Sander van Harmelen
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
//

package cloudstack

import (
	"encoding/json"
	"net/url"
	"strconv"
)

type ChangeOutOfBandManagementPasswordParams struct {
	p map[string]interface{}
}

func (p *ChangeOutOfBandManagementPasswordParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	return u
}

func (p *ChangeOutOfBandManagementPasswordParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *ChangeOutOfBandManagementPasswordParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

// You should always use this function to get a new ChangeOutOfBandManagementPasswordParams instance,
// as then you are sure you have configured all required params
func (s *OutofbandManagementService) NewChangeOutOfBandManagementPasswordParams(hostid string) *ChangeOutOfBandManagementPasswordParams {
	p := &ChangeOutOfBandManagementPasswordParams{}
	p.p = make(map[string]interface{})
	p.p["hostid"] = hostid
	return p
}

// Changes out-of-band management interface password on the host and updates the interface configuration in CloudStack if the operation succeeds, else reverts the old password
func (s *OutofbandManagementService) ChangeOutOfBandManagementPassword(p *ChangeOutOfBandManagementPasswordParams) (*ChangeOutOfBandManagementPasswordResponse, error) {
	resp, err := s.cs.newRequest("changeOutOfBandManagementPassword", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ChangeOutOfBandManagementPasswordResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	// If we have a async client, we need to wait for the async result
	if s.cs.async {
		b, err := s.cs.GetAsyncJobResult(r.JobID, s.cs.timeout)
		if err != nil {
			if err == AsyncTimeoutErr {
				return &r, err
			}
			return nil, err
		}

		b, err = getRawValue(b)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type ChangeOutOfBandManagementPasswordResponse struct {
	Action      string `json:"action"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Driver      string `json:"driver"`
	Enabled     bool   `json:"enabled"`
	Hostid      string `json:"hostid"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Password    string `json:"password"`
	Port        string `json:"port"`
	Powerstate  string `json:"powerstate"`
	Status      bool   `json:"status"`
	Username    string `json:"username"`
}

type ConfigureOutOfBandManagementParams struct {
	p map[string]interface{}
}

func (p *ConfigureOutOfBandManagementParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["address"]; found {
		u.Set("address", v.(string))
	}
	if v, found := p.p["driver"]; found {
		u.Set("driver", v.(string))
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	if v, found := p.p["port"]; found {
		u.Set("port", v.(string))
	}
	if v, found := p.p["username"]; found {
		u.Set("username", v.(string))
	}
	return u
}

func (p *ConfigureOutOfBandManagementParams) SetAddress(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["address"] = v
	return
}

func (p *ConfigureOutOfBandManagementParams) SetDriver(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["driver"] = v
	return
}

func (p *ConfigureOutOfBandManagementParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *ConfigureOutOfBandManagementParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

func (p *ConfigureOutOfBandManagementParams) SetPort(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["port"] = v
	return
}

func (p *ConfigureOutOfBandManagementParams) SetUsername(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["username"] = v
	return
}

// You should always use this function to get a new ConfigureOutOfBandManagementParams instance,
// as then you are sure you have configured all required params
func (s *OutofbandManagementService) NewConfigureOutOfBandManagementParams(address string, driver string, hostid string, password string, port string, username string) *ConfigureOutOfBandManagementParams {
	p := &ConfigureOutOfBandManagementParams{}
	p.p = make(map[string]interface{})
	p.p["address"] = address
	p.p["driver"] = driver
	p.p["hostid"] = hostid
	p.p["password"] = password
	p.p["port"] = port
	p.p["username"] = username
	return p
}

// Configures a host's out-of-band management interface
func (s *OutofbandManagementService) ConfigureOutOfBandManagement(p *ConfigureOutOfBandManagementParams) (*OutOfBandManagementResponse, error) {
	resp, err := s.cs.newRequest("configureOutOfBandManagement", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r OutOfBandManagementResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type OutOfBandManagementResponse struct {
	Action      string `json:"action"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Driver      string `json:"driver"`
	Enabled     bool   `json:"enabled"`
	Hostid      string `json:"hostid"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Password    string `json:"password"`
	Port        string `json:"port"`
	Powerstate  string `json:"powerstate"`
	Status      bool   `json:"status"`
	Username    string `json:"username"`
}

type IssueOutOfBandManagementPowerActionParams struct {
	p map[string]interface{}
}

func (p *IssueOutOfBandManagementPowerActionParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["action"]; found {
		u.Set("action", v.(string))
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["timeout"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("timeout", vv)
	}
	return u
}

func (p *IssueOutOfBandManagementPowerActionParams) SetAction(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["action"] = v
	return
}

func (p *IssueOutOfBandManagementPowerActionParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *IssueOutOfBandManagementPowerActionParams) SetTimeout(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["timeout"] = v
	return
}

// You should always use this function to get a new IssueOutOfBandManagementPowerActionParams instance,
// as then you are sure you have configured all required params
func (s *OutofbandManagementService) NewIssueOutOfBandManagementPowerActionParams(action string, hostid string) *IssueOutOfBandManagementPowerActionParams {
	p := &IssueOutOfBandManagementPowerActionParams{}
	p.p = make(map[string]interface{})
	p.p["action"] = action
	p.p["hostid"] = hostid
	return p
}

// Initiates the specified power action to the host's out-of-band management interface
func (s *OutofbandManagementService) IssueOutOfBandManagementPowerAction(p *IssueOutOfBandManagementPowerActionParams) (*IssueOutOfBandManagementPowerActionResponse, error) {
	resp, err := s.cs.newRequest("issueOutOfBandManagementPowerAction", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r IssueOutOfBandManagementPowerActionResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	// If we have a async client, we need to wait for the async result
	if s.cs.async {
		b, err := s.cs.GetAsyncJobResult(r.JobID, s.cs.timeout)
		if err != nil {
			if err == AsyncTimeoutErr {
				return &r, err
			}
			return nil, err
		}

		b, err = getRawValue(b)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type IssueOutOfBandManagementPowerActionResponse struct {
	Action      string `json:"action"`
	Address     string `json:"address"`
	Description string `json:"description"`
	Driver      string `json:"driver"`
	Enabled     bool   `json:"enabled"`
	Hostid      string `json:"hostid"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Password    string `json:"password"`
	Port        string `json:"port"`
	Powerstate  string `json:"powerstate"`
	Status      bool   `json:"status"`
	Username    string `json:"username"`
}

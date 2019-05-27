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

type AddNuageVspDeviceParams struct {
	p map[string]interface{}
}

func (p *AddNuageVspDeviceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["apiversion"]; found {
		u.Set("apiversion", v.(string))
	}
	if v, found := p.p["hostname"]; found {
		u.Set("hostname", v.(string))
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["port"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("port", vv)
	}
	if v, found := p.p["retrycount"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("retrycount", vv)
	}
	if v, found := p.p["retryinterval"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("retryinterval", vv)
	}
	if v, found := p.p["username"]; found {
		u.Set("username", v.(string))
	}
	return u
}

func (p *AddNuageVspDeviceParams) SetApiversion(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["apiversion"] = v
	return
}

func (p *AddNuageVspDeviceParams) SetHostname(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostname"] = v
	return
}

func (p *AddNuageVspDeviceParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

func (p *AddNuageVspDeviceParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *AddNuageVspDeviceParams) SetPort(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["port"] = v
	return
}

func (p *AddNuageVspDeviceParams) SetRetrycount(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["retrycount"] = v
	return
}

func (p *AddNuageVspDeviceParams) SetRetryinterval(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["retryinterval"] = v
	return
}

func (p *AddNuageVspDeviceParams) SetUsername(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["username"] = v
	return
}

// You should always use this function to get a new AddNuageVspDeviceParams instance,
// as then you are sure you have configured all required params
func (s *NuageVSPService) NewAddNuageVspDeviceParams(hostname string, password string, physicalnetworkid string, port int, username string) *AddNuageVspDeviceParams {
	p := &AddNuageVspDeviceParams{}
	p.p = make(map[string]interface{})
	p.p["hostname"] = hostname
	p.p["password"] = password
	p.p["physicalnetworkid"] = physicalnetworkid
	p.p["port"] = port
	p.p["username"] = username
	return p
}

// Adds a Nuage VSP device
func (s *NuageVSPService) AddNuageVspDevice(p *AddNuageVspDeviceParams) (*AddNuageVspDeviceResponse, error) {
	resp, err := s.cs.newRequest("addNuageVspDevice", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddNuageVspDeviceResponse
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

type AddNuageVspDeviceResponse struct {
	Apiversion        string `json:"apiversion"`
	Cmsid             string `json:"cmsid"`
	Hostname          string `json:"hostname"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
	Nuagedevicename   string `json:"nuagedevicename"`
	Physicalnetworkid string `json:"physicalnetworkid"`
	Port              int    `json:"port"`
	Provider          string `json:"provider"`
	Retrycount        int    `json:"retrycount"`
	Retryinterval     int64  `json:"retryinterval"`
	Vspdeviceid       string `json:"vspdeviceid"`
}

type DeleteNuageVspDeviceParams struct {
	p map[string]interface{}
}

func (p *DeleteNuageVspDeviceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["vspdeviceid"]; found {
		u.Set("vspdeviceid", v.(string))
	}
	return u
}

func (p *DeleteNuageVspDeviceParams) SetVspdeviceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vspdeviceid"] = v
	return
}

// You should always use this function to get a new DeleteNuageVspDeviceParams instance,
// as then you are sure you have configured all required params
func (s *NuageVSPService) NewDeleteNuageVspDeviceParams(vspdeviceid string) *DeleteNuageVspDeviceParams {
	p := &DeleteNuageVspDeviceParams{}
	p.p = make(map[string]interface{})
	p.p["vspdeviceid"] = vspdeviceid
	return p
}

// delete a nuage vsp device
func (s *NuageVSPService) DeleteNuageVspDevice(p *DeleteNuageVspDeviceParams) (*DeleteNuageVspDeviceResponse, error) {
	resp, err := s.cs.newRequest("deleteNuageVspDevice", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteNuageVspDeviceResponse
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

		if err := json.Unmarshal(b, &r); err != nil {
			return nil, err
		}
	}

	return &r, nil
}

type DeleteNuageVspDeviceResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type ListNuageVspDevicesParams struct {
	p map[string]interface{}
}

func (p *ListNuageVspDevicesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["page"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("page", vv)
	}
	if v, found := p.p["pagesize"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("pagesize", vv)
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["vspdeviceid"]; found {
		u.Set("vspdeviceid", v.(string))
	}
	return u
}

func (p *ListNuageVspDevicesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListNuageVspDevicesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListNuageVspDevicesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListNuageVspDevicesParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *ListNuageVspDevicesParams) SetVspdeviceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vspdeviceid"] = v
	return
}

// You should always use this function to get a new ListNuageVspDevicesParams instance,
// as then you are sure you have configured all required params
func (s *NuageVSPService) NewListNuageVspDevicesParams() *ListNuageVspDevicesParams {
	p := &ListNuageVspDevicesParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists Nuage VSP devices
func (s *NuageVSPService) ListNuageVspDevices(p *ListNuageVspDevicesParams) (*ListNuageVspDevicesResponse, error) {
	resp, err := s.cs.newRequest("listNuageVspDevices", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListNuageVspDevicesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListNuageVspDevicesResponse struct {
	Count           int               `json:"count"`
	NuageVspDevices []*NuageVspDevice `json:"nuagevspdevice"`
}

type NuageVspDevice struct {
	Apiversion        string `json:"apiversion"`
	Cmsid             string `json:"cmsid"`
	Hostname          string `json:"hostname"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
	Nuagedevicename   string `json:"nuagedevicename"`
	Physicalnetworkid string `json:"physicalnetworkid"`
	Port              int    `json:"port"`
	Provider          string `json:"provider"`
	Retrycount        int    `json:"retrycount"`
	Retryinterval     int64  `json:"retryinterval"`
	Vspdeviceid       string `json:"vspdeviceid"`
}

type UpdateNuageVspDeviceParams struct {
	p map[string]interface{}
}

func (p *UpdateNuageVspDeviceParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["apiversion"]; found {
		u.Set("apiversion", v.(string))
	}
	if v, found := p.p["hostname"]; found {
		u.Set("hostname", v.(string))
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["port"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("port", vv)
	}
	if v, found := p.p["retrycount"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("retrycount", vv)
	}
	if v, found := p.p["retryinterval"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("retryinterval", vv)
	}
	if v, found := p.p["username"]; found {
		u.Set("username", v.(string))
	}
	return u
}

func (p *UpdateNuageVspDeviceParams) SetApiversion(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["apiversion"] = v
	return
}

func (p *UpdateNuageVspDeviceParams) SetHostname(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostname"] = v
	return
}

func (p *UpdateNuageVspDeviceParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

func (p *UpdateNuageVspDeviceParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *UpdateNuageVspDeviceParams) SetPort(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["port"] = v
	return
}

func (p *UpdateNuageVspDeviceParams) SetRetrycount(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["retrycount"] = v
	return
}

func (p *UpdateNuageVspDeviceParams) SetRetryinterval(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["retryinterval"] = v
	return
}

func (p *UpdateNuageVspDeviceParams) SetUsername(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["username"] = v
	return
}

// You should always use this function to get a new UpdateNuageVspDeviceParams instance,
// as then you are sure you have configured all required params
func (s *NuageVSPService) NewUpdateNuageVspDeviceParams(physicalnetworkid string) *UpdateNuageVspDeviceParams {
	p := &UpdateNuageVspDeviceParams{}
	p.p = make(map[string]interface{})
	p.p["physicalnetworkid"] = physicalnetworkid
	return p
}

// Update a Nuage VSP device
func (s *NuageVSPService) UpdateNuageVspDevice(p *UpdateNuageVspDeviceParams) (*UpdateNuageVspDeviceResponse, error) {
	resp, err := s.cs.newRequest("updateNuageVspDevice", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateNuageVspDeviceResponse
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

type UpdateNuageVspDeviceResponse struct {
	Apiversion        string `json:"apiversion"`
	Cmsid             string `json:"cmsid"`
	Hostname          string `json:"hostname"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
	Nuagedevicename   string `json:"nuagedevicename"`
	Physicalnetworkid string `json:"physicalnetworkid"`
	Port              int    `json:"port"`
	Provider          string `json:"provider"`
	Retrycount        int    `json:"retrycount"`
	Retryinterval     int64  `json:"retryinterval"`
	Vspdeviceid       string `json:"vspdeviceid"`
}

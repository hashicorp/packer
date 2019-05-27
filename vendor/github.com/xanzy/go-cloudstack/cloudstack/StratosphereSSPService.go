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

type AddStratosphereSspParams struct {
	p map[string]interface{}
}

func (p *AddStratosphereSspParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	if v, found := p.p["tenantuuid"]; found {
		u.Set("tenantuuid", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["username"]; found {
		u.Set("username", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *AddStratosphereSspParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *AddStratosphereSspParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

func (p *AddStratosphereSspParams) SetTenantuuid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tenantuuid"] = v
	return
}

func (p *AddStratosphereSspParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *AddStratosphereSspParams) SetUsername(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["username"] = v
	return
}

func (p *AddStratosphereSspParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new AddStratosphereSspParams instance,
// as then you are sure you have configured all required params
func (s *StratosphereSSPService) NewAddStratosphereSspParams(name string, url string, zoneid string) *AddStratosphereSspParams {
	p := &AddStratosphereSspParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	p.p["url"] = url
	p.p["zoneid"] = zoneid
	return p
}

// Adds stratosphere ssp server
func (s *StratosphereSSPService) AddStratosphereSsp(p *AddStratosphereSspParams) (*AddStratosphereSspResponse, error) {
	resp, err := s.cs.newRequest("addStratosphereSsp", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddStratosphereSspResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type AddStratosphereSspResponse struct {
	Hostid    string `json:"hostid"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Zoneid    string `json:"zoneid"`
}

type DeleteStratosphereSspParams struct {
	p map[string]interface{}
}

func (p *DeleteStratosphereSspParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	return u
}

func (p *DeleteStratosphereSspParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

// You should always use this function to get a new DeleteStratosphereSspParams instance,
// as then you are sure you have configured all required params
func (s *StratosphereSSPService) NewDeleteStratosphereSspParams(hostid string) *DeleteStratosphereSspParams {
	p := &DeleteStratosphereSspParams{}
	p.p = make(map[string]interface{})
	p.p["hostid"] = hostid
	return p
}

// Removes stratosphere ssp server
func (s *StratosphereSSPService) DeleteStratosphereSsp(p *DeleteStratosphereSspParams) (*DeleteStratosphereSspResponse, error) {
	resp, err := s.cs.newRequest("deleteStratosphereSsp", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteStratosphereSspResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DeleteStratosphereSspResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *DeleteStratosphereSspResponse) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}

	if success, ok := m["success"].(string); ok {
		m["success"] = success == "true"
		b, err = json.Marshal(m)
		if err != nil {
			return err
		}
	}

	if ostypeid, ok := m["ostypeid"].(float64); ok {
		m["ostypeid"] = strconv.Itoa(int(ostypeid))
		b, err = json.Marshal(m)
		if err != nil {
			return err
		}
	}

	type alias DeleteStratosphereSspResponse
	return json.Unmarshal(b, (*alias)(r))
}

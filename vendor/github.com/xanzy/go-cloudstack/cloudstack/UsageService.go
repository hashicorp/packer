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
	"fmt"
	"net/url"
	"strconv"
)

type AddTrafficMonitorParams struct {
	p map[string]interface{}
}

func (p *AddTrafficMonitorParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["excludezones"]; found {
		u.Set("excludezones", v.(string))
	}
	if v, found := p.p["includezones"]; found {
		u.Set("includezones", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *AddTrafficMonitorParams) SetExcludezones(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["excludezones"] = v
	return
}

func (p *AddTrafficMonitorParams) SetIncludezones(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["includezones"] = v
	return
}

func (p *AddTrafficMonitorParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *AddTrafficMonitorParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new AddTrafficMonitorParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewAddTrafficMonitorParams(url string, zoneid string) *AddTrafficMonitorParams {
	p := &AddTrafficMonitorParams{}
	p.p = make(map[string]interface{})
	p.p["url"] = url
	p.p["zoneid"] = zoneid
	return p
}

// Adds Traffic Monitor Host for Direct Network Usage
func (s *UsageService) AddTrafficMonitor(p *AddTrafficMonitorParams) (*AddTrafficMonitorResponse, error) {
	resp, err := s.cs.newRequest("addTrafficMonitor", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddTrafficMonitorResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type AddTrafficMonitorResponse struct {
	Id         string `json:"id"`
	Ipaddress  string `json:"ipaddress"`
	JobID      string `json:"jobid"`
	Jobstatus  int    `json:"jobstatus"`
	Numretries string `json:"numretries"`
	Timeout    string `json:"timeout"`
	Zoneid     string `json:"zoneid"`
}

type AddTrafficTypeParams struct {
	p map[string]interface{}
}

func (p *AddTrafficTypeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hypervnetworklabel"]; found {
		u.Set("hypervnetworklabel", v.(string))
	}
	if v, found := p.p["isolationmethod"]; found {
		u.Set("isolationmethod", v.(string))
	}
	if v, found := p.p["kvmnetworklabel"]; found {
		u.Set("kvmnetworklabel", v.(string))
	}
	if v, found := p.p["ovm3networklabel"]; found {
		u.Set("ovm3networklabel", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["traffictype"]; found {
		u.Set("traffictype", v.(string))
	}
	if v, found := p.p["vlan"]; found {
		u.Set("vlan", v.(string))
	}
	if v, found := p.p["vmwarenetworklabel"]; found {
		u.Set("vmwarenetworklabel", v.(string))
	}
	if v, found := p.p["xennetworklabel"]; found {
		u.Set("xennetworklabel", v.(string))
	}
	return u
}

func (p *AddTrafficTypeParams) SetHypervnetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hypervnetworklabel"] = v
	return
}

func (p *AddTrafficTypeParams) SetIsolationmethod(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isolationmethod"] = v
	return
}

func (p *AddTrafficTypeParams) SetKvmnetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["kvmnetworklabel"] = v
	return
}

func (p *AddTrafficTypeParams) SetOvm3networklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ovm3networklabel"] = v
	return
}

func (p *AddTrafficTypeParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *AddTrafficTypeParams) SetTraffictype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["traffictype"] = v
	return
}

func (p *AddTrafficTypeParams) SetVlan(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

func (p *AddTrafficTypeParams) SetVmwarenetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vmwarenetworklabel"] = v
	return
}

func (p *AddTrafficTypeParams) SetXennetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["xennetworklabel"] = v
	return
}

// You should always use this function to get a new AddTrafficTypeParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewAddTrafficTypeParams(physicalnetworkid string, traffictype string) *AddTrafficTypeParams {
	p := &AddTrafficTypeParams{}
	p.p = make(map[string]interface{})
	p.p["physicalnetworkid"] = physicalnetworkid
	p.p["traffictype"] = traffictype
	return p
}

// Adds traffic type to a physical network
func (s *UsageService) AddTrafficType(p *AddTrafficTypeParams) (*AddTrafficTypeResponse, error) {
	resp, err := s.cs.newRequest("addTrafficType", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddTrafficTypeResponse
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

type AddTrafficTypeResponse struct {
	Hypervnetworklabel string `json:"hypervnetworklabel"`
	Id                 string `json:"id"`
	JobID              string `json:"jobid"`
	Jobstatus          int    `json:"jobstatus"`
	Kvmnetworklabel    string `json:"kvmnetworklabel"`
	Ovm3networklabel   string `json:"ovm3networklabel"`
	Physicalnetworkid  string `json:"physicalnetworkid"`
	Traffictype        string `json:"traffictype"`
	Vmwarenetworklabel string `json:"vmwarenetworklabel"`
	Xennetworklabel    string `json:"xennetworklabel"`
}

type DeleteTrafficMonitorParams struct {
	p map[string]interface{}
}

func (p *DeleteTrafficMonitorParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteTrafficMonitorParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteTrafficMonitorParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewDeleteTrafficMonitorParams(id string) *DeleteTrafficMonitorParams {
	p := &DeleteTrafficMonitorParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes an traffic monitor host.
func (s *UsageService) DeleteTrafficMonitor(p *DeleteTrafficMonitorParams) (*DeleteTrafficMonitorResponse, error) {
	resp, err := s.cs.newRequest("deleteTrafficMonitor", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteTrafficMonitorResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DeleteTrafficMonitorResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *DeleteTrafficMonitorResponse) UnmarshalJSON(b []byte) error {
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

	type alias DeleteTrafficMonitorResponse
	return json.Unmarshal(b, (*alias)(r))
}

type DeleteTrafficTypeParams struct {
	p map[string]interface{}
}

func (p *DeleteTrafficTypeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteTrafficTypeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteTrafficTypeParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewDeleteTrafficTypeParams(id string) *DeleteTrafficTypeParams {
	p := &DeleteTrafficTypeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes traffic type of a physical network
func (s *UsageService) DeleteTrafficType(p *DeleteTrafficTypeParams) (*DeleteTrafficTypeResponse, error) {
	resp, err := s.cs.newRequest("deleteTrafficType", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteTrafficTypeResponse
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

type DeleteTrafficTypeResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type GenerateUsageRecordsParams struct {
	p map[string]interface{}
}

func (p *GenerateUsageRecordsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["enddate"]; found {
		u.Set("enddate", v.(string))
	}
	if v, found := p.p["startdate"]; found {
		u.Set("startdate", v.(string))
	}
	return u
}

func (p *GenerateUsageRecordsParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *GenerateUsageRecordsParams) SetEnddate(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["enddate"] = v
	return
}

func (p *GenerateUsageRecordsParams) SetStartdate(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["startdate"] = v
	return
}

// You should always use this function to get a new GenerateUsageRecordsParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewGenerateUsageRecordsParams(enddate string, startdate string) *GenerateUsageRecordsParams {
	p := &GenerateUsageRecordsParams{}
	p.p = make(map[string]interface{})
	p.p["enddate"] = enddate
	p.p["startdate"] = startdate
	return p
}

// Generates usage records. This will generate records only if there any records to be generated, i.e if the scheduled usage job was not run or failed
func (s *UsageService) GenerateUsageRecords(p *GenerateUsageRecordsParams) (*GenerateUsageRecordsResponse, error) {
	resp, err := s.cs.newRequest("generateUsageRecords", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r GenerateUsageRecordsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type GenerateUsageRecordsResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *GenerateUsageRecordsResponse) UnmarshalJSON(b []byte) error {
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

	type alias GenerateUsageRecordsResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ListTrafficMonitorsParams struct {
	p map[string]interface{}
}

func (p *ListTrafficMonitorsParams) toURLValues() url.Values {
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
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListTrafficMonitorsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListTrafficMonitorsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListTrafficMonitorsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListTrafficMonitorsParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListTrafficMonitorsParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewListTrafficMonitorsParams(zoneid string) *ListTrafficMonitorsParams {
	p := &ListTrafficMonitorsParams{}
	p.p = make(map[string]interface{})
	p.p["zoneid"] = zoneid
	return p
}

// List traffic monitor Hosts.
func (s *UsageService) ListTrafficMonitors(p *ListTrafficMonitorsParams) (*ListTrafficMonitorsResponse, error) {
	resp, err := s.cs.newRequest("listTrafficMonitors", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListTrafficMonitorsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListTrafficMonitorsResponse struct {
	Count           int               `json:"count"`
	TrafficMonitors []*TrafficMonitor `json:"trafficmonitor"`
}

type TrafficMonitor struct {
	Id         string `json:"id"`
	Ipaddress  string `json:"ipaddress"`
	JobID      string `json:"jobid"`
	Jobstatus  int    `json:"jobstatus"`
	Numretries string `json:"numretries"`
	Timeout    string `json:"timeout"`
	Zoneid     string `json:"zoneid"`
}

type ListTrafficTypeImplementorsParams struct {
	p map[string]interface{}
}

func (p *ListTrafficTypeImplementorsParams) toURLValues() url.Values {
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
	if v, found := p.p["traffictype"]; found {
		u.Set("traffictype", v.(string))
	}
	return u
}

func (p *ListTrafficTypeImplementorsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListTrafficTypeImplementorsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListTrafficTypeImplementorsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListTrafficTypeImplementorsParams) SetTraffictype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["traffictype"] = v
	return
}

// You should always use this function to get a new ListTrafficTypeImplementorsParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewListTrafficTypeImplementorsParams() *ListTrafficTypeImplementorsParams {
	p := &ListTrafficTypeImplementorsParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists implementors of implementor of a network traffic type or implementors of all network traffic types
func (s *UsageService) ListTrafficTypeImplementors(p *ListTrafficTypeImplementorsParams) (*ListTrafficTypeImplementorsResponse, error) {
	resp, err := s.cs.newRequest("listTrafficTypeImplementors", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListTrafficTypeImplementorsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListTrafficTypeImplementorsResponse struct {
	Count                   int                       `json:"count"`
	TrafficTypeImplementors []*TrafficTypeImplementor `json:"traffictypeimplementor"`
}

type TrafficTypeImplementor struct {
	JobID                  string `json:"jobid"`
	Jobstatus              int    `json:"jobstatus"`
	Traffictype            string `json:"traffictype"`
	Traffictypeimplementor string `json:"traffictypeimplementor"`
}

type ListTrafficTypesParams struct {
	p map[string]interface{}
}

func (p *ListTrafficTypesParams) toURLValues() url.Values {
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
	return u
}

func (p *ListTrafficTypesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListTrafficTypesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListTrafficTypesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListTrafficTypesParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

// You should always use this function to get a new ListTrafficTypesParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewListTrafficTypesParams(physicalnetworkid string) *ListTrafficTypesParams {
	p := &ListTrafficTypesParams{}
	p.p = make(map[string]interface{})
	p.p["physicalnetworkid"] = physicalnetworkid
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *UsageService) GetTrafficTypeID(keyword string, physicalnetworkid string, opts ...OptionFunc) (string, int, error) {
	p := &ListTrafficTypesParams{}
	p.p = make(map[string]interface{})

	p.p["keyword"] = keyword
	p.p["physicalnetworkid"] = physicalnetworkid

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListTrafficTypes(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", keyword, l)
	}

	if l.Count == 1 {
		return l.TrafficTypes[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.TrafficTypes {
			if v.Name == keyword {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", keyword, l)
}

// Lists traffic types of a given physical network.
func (s *UsageService) ListTrafficTypes(p *ListTrafficTypesParams) (*ListTrafficTypesResponse, error) {
	resp, err := s.cs.newRequest("listTrafficTypes", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListTrafficTypesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListTrafficTypesResponse struct {
	Count        int            `json:"count"`
	TrafficTypes []*TrafficType `json:"traffictype"`
}

type TrafficType struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	JobID                        string   `json:"jobid"`
	Jobstatus                    int      `json:"jobstatus"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type ListUsageRecordsParams struct {
	p map[string]interface{}
}

func (p *ListUsageRecordsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["accountid"]; found {
		u.Set("accountid", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["enddate"]; found {
		u.Set("enddate", v.(string))
	}
	if v, found := p.p["includetags"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("includetags", vv)
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
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["startdate"]; found {
		u.Set("startdate", v.(string))
	}
	if v, found := p.p["type"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("type", vv)
	}
	if v, found := p.p["usageid"]; found {
		u.Set("usageid", v.(string))
	}
	return u
}

func (p *ListUsageRecordsParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListUsageRecordsParams) SetAccountid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["accountid"] = v
	return
}

func (p *ListUsageRecordsParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListUsageRecordsParams) SetEnddate(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["enddate"] = v
	return
}

func (p *ListUsageRecordsParams) SetIncludetags(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["includetags"] = v
	return
}

func (p *ListUsageRecordsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListUsageRecordsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListUsageRecordsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListUsageRecordsParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListUsageRecordsParams) SetStartdate(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["startdate"] = v
	return
}

func (p *ListUsageRecordsParams) SetType(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

func (p *ListUsageRecordsParams) SetUsageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["usageid"] = v
	return
}

// You should always use this function to get a new ListUsageRecordsParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewListUsageRecordsParams(enddate string, startdate string) *ListUsageRecordsParams {
	p := &ListUsageRecordsParams{}
	p.p = make(map[string]interface{})
	p.p["enddate"] = enddate
	p.p["startdate"] = startdate
	return p
}

// Lists usage records for accounts
func (s *UsageService) ListUsageRecords(p *ListUsageRecordsParams) (*ListUsageRecordsResponse, error) {
	resp, err := s.cs.newRequest("listUsageRecords", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListUsageRecordsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListUsageRecordsResponse struct {
	Count        int            `json:"count"`
	UsageRecords []*UsageRecord `json:"usagerecord"`
}

type UsageRecord struct {
	Account          string `json:"account"`
	Accountid        string `json:"accountid"`
	Cpunumber        int64  `json:"cpunumber"`
	Cpuspeed         int64  `json:"cpuspeed"`
	Description      string `json:"description"`
	Domain           string `json:"domain"`
	Domainid         string `json:"domainid"`
	Enddate          string `json:"enddate"`
	Isdefault        bool   `json:"isdefault"`
	Issourcenat      bool   `json:"issourcenat"`
	Issystem         bool   `json:"issystem"`
	JobID            string `json:"jobid"`
	Jobstatus        int    `json:"jobstatus"`
	Memory           int64  `json:"memory"`
	Name             string `json:"name"`
	Networkid        string `json:"networkid"`
	Offeringid       string `json:"offeringid"`
	Project          string `json:"project"`
	Projectid        string `json:"projectid"`
	Rawusage         string `json:"rawusage"`
	Size             int64  `json:"size"`
	Startdate        string `json:"startdate"`
	Tags             []Tags `json:"tags"`
	Templateid       string `json:"templateid"`
	Type             string `json:"type"`
	Usage            string `json:"usage"`
	Usageid          string `json:"usageid"`
	Usagetype        int    `json:"usagetype"`
	Virtualmachineid string `json:"virtualmachineid"`
	Virtualsize      int64  `json:"virtualsize"`
	Zoneid           string `json:"zoneid"`
}

type ListUsageTypesParams struct {
	p map[string]interface{}
}

func (p *ListUsageTypesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	return u
}

// You should always use this function to get a new ListUsageTypesParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewListUsageTypesParams() *ListUsageTypesParams {
	p := &ListUsageTypesParams{}
	p.p = make(map[string]interface{})
	return p
}

// List Usage Types
func (s *UsageService) ListUsageTypes(p *ListUsageTypesParams) (*ListUsageTypesResponse, error) {
	resp, err := s.cs.newRequest("listUsageTypes", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListUsageTypesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListUsageTypesResponse struct {
	Count      int          `json:"count"`
	UsageTypes []*UsageType `json:"usagetype"`
}

type UsageType struct {
	Description string `json:"description"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Usagetypeid int    `json:"usagetypeid"`
}

type RemoveRawUsageRecordsParams struct {
	p map[string]interface{}
}

func (p *RemoveRawUsageRecordsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["interval"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("interval", vv)
	}
	return u
}

func (p *RemoveRawUsageRecordsParams) SetInterval(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["interval"] = v
	return
}

// You should always use this function to get a new RemoveRawUsageRecordsParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewRemoveRawUsageRecordsParams(interval int) *RemoveRawUsageRecordsParams {
	p := &RemoveRawUsageRecordsParams{}
	p.p = make(map[string]interface{})
	p.p["interval"] = interval
	return p
}

// Safely removes raw records from cloud_usage table
func (s *UsageService) RemoveRawUsageRecords(p *RemoveRawUsageRecordsParams) (*RemoveRawUsageRecordsResponse, error) {
	resp, err := s.cs.newRequest("removeRawUsageRecords", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r RemoveRawUsageRecordsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type RemoveRawUsageRecordsResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *RemoveRawUsageRecordsResponse) UnmarshalJSON(b []byte) error {
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

	type alias RemoveRawUsageRecordsResponse
	return json.Unmarshal(b, (*alias)(r))
}

type UpdateTrafficTypeParams struct {
	p map[string]interface{}
}

func (p *UpdateTrafficTypeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hypervnetworklabel"]; found {
		u.Set("hypervnetworklabel", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["kvmnetworklabel"]; found {
		u.Set("kvmnetworklabel", v.(string))
	}
	if v, found := p.p["ovm3networklabel"]; found {
		u.Set("ovm3networklabel", v.(string))
	}
	if v, found := p.p["vmwarenetworklabel"]; found {
		u.Set("vmwarenetworklabel", v.(string))
	}
	if v, found := p.p["xennetworklabel"]; found {
		u.Set("xennetworklabel", v.(string))
	}
	return u
}

func (p *UpdateTrafficTypeParams) SetHypervnetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hypervnetworklabel"] = v
	return
}

func (p *UpdateTrafficTypeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateTrafficTypeParams) SetKvmnetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["kvmnetworklabel"] = v
	return
}

func (p *UpdateTrafficTypeParams) SetOvm3networklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ovm3networklabel"] = v
	return
}

func (p *UpdateTrafficTypeParams) SetVmwarenetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vmwarenetworklabel"] = v
	return
}

func (p *UpdateTrafficTypeParams) SetXennetworklabel(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["xennetworklabel"] = v
	return
}

// You should always use this function to get a new UpdateTrafficTypeParams instance,
// as then you are sure you have configured all required params
func (s *UsageService) NewUpdateTrafficTypeParams(id string) *UpdateTrafficTypeParams {
	p := &UpdateTrafficTypeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates traffic type of a physical network
func (s *UsageService) UpdateTrafficType(p *UpdateTrafficTypeParams) (*UpdateTrafficTypeResponse, error) {
	resp, err := s.cs.newRequest("updateTrafficType", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateTrafficTypeResponse
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

type UpdateTrafficTypeResponse struct {
	Hypervnetworklabel string `json:"hypervnetworklabel"`
	Id                 string `json:"id"`
	JobID              string `json:"jobid"`
	Jobstatus          int    `json:"jobstatus"`
	Kvmnetworklabel    string `json:"kvmnetworklabel"`
	Ovm3networklabel   string `json:"ovm3networklabel"`
	Physicalnetworkid  string `json:"physicalnetworkid"`
	Traffictype        string `json:"traffictype"`
	Vmwarenetworklabel string `json:"vmwarenetworklabel"`
	Xennetworklabel    string `json:"xennetworklabel"`
}

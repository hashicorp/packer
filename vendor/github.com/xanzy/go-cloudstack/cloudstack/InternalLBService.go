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
	"strings"
)

type ConfigureInternalLoadBalancerElementParams struct {
	p map[string]interface{}
}

func (p *ConfigureInternalLoadBalancerElementParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["enabled"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("enabled", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *ConfigureInternalLoadBalancerElementParams) SetEnabled(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["enabled"] = v
	return
}

func (p *ConfigureInternalLoadBalancerElementParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new ConfigureInternalLoadBalancerElementParams instance,
// as then you are sure you have configured all required params
func (s *InternalLBService) NewConfigureInternalLoadBalancerElementParams(enabled bool, id string) *ConfigureInternalLoadBalancerElementParams {
	p := &ConfigureInternalLoadBalancerElementParams{}
	p.p = make(map[string]interface{})
	p.p["enabled"] = enabled
	p.p["id"] = id
	return p
}

// Configures an Internal Load Balancer element.
func (s *InternalLBService) ConfigureInternalLoadBalancerElement(p *ConfigureInternalLoadBalancerElementParams) (*InternalLoadBalancerElementResponse, error) {
	resp, err := s.cs.newRequest("configureInternalLoadBalancerElement", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r InternalLoadBalancerElementResponse
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

type InternalLoadBalancerElementResponse struct {
	Enabled   bool   `json:"enabled"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Nspid     string `json:"nspid"`
}

type CreateInternalLoadBalancerElementParams struct {
	p map[string]interface{}
}

func (p *CreateInternalLoadBalancerElementParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["nspid"]; found {
		u.Set("nspid", v.(string))
	}
	return u
}

func (p *CreateInternalLoadBalancerElementParams) SetNspid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["nspid"] = v
	return
}

// You should always use this function to get a new CreateInternalLoadBalancerElementParams instance,
// as then you are sure you have configured all required params
func (s *InternalLBService) NewCreateInternalLoadBalancerElementParams(nspid string) *CreateInternalLoadBalancerElementParams {
	p := &CreateInternalLoadBalancerElementParams{}
	p.p = make(map[string]interface{})
	p.p["nspid"] = nspid
	return p
}

// Create an Internal Load Balancer element.
func (s *InternalLBService) CreateInternalLoadBalancerElement(p *CreateInternalLoadBalancerElementParams) (*CreateInternalLoadBalancerElementResponse, error) {
	resp, err := s.cs.newRequest("createInternalLoadBalancerElement", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateInternalLoadBalancerElementResponse
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

type CreateInternalLoadBalancerElementResponse struct {
	Enabled   bool   `json:"enabled"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Nspid     string `json:"nspid"`
}

type ListInternalLoadBalancerElementsParams struct {
	p map[string]interface{}
}

func (p *ListInternalLoadBalancerElementsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["enabled"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("enabled", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["nspid"]; found {
		u.Set("nspid", v.(string))
	}
	if v, found := p.p["page"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("page", vv)
	}
	if v, found := p.p["pagesize"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("pagesize", vv)
	}
	return u
}

func (p *ListInternalLoadBalancerElementsParams) SetEnabled(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["enabled"] = v
	return
}

func (p *ListInternalLoadBalancerElementsParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListInternalLoadBalancerElementsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListInternalLoadBalancerElementsParams) SetNspid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["nspid"] = v
	return
}

func (p *ListInternalLoadBalancerElementsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListInternalLoadBalancerElementsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

// You should always use this function to get a new ListInternalLoadBalancerElementsParams instance,
// as then you are sure you have configured all required params
func (s *InternalLBService) NewListInternalLoadBalancerElementsParams() *ListInternalLoadBalancerElementsParams {
	p := &ListInternalLoadBalancerElementsParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *InternalLBService) GetInternalLoadBalancerElementByID(id string, opts ...OptionFunc) (*InternalLoadBalancerElement, int, error) {
	p := &ListInternalLoadBalancerElementsParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListInternalLoadBalancerElements(p)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", id)) {
			return nil, 0, fmt.Errorf("No match found for %s: %+v", id, l)
		}
		return nil, -1, err
	}

	if l.Count == 0 {
		return nil, l.Count, fmt.Errorf("No match found for %s: %+v", id, l)
	}

	if l.Count == 1 {
		return l.InternalLoadBalancerElements[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for InternalLoadBalancerElement UUID: %s!", id)
}

// Lists all available Internal Load Balancer elements.
func (s *InternalLBService) ListInternalLoadBalancerElements(p *ListInternalLoadBalancerElementsParams) (*ListInternalLoadBalancerElementsResponse, error) {
	resp, err := s.cs.newRequest("listInternalLoadBalancerElements", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListInternalLoadBalancerElementsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListInternalLoadBalancerElementsResponse struct {
	Count                        int                            `json:"count"`
	InternalLoadBalancerElements []*InternalLoadBalancerElement `json:"internalloadbalancerelement"`
}

type InternalLoadBalancerElement struct {
	Enabled   bool   `json:"enabled"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Nspid     string `json:"nspid"`
}

type ListInternalLoadBalancerVMsParams struct {
	p map[string]interface{}
}

func (p *ListInternalLoadBalancerVMsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["forvpc"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("forvpc", vv)
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["isrecursive"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("isrecursive", vv)
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["listall"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("listall", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["networkid"]; found {
		u.Set("networkid", v.(string))
	}
	if v, found := p.p["page"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("page", vv)
	}
	if v, found := p.p["pagesize"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("pagesize", vv)
	}
	if v, found := p.p["podid"]; found {
		u.Set("podid", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["vpcid"]; found {
		u.Set("vpcid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListInternalLoadBalancerVMsParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetForvpc(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forvpc"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetNetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkid"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetPodid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["podid"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetVpcid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcid"] = v
	return
}

func (p *ListInternalLoadBalancerVMsParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListInternalLoadBalancerVMsParams instance,
// as then you are sure you have configured all required params
func (s *InternalLBService) NewListInternalLoadBalancerVMsParams() *ListInternalLoadBalancerVMsParams {
	p := &ListInternalLoadBalancerVMsParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *InternalLBService) GetInternalLoadBalancerVMID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListInternalLoadBalancerVMsParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListInternalLoadBalancerVMs(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.InternalLoadBalancerVMs[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.InternalLoadBalancerVMs {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *InternalLBService) GetInternalLoadBalancerVMByName(name string, opts ...OptionFunc) (*InternalLoadBalancerVM, int, error) {
	id, count, err := s.GetInternalLoadBalancerVMID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetInternalLoadBalancerVMByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *InternalLBService) GetInternalLoadBalancerVMByID(id string, opts ...OptionFunc) (*InternalLoadBalancerVM, int, error) {
	p := &ListInternalLoadBalancerVMsParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListInternalLoadBalancerVMs(p)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", id)) {
			return nil, 0, fmt.Errorf("No match found for %s: %+v", id, l)
		}
		return nil, -1, err
	}

	if l.Count == 0 {
		return nil, l.Count, fmt.Errorf("No match found for %s: %+v", id, l)
	}

	if l.Count == 1 {
		return l.InternalLoadBalancerVMs[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for InternalLoadBalancerVM UUID: %s!", id)
}

// List internal LB VMs.
func (s *InternalLBService) ListInternalLoadBalancerVMs(p *ListInternalLoadBalancerVMsParams) (*ListInternalLoadBalancerVMsResponse, error) {
	resp, err := s.cs.newRequest("listInternalLoadBalancerVMs", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListInternalLoadBalancerVMsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListInternalLoadBalancerVMsResponse struct {
	Count                   int                       `json:"count"`
	InternalLoadBalancerVMs []*InternalLoadBalancerVM `json:"internalloadbalancervm"`
}

type InternalLoadBalancerVM struct {
	Account             string `json:"account"`
	Created             string `json:"created"`
	Dns1                string `json:"dns1"`
	Dns2                string `json:"dns2"`
	Domain              string `json:"domain"`
	Domainid            string `json:"domainid"`
	Gateway             string `json:"gateway"`
	Guestipaddress      string `json:"guestipaddress"`
	Guestmacaddress     string `json:"guestmacaddress"`
	Guestnetmask        string `json:"guestnetmask"`
	Guestnetworkid      string `json:"guestnetworkid"`
	Guestnetworkname    string `json:"guestnetworkname"`
	Hostid              string `json:"hostid"`
	Hostname            string `json:"hostname"`
	Hypervisor          string `json:"hypervisor"`
	Id                  string `json:"id"`
	Ip6dns1             string `json:"ip6dns1"`
	Ip6dns2             string `json:"ip6dns2"`
	Isredundantrouter   bool   `json:"isredundantrouter"`
	JobID               string `json:"jobid"`
	Jobstatus           int    `json:"jobstatus"`
	Linklocalip         string `json:"linklocalip"`
	Linklocalmacaddress string `json:"linklocalmacaddress"`
	Linklocalnetmask    string `json:"linklocalnetmask"`
	Linklocalnetworkid  string `json:"linklocalnetworkid"`
	Name                string `json:"name"`
	Networkdomain       string `json:"networkdomain"`
	Nic                 []Nic  `json:"nic"`
	Podid               string `json:"podid"`
	Project             string `json:"project"`
	Projectid           string `json:"projectid"`
	Publicip            string `json:"publicip"`
	Publicmacaddress    string `json:"publicmacaddress"`
	Publicnetmask       string `json:"publicnetmask"`
	Publicnetworkid     string `json:"publicnetworkid"`
	Redundantstate      string `json:"redundantstate"`
	Requiresupgrade     bool   `json:"requiresupgrade"`
	Role                string `json:"role"`
	Scriptsversion      string `json:"scriptsversion"`
	Serviceofferingid   string `json:"serviceofferingid"`
	Serviceofferingname string `json:"serviceofferingname"`
	State               string `json:"state"`
	Templateid          string `json:"templateid"`
	Version             string `json:"version"`
	Vpcid               string `json:"vpcid"`
	Vpcname             string `json:"vpcname"`
	Zoneid              string `json:"zoneid"`
	Zonename            string `json:"zonename"`
}

type StartInternalLoadBalancerVMParams struct {
	p map[string]interface{}
}

func (p *StartInternalLoadBalancerVMParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *StartInternalLoadBalancerVMParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new StartInternalLoadBalancerVMParams instance,
// as then you are sure you have configured all required params
func (s *InternalLBService) NewStartInternalLoadBalancerVMParams(id string) *StartInternalLoadBalancerVMParams {
	p := &StartInternalLoadBalancerVMParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Starts an existing internal lb vm.
func (s *InternalLBService) StartInternalLoadBalancerVM(p *StartInternalLoadBalancerVMParams) (*StartInternalLoadBalancerVMResponse, error) {
	resp, err := s.cs.newRequest("startInternalLoadBalancerVM", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r StartInternalLoadBalancerVMResponse
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

type StartInternalLoadBalancerVMResponse struct {
	Account             string `json:"account"`
	Created             string `json:"created"`
	Dns1                string `json:"dns1"`
	Dns2                string `json:"dns2"`
	Domain              string `json:"domain"`
	Domainid            string `json:"domainid"`
	Gateway             string `json:"gateway"`
	Guestipaddress      string `json:"guestipaddress"`
	Guestmacaddress     string `json:"guestmacaddress"`
	Guestnetmask        string `json:"guestnetmask"`
	Guestnetworkid      string `json:"guestnetworkid"`
	Guestnetworkname    string `json:"guestnetworkname"`
	Hostid              string `json:"hostid"`
	Hostname            string `json:"hostname"`
	Hypervisor          string `json:"hypervisor"`
	Id                  string `json:"id"`
	Ip6dns1             string `json:"ip6dns1"`
	Ip6dns2             string `json:"ip6dns2"`
	Isredundantrouter   bool   `json:"isredundantrouter"`
	JobID               string `json:"jobid"`
	Jobstatus           int    `json:"jobstatus"`
	Linklocalip         string `json:"linklocalip"`
	Linklocalmacaddress string `json:"linklocalmacaddress"`
	Linklocalnetmask    string `json:"linklocalnetmask"`
	Linklocalnetworkid  string `json:"linklocalnetworkid"`
	Name                string `json:"name"`
	Networkdomain       string `json:"networkdomain"`
	Nic                 []Nic  `json:"nic"`
	Podid               string `json:"podid"`
	Project             string `json:"project"`
	Projectid           string `json:"projectid"`
	Publicip            string `json:"publicip"`
	Publicmacaddress    string `json:"publicmacaddress"`
	Publicnetmask       string `json:"publicnetmask"`
	Publicnetworkid     string `json:"publicnetworkid"`
	Redundantstate      string `json:"redundantstate"`
	Requiresupgrade     bool   `json:"requiresupgrade"`
	Role                string `json:"role"`
	Scriptsversion      string `json:"scriptsversion"`
	Serviceofferingid   string `json:"serviceofferingid"`
	Serviceofferingname string `json:"serviceofferingname"`
	State               string `json:"state"`
	Templateid          string `json:"templateid"`
	Version             string `json:"version"`
	Vpcid               string `json:"vpcid"`
	Vpcname             string `json:"vpcname"`
	Zoneid              string `json:"zoneid"`
	Zonename            string `json:"zonename"`
}

type StopInternalLoadBalancerVMParams struct {
	p map[string]interface{}
}

func (p *StopInternalLoadBalancerVMParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["forced"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("forced", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *StopInternalLoadBalancerVMParams) SetForced(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forced"] = v
	return
}

func (p *StopInternalLoadBalancerVMParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new StopInternalLoadBalancerVMParams instance,
// as then you are sure you have configured all required params
func (s *InternalLBService) NewStopInternalLoadBalancerVMParams(id string) *StopInternalLoadBalancerVMParams {
	p := &StopInternalLoadBalancerVMParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Stops an Internal LB vm.
func (s *InternalLBService) StopInternalLoadBalancerVM(p *StopInternalLoadBalancerVMParams) (*StopInternalLoadBalancerVMResponse, error) {
	resp, err := s.cs.newRequest("stopInternalLoadBalancerVM", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r StopInternalLoadBalancerVMResponse
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

type StopInternalLoadBalancerVMResponse struct {
	Account             string `json:"account"`
	Created             string `json:"created"`
	Dns1                string `json:"dns1"`
	Dns2                string `json:"dns2"`
	Domain              string `json:"domain"`
	Domainid            string `json:"domainid"`
	Gateway             string `json:"gateway"`
	Guestipaddress      string `json:"guestipaddress"`
	Guestmacaddress     string `json:"guestmacaddress"`
	Guestnetmask        string `json:"guestnetmask"`
	Guestnetworkid      string `json:"guestnetworkid"`
	Guestnetworkname    string `json:"guestnetworkname"`
	Hostid              string `json:"hostid"`
	Hostname            string `json:"hostname"`
	Hypervisor          string `json:"hypervisor"`
	Id                  string `json:"id"`
	Ip6dns1             string `json:"ip6dns1"`
	Ip6dns2             string `json:"ip6dns2"`
	Isredundantrouter   bool   `json:"isredundantrouter"`
	JobID               string `json:"jobid"`
	Jobstatus           int    `json:"jobstatus"`
	Linklocalip         string `json:"linklocalip"`
	Linklocalmacaddress string `json:"linklocalmacaddress"`
	Linklocalnetmask    string `json:"linklocalnetmask"`
	Linklocalnetworkid  string `json:"linklocalnetworkid"`
	Name                string `json:"name"`
	Networkdomain       string `json:"networkdomain"`
	Nic                 []Nic  `json:"nic"`
	Podid               string `json:"podid"`
	Project             string `json:"project"`
	Projectid           string `json:"projectid"`
	Publicip            string `json:"publicip"`
	Publicmacaddress    string `json:"publicmacaddress"`
	Publicnetmask       string `json:"publicnetmask"`
	Publicnetworkid     string `json:"publicnetworkid"`
	Redundantstate      string `json:"redundantstate"`
	Requiresupgrade     bool   `json:"requiresupgrade"`
	Role                string `json:"role"`
	Scriptsversion      string `json:"scriptsversion"`
	Serviceofferingid   string `json:"serviceofferingid"`
	Serviceofferingname string `json:"serviceofferingname"`
	State               string `json:"state"`
	Templateid          string `json:"templateid"`
	Version             string `json:"version"`
	Vpcid               string `json:"vpcid"`
	Vpcname             string `json:"vpcname"`
	Zoneid              string `json:"zoneid"`
	Zonename            string `json:"zonename"`
}

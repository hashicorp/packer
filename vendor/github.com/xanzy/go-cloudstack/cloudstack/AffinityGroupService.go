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

type CreateAffinityGroupParams struct {
	p map[string]interface{}
}

func (p *CreateAffinityGroupParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["description"]; found {
		u.Set("description", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	return u
}

func (p *CreateAffinityGroupParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *CreateAffinityGroupParams) SetDescription(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["description"] = v
	return
}

func (p *CreateAffinityGroupParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreateAffinityGroupParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateAffinityGroupParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *CreateAffinityGroupParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

// You should always use this function to get a new CreateAffinityGroupParams instance,
// as then you are sure you have configured all required params
func (s *AffinityGroupService) NewCreateAffinityGroupParams(name string, affinityGroupType string) *CreateAffinityGroupParams {
	p := &CreateAffinityGroupParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	p.p["type"] = affinityGroupType
	return p
}

// Creates an affinity/anti-affinity group
func (s *AffinityGroupService) CreateAffinityGroup(p *CreateAffinityGroupParams) (*CreateAffinityGroupResponse, error) {
	resp, err := s.cs.newRequest("createAffinityGroup", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateAffinityGroupResponse
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

type CreateAffinityGroupResponse struct {
	Account           string   `json:"account"`
	Description       string   `json:"description"`
	Domain            string   `json:"domain"`
	Domainid          string   `json:"domainid"`
	Id                string   `json:"id"`
	JobID             string   `json:"jobid"`
	Jobstatus         int      `json:"jobstatus"`
	Name              string   `json:"name"`
	Project           string   `json:"project"`
	Projectid         string   `json:"projectid"`
	Type              string   `json:"type"`
	VirtualmachineIds []string `json:"virtualmachineIds"`
}

type DeleteAffinityGroupParams struct {
	p map[string]interface{}
}

func (p *DeleteAffinityGroupParams) toURLValues() url.Values {
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
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	return u
}

func (p *DeleteAffinityGroupParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *DeleteAffinityGroupParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *DeleteAffinityGroupParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *DeleteAffinityGroupParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *DeleteAffinityGroupParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

// You should always use this function to get a new DeleteAffinityGroupParams instance,
// as then you are sure you have configured all required params
func (s *AffinityGroupService) NewDeleteAffinityGroupParams() *DeleteAffinityGroupParams {
	p := &DeleteAffinityGroupParams{}
	p.p = make(map[string]interface{})
	return p
}

// Deletes affinity group
func (s *AffinityGroupService) DeleteAffinityGroup(p *DeleteAffinityGroupParams) (*DeleteAffinityGroupResponse, error) {
	resp, err := s.cs.newRequest("deleteAffinityGroup", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteAffinityGroupResponse
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

type DeleteAffinityGroupResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type ListAffinityGroupTypesParams struct {
	p map[string]interface{}
}

func (p *ListAffinityGroupTypesParams) toURLValues() url.Values {
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
	return u
}

func (p *ListAffinityGroupTypesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListAffinityGroupTypesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListAffinityGroupTypesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

// You should always use this function to get a new ListAffinityGroupTypesParams instance,
// as then you are sure you have configured all required params
func (s *AffinityGroupService) NewListAffinityGroupTypesParams() *ListAffinityGroupTypesParams {
	p := &ListAffinityGroupTypesParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists affinity group types available
func (s *AffinityGroupService) ListAffinityGroupTypes(p *ListAffinityGroupTypesParams) (*ListAffinityGroupTypesResponse, error) {
	resp, err := s.cs.newRequest("listAffinityGroupTypes", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListAffinityGroupTypesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListAffinityGroupTypesResponse struct {
	Count              int                  `json:"count"`
	AffinityGroupTypes []*AffinityGroupType `json:"affinitygrouptype"`
}

type AffinityGroupType struct {
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Type      string `json:"type"`
}

type ListAffinityGroupsParams struct {
	p map[string]interface{}
}

func (p *ListAffinityGroupsParams) toURLValues() url.Values {
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
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *ListAffinityGroupsParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListAffinityGroupsParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListAffinityGroupsParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListAffinityGroupsParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListAffinityGroupsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListAffinityGroupsParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListAffinityGroupsParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListAffinityGroupsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListAffinityGroupsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListAffinityGroupsParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListAffinityGroupsParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

func (p *ListAffinityGroupsParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new ListAffinityGroupsParams instance,
// as then you are sure you have configured all required params
func (s *AffinityGroupService) NewListAffinityGroupsParams() *ListAffinityGroupsParams {
	p := &ListAffinityGroupsParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *AffinityGroupService) GetAffinityGroupID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListAffinityGroupsParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListAffinityGroups(p)
	if err != nil {
		return "", -1, err
	}

	// This is needed because of a bug with the listAffinityGroup call. It reports the
	// number of VirtualMachines in the groups as being the number of groups found.
	l.Count = len(l.AffinityGroups)

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.AffinityGroups[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.AffinityGroups {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *AffinityGroupService) GetAffinityGroupByName(name string, opts ...OptionFunc) (*AffinityGroup, int, error) {
	id, count, err := s.GetAffinityGroupID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetAffinityGroupByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *AffinityGroupService) GetAffinityGroupByID(id string, opts ...OptionFunc) (*AffinityGroup, int, error) {
	p := &ListAffinityGroupsParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListAffinityGroups(p)
	if err != nil {
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", id)) {
			return nil, 0, fmt.Errorf("No match found for %s: %+v", id, l)
		}
		return nil, -1, err
	}

	// This is needed because of a bug with the listAffinityGroup call. It reports the
	// number of VirtualMachines in the groups as being the number of groups found.
	l.Count = len(l.AffinityGroups)

	if l.Count == 0 {
		return nil, l.Count, fmt.Errorf("No match found for %s: %+v", id, l)
	}

	if l.Count == 1 {
		return l.AffinityGroups[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for AffinityGroup UUID: %s!", id)
}

// Lists affinity groups
func (s *AffinityGroupService) ListAffinityGroups(p *ListAffinityGroupsParams) (*ListAffinityGroupsResponse, error) {
	resp, err := s.cs.newRequest("listAffinityGroups", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListAffinityGroupsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListAffinityGroupsResponse struct {
	Count          int              `json:"count"`
	AffinityGroups []*AffinityGroup `json:"affinitygroup"`
}

type AffinityGroup struct {
	Account           string   `json:"account"`
	Description       string   `json:"description"`
	Domain            string   `json:"domain"`
	Domainid          string   `json:"domainid"`
	Id                string   `json:"id"`
	JobID             string   `json:"jobid"`
	Jobstatus         int      `json:"jobstatus"`
	Name              string   `json:"name"`
	Project           string   `json:"project"`
	Projectid         string   `json:"projectid"`
	Type              string   `json:"type"`
	VirtualmachineIds []string `json:"virtualmachineIds"`
}

type UpdateVMAffinityGroupParams struct {
	p map[string]interface{}
}

func (p *UpdateVMAffinityGroupParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["affinitygroupids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("affinitygroupids", vv)
	}
	if v, found := p.p["affinitygroupnames"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("affinitygroupnames", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *UpdateVMAffinityGroupParams) SetAffinitygroupids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["affinitygroupids"] = v
	return
}

func (p *UpdateVMAffinityGroupParams) SetAffinitygroupnames(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["affinitygroupnames"] = v
	return
}

func (p *UpdateVMAffinityGroupParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new UpdateVMAffinityGroupParams instance,
// as then you are sure you have configured all required params
func (s *AffinityGroupService) NewUpdateVMAffinityGroupParams(id string) *UpdateVMAffinityGroupParams {
	p := &UpdateVMAffinityGroupParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates the affinity/anti-affinity group associations of a virtual machine. The VM has to be stopped and restarted for the new properties to take effect.
func (s *AffinityGroupService) UpdateVMAffinityGroup(p *UpdateVMAffinityGroupParams) (*UpdateVMAffinityGroupResponse, error) {
	resp, err := s.cs.newRequest("updateVMAffinityGroup", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateVMAffinityGroupResponse
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

type UpdateVMAffinityGroupResponse struct {
	Account               string                                       `json:"account"`
	Affinitygroup         []UpdateVMAffinityGroupResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                          `json:"cpunumber"`
	Cpuspeed              int                                          `json:"cpuspeed"`
	Cpuused               string                                       `json:"cpuused"`
	Created               string                                       `json:"created"`
	Details               map[string]string                            `json:"details"`
	Diskioread            int64                                        `json:"diskioread"`
	Diskiowrite           int64                                        `json:"diskiowrite"`
	Diskkbsread           int64                                        `json:"diskkbsread"`
	Diskkbswrite          int64                                        `json:"diskkbswrite"`
	Diskofferingid        string                                       `json:"diskofferingid"`
	Diskofferingname      string                                       `json:"diskofferingname"`
	Displayname           string                                       `json:"displayname"`
	Displayvm             bool                                         `json:"displayvm"`
	Domain                string                                       `json:"domain"`
	Domainid              string                                       `json:"domainid"`
	Forvirtualnetwork     bool                                         `json:"forvirtualnetwork"`
	Group                 string                                       `json:"group"`
	Groupid               string                                       `json:"groupid"`
	Guestosid             string                                       `json:"guestosid"`
	Haenable              bool                                         `json:"haenable"`
	Hostid                string                                       `json:"hostid"`
	Hostname              string                                       `json:"hostname"`
	Hypervisor            string                                       `json:"hypervisor"`
	Id                    string                                       `json:"id"`
	Instancename          string                                       `json:"instancename"`
	Isdynamicallyscalable bool                                         `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                       `json:"isodisplaytext"`
	Isoid                 string                                       `json:"isoid"`
	Isoname               string                                       `json:"isoname"`
	JobID                 string                                       `json:"jobid"`
	Jobstatus             int                                          `json:"jobstatus"`
	Keypair               string                                       `json:"keypair"`
	Memory                int                                          `json:"memory"`
	Memoryintfreekbs      int64                                        `json:"memoryintfreekbs"`
	Memorykbs             int64                                        `json:"memorykbs"`
	Memorytargetkbs       int64                                        `json:"memorytargetkbs"`
	Name                  string                                       `json:"name"`
	Networkkbsread        int64                                        `json:"networkkbsread"`
	Networkkbswrite       int64                                        `json:"networkkbswrite"`
	Nic                   []Nic                                        `json:"nic"`
	Ostypeid              string                                       `json:"ostypeid"`
	Password              string                                       `json:"password"`
	Passwordenabled       bool                                         `json:"passwordenabled"`
	Project               string                                       `json:"project"`
	Projectid             string                                       `json:"projectid"`
	Publicip              string                                       `json:"publicip"`
	Publicipid            string                                       `json:"publicipid"`
	Rootdeviceid          int64                                        `json:"rootdeviceid"`
	Rootdevicetype        string                                       `json:"rootdevicetype"`
	Securitygroup         []UpdateVMAffinityGroupResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                       `json:"serviceofferingid"`
	Serviceofferingname   string                                       `json:"serviceofferingname"`
	Servicestate          string                                       `json:"servicestate"`
	State                 string                                       `json:"state"`
	Tags                  []Tags                                       `json:"tags"`
	Templatedisplaytext   string                                       `json:"templatedisplaytext"`
	Templateid            string                                       `json:"templateid"`
	Templatename          string                                       `json:"templatename"`
	Userid                string                                       `json:"userid"`
	Username              string                                       `json:"username"`
	Vgpu                  string                                       `json:"vgpu"`
	Zoneid                string                                       `json:"zoneid"`
	Zonename              string                                       `json:"zonename"`
}

type UpdateVMAffinityGroupResponseSecuritygroup struct {
	Account             string                                           `json:"account"`
	Description         string                                           `json:"description"`
	Domain              string                                           `json:"domain"`
	Domainid            string                                           `json:"domainid"`
	Egressrule          []UpdateVMAffinityGroupResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                           `json:"id"`
	Ingressrule         []UpdateVMAffinityGroupResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                           `json:"name"`
	Project             string                                           `json:"project"`
	Projectid           string                                           `json:"projectid"`
	Tags                []Tags                                           `json:"tags"`
	Virtualmachinecount int                                              `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                    `json:"virtualmachineids"`
}

type UpdateVMAffinityGroupResponseSecuritygroupRule struct {
	Account           string `json:"account"`
	Cidr              string `json:"cidr"`
	Endport           int    `json:"endport"`
	Icmpcode          int    `json:"icmpcode"`
	Icmptype          int    `json:"icmptype"`
	Protocol          string `json:"protocol"`
	Ruleid            string `json:"ruleid"`
	Securitygroupname string `json:"securitygroupname"`
	Startport         int    `json:"startport"`
	Tags              []Tags `json:"tags"`
}

type UpdateVMAffinityGroupResponseAffinitygroup struct {
	Account           string   `json:"account"`
	Description       string   `json:"description"`
	Domain            string   `json:"domain"`
	Domainid          string   `json:"domainid"`
	Id                string   `json:"id"`
	Name              string   `json:"name"`
	Project           string   `json:"project"`
	Projectid         string   `json:"projectid"`
	Type              string   `json:"type"`
	VirtualmachineIds []string `json:"virtualmachineIds"`
}

func (r *UpdateVMAffinityGroupResponse) UnmarshalJSON(b []byte) error {
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

	type alias UpdateVMAffinityGroupResponse
	return json.Unmarshal(b, (*alias)(r))
}

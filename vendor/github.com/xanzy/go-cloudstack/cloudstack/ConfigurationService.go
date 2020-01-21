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

type ListCapabilitiesParams struct {
	p map[string]interface{}
}

func (p *ListCapabilitiesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	return u
}

// You should always use this function to get a new ListCapabilitiesParams instance,
// as then you are sure you have configured all required params
func (s *ConfigurationService) NewListCapabilitiesParams() *ListCapabilitiesParams {
	p := &ListCapabilitiesParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists capabilities
func (s *ConfigurationService) ListCapabilities(p *ListCapabilitiesParams) (*ListCapabilitiesResponse, error) {
	resp, err := s.cs.newRequest("listCapabilities", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListCapabilitiesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListCapabilitiesResponse struct {
	Count        int           `json:"count"`
	Capabilities []*Capability `json:"capability"`
}

type Capability struct {
	Allowusercreateprojects   bool   `json:"allowusercreateprojects"`
	Allowuserexpungerecovervm bool   `json:"allowuserexpungerecovervm"`
	Allowuserviewdestroyedvm  bool   `json:"allowuserviewdestroyedvm"`
	Apilimitinterval          int    `json:"apilimitinterval"`
	Apilimitmax               int    `json:"apilimitmax"`
	Cloudstackversion         string `json:"cloudstackversion"`
	Customdiskofferingmaxsize int64  `json:"customdiskofferingmaxsize"`
	Customdiskofferingminsize int64  `json:"customdiskofferingminsize"`
	Dynamicrolesenabled       bool   `json:"dynamicrolesenabled"`
	JobID                     string `json:"jobid"`
	Jobstatus                 int    `json:"jobstatus"`
	Kvmsnapshotenabled        bool   `json:"kvmsnapshotenabled"`
	Projectinviterequired     bool   `json:"projectinviterequired"`
	Regionsecondaryenabled    bool   `json:"regionsecondaryenabled"`
	Securitygroupsenabled     bool   `json:"securitygroupsenabled"`
	SupportELB                string `json:"supportELB"`
	Userpublictemplateenabled bool   `json:"userpublictemplateenabled"`
}

type ListConfigurationsParams struct {
	p map[string]interface{}
}

func (p *ListConfigurationsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["accountid"]; found {
		u.Set("accountid", v.(string))
	}
	if v, found := p.p["category"]; found {
		u.Set("category", v.(string))
	}
	if v, found := p.p["clusterid"]; found {
		u.Set("clusterid", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["imagestoreuuid"]; found {
		u.Set("imagestoreuuid", v.(string))
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
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
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListConfigurationsParams) SetAccountid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["accountid"] = v
	return
}

func (p *ListConfigurationsParams) SetCategory(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["category"] = v
	return
}

func (p *ListConfigurationsParams) SetClusterid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["clusterid"] = v
	return
}

func (p *ListConfigurationsParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListConfigurationsParams) SetImagestoreuuid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["imagestoreuuid"] = v
	return
}

func (p *ListConfigurationsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListConfigurationsParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListConfigurationsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListConfigurationsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListConfigurationsParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *ListConfigurationsParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListConfigurationsParams instance,
// as then you are sure you have configured all required params
func (s *ConfigurationService) NewListConfigurationsParams() *ListConfigurationsParams {
	p := &ListConfigurationsParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists all configurations.
func (s *ConfigurationService) ListConfigurations(p *ListConfigurationsParams) (*ListConfigurationsResponse, error) {
	resp, err := s.cs.newRequest("listConfigurations", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListConfigurationsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListConfigurationsResponse struct {
	Count          int              `json:"count"`
	Configurations []*Configuration `json:"configuration"`
}

type Configuration struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Id          int64  `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	Value       string `json:"value"`
}

type ListDeploymentPlannersParams struct {
	p map[string]interface{}
}

func (p *ListDeploymentPlannersParams) toURLValues() url.Values {
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

func (p *ListDeploymentPlannersParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListDeploymentPlannersParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListDeploymentPlannersParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

// You should always use this function to get a new ListDeploymentPlannersParams instance,
// as then you are sure you have configured all required params
func (s *ConfigurationService) NewListDeploymentPlannersParams() *ListDeploymentPlannersParams {
	p := &ListDeploymentPlannersParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists all DeploymentPlanners available.
func (s *ConfigurationService) ListDeploymentPlanners(p *ListDeploymentPlannersParams) (*ListDeploymentPlannersResponse, error) {
	resp, err := s.cs.newRequest("listDeploymentPlanners", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListDeploymentPlannersResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListDeploymentPlannersResponse struct {
	Count              int                  `json:"count"`
	DeploymentPlanners []*DeploymentPlanner `json:"deploymentplanner"`
}

type DeploymentPlanner struct {
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Name      string `json:"name"`
}

type UpdateConfigurationParams struct {
	p map[string]interface{}
}

func (p *UpdateConfigurationParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["accountid"]; found {
		u.Set("accountid", v.(string))
	}
	if v, found := p.p["clusterid"]; found {
		u.Set("clusterid", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["imagestoreuuid"]; found {
		u.Set("imagestoreuuid", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["value"]; found {
		u.Set("value", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *UpdateConfigurationParams) SetAccountid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["accountid"] = v
	return
}

func (p *UpdateConfigurationParams) SetClusterid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["clusterid"] = v
	return
}

func (p *UpdateConfigurationParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *UpdateConfigurationParams) SetImagestoreuuid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["imagestoreuuid"] = v
	return
}

func (p *UpdateConfigurationParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UpdateConfigurationParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *UpdateConfigurationParams) SetValue(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["value"] = v
	return
}

func (p *UpdateConfigurationParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new UpdateConfigurationParams instance,
// as then you are sure you have configured all required params
func (s *ConfigurationService) NewUpdateConfigurationParams(name string) *UpdateConfigurationParams {
	p := &UpdateConfigurationParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	return p
}

// Updates a configuration.
func (s *ConfigurationService) UpdateConfiguration(p *UpdateConfigurationParams) (*UpdateConfigurationResponse, error) {
	resp, err := s.cs.newRequest("updateConfiguration", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateConfigurationResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type UpdateConfigurationResponse struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Id          int64  `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	Value       string `json:"value"`
}

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

type CreateSSHKeyPairParams struct {
	p map[string]interface{}
}

func (p *CreateSSHKeyPairParams) toURLValues() url.Values {
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
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	return u
}

func (p *CreateSSHKeyPairParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *CreateSSHKeyPairParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreateSSHKeyPairParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateSSHKeyPairParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

// You should always use this function to get a new CreateSSHKeyPairParams instance,
// as then you are sure you have configured all required params
func (s *SSHService) NewCreateSSHKeyPairParams(name string) *CreateSSHKeyPairParams {
	p := &CreateSSHKeyPairParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	return p
}

// Create a new keypair and returns the private key
func (s *SSHService) CreateSSHKeyPair(p *CreateSSHKeyPairParams) (*CreateSSHKeyPairResponse, error) {
	resp, err := s.cs.newRequest("createSSHKeyPair", p.toURLValues())
	if err != nil {
		return nil, err
	}

	if resp, err = getRawValue(resp); err != nil {
		return nil, err
	}

	var r CreateSSHKeyPairResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type CreateSSHKeyPairResponse struct {
	Account     string `json:"account"`
	Domain      string `json:"domain"`
	Domainid    string `json:"domainid"`
	Fingerprint string `json:"fingerprint"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
	Privatekey  string `json:"privatekey"`
}

type DeleteSSHKeyPairParams struct {
	p map[string]interface{}
}

func (p *DeleteSSHKeyPairParams) toURLValues() url.Values {
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
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	return u
}

func (p *DeleteSSHKeyPairParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *DeleteSSHKeyPairParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *DeleteSSHKeyPairParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *DeleteSSHKeyPairParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

// You should always use this function to get a new DeleteSSHKeyPairParams instance,
// as then you are sure you have configured all required params
func (s *SSHService) NewDeleteSSHKeyPairParams(name string) *DeleteSSHKeyPairParams {
	p := &DeleteSSHKeyPairParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	return p
}

// Deletes a keypair by name
func (s *SSHService) DeleteSSHKeyPair(p *DeleteSSHKeyPairParams) (*DeleteSSHKeyPairResponse, error) {
	resp, err := s.cs.newRequest("deleteSSHKeyPair", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteSSHKeyPairResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DeleteSSHKeyPairResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *DeleteSSHKeyPairResponse) UnmarshalJSON(b []byte) error {
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

	type alias DeleteSSHKeyPairResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ListSSHKeyPairsParams struct {
	p map[string]interface{}
}

func (p *ListSSHKeyPairsParams) toURLValues() url.Values {
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
	if v, found := p.p["fingerprint"]; found {
		u.Set("fingerprint", v.(string))
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
	return u
}

func (p *ListSSHKeyPairsParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetFingerprint(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["fingerprint"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListSSHKeyPairsParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

// You should always use this function to get a new ListSSHKeyPairsParams instance,
// as then you are sure you have configured all required params
func (s *SSHService) NewListSSHKeyPairsParams() *ListSSHKeyPairsParams {
	p := &ListSSHKeyPairsParams{}
	p.p = make(map[string]interface{})
	return p
}

// List registered keypairs
func (s *SSHService) ListSSHKeyPairs(p *ListSSHKeyPairsParams) (*ListSSHKeyPairsResponse, error) {
	resp, err := s.cs.newRequest("listSSHKeyPairs", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListSSHKeyPairsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListSSHKeyPairsResponse struct {
	Count       int           `json:"count"`
	SSHKeyPairs []*SSHKeyPair `json:"sshkeypair"`
}

type SSHKeyPair struct {
	Account     string `json:"account"`
	Domain      string `json:"domain"`
	Domainid    string `json:"domainid"`
	Fingerprint string `json:"fingerprint"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
}

type RegisterSSHKeyPairParams struct {
	p map[string]interface{}
}

func (p *RegisterSSHKeyPairParams) toURLValues() url.Values {
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
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["publickey"]; found {
		u.Set("publickey", v.(string))
	}
	return u
}

func (p *RegisterSSHKeyPairParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *RegisterSSHKeyPairParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *RegisterSSHKeyPairParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *RegisterSSHKeyPairParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *RegisterSSHKeyPairParams) SetPublickey(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["publickey"] = v
	return
}

// You should always use this function to get a new RegisterSSHKeyPairParams instance,
// as then you are sure you have configured all required params
func (s *SSHService) NewRegisterSSHKeyPairParams(name string, publickey string) *RegisterSSHKeyPairParams {
	p := &RegisterSSHKeyPairParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	p.p["publickey"] = publickey
	return p
}

// Register a public key in a keypair under a certain name
func (s *SSHService) RegisterSSHKeyPair(p *RegisterSSHKeyPairParams) (*RegisterSSHKeyPairResponse, error) {
	resp, err := s.cs.newRequest("registerSSHKeyPair", p.toURLValues())
	if err != nil {
		return nil, err
	}

	if resp, err = getRawValue(resp); err != nil {
		return nil, err
	}

	var r RegisterSSHKeyPairResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type RegisterSSHKeyPairResponse struct {
	Account     string `json:"account"`
	Domain      string `json:"domain"`
	Domainid    string `json:"domainid"`
	Fingerprint string `json:"fingerprint"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
}

type ResetSSHKeyForVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *ResetSSHKeyForVirtualMachineParams) toURLValues() url.Values {
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
	if v, found := p.p["keypair"]; found {
		u.Set("keypair", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	return u
}

func (p *ResetSSHKeyForVirtualMachineParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ResetSSHKeyForVirtualMachineParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ResetSSHKeyForVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ResetSSHKeyForVirtualMachineParams) SetKeypair(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keypair"] = v
	return
}

func (p *ResetSSHKeyForVirtualMachineParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

// You should always use this function to get a new ResetSSHKeyForVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *SSHService) NewResetSSHKeyForVirtualMachineParams(id string, keypair string) *ResetSSHKeyForVirtualMachineParams {
	p := &ResetSSHKeyForVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	p.p["keypair"] = keypair
	return p
}

// Resets the SSH Key for virtual machine. The virtual machine must be in a "Stopped" state. [async]
func (s *SSHService) ResetSSHKeyForVirtualMachine(p *ResetSSHKeyForVirtualMachineParams) (*ResetSSHKeyForVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("resetSSHKeyForVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ResetSSHKeyForVirtualMachineResponse
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

type ResetSSHKeyForVirtualMachineResponse struct {
	Account               string                                              `json:"account"`
	Affinitygroup         []ResetSSHKeyForVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                                 `json:"cpunumber"`
	Cpuspeed              int                                                 `json:"cpuspeed"`
	Cpuused               string                                              `json:"cpuused"`
	Created               string                                              `json:"created"`
	Details               map[string]string                                   `json:"details"`
	Diskioread            int64                                               `json:"diskioread"`
	Diskiowrite           int64                                               `json:"diskiowrite"`
	Diskkbsread           int64                                               `json:"diskkbsread"`
	Diskkbswrite          int64                                               `json:"diskkbswrite"`
	Diskofferingid        string                                              `json:"diskofferingid"`
	Diskofferingname      string                                              `json:"diskofferingname"`
	Displayname           string                                              `json:"displayname"`
	Displayvm             bool                                                `json:"displayvm"`
	Domain                string                                              `json:"domain"`
	Domainid              string                                              `json:"domainid"`
	Forvirtualnetwork     bool                                                `json:"forvirtualnetwork"`
	Group                 string                                              `json:"group"`
	Groupid               string                                              `json:"groupid"`
	Guestosid             string                                              `json:"guestosid"`
	Haenable              bool                                                `json:"haenable"`
	Hostid                string                                              `json:"hostid"`
	Hostname              string                                              `json:"hostname"`
	Hypervisor            string                                              `json:"hypervisor"`
	Id                    string                                              `json:"id"`
	Instancename          string                                              `json:"instancename"`
	Isdynamicallyscalable bool                                                `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                              `json:"isodisplaytext"`
	Isoid                 string                                              `json:"isoid"`
	Isoname               string                                              `json:"isoname"`
	JobID                 string                                              `json:"jobid"`
	Jobstatus             int                                                 `json:"jobstatus"`
	Keypair               string                                              `json:"keypair"`
	Memory                int                                                 `json:"memory"`
	Memoryintfreekbs      int64                                               `json:"memoryintfreekbs"`
	Memorykbs             int64                                               `json:"memorykbs"`
	Memorytargetkbs       int64                                               `json:"memorytargetkbs"`
	Name                  string                                              `json:"name"`
	Networkkbsread        int64                                               `json:"networkkbsread"`
	Networkkbswrite       int64                                               `json:"networkkbswrite"`
	Nic                   []Nic                                               `json:"nic"`
	Ostypeid              string                                              `json:"ostypeid"`
	Password              string                                              `json:"password"`
	Passwordenabled       bool                                                `json:"passwordenabled"`
	Project               string                                              `json:"project"`
	Projectid             string                                              `json:"projectid"`
	Publicip              string                                              `json:"publicip"`
	Publicipid            string                                              `json:"publicipid"`
	Rootdeviceid          int64                                               `json:"rootdeviceid"`
	Rootdevicetype        string                                              `json:"rootdevicetype"`
	Securitygroup         []ResetSSHKeyForVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                              `json:"serviceofferingid"`
	Serviceofferingname   string                                              `json:"serviceofferingname"`
	Servicestate          string                                              `json:"servicestate"`
	State                 string                                              `json:"state"`
	Tags                  []Tags                                              `json:"tags"`
	Templatedisplaytext   string                                              `json:"templatedisplaytext"`
	Templateid            string                                              `json:"templateid"`
	Templatename          string                                              `json:"templatename"`
	Userid                string                                              `json:"userid"`
	Username              string                                              `json:"username"`
	Vgpu                  string                                              `json:"vgpu"`
	Zoneid                string                                              `json:"zoneid"`
	Zonename              string                                              `json:"zonename"`
}

type ResetSSHKeyForVirtualMachineResponseSecuritygroup struct {
	Account             string                                                  `json:"account"`
	Description         string                                                  `json:"description"`
	Domain              string                                                  `json:"domain"`
	Domainid            string                                                  `json:"domainid"`
	Egressrule          []ResetSSHKeyForVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                                  `json:"id"`
	Ingressrule         []ResetSSHKeyForVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                                  `json:"name"`
	Project             string                                                  `json:"project"`
	Projectid           string                                                  `json:"projectid"`
	Tags                []Tags                                                  `json:"tags"`
	Virtualmachinecount int                                                     `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                           `json:"virtualmachineids"`
}

type ResetSSHKeyForVirtualMachineResponseSecuritygroupRule struct {
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

type ResetSSHKeyForVirtualMachineResponseAffinitygroup struct {
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

func (r *ResetSSHKeyForVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias ResetSSHKeyForVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

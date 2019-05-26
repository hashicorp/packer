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

type AddNicToVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *AddNicToVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["dhcpoptions"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("dhcpoptions[%d].key", i), k)
			u.Set(fmt.Sprintf("dhcpoptions[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["ipaddress"]; found {
		u.Set("ipaddress", v.(string))
	}
	if v, found := p.p["macaddress"]; found {
		u.Set("macaddress", v.(string))
	}
	if v, found := p.p["networkid"]; found {
		u.Set("networkid", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *AddNicToVirtualMachineParams) SetDhcpoptions(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["dhcpoptions"] = v
	return
}

func (p *AddNicToVirtualMachineParams) SetIpaddress(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ipaddress"] = v
	return
}

func (p *AddNicToVirtualMachineParams) SetMacaddress(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["macaddress"] = v
	return
}

func (p *AddNicToVirtualMachineParams) SetNetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkid"] = v
	return
}

func (p *AddNicToVirtualMachineParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new AddNicToVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewAddNicToVirtualMachineParams(networkid string, virtualmachineid string) *AddNicToVirtualMachineParams {
	p := &AddNicToVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["networkid"] = networkid
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Adds VM to specified network by creating a NIC
func (s *VirtualMachineService) AddNicToVirtualMachine(p *AddNicToVirtualMachineParams) (*AddNicToVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("addNicToVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddNicToVirtualMachineResponse
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

type AddNicToVirtualMachineResponse struct {
	Account               string                                        `json:"account"`
	Affinitygroup         []AddNicToVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                           `json:"cpunumber"`
	Cpuspeed              int                                           `json:"cpuspeed"`
	Cpuused               string                                        `json:"cpuused"`
	Created               string                                        `json:"created"`
	Details               map[string]string                             `json:"details"`
	Diskioread            int64                                         `json:"diskioread"`
	Diskiowrite           int64                                         `json:"diskiowrite"`
	Diskkbsread           int64                                         `json:"diskkbsread"`
	Diskkbswrite          int64                                         `json:"diskkbswrite"`
	Diskofferingid        string                                        `json:"diskofferingid"`
	Diskofferingname      string                                        `json:"diskofferingname"`
	Displayname           string                                        `json:"displayname"`
	Displayvm             bool                                          `json:"displayvm"`
	Domain                string                                        `json:"domain"`
	Domainid              string                                        `json:"domainid"`
	Forvirtualnetwork     bool                                          `json:"forvirtualnetwork"`
	Group                 string                                        `json:"group"`
	Groupid               string                                        `json:"groupid"`
	Guestosid             string                                        `json:"guestosid"`
	Haenable              bool                                          `json:"haenable"`
	Hostid                string                                        `json:"hostid"`
	Hostname              string                                        `json:"hostname"`
	Hypervisor            string                                        `json:"hypervisor"`
	Id                    string                                        `json:"id"`
	Instancename          string                                        `json:"instancename"`
	Isdynamicallyscalable bool                                          `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                        `json:"isodisplaytext"`
	Isoid                 string                                        `json:"isoid"`
	Isoname               string                                        `json:"isoname"`
	JobID                 string                                        `json:"jobid"`
	Jobstatus             int                                           `json:"jobstatus"`
	Keypair               string                                        `json:"keypair"`
	Memory                int                                           `json:"memory"`
	Memoryintfreekbs      int64                                         `json:"memoryintfreekbs"`
	Memorykbs             int64                                         `json:"memorykbs"`
	Memorytargetkbs       int64                                         `json:"memorytargetkbs"`
	Name                  string                                        `json:"name"`
	Networkkbsread        int64                                         `json:"networkkbsread"`
	Networkkbswrite       int64                                         `json:"networkkbswrite"`
	Nic                   []Nic                                         `json:"nic"`
	Ostypeid              string                                        `json:"ostypeid"`
	Password              string                                        `json:"password"`
	Passwordenabled       bool                                          `json:"passwordenabled"`
	Project               string                                        `json:"project"`
	Projectid             string                                        `json:"projectid"`
	Publicip              string                                        `json:"publicip"`
	Publicipid            string                                        `json:"publicipid"`
	Rootdeviceid          int64                                         `json:"rootdeviceid"`
	Rootdevicetype        string                                        `json:"rootdevicetype"`
	Securitygroup         []AddNicToVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                        `json:"serviceofferingid"`
	Serviceofferingname   string                                        `json:"serviceofferingname"`
	Servicestate          string                                        `json:"servicestate"`
	State                 string                                        `json:"state"`
	Tags                  []Tags                                        `json:"tags"`
	Templatedisplaytext   string                                        `json:"templatedisplaytext"`
	Templateid            string                                        `json:"templateid"`
	Templatename          string                                        `json:"templatename"`
	Userid                string                                        `json:"userid"`
	Username              string                                        `json:"username"`
	Vgpu                  string                                        `json:"vgpu"`
	Zoneid                string                                        `json:"zoneid"`
	Zonename              string                                        `json:"zonename"`
}

type AddNicToVirtualMachineResponseSecuritygroup struct {
	Account             string                                            `json:"account"`
	Description         string                                            `json:"description"`
	Domain              string                                            `json:"domain"`
	Domainid            string                                            `json:"domainid"`
	Egressrule          []AddNicToVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                            `json:"id"`
	Ingressrule         []AddNicToVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                            `json:"name"`
	Project             string                                            `json:"project"`
	Projectid           string                                            `json:"projectid"`
	Tags                []Tags                                            `json:"tags"`
	Virtualmachinecount int                                               `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                     `json:"virtualmachineids"`
}

type AddNicToVirtualMachineResponseSecuritygroupRule struct {
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

type AddNicToVirtualMachineResponseAffinitygroup struct {
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

func (r *AddNicToVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias AddNicToVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type AssignVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *AssignVirtualMachineParams) toURLValues() url.Values {
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
	if v, found := p.p["networkids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("networkids", vv)
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["securitygroupids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("securitygroupids", vv)
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *AssignVirtualMachineParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *AssignVirtualMachineParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *AssignVirtualMachineParams) SetNetworkids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkids"] = v
	return
}

func (p *AssignVirtualMachineParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *AssignVirtualMachineParams) SetSecuritygroupids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["securitygroupids"] = v
	return
}

func (p *AssignVirtualMachineParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new AssignVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewAssignVirtualMachineParams(virtualmachineid string) *AssignVirtualMachineParams {
	p := &AssignVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Change ownership of a VM from one account to another. This API is available for Basic zones with security groups and Advanced zones with guest networks. A root administrator can reassign a VM from any account to any other account in any domain. A domain administrator can reassign a VM to any account in the same domain.
func (s *VirtualMachineService) AssignVirtualMachine(p *AssignVirtualMachineParams) (*AssignVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("assignVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AssignVirtualMachineResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type AssignVirtualMachineResponse struct {
	Account               string                                      `json:"account"`
	Affinitygroup         []AssignVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                         `json:"cpunumber"`
	Cpuspeed              int                                         `json:"cpuspeed"`
	Cpuused               string                                      `json:"cpuused"`
	Created               string                                      `json:"created"`
	Details               map[string]string                           `json:"details"`
	Diskioread            int64                                       `json:"diskioread"`
	Diskiowrite           int64                                       `json:"diskiowrite"`
	Diskkbsread           int64                                       `json:"diskkbsread"`
	Diskkbswrite          int64                                       `json:"diskkbswrite"`
	Diskofferingid        string                                      `json:"diskofferingid"`
	Diskofferingname      string                                      `json:"diskofferingname"`
	Displayname           string                                      `json:"displayname"`
	Displayvm             bool                                        `json:"displayvm"`
	Domain                string                                      `json:"domain"`
	Domainid              string                                      `json:"domainid"`
	Forvirtualnetwork     bool                                        `json:"forvirtualnetwork"`
	Group                 string                                      `json:"group"`
	Groupid               string                                      `json:"groupid"`
	Guestosid             string                                      `json:"guestosid"`
	Haenable              bool                                        `json:"haenable"`
	Hostid                string                                      `json:"hostid"`
	Hostname              string                                      `json:"hostname"`
	Hypervisor            string                                      `json:"hypervisor"`
	Id                    string                                      `json:"id"`
	Instancename          string                                      `json:"instancename"`
	Isdynamicallyscalable bool                                        `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                      `json:"isodisplaytext"`
	Isoid                 string                                      `json:"isoid"`
	Isoname               string                                      `json:"isoname"`
	JobID                 string                                      `json:"jobid"`
	Jobstatus             int                                         `json:"jobstatus"`
	Keypair               string                                      `json:"keypair"`
	Memory                int                                         `json:"memory"`
	Memoryintfreekbs      int64                                       `json:"memoryintfreekbs"`
	Memorykbs             int64                                       `json:"memorykbs"`
	Memorytargetkbs       int64                                       `json:"memorytargetkbs"`
	Name                  string                                      `json:"name"`
	Networkkbsread        int64                                       `json:"networkkbsread"`
	Networkkbswrite       int64                                       `json:"networkkbswrite"`
	Nic                   []Nic                                       `json:"nic"`
	Ostypeid              string                                      `json:"ostypeid"`
	Password              string                                      `json:"password"`
	Passwordenabled       bool                                        `json:"passwordenabled"`
	Project               string                                      `json:"project"`
	Projectid             string                                      `json:"projectid"`
	Publicip              string                                      `json:"publicip"`
	Publicipid            string                                      `json:"publicipid"`
	Rootdeviceid          int64                                       `json:"rootdeviceid"`
	Rootdevicetype        string                                      `json:"rootdevicetype"`
	Securitygroup         []AssignVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                      `json:"serviceofferingid"`
	Serviceofferingname   string                                      `json:"serviceofferingname"`
	Servicestate          string                                      `json:"servicestate"`
	State                 string                                      `json:"state"`
	Tags                  []Tags                                      `json:"tags"`
	Templatedisplaytext   string                                      `json:"templatedisplaytext"`
	Templateid            string                                      `json:"templateid"`
	Templatename          string                                      `json:"templatename"`
	Userid                string                                      `json:"userid"`
	Username              string                                      `json:"username"`
	Vgpu                  string                                      `json:"vgpu"`
	Zoneid                string                                      `json:"zoneid"`
	Zonename              string                                      `json:"zonename"`
}

type AssignVirtualMachineResponseSecuritygroup struct {
	Account             string                                          `json:"account"`
	Description         string                                          `json:"description"`
	Domain              string                                          `json:"domain"`
	Domainid            string                                          `json:"domainid"`
	Egressrule          []AssignVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                          `json:"id"`
	Ingressrule         []AssignVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                          `json:"name"`
	Project             string                                          `json:"project"`
	Projectid           string                                          `json:"projectid"`
	Tags                []Tags                                          `json:"tags"`
	Virtualmachinecount int                                             `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                   `json:"virtualmachineids"`
}

type AssignVirtualMachineResponseSecuritygroupRule struct {
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

type AssignVirtualMachineResponseAffinitygroup struct {
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

func (r *AssignVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias AssignVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ChangeServiceForVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *ChangeServiceForVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["details"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("details[%d].%s", i, k), vv)
			i++
		}
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["serviceofferingid"]; found {
		u.Set("serviceofferingid", v.(string))
	}
	return u
}

func (p *ChangeServiceForVirtualMachineParams) SetDetails(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["details"] = v
	return
}

func (p *ChangeServiceForVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ChangeServiceForVirtualMachineParams) SetServiceofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceofferingid"] = v
	return
}

// You should always use this function to get a new ChangeServiceForVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewChangeServiceForVirtualMachineParams(id string, serviceofferingid string) *ChangeServiceForVirtualMachineParams {
	p := &ChangeServiceForVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	p.p["serviceofferingid"] = serviceofferingid
	return p
}

// Changes the service offering for a virtual machine. The virtual machine must be in a "Stopped" state for this command to take effect.
func (s *VirtualMachineService) ChangeServiceForVirtualMachine(p *ChangeServiceForVirtualMachineParams) (*ChangeServiceForVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("changeServiceForVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ChangeServiceForVirtualMachineResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ChangeServiceForVirtualMachineResponse struct {
	Account               string                                                `json:"account"`
	Affinitygroup         []ChangeServiceForVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                                   `json:"cpunumber"`
	Cpuspeed              int                                                   `json:"cpuspeed"`
	Cpuused               string                                                `json:"cpuused"`
	Created               string                                                `json:"created"`
	Details               map[string]string                                     `json:"details"`
	Diskioread            int64                                                 `json:"diskioread"`
	Diskiowrite           int64                                                 `json:"diskiowrite"`
	Diskkbsread           int64                                                 `json:"diskkbsread"`
	Diskkbswrite          int64                                                 `json:"diskkbswrite"`
	Diskofferingid        string                                                `json:"diskofferingid"`
	Diskofferingname      string                                                `json:"diskofferingname"`
	Displayname           string                                                `json:"displayname"`
	Displayvm             bool                                                  `json:"displayvm"`
	Domain                string                                                `json:"domain"`
	Domainid              string                                                `json:"domainid"`
	Forvirtualnetwork     bool                                                  `json:"forvirtualnetwork"`
	Group                 string                                                `json:"group"`
	Groupid               string                                                `json:"groupid"`
	Guestosid             string                                                `json:"guestosid"`
	Haenable              bool                                                  `json:"haenable"`
	Hostid                string                                                `json:"hostid"`
	Hostname              string                                                `json:"hostname"`
	Hypervisor            string                                                `json:"hypervisor"`
	Id                    string                                                `json:"id"`
	Instancename          string                                                `json:"instancename"`
	Isdynamicallyscalable bool                                                  `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                                `json:"isodisplaytext"`
	Isoid                 string                                                `json:"isoid"`
	Isoname               string                                                `json:"isoname"`
	JobID                 string                                                `json:"jobid"`
	Jobstatus             int                                                   `json:"jobstatus"`
	Keypair               string                                                `json:"keypair"`
	Memory                int                                                   `json:"memory"`
	Memoryintfreekbs      int64                                                 `json:"memoryintfreekbs"`
	Memorykbs             int64                                                 `json:"memorykbs"`
	Memorytargetkbs       int64                                                 `json:"memorytargetkbs"`
	Name                  string                                                `json:"name"`
	Networkkbsread        int64                                                 `json:"networkkbsread"`
	Networkkbswrite       int64                                                 `json:"networkkbswrite"`
	Nic                   []Nic                                                 `json:"nic"`
	Ostypeid              string                                                `json:"ostypeid"`
	Password              string                                                `json:"password"`
	Passwordenabled       bool                                                  `json:"passwordenabled"`
	Project               string                                                `json:"project"`
	Projectid             string                                                `json:"projectid"`
	Publicip              string                                                `json:"publicip"`
	Publicipid            string                                                `json:"publicipid"`
	Rootdeviceid          int64                                                 `json:"rootdeviceid"`
	Rootdevicetype        string                                                `json:"rootdevicetype"`
	Securitygroup         []ChangeServiceForVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                                `json:"serviceofferingid"`
	Serviceofferingname   string                                                `json:"serviceofferingname"`
	Servicestate          string                                                `json:"servicestate"`
	State                 string                                                `json:"state"`
	Tags                  []Tags                                                `json:"tags"`
	Templatedisplaytext   string                                                `json:"templatedisplaytext"`
	Templateid            string                                                `json:"templateid"`
	Templatename          string                                                `json:"templatename"`
	Userid                string                                                `json:"userid"`
	Username              string                                                `json:"username"`
	Vgpu                  string                                                `json:"vgpu"`
	Zoneid                string                                                `json:"zoneid"`
	Zonename              string                                                `json:"zonename"`
}

type ChangeServiceForVirtualMachineResponseSecuritygroup struct {
	Account             string                                                    `json:"account"`
	Description         string                                                    `json:"description"`
	Domain              string                                                    `json:"domain"`
	Domainid            string                                                    `json:"domainid"`
	Egressrule          []ChangeServiceForVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                                    `json:"id"`
	Ingressrule         []ChangeServiceForVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                                    `json:"name"`
	Project             string                                                    `json:"project"`
	Projectid           string                                                    `json:"projectid"`
	Tags                []Tags                                                    `json:"tags"`
	Virtualmachinecount int                                                       `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                             `json:"virtualmachineids"`
}

type ChangeServiceForVirtualMachineResponseSecuritygroupRule struct {
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

type ChangeServiceForVirtualMachineResponseAffinitygroup struct {
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

func (r *ChangeServiceForVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias ChangeServiceForVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type CleanVMReservationsParams struct {
	p map[string]interface{}
}

func (p *CleanVMReservationsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	return u
}

// You should always use this function to get a new CleanVMReservationsParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewCleanVMReservationsParams() *CleanVMReservationsParams {
	p := &CleanVMReservationsParams{}
	p.p = make(map[string]interface{})
	return p
}

// Cleanups VM reservations in the database.
func (s *VirtualMachineService) CleanVMReservations(p *CleanVMReservationsParams) (*CleanVMReservationsResponse, error) {
	resp, err := s.cs.newRequest("cleanVMReservations", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CleanVMReservationsResponse
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

type CleanVMReservationsResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type DeployVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *DeployVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["affinitygroupids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("affinitygroupids", vv)
	}
	if v, found := p.p["affinitygroupnames"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("affinitygroupnames", vv)
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["datadiskofferinglist"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("datadiskofferinglist[%d].key", i), k)
			u.Set(fmt.Sprintf("datadiskofferinglist[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["deploymentplanner"]; found {
		u.Set("deploymentplanner", v.(string))
	}
	if v, found := p.p["details"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("details[%d].%s", i, k), vv)
			i++
		}
	}
	if v, found := p.p["dhcpoptionsnetworklist"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("dhcpoptionsnetworklist[%d].key", i), k)
			u.Set(fmt.Sprintf("dhcpoptionsnetworklist[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["displayname"]; found {
		u.Set("displayname", v.(string))
	}
	if v, found := p.p["displayvm"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvm", vv)
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["extraconfig"]; found {
		u.Set("extraconfig", v.(string))
	}
	if v, found := p.p["group"]; found {
		u.Set("group", v.(string))
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["hypervisor"]; found {
		u.Set("hypervisor", v.(string))
	}
	if v, found := p.p["ip6address"]; found {
		u.Set("ip6address", v.(string))
	}
	if v, found := p.p["ipaddress"]; found {
		u.Set("ipaddress", v.(string))
	}
	if v, found := p.p["iptonetworklist"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("iptonetworklist[%d].key", i), k)
			u.Set(fmt.Sprintf("iptonetworklist[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["keyboard"]; found {
		u.Set("keyboard", v.(string))
	}
	if v, found := p.p["keypair"]; found {
		u.Set("keypair", v.(string))
	}
	if v, found := p.p["macaddress"]; found {
		u.Set("macaddress", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["networkids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("networkids", vv)
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["rootdisksize"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("rootdisksize", vv)
	}
	if v, found := p.p["securitygroupids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("securitygroupids", vv)
	}
	if v, found := p.p["securitygroupnames"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("securitygroupnames", vv)
	}
	if v, found := p.p["serviceofferingid"]; found {
		u.Set("serviceofferingid", v.(string))
	}
	if v, found := p.p["size"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("size", vv)
	}
	if v, found := p.p["startvm"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("startvm", vv)
	}
	if v, found := p.p["templateid"]; found {
		u.Set("templateid", v.(string))
	}
	if v, found := p.p["userdata"]; found {
		u.Set("userdata", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *DeployVirtualMachineParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *DeployVirtualMachineParams) SetAffinitygroupids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["affinitygroupids"] = v
	return
}

func (p *DeployVirtualMachineParams) SetAffinitygroupnames(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["affinitygroupnames"] = v
	return
}

func (p *DeployVirtualMachineParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDatadiskofferinglist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["datadiskofferinglist"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDeploymentplanner(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["deploymentplanner"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDetails(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["details"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDhcpoptionsnetworklist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["dhcpoptionsnetworklist"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDisplayname(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayname"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDisplayvm(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvm"] = v
	return
}

func (p *DeployVirtualMachineParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *DeployVirtualMachineParams) SetExtraconfig(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["extraconfig"] = v
	return
}

func (p *DeployVirtualMachineParams) SetGroup(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["group"] = v
	return
}

func (p *DeployVirtualMachineParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *DeployVirtualMachineParams) SetHypervisor(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hypervisor"] = v
	return
}

func (p *DeployVirtualMachineParams) SetIp6address(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ip6address"] = v
	return
}

func (p *DeployVirtualMachineParams) SetIpaddress(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ipaddress"] = v
	return
}

func (p *DeployVirtualMachineParams) SetIptonetworklist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["iptonetworklist"] = v
	return
}

func (p *DeployVirtualMachineParams) SetKeyboard(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyboard"] = v
	return
}

func (p *DeployVirtualMachineParams) SetKeypair(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keypair"] = v
	return
}

func (p *DeployVirtualMachineParams) SetMacaddress(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["macaddress"] = v
	return
}

func (p *DeployVirtualMachineParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *DeployVirtualMachineParams) SetNetworkids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkids"] = v
	return
}

func (p *DeployVirtualMachineParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *DeployVirtualMachineParams) SetRootdisksize(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["rootdisksize"] = v
	return
}

func (p *DeployVirtualMachineParams) SetSecuritygroupids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["securitygroupids"] = v
	return
}

func (p *DeployVirtualMachineParams) SetSecuritygroupnames(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["securitygroupnames"] = v
	return
}

func (p *DeployVirtualMachineParams) SetServiceofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceofferingid"] = v
	return
}

func (p *DeployVirtualMachineParams) SetSize(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["size"] = v
	return
}

func (p *DeployVirtualMachineParams) SetStartvm(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["startvm"] = v
	return
}

func (p *DeployVirtualMachineParams) SetTemplateid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["templateid"] = v
	return
}

func (p *DeployVirtualMachineParams) SetUserdata(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["userdata"] = v
	return
}

func (p *DeployVirtualMachineParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new DeployVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewDeployVirtualMachineParams(serviceofferingid string, templateid string, zoneid string) *DeployVirtualMachineParams {
	p := &DeployVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["serviceofferingid"] = serviceofferingid
	p.p["templateid"] = templateid
	p.p["zoneid"] = zoneid
	return p
}

// Creates and automatically starts a virtual machine based on a service offering, disk offering, and template.
func (s *VirtualMachineService) DeployVirtualMachine(p *DeployVirtualMachineParams) (*DeployVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("deployVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeployVirtualMachineResponse
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

type DeployVirtualMachineResponse struct {
	Account               string                                      `json:"account"`
	Affinitygroup         []DeployVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                         `json:"cpunumber"`
	Cpuspeed              int                                         `json:"cpuspeed"`
	Cpuused               string                                      `json:"cpuused"`
	Created               string                                      `json:"created"`
	Details               map[string]string                           `json:"details"`
	Diskioread            int64                                       `json:"diskioread"`
	Diskiowrite           int64                                       `json:"diskiowrite"`
	Diskkbsread           int64                                       `json:"diskkbsread"`
	Diskkbswrite          int64                                       `json:"diskkbswrite"`
	Diskofferingid        string                                      `json:"diskofferingid"`
	Diskofferingname      string                                      `json:"diskofferingname"`
	Displayname           string                                      `json:"displayname"`
	Displayvm             bool                                        `json:"displayvm"`
	Domain                string                                      `json:"domain"`
	Domainid              string                                      `json:"domainid"`
	Forvirtualnetwork     bool                                        `json:"forvirtualnetwork"`
	Group                 string                                      `json:"group"`
	Groupid               string                                      `json:"groupid"`
	Guestosid             string                                      `json:"guestosid"`
	Haenable              bool                                        `json:"haenable"`
	Hostid                string                                      `json:"hostid"`
	Hostname              string                                      `json:"hostname"`
	Hypervisor            string                                      `json:"hypervisor"`
	Id                    string                                      `json:"id"`
	Instancename          string                                      `json:"instancename"`
	Isdynamicallyscalable bool                                        `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                      `json:"isodisplaytext"`
	Isoid                 string                                      `json:"isoid"`
	Isoname               string                                      `json:"isoname"`
	JobID                 string                                      `json:"jobid"`
	Jobstatus             int                                         `json:"jobstatus"`
	Keypair               string                                      `json:"keypair"`
	Memory                int                                         `json:"memory"`
	Memoryintfreekbs      int64                                       `json:"memoryintfreekbs"`
	Memorykbs             int64                                       `json:"memorykbs"`
	Memorytargetkbs       int64                                       `json:"memorytargetkbs"`
	Name                  string                                      `json:"name"`
	Networkkbsread        int64                                       `json:"networkkbsread"`
	Networkkbswrite       int64                                       `json:"networkkbswrite"`
	Nic                   []Nic                                       `json:"nic"`
	Ostypeid              string                                      `json:"ostypeid"`
	Password              string                                      `json:"password"`
	Passwordenabled       bool                                        `json:"passwordenabled"`
	Project               string                                      `json:"project"`
	Projectid             string                                      `json:"projectid"`
	Publicip              string                                      `json:"publicip"`
	Publicipid            string                                      `json:"publicipid"`
	Rootdeviceid          int64                                       `json:"rootdeviceid"`
	Rootdevicetype        string                                      `json:"rootdevicetype"`
	Securitygroup         []DeployVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                      `json:"serviceofferingid"`
	Serviceofferingname   string                                      `json:"serviceofferingname"`
	Servicestate          string                                      `json:"servicestate"`
	State                 string                                      `json:"state"`
	Tags                  []Tags                                      `json:"tags"`
	Templatedisplaytext   string                                      `json:"templatedisplaytext"`
	Templateid            string                                      `json:"templateid"`
	Templatename          string                                      `json:"templatename"`
	Userid                string                                      `json:"userid"`
	Username              string                                      `json:"username"`
	Vgpu                  string                                      `json:"vgpu"`
	Zoneid                string                                      `json:"zoneid"`
	Zonename              string                                      `json:"zonename"`
}

type DeployVirtualMachineResponseSecuritygroup struct {
	Account             string                                          `json:"account"`
	Description         string                                          `json:"description"`
	Domain              string                                          `json:"domain"`
	Domainid            string                                          `json:"domainid"`
	Egressrule          []DeployVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                          `json:"id"`
	Ingressrule         []DeployVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                          `json:"name"`
	Project             string                                          `json:"project"`
	Projectid           string                                          `json:"projectid"`
	Tags                []Tags                                          `json:"tags"`
	Virtualmachinecount int                                             `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                   `json:"virtualmachineids"`
}

type DeployVirtualMachineResponseSecuritygroupRule struct {
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

type DeployVirtualMachineResponseAffinitygroup struct {
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

func (r *DeployVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias DeployVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type DestroyVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *DestroyVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["expunge"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("expunge", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["volumeids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("volumeids", vv)
	}
	return u
}

func (p *DestroyVirtualMachineParams) SetExpunge(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["expunge"] = v
	return
}

func (p *DestroyVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *DestroyVirtualMachineParams) SetVolumeids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["volumeids"] = v
	return
}

// You should always use this function to get a new DestroyVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewDestroyVirtualMachineParams(id string) *DestroyVirtualMachineParams {
	p := &DestroyVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Destroys a virtual machine.
func (s *VirtualMachineService) DestroyVirtualMachine(p *DestroyVirtualMachineParams) (*DestroyVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("destroyVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DestroyVirtualMachineResponse
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

type DestroyVirtualMachineResponse struct {
	Account               string                                       `json:"account"`
	Affinitygroup         []DestroyVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
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
	Securitygroup         []DestroyVirtualMachineResponseSecuritygroup `json:"securitygroup"`
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

type DestroyVirtualMachineResponseSecuritygroup struct {
	Account             string                                           `json:"account"`
	Description         string                                           `json:"description"`
	Domain              string                                           `json:"domain"`
	Domainid            string                                           `json:"domainid"`
	Egressrule          []DestroyVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                           `json:"id"`
	Ingressrule         []DestroyVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                           `json:"name"`
	Project             string                                           `json:"project"`
	Projectid           string                                           `json:"projectid"`
	Tags                []Tags                                           `json:"tags"`
	Virtualmachinecount int                                              `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                    `json:"virtualmachineids"`
}

type DestroyVirtualMachineResponseSecuritygroupRule struct {
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

type DestroyVirtualMachineResponseAffinitygroup struct {
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

func (r *DestroyVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias DestroyVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ExpungeVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *ExpungeVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *ExpungeVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new ExpungeVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewExpungeVirtualMachineParams(id string) *ExpungeVirtualMachineParams {
	p := &ExpungeVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Expunge a virtual machine. Once expunged, it cannot be recoverd.
func (s *VirtualMachineService) ExpungeVirtualMachine(p *ExpungeVirtualMachineParams) (*ExpungeVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("expungeVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ExpungeVirtualMachineResponse
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

type ExpungeVirtualMachineResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type GetVMPasswordParams struct {
	p map[string]interface{}
}

func (p *GetVMPasswordParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *GetVMPasswordParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new GetVMPasswordParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewGetVMPasswordParams(id string) *GetVMPasswordParams {
	p := &GetVMPasswordParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Returns an encrypted password for the VM
func (s *VirtualMachineService) GetVMPassword(p *GetVMPasswordParams) (*GetVMPasswordResponse, error) {
	resp, err := s.cs.newRequest("getVMPassword", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r GetVMPasswordResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type GetVMPasswordResponse struct {
	Encryptedpassword string `json:"encryptedpassword"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
}

type ListVirtualMachinesParams struct {
	p map[string]interface{}
}

func (p *ListVirtualMachinesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["affinitygroupid"]; found {
		u.Set("affinitygroupid", v.(string))
	}
	if v, found := p.p["details"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("details", vv)
	}
	if v, found := p.p["displayvm"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvm", vv)
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["forvirtualnetwork"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("forvirtualnetwork", vv)
	}
	if v, found := p.p["groupid"]; found {
		u.Set("groupid", v.(string))
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["hypervisor"]; found {
		u.Set("hypervisor", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["ids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("ids", vv)
	}
	if v, found := p.p["isoid"]; found {
		u.Set("isoid", v.(string))
	}
	if v, found := p.p["isrecursive"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("isrecursive", vv)
	}
	if v, found := p.p["keypair"]; found {
		u.Set("keypair", v.(string))
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
	if v, found := p.p["serviceofferingid"]; found {
		u.Set("serviceofferingid", v.(string))
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["tags"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("tags[%d].key", i), k)
			u.Set(fmt.Sprintf("tags[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["templateid"]; found {
		u.Set("templateid", v.(string))
	}
	if v, found := p.p["userid"]; found {
		u.Set("userid", v.(string))
	}
	if v, found := p.p["vpcid"]; found {
		u.Set("vpcid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListVirtualMachinesParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListVirtualMachinesParams) SetAffinitygroupid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["affinitygroupid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetDetails(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["details"] = v
	return
}

func (p *ListVirtualMachinesParams) SetDisplayvm(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvm"] = v
	return
}

func (p *ListVirtualMachinesParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetForvirtualnetwork(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forvirtualnetwork"] = v
	return
}

func (p *ListVirtualMachinesParams) SetGroupid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["groupid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetHypervisor(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hypervisor"] = v
	return
}

func (p *ListVirtualMachinesParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListVirtualMachinesParams) SetIds(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ids"] = v
	return
}

func (p *ListVirtualMachinesParams) SetIsoid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isoid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListVirtualMachinesParams) SetKeypair(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keypair"] = v
	return
}

func (p *ListVirtualMachinesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListVirtualMachinesParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListVirtualMachinesParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListVirtualMachinesParams) SetNetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListVirtualMachinesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListVirtualMachinesParams) SetPodid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["podid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetServiceofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceofferingid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *ListVirtualMachinesParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetTags(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *ListVirtualMachinesParams) SetTemplateid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["templateid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetUserid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["userid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetVpcid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcid"] = v
	return
}

func (p *ListVirtualMachinesParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListVirtualMachinesParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewListVirtualMachinesParams() *ListVirtualMachinesParams {
	p := &ListVirtualMachinesParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VirtualMachineService) GetVirtualMachineID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListVirtualMachinesParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListVirtualMachines(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.VirtualMachines[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.VirtualMachines {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VirtualMachineService) GetVirtualMachineByName(name string, opts ...OptionFunc) (*VirtualMachine, int, error) {
	id, count, err := s.GetVirtualMachineID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetVirtualMachineByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VirtualMachineService) GetVirtualMachineByID(id string, opts ...OptionFunc) (*VirtualMachine, int, error) {
	p := &ListVirtualMachinesParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListVirtualMachines(p)
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
		return l.VirtualMachines[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for VirtualMachine UUID: %s!", id)
}

// List the virtual machines owned by the account.
func (s *VirtualMachineService) ListVirtualMachines(p *ListVirtualMachinesParams) (*ListVirtualMachinesResponse, error) {
	resp, err := s.cs.newRequest("listVirtualMachines", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListVirtualMachinesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListVirtualMachinesResponse struct {
	Count           int               `json:"count"`
	VirtualMachines []*VirtualMachine `json:"virtualmachine"`
}

type VirtualMachine struct {
	Account               string                        `json:"account"`
	Affinitygroup         []VirtualMachineAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                           `json:"cpunumber"`
	Cpuspeed              int                           `json:"cpuspeed"`
	Cpuused               string                        `json:"cpuused"`
	Created               string                        `json:"created"`
	Details               map[string]string             `json:"details"`
	Diskioread            int64                         `json:"diskioread"`
	Diskiowrite           int64                         `json:"diskiowrite"`
	Diskkbsread           int64                         `json:"diskkbsread"`
	Diskkbswrite          int64                         `json:"diskkbswrite"`
	Diskofferingid        string                        `json:"diskofferingid"`
	Diskofferingname      string                        `json:"diskofferingname"`
	Displayname           string                        `json:"displayname"`
	Displayvm             bool                          `json:"displayvm"`
	Domain                string                        `json:"domain"`
	Domainid              string                        `json:"domainid"`
	Forvirtualnetwork     bool                          `json:"forvirtualnetwork"`
	Group                 string                        `json:"group"`
	Groupid               string                        `json:"groupid"`
	Guestosid             string                        `json:"guestosid"`
	Haenable              bool                          `json:"haenable"`
	Hostid                string                        `json:"hostid"`
	Hostname              string                        `json:"hostname"`
	Hypervisor            string                        `json:"hypervisor"`
	Id                    string                        `json:"id"`
	Instancename          string                        `json:"instancename"`
	Isdynamicallyscalable bool                          `json:"isdynamicallyscalable"`
	Isodisplaytext        string                        `json:"isodisplaytext"`
	Isoid                 string                        `json:"isoid"`
	Isoname               string                        `json:"isoname"`
	JobID                 string                        `json:"jobid"`
	Jobstatus             int                           `json:"jobstatus"`
	Keypair               string                        `json:"keypair"`
	Memory                int                           `json:"memory"`
	Memoryintfreekbs      int64                         `json:"memoryintfreekbs"`
	Memorykbs             int64                         `json:"memorykbs"`
	Memorytargetkbs       int64                         `json:"memorytargetkbs"`
	Name                  string                        `json:"name"`
	Networkkbsread        int64                         `json:"networkkbsread"`
	Networkkbswrite       int64                         `json:"networkkbswrite"`
	Nic                   []Nic                         `json:"nic"`
	Ostypeid              string                        `json:"ostypeid"`
	Password              string                        `json:"password"`
	Passwordenabled       bool                          `json:"passwordenabled"`
	Project               string                        `json:"project"`
	Projectid             string                        `json:"projectid"`
	Publicip              string                        `json:"publicip"`
	Publicipid            string                        `json:"publicipid"`
	Rootdeviceid          int64                         `json:"rootdeviceid"`
	Rootdevicetype        string                        `json:"rootdevicetype"`
	Securitygroup         []VirtualMachineSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                        `json:"serviceofferingid"`
	Serviceofferingname   string                        `json:"serviceofferingname"`
	Servicestate          string                        `json:"servicestate"`
	State                 string                        `json:"state"`
	Tags                  []Tags                        `json:"tags"`
	Templatedisplaytext   string                        `json:"templatedisplaytext"`
	Templateid            string                        `json:"templateid"`
	Templatename          string                        `json:"templatename"`
	Userid                string                        `json:"userid"`
	Username              string                        `json:"username"`
	Vgpu                  string                        `json:"vgpu"`
	Zoneid                string                        `json:"zoneid"`
	Zonename              string                        `json:"zonename"`
}

type VirtualMachineSecuritygroup struct {
	Account             string                            `json:"account"`
	Description         string                            `json:"description"`
	Domain              string                            `json:"domain"`
	Domainid            string                            `json:"domainid"`
	Egressrule          []VirtualMachineSecuritygroupRule `json:"egressrule"`
	Id                  string                            `json:"id"`
	Ingressrule         []VirtualMachineSecuritygroupRule `json:"ingressrule"`
	Name                string                            `json:"name"`
	Project             string                            `json:"project"`
	Projectid           string                            `json:"projectid"`
	Tags                []Tags                            `json:"tags"`
	Virtualmachinecount int                               `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                     `json:"virtualmachineids"`
}

type VirtualMachineSecuritygroupRule struct {
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

type VirtualMachineAffinitygroup struct {
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

func (r *VirtualMachine) UnmarshalJSON(b []byte) error {
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

	type alias VirtualMachine
	return json.Unmarshal(b, (*alias)(r))
}

type MigrateVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *MigrateVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *MigrateVirtualMachineParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *MigrateVirtualMachineParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *MigrateVirtualMachineParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new MigrateVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewMigrateVirtualMachineParams(virtualmachineid string) *MigrateVirtualMachineParams {
	p := &MigrateVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Attempts Migration of a VM to a different host or Root volume of the vm to a different storage pool
func (s *VirtualMachineService) MigrateVirtualMachine(p *MigrateVirtualMachineParams) (*MigrateVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("migrateVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r MigrateVirtualMachineResponse
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

type MigrateVirtualMachineResponse struct {
	Account               string                                       `json:"account"`
	Affinitygroup         []MigrateVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
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
	Securitygroup         []MigrateVirtualMachineResponseSecuritygroup `json:"securitygroup"`
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

type MigrateVirtualMachineResponseSecuritygroup struct {
	Account             string                                           `json:"account"`
	Description         string                                           `json:"description"`
	Domain              string                                           `json:"domain"`
	Domainid            string                                           `json:"domainid"`
	Egressrule          []MigrateVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                           `json:"id"`
	Ingressrule         []MigrateVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                           `json:"name"`
	Project             string                                           `json:"project"`
	Projectid           string                                           `json:"projectid"`
	Tags                []Tags                                           `json:"tags"`
	Virtualmachinecount int                                              `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                    `json:"virtualmachineids"`
}

type MigrateVirtualMachineResponseSecuritygroupRule struct {
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

type MigrateVirtualMachineResponseAffinitygroup struct {
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

func (r *MigrateVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias MigrateVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type MigrateVirtualMachineWithVolumeParams struct {
	p map[string]interface{}
}

func (p *MigrateVirtualMachineWithVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["migrateto"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("migrateto[%d].key", i), k)
			u.Set(fmt.Sprintf("migrateto[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *MigrateVirtualMachineWithVolumeParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *MigrateVirtualMachineWithVolumeParams) SetMigrateto(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["migrateto"] = v
	return
}

func (p *MigrateVirtualMachineWithVolumeParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new MigrateVirtualMachineWithVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewMigrateVirtualMachineWithVolumeParams(hostid string, virtualmachineid string) *MigrateVirtualMachineWithVolumeParams {
	p := &MigrateVirtualMachineWithVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["hostid"] = hostid
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Attempts Migration of a VM with its volumes to a different host
func (s *VirtualMachineService) MigrateVirtualMachineWithVolume(p *MigrateVirtualMachineWithVolumeParams) (*MigrateVirtualMachineWithVolumeResponse, error) {
	resp, err := s.cs.newRequest("migrateVirtualMachineWithVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r MigrateVirtualMachineWithVolumeResponse
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

type MigrateVirtualMachineWithVolumeResponse struct {
	Account               string                                                 `json:"account"`
	Affinitygroup         []MigrateVirtualMachineWithVolumeResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                                    `json:"cpunumber"`
	Cpuspeed              int                                                    `json:"cpuspeed"`
	Cpuused               string                                                 `json:"cpuused"`
	Created               string                                                 `json:"created"`
	Details               map[string]string                                      `json:"details"`
	Diskioread            int64                                                  `json:"diskioread"`
	Diskiowrite           int64                                                  `json:"diskiowrite"`
	Diskkbsread           int64                                                  `json:"diskkbsread"`
	Diskkbswrite          int64                                                  `json:"diskkbswrite"`
	Diskofferingid        string                                                 `json:"diskofferingid"`
	Diskofferingname      string                                                 `json:"diskofferingname"`
	Displayname           string                                                 `json:"displayname"`
	Displayvm             bool                                                   `json:"displayvm"`
	Domain                string                                                 `json:"domain"`
	Domainid              string                                                 `json:"domainid"`
	Forvirtualnetwork     bool                                                   `json:"forvirtualnetwork"`
	Group                 string                                                 `json:"group"`
	Groupid               string                                                 `json:"groupid"`
	Guestosid             string                                                 `json:"guestosid"`
	Haenable              bool                                                   `json:"haenable"`
	Hostid                string                                                 `json:"hostid"`
	Hostname              string                                                 `json:"hostname"`
	Hypervisor            string                                                 `json:"hypervisor"`
	Id                    string                                                 `json:"id"`
	Instancename          string                                                 `json:"instancename"`
	Isdynamicallyscalable bool                                                   `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                                 `json:"isodisplaytext"`
	Isoid                 string                                                 `json:"isoid"`
	Isoname               string                                                 `json:"isoname"`
	JobID                 string                                                 `json:"jobid"`
	Jobstatus             int                                                    `json:"jobstatus"`
	Keypair               string                                                 `json:"keypair"`
	Memory                int                                                    `json:"memory"`
	Memoryintfreekbs      int64                                                  `json:"memoryintfreekbs"`
	Memorykbs             int64                                                  `json:"memorykbs"`
	Memorytargetkbs       int64                                                  `json:"memorytargetkbs"`
	Name                  string                                                 `json:"name"`
	Networkkbsread        int64                                                  `json:"networkkbsread"`
	Networkkbswrite       int64                                                  `json:"networkkbswrite"`
	Nic                   []Nic                                                  `json:"nic"`
	Ostypeid              string                                                 `json:"ostypeid"`
	Password              string                                                 `json:"password"`
	Passwordenabled       bool                                                   `json:"passwordenabled"`
	Project               string                                                 `json:"project"`
	Projectid             string                                                 `json:"projectid"`
	Publicip              string                                                 `json:"publicip"`
	Publicipid            string                                                 `json:"publicipid"`
	Rootdeviceid          int64                                                  `json:"rootdeviceid"`
	Rootdevicetype        string                                                 `json:"rootdevicetype"`
	Securitygroup         []MigrateVirtualMachineWithVolumeResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                                 `json:"serviceofferingid"`
	Serviceofferingname   string                                                 `json:"serviceofferingname"`
	Servicestate          string                                                 `json:"servicestate"`
	State                 string                                                 `json:"state"`
	Tags                  []Tags                                                 `json:"tags"`
	Templatedisplaytext   string                                                 `json:"templatedisplaytext"`
	Templateid            string                                                 `json:"templateid"`
	Templatename          string                                                 `json:"templatename"`
	Userid                string                                                 `json:"userid"`
	Username              string                                                 `json:"username"`
	Vgpu                  string                                                 `json:"vgpu"`
	Zoneid                string                                                 `json:"zoneid"`
	Zonename              string                                                 `json:"zonename"`
}

type MigrateVirtualMachineWithVolumeResponseSecuritygroup struct {
	Account             string                                                     `json:"account"`
	Description         string                                                     `json:"description"`
	Domain              string                                                     `json:"domain"`
	Domainid            string                                                     `json:"domainid"`
	Egressrule          []MigrateVirtualMachineWithVolumeResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                                     `json:"id"`
	Ingressrule         []MigrateVirtualMachineWithVolumeResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                                     `json:"name"`
	Project             string                                                     `json:"project"`
	Projectid           string                                                     `json:"projectid"`
	Tags                []Tags                                                     `json:"tags"`
	Virtualmachinecount int                                                        `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                              `json:"virtualmachineids"`
}

type MigrateVirtualMachineWithVolumeResponseSecuritygroupRule struct {
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

type MigrateVirtualMachineWithVolumeResponseAffinitygroup struct {
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

func (r *MigrateVirtualMachineWithVolumeResponse) UnmarshalJSON(b []byte) error {
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

	type alias MigrateVirtualMachineWithVolumeResponse
	return json.Unmarshal(b, (*alias)(r))
}

type RebootVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *RebootVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *RebootVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new RebootVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewRebootVirtualMachineParams(id string) *RebootVirtualMachineParams {
	p := &RebootVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Reboots a virtual machine.
func (s *VirtualMachineService) RebootVirtualMachine(p *RebootVirtualMachineParams) (*RebootVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("rebootVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r RebootVirtualMachineResponse
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

type RebootVirtualMachineResponse struct {
	Account               string                                      `json:"account"`
	Affinitygroup         []RebootVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                         `json:"cpunumber"`
	Cpuspeed              int                                         `json:"cpuspeed"`
	Cpuused               string                                      `json:"cpuused"`
	Created               string                                      `json:"created"`
	Details               map[string]string                           `json:"details"`
	Diskioread            int64                                       `json:"diskioread"`
	Diskiowrite           int64                                       `json:"diskiowrite"`
	Diskkbsread           int64                                       `json:"diskkbsread"`
	Diskkbswrite          int64                                       `json:"diskkbswrite"`
	Diskofferingid        string                                      `json:"diskofferingid"`
	Diskofferingname      string                                      `json:"diskofferingname"`
	Displayname           string                                      `json:"displayname"`
	Displayvm             bool                                        `json:"displayvm"`
	Domain                string                                      `json:"domain"`
	Domainid              string                                      `json:"domainid"`
	Forvirtualnetwork     bool                                        `json:"forvirtualnetwork"`
	Group                 string                                      `json:"group"`
	Groupid               string                                      `json:"groupid"`
	Guestosid             string                                      `json:"guestosid"`
	Haenable              bool                                        `json:"haenable"`
	Hostid                string                                      `json:"hostid"`
	Hostname              string                                      `json:"hostname"`
	Hypervisor            string                                      `json:"hypervisor"`
	Id                    string                                      `json:"id"`
	Instancename          string                                      `json:"instancename"`
	Isdynamicallyscalable bool                                        `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                      `json:"isodisplaytext"`
	Isoid                 string                                      `json:"isoid"`
	Isoname               string                                      `json:"isoname"`
	JobID                 string                                      `json:"jobid"`
	Jobstatus             int                                         `json:"jobstatus"`
	Keypair               string                                      `json:"keypair"`
	Memory                int                                         `json:"memory"`
	Memoryintfreekbs      int64                                       `json:"memoryintfreekbs"`
	Memorykbs             int64                                       `json:"memorykbs"`
	Memorytargetkbs       int64                                       `json:"memorytargetkbs"`
	Name                  string                                      `json:"name"`
	Networkkbsread        int64                                       `json:"networkkbsread"`
	Networkkbswrite       int64                                       `json:"networkkbswrite"`
	Nic                   []Nic                                       `json:"nic"`
	Ostypeid              string                                      `json:"ostypeid"`
	Password              string                                      `json:"password"`
	Passwordenabled       bool                                        `json:"passwordenabled"`
	Project               string                                      `json:"project"`
	Projectid             string                                      `json:"projectid"`
	Publicip              string                                      `json:"publicip"`
	Publicipid            string                                      `json:"publicipid"`
	Rootdeviceid          int64                                       `json:"rootdeviceid"`
	Rootdevicetype        string                                      `json:"rootdevicetype"`
	Securitygroup         []RebootVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                      `json:"serviceofferingid"`
	Serviceofferingname   string                                      `json:"serviceofferingname"`
	Servicestate          string                                      `json:"servicestate"`
	State                 string                                      `json:"state"`
	Tags                  []Tags                                      `json:"tags"`
	Templatedisplaytext   string                                      `json:"templatedisplaytext"`
	Templateid            string                                      `json:"templateid"`
	Templatename          string                                      `json:"templatename"`
	Userid                string                                      `json:"userid"`
	Username              string                                      `json:"username"`
	Vgpu                  string                                      `json:"vgpu"`
	Zoneid                string                                      `json:"zoneid"`
	Zonename              string                                      `json:"zonename"`
}

type RebootVirtualMachineResponseSecuritygroup struct {
	Account             string                                          `json:"account"`
	Description         string                                          `json:"description"`
	Domain              string                                          `json:"domain"`
	Domainid            string                                          `json:"domainid"`
	Egressrule          []RebootVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                          `json:"id"`
	Ingressrule         []RebootVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                          `json:"name"`
	Project             string                                          `json:"project"`
	Projectid           string                                          `json:"projectid"`
	Tags                []Tags                                          `json:"tags"`
	Virtualmachinecount int                                             `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                   `json:"virtualmachineids"`
}

type RebootVirtualMachineResponseSecuritygroupRule struct {
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

type RebootVirtualMachineResponseAffinitygroup struct {
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

func (r *RebootVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias RebootVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type RecoverVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *RecoverVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *RecoverVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new RecoverVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewRecoverVirtualMachineParams(id string) *RecoverVirtualMachineParams {
	p := &RecoverVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Recovers a virtual machine.
func (s *VirtualMachineService) RecoverVirtualMachine(p *RecoverVirtualMachineParams) (*RecoverVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("recoverVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r RecoverVirtualMachineResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type RecoverVirtualMachineResponse struct {
	Account               string                                       `json:"account"`
	Affinitygroup         []RecoverVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
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
	Securitygroup         []RecoverVirtualMachineResponseSecuritygroup `json:"securitygroup"`
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

type RecoverVirtualMachineResponseSecuritygroup struct {
	Account             string                                           `json:"account"`
	Description         string                                           `json:"description"`
	Domain              string                                           `json:"domain"`
	Domainid            string                                           `json:"domainid"`
	Egressrule          []RecoverVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                           `json:"id"`
	Ingressrule         []RecoverVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                           `json:"name"`
	Project             string                                           `json:"project"`
	Projectid           string                                           `json:"projectid"`
	Tags                []Tags                                           `json:"tags"`
	Virtualmachinecount int                                              `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                    `json:"virtualmachineids"`
}

type RecoverVirtualMachineResponseSecuritygroupRule struct {
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

type RecoverVirtualMachineResponseAffinitygroup struct {
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

func (r *RecoverVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias RecoverVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type RemoveNicFromVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *RemoveNicFromVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["nicid"]; found {
		u.Set("nicid", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *RemoveNicFromVirtualMachineParams) SetNicid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["nicid"] = v
	return
}

func (p *RemoveNicFromVirtualMachineParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new RemoveNicFromVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewRemoveNicFromVirtualMachineParams(nicid string, virtualmachineid string) *RemoveNicFromVirtualMachineParams {
	p := &RemoveNicFromVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["nicid"] = nicid
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Removes VM from specified network by deleting a NIC
func (s *VirtualMachineService) RemoveNicFromVirtualMachine(p *RemoveNicFromVirtualMachineParams) (*RemoveNicFromVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("removeNicFromVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r RemoveNicFromVirtualMachineResponse
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

type RemoveNicFromVirtualMachineResponse struct {
	Account               string                                             `json:"account"`
	Affinitygroup         []RemoveNicFromVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                                `json:"cpunumber"`
	Cpuspeed              int                                                `json:"cpuspeed"`
	Cpuused               string                                             `json:"cpuused"`
	Created               string                                             `json:"created"`
	Details               map[string]string                                  `json:"details"`
	Diskioread            int64                                              `json:"diskioread"`
	Diskiowrite           int64                                              `json:"diskiowrite"`
	Diskkbsread           int64                                              `json:"diskkbsread"`
	Diskkbswrite          int64                                              `json:"diskkbswrite"`
	Diskofferingid        string                                             `json:"diskofferingid"`
	Diskofferingname      string                                             `json:"diskofferingname"`
	Displayname           string                                             `json:"displayname"`
	Displayvm             bool                                               `json:"displayvm"`
	Domain                string                                             `json:"domain"`
	Domainid              string                                             `json:"domainid"`
	Forvirtualnetwork     bool                                               `json:"forvirtualnetwork"`
	Group                 string                                             `json:"group"`
	Groupid               string                                             `json:"groupid"`
	Guestosid             string                                             `json:"guestosid"`
	Haenable              bool                                               `json:"haenable"`
	Hostid                string                                             `json:"hostid"`
	Hostname              string                                             `json:"hostname"`
	Hypervisor            string                                             `json:"hypervisor"`
	Id                    string                                             `json:"id"`
	Instancename          string                                             `json:"instancename"`
	Isdynamicallyscalable bool                                               `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                             `json:"isodisplaytext"`
	Isoid                 string                                             `json:"isoid"`
	Isoname               string                                             `json:"isoname"`
	JobID                 string                                             `json:"jobid"`
	Jobstatus             int                                                `json:"jobstatus"`
	Keypair               string                                             `json:"keypair"`
	Memory                int                                                `json:"memory"`
	Memoryintfreekbs      int64                                              `json:"memoryintfreekbs"`
	Memorykbs             int64                                              `json:"memorykbs"`
	Memorytargetkbs       int64                                              `json:"memorytargetkbs"`
	Name                  string                                             `json:"name"`
	Networkkbsread        int64                                              `json:"networkkbsread"`
	Networkkbswrite       int64                                              `json:"networkkbswrite"`
	Nic                   []Nic                                              `json:"nic"`
	Ostypeid              string                                             `json:"ostypeid"`
	Password              string                                             `json:"password"`
	Passwordenabled       bool                                               `json:"passwordenabled"`
	Project               string                                             `json:"project"`
	Projectid             string                                             `json:"projectid"`
	Publicip              string                                             `json:"publicip"`
	Publicipid            string                                             `json:"publicipid"`
	Rootdeviceid          int64                                              `json:"rootdeviceid"`
	Rootdevicetype        string                                             `json:"rootdevicetype"`
	Securitygroup         []RemoveNicFromVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                             `json:"serviceofferingid"`
	Serviceofferingname   string                                             `json:"serviceofferingname"`
	Servicestate          string                                             `json:"servicestate"`
	State                 string                                             `json:"state"`
	Tags                  []Tags                                             `json:"tags"`
	Templatedisplaytext   string                                             `json:"templatedisplaytext"`
	Templateid            string                                             `json:"templateid"`
	Templatename          string                                             `json:"templatename"`
	Userid                string                                             `json:"userid"`
	Username              string                                             `json:"username"`
	Vgpu                  string                                             `json:"vgpu"`
	Zoneid                string                                             `json:"zoneid"`
	Zonename              string                                             `json:"zonename"`
}

type RemoveNicFromVirtualMachineResponseSecuritygroup struct {
	Account             string                                                 `json:"account"`
	Description         string                                                 `json:"description"`
	Domain              string                                                 `json:"domain"`
	Domainid            string                                                 `json:"domainid"`
	Egressrule          []RemoveNicFromVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                                 `json:"id"`
	Ingressrule         []RemoveNicFromVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                                 `json:"name"`
	Project             string                                                 `json:"project"`
	Projectid           string                                                 `json:"projectid"`
	Tags                []Tags                                                 `json:"tags"`
	Virtualmachinecount int                                                    `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                          `json:"virtualmachineids"`
}

type RemoveNicFromVirtualMachineResponseSecuritygroupRule struct {
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

type RemoveNicFromVirtualMachineResponseAffinitygroup struct {
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

func (r *RemoveNicFromVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias RemoveNicFromVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ResetPasswordForVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *ResetPasswordForVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *ResetPasswordForVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new ResetPasswordForVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewResetPasswordForVirtualMachineParams(id string) *ResetPasswordForVirtualMachineParams {
	p := &ResetPasswordForVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Resets the password for virtual machine. The virtual machine must be in a "Stopped" state and the template must already support this feature for this command to take effect. [async]
func (s *VirtualMachineService) ResetPasswordForVirtualMachine(p *ResetPasswordForVirtualMachineParams) (*ResetPasswordForVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("resetPasswordForVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ResetPasswordForVirtualMachineResponse
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

type ResetPasswordForVirtualMachineResponse struct {
	Account               string                                                `json:"account"`
	Affinitygroup         []ResetPasswordForVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                                   `json:"cpunumber"`
	Cpuspeed              int                                                   `json:"cpuspeed"`
	Cpuused               string                                                `json:"cpuused"`
	Created               string                                                `json:"created"`
	Details               map[string]string                                     `json:"details"`
	Diskioread            int64                                                 `json:"diskioread"`
	Diskiowrite           int64                                                 `json:"diskiowrite"`
	Diskkbsread           int64                                                 `json:"diskkbsread"`
	Diskkbswrite          int64                                                 `json:"diskkbswrite"`
	Diskofferingid        string                                                `json:"diskofferingid"`
	Diskofferingname      string                                                `json:"diskofferingname"`
	Displayname           string                                                `json:"displayname"`
	Displayvm             bool                                                  `json:"displayvm"`
	Domain                string                                                `json:"domain"`
	Domainid              string                                                `json:"domainid"`
	Forvirtualnetwork     bool                                                  `json:"forvirtualnetwork"`
	Group                 string                                                `json:"group"`
	Groupid               string                                                `json:"groupid"`
	Guestosid             string                                                `json:"guestosid"`
	Haenable              bool                                                  `json:"haenable"`
	Hostid                string                                                `json:"hostid"`
	Hostname              string                                                `json:"hostname"`
	Hypervisor            string                                                `json:"hypervisor"`
	Id                    string                                                `json:"id"`
	Instancename          string                                                `json:"instancename"`
	Isdynamicallyscalable bool                                                  `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                                `json:"isodisplaytext"`
	Isoid                 string                                                `json:"isoid"`
	Isoname               string                                                `json:"isoname"`
	JobID                 string                                                `json:"jobid"`
	Jobstatus             int                                                   `json:"jobstatus"`
	Keypair               string                                                `json:"keypair"`
	Memory                int                                                   `json:"memory"`
	Memoryintfreekbs      int64                                                 `json:"memoryintfreekbs"`
	Memorykbs             int64                                                 `json:"memorykbs"`
	Memorytargetkbs       int64                                                 `json:"memorytargetkbs"`
	Name                  string                                                `json:"name"`
	Networkkbsread        int64                                                 `json:"networkkbsread"`
	Networkkbswrite       int64                                                 `json:"networkkbswrite"`
	Nic                   []Nic                                                 `json:"nic"`
	Ostypeid              string                                                `json:"ostypeid"`
	Password              string                                                `json:"password"`
	Passwordenabled       bool                                                  `json:"passwordenabled"`
	Project               string                                                `json:"project"`
	Projectid             string                                                `json:"projectid"`
	Publicip              string                                                `json:"publicip"`
	Publicipid            string                                                `json:"publicipid"`
	Rootdeviceid          int64                                                 `json:"rootdeviceid"`
	Rootdevicetype        string                                                `json:"rootdevicetype"`
	Securitygroup         []ResetPasswordForVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                                `json:"serviceofferingid"`
	Serviceofferingname   string                                                `json:"serviceofferingname"`
	Servicestate          string                                                `json:"servicestate"`
	State                 string                                                `json:"state"`
	Tags                  []Tags                                                `json:"tags"`
	Templatedisplaytext   string                                                `json:"templatedisplaytext"`
	Templateid            string                                                `json:"templateid"`
	Templatename          string                                                `json:"templatename"`
	Userid                string                                                `json:"userid"`
	Username              string                                                `json:"username"`
	Vgpu                  string                                                `json:"vgpu"`
	Zoneid                string                                                `json:"zoneid"`
	Zonename              string                                                `json:"zonename"`
}

type ResetPasswordForVirtualMachineResponseSecuritygroup struct {
	Account             string                                                    `json:"account"`
	Description         string                                                    `json:"description"`
	Domain              string                                                    `json:"domain"`
	Domainid            string                                                    `json:"domainid"`
	Egressrule          []ResetPasswordForVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                                    `json:"id"`
	Ingressrule         []ResetPasswordForVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                                    `json:"name"`
	Project             string                                                    `json:"project"`
	Projectid           string                                                    `json:"projectid"`
	Tags                []Tags                                                    `json:"tags"`
	Virtualmachinecount int                                                       `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                             `json:"virtualmachineids"`
}

type ResetPasswordForVirtualMachineResponseSecuritygroupRule struct {
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

type ResetPasswordForVirtualMachineResponseAffinitygroup struct {
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

func (r *ResetPasswordForVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias ResetPasswordForVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type RestoreVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *RestoreVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["templateid"]; found {
		u.Set("templateid", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *RestoreVirtualMachineParams) SetTemplateid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["templateid"] = v
	return
}

func (p *RestoreVirtualMachineParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new RestoreVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewRestoreVirtualMachineParams(virtualmachineid string) *RestoreVirtualMachineParams {
	p := &RestoreVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Restore a VM to original template/ISO or new template/ISO
func (s *VirtualMachineService) RestoreVirtualMachine(p *RestoreVirtualMachineParams) (*RestoreVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("restoreVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r RestoreVirtualMachineResponse
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

type RestoreVirtualMachineResponse struct {
	Account               string                                       `json:"account"`
	Affinitygroup         []RestoreVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
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
	Securitygroup         []RestoreVirtualMachineResponseSecuritygroup `json:"securitygroup"`
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

type RestoreVirtualMachineResponseSecuritygroup struct {
	Account             string                                           `json:"account"`
	Description         string                                           `json:"description"`
	Domain              string                                           `json:"domain"`
	Domainid            string                                           `json:"domainid"`
	Egressrule          []RestoreVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                           `json:"id"`
	Ingressrule         []RestoreVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                           `json:"name"`
	Project             string                                           `json:"project"`
	Projectid           string                                           `json:"projectid"`
	Tags                []Tags                                           `json:"tags"`
	Virtualmachinecount int                                              `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                    `json:"virtualmachineids"`
}

type RestoreVirtualMachineResponseSecuritygroupRule struct {
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

type RestoreVirtualMachineResponseAffinitygroup struct {
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

func (r *RestoreVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias RestoreVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ScaleVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *ScaleVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["details"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("details[%d].%s", i, k), vv)
			i++
		}
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["serviceofferingid"]; found {
		u.Set("serviceofferingid", v.(string))
	}
	return u
}

func (p *ScaleVirtualMachineParams) SetDetails(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["details"] = v
	return
}

func (p *ScaleVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ScaleVirtualMachineParams) SetServiceofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceofferingid"] = v
	return
}

// You should always use this function to get a new ScaleVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewScaleVirtualMachineParams(id string, serviceofferingid string) *ScaleVirtualMachineParams {
	p := &ScaleVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	p.p["serviceofferingid"] = serviceofferingid
	return p
}

// Scales the virtual machine to a new service offering.
func (s *VirtualMachineService) ScaleVirtualMachine(p *ScaleVirtualMachineParams) (*ScaleVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("scaleVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ScaleVirtualMachineResponse
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

type ScaleVirtualMachineResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type StartVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *StartVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["deploymentplanner"]; found {
		u.Set("deploymentplanner", v.(string))
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *StartVirtualMachineParams) SetDeploymentplanner(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["deploymentplanner"] = v
	return
}

func (p *StartVirtualMachineParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *StartVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new StartVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewStartVirtualMachineParams(id string) *StartVirtualMachineParams {
	p := &StartVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Starts a virtual machine.
func (s *VirtualMachineService) StartVirtualMachine(p *StartVirtualMachineParams) (*StartVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("startVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r StartVirtualMachineResponse
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

type StartVirtualMachineResponse struct {
	Account               string                                     `json:"account"`
	Affinitygroup         []StartVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                        `json:"cpunumber"`
	Cpuspeed              int                                        `json:"cpuspeed"`
	Cpuused               string                                     `json:"cpuused"`
	Created               string                                     `json:"created"`
	Details               map[string]string                          `json:"details"`
	Diskioread            int64                                      `json:"diskioread"`
	Diskiowrite           int64                                      `json:"diskiowrite"`
	Diskkbsread           int64                                      `json:"diskkbsread"`
	Diskkbswrite          int64                                      `json:"diskkbswrite"`
	Diskofferingid        string                                     `json:"diskofferingid"`
	Diskofferingname      string                                     `json:"diskofferingname"`
	Displayname           string                                     `json:"displayname"`
	Displayvm             bool                                       `json:"displayvm"`
	Domain                string                                     `json:"domain"`
	Domainid              string                                     `json:"domainid"`
	Forvirtualnetwork     bool                                       `json:"forvirtualnetwork"`
	Group                 string                                     `json:"group"`
	Groupid               string                                     `json:"groupid"`
	Guestosid             string                                     `json:"guestosid"`
	Haenable              bool                                       `json:"haenable"`
	Hostid                string                                     `json:"hostid"`
	Hostname              string                                     `json:"hostname"`
	Hypervisor            string                                     `json:"hypervisor"`
	Id                    string                                     `json:"id"`
	Instancename          string                                     `json:"instancename"`
	Isdynamicallyscalable bool                                       `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                     `json:"isodisplaytext"`
	Isoid                 string                                     `json:"isoid"`
	Isoname               string                                     `json:"isoname"`
	JobID                 string                                     `json:"jobid"`
	Jobstatus             int                                        `json:"jobstatus"`
	Keypair               string                                     `json:"keypair"`
	Memory                int                                        `json:"memory"`
	Memoryintfreekbs      int64                                      `json:"memoryintfreekbs"`
	Memorykbs             int64                                      `json:"memorykbs"`
	Memorytargetkbs       int64                                      `json:"memorytargetkbs"`
	Name                  string                                     `json:"name"`
	Networkkbsread        int64                                      `json:"networkkbsread"`
	Networkkbswrite       int64                                      `json:"networkkbswrite"`
	Nic                   []Nic                                      `json:"nic"`
	Ostypeid              string                                     `json:"ostypeid"`
	Password              string                                     `json:"password"`
	Passwordenabled       bool                                       `json:"passwordenabled"`
	Project               string                                     `json:"project"`
	Projectid             string                                     `json:"projectid"`
	Publicip              string                                     `json:"publicip"`
	Publicipid            string                                     `json:"publicipid"`
	Rootdeviceid          int64                                      `json:"rootdeviceid"`
	Rootdevicetype        string                                     `json:"rootdevicetype"`
	Securitygroup         []StartVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                     `json:"serviceofferingid"`
	Serviceofferingname   string                                     `json:"serviceofferingname"`
	Servicestate          string                                     `json:"servicestate"`
	State                 string                                     `json:"state"`
	Tags                  []Tags                                     `json:"tags"`
	Templatedisplaytext   string                                     `json:"templatedisplaytext"`
	Templateid            string                                     `json:"templateid"`
	Templatename          string                                     `json:"templatename"`
	Userid                string                                     `json:"userid"`
	Username              string                                     `json:"username"`
	Vgpu                  string                                     `json:"vgpu"`
	Zoneid                string                                     `json:"zoneid"`
	Zonename              string                                     `json:"zonename"`
}

type StartVirtualMachineResponseSecuritygroup struct {
	Account             string                                         `json:"account"`
	Description         string                                         `json:"description"`
	Domain              string                                         `json:"domain"`
	Domainid            string                                         `json:"domainid"`
	Egressrule          []StartVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                         `json:"id"`
	Ingressrule         []StartVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                         `json:"name"`
	Project             string                                         `json:"project"`
	Projectid           string                                         `json:"projectid"`
	Tags                []Tags                                         `json:"tags"`
	Virtualmachinecount int                                            `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                  `json:"virtualmachineids"`
}

type StartVirtualMachineResponseSecuritygroupRule struct {
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

type StartVirtualMachineResponseAffinitygroup struct {
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

func (r *StartVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias StartVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type StopVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *StopVirtualMachineParams) toURLValues() url.Values {
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

func (p *StopVirtualMachineParams) SetForced(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forced"] = v
	return
}

func (p *StopVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new StopVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewStopVirtualMachineParams(id string) *StopVirtualMachineParams {
	p := &StopVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Stops a virtual machine.
func (s *VirtualMachineService) StopVirtualMachine(p *StopVirtualMachineParams) (*StopVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("stopVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r StopVirtualMachineResponse
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

type StopVirtualMachineResponse struct {
	Account               string                                    `json:"account"`
	Affinitygroup         []StopVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                       `json:"cpunumber"`
	Cpuspeed              int                                       `json:"cpuspeed"`
	Cpuused               string                                    `json:"cpuused"`
	Created               string                                    `json:"created"`
	Details               map[string]string                         `json:"details"`
	Diskioread            int64                                     `json:"diskioread"`
	Diskiowrite           int64                                     `json:"diskiowrite"`
	Diskkbsread           int64                                     `json:"diskkbsread"`
	Diskkbswrite          int64                                     `json:"diskkbswrite"`
	Diskofferingid        string                                    `json:"diskofferingid"`
	Diskofferingname      string                                    `json:"diskofferingname"`
	Displayname           string                                    `json:"displayname"`
	Displayvm             bool                                      `json:"displayvm"`
	Domain                string                                    `json:"domain"`
	Domainid              string                                    `json:"domainid"`
	Forvirtualnetwork     bool                                      `json:"forvirtualnetwork"`
	Group                 string                                    `json:"group"`
	Groupid               string                                    `json:"groupid"`
	Guestosid             string                                    `json:"guestosid"`
	Haenable              bool                                      `json:"haenable"`
	Hostid                string                                    `json:"hostid"`
	Hostname              string                                    `json:"hostname"`
	Hypervisor            string                                    `json:"hypervisor"`
	Id                    string                                    `json:"id"`
	Instancename          string                                    `json:"instancename"`
	Isdynamicallyscalable bool                                      `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                    `json:"isodisplaytext"`
	Isoid                 string                                    `json:"isoid"`
	Isoname               string                                    `json:"isoname"`
	JobID                 string                                    `json:"jobid"`
	Jobstatus             int                                       `json:"jobstatus"`
	Keypair               string                                    `json:"keypair"`
	Memory                int                                       `json:"memory"`
	Memoryintfreekbs      int64                                     `json:"memoryintfreekbs"`
	Memorykbs             int64                                     `json:"memorykbs"`
	Memorytargetkbs       int64                                     `json:"memorytargetkbs"`
	Name                  string                                    `json:"name"`
	Networkkbsread        int64                                     `json:"networkkbsread"`
	Networkkbswrite       int64                                     `json:"networkkbswrite"`
	Nic                   []Nic                                     `json:"nic"`
	Ostypeid              string                                    `json:"ostypeid"`
	Password              string                                    `json:"password"`
	Passwordenabled       bool                                      `json:"passwordenabled"`
	Project               string                                    `json:"project"`
	Projectid             string                                    `json:"projectid"`
	Publicip              string                                    `json:"publicip"`
	Publicipid            string                                    `json:"publicipid"`
	Rootdeviceid          int64                                     `json:"rootdeviceid"`
	Rootdevicetype        string                                    `json:"rootdevicetype"`
	Securitygroup         []StopVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                    `json:"serviceofferingid"`
	Serviceofferingname   string                                    `json:"serviceofferingname"`
	Servicestate          string                                    `json:"servicestate"`
	State                 string                                    `json:"state"`
	Tags                  []Tags                                    `json:"tags"`
	Templatedisplaytext   string                                    `json:"templatedisplaytext"`
	Templateid            string                                    `json:"templateid"`
	Templatename          string                                    `json:"templatename"`
	Userid                string                                    `json:"userid"`
	Username              string                                    `json:"username"`
	Vgpu                  string                                    `json:"vgpu"`
	Zoneid                string                                    `json:"zoneid"`
	Zonename              string                                    `json:"zonename"`
}

type StopVirtualMachineResponseSecuritygroup struct {
	Account             string                                        `json:"account"`
	Description         string                                        `json:"description"`
	Domain              string                                        `json:"domain"`
	Domainid            string                                        `json:"domainid"`
	Egressrule          []StopVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                        `json:"id"`
	Ingressrule         []StopVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                        `json:"name"`
	Project             string                                        `json:"project"`
	Projectid           string                                        `json:"projectid"`
	Tags                []Tags                                        `json:"tags"`
	Virtualmachinecount int                                           `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                 `json:"virtualmachineids"`
}

type StopVirtualMachineResponseSecuritygroupRule struct {
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

type StopVirtualMachineResponseAffinitygroup struct {
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

func (r *StopVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias StopVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type UpdateDefaultNicForVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *UpdateDefaultNicForVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["nicid"]; found {
		u.Set("nicid", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *UpdateDefaultNicForVirtualMachineParams) SetNicid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["nicid"] = v
	return
}

func (p *UpdateDefaultNicForVirtualMachineParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new UpdateDefaultNicForVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewUpdateDefaultNicForVirtualMachineParams(nicid string, virtualmachineid string) *UpdateDefaultNicForVirtualMachineParams {
	p := &UpdateDefaultNicForVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["nicid"] = nicid
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Changes the default NIC on a VM
func (s *VirtualMachineService) UpdateDefaultNicForVirtualMachine(p *UpdateDefaultNicForVirtualMachineParams) (*UpdateDefaultNicForVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("updateDefaultNicForVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateDefaultNicForVirtualMachineResponse
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

type UpdateDefaultNicForVirtualMachineResponse struct {
	Account               string                                                   `json:"account"`
	Affinitygroup         []UpdateDefaultNicForVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                                      `json:"cpunumber"`
	Cpuspeed              int                                                      `json:"cpuspeed"`
	Cpuused               string                                                   `json:"cpuused"`
	Created               string                                                   `json:"created"`
	Details               map[string]string                                        `json:"details"`
	Diskioread            int64                                                    `json:"diskioread"`
	Diskiowrite           int64                                                    `json:"diskiowrite"`
	Diskkbsread           int64                                                    `json:"diskkbsread"`
	Diskkbswrite          int64                                                    `json:"diskkbswrite"`
	Diskofferingid        string                                                   `json:"diskofferingid"`
	Diskofferingname      string                                                   `json:"diskofferingname"`
	Displayname           string                                                   `json:"displayname"`
	Displayvm             bool                                                     `json:"displayvm"`
	Domain                string                                                   `json:"domain"`
	Domainid              string                                                   `json:"domainid"`
	Forvirtualnetwork     bool                                                     `json:"forvirtualnetwork"`
	Group                 string                                                   `json:"group"`
	Groupid               string                                                   `json:"groupid"`
	Guestosid             string                                                   `json:"guestosid"`
	Haenable              bool                                                     `json:"haenable"`
	Hostid                string                                                   `json:"hostid"`
	Hostname              string                                                   `json:"hostname"`
	Hypervisor            string                                                   `json:"hypervisor"`
	Id                    string                                                   `json:"id"`
	Instancename          string                                                   `json:"instancename"`
	Isdynamicallyscalable bool                                                     `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                                   `json:"isodisplaytext"`
	Isoid                 string                                                   `json:"isoid"`
	Isoname               string                                                   `json:"isoname"`
	JobID                 string                                                   `json:"jobid"`
	Jobstatus             int                                                      `json:"jobstatus"`
	Keypair               string                                                   `json:"keypair"`
	Memory                int                                                      `json:"memory"`
	Memoryintfreekbs      int64                                                    `json:"memoryintfreekbs"`
	Memorykbs             int64                                                    `json:"memorykbs"`
	Memorytargetkbs       int64                                                    `json:"memorytargetkbs"`
	Name                  string                                                   `json:"name"`
	Networkkbsread        int64                                                    `json:"networkkbsread"`
	Networkkbswrite       int64                                                    `json:"networkkbswrite"`
	Nic                   []Nic                                                    `json:"nic"`
	Ostypeid              string                                                   `json:"ostypeid"`
	Password              string                                                   `json:"password"`
	Passwordenabled       bool                                                     `json:"passwordenabled"`
	Project               string                                                   `json:"project"`
	Projectid             string                                                   `json:"projectid"`
	Publicip              string                                                   `json:"publicip"`
	Publicipid            string                                                   `json:"publicipid"`
	Rootdeviceid          int64                                                    `json:"rootdeviceid"`
	Rootdevicetype        string                                                   `json:"rootdevicetype"`
	Securitygroup         []UpdateDefaultNicForVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                                   `json:"serviceofferingid"`
	Serviceofferingname   string                                                   `json:"serviceofferingname"`
	Servicestate          string                                                   `json:"servicestate"`
	State                 string                                                   `json:"state"`
	Tags                  []Tags                                                   `json:"tags"`
	Templatedisplaytext   string                                                   `json:"templatedisplaytext"`
	Templateid            string                                                   `json:"templateid"`
	Templatename          string                                                   `json:"templatename"`
	Userid                string                                                   `json:"userid"`
	Username              string                                                   `json:"username"`
	Vgpu                  string                                                   `json:"vgpu"`
	Zoneid                string                                                   `json:"zoneid"`
	Zonename              string                                                   `json:"zonename"`
}

type UpdateDefaultNicForVirtualMachineResponseSecuritygroup struct {
	Account             string                                                       `json:"account"`
	Description         string                                                       `json:"description"`
	Domain              string                                                       `json:"domain"`
	Domainid            string                                                       `json:"domainid"`
	Egressrule          []UpdateDefaultNicForVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                                       `json:"id"`
	Ingressrule         []UpdateDefaultNicForVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                                       `json:"name"`
	Project             string                                                       `json:"project"`
	Projectid           string                                                       `json:"projectid"`
	Tags                []Tags                                                       `json:"tags"`
	Virtualmachinecount int                                                          `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                                `json:"virtualmachineids"`
}

type UpdateDefaultNicForVirtualMachineResponseSecuritygroupRule struct {
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

type UpdateDefaultNicForVirtualMachineResponseAffinitygroup struct {
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

func (r *UpdateDefaultNicForVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias UpdateDefaultNicForVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

type UpdateVirtualMachineParams struct {
	p map[string]interface{}
}

func (p *UpdateVirtualMachineParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["cleanupdetails"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("cleanupdetails", vv)
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["details"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("details[%d].%s", i, k), vv)
			i++
		}
	}
	if v, found := p.p["dhcpoptionsnetworklist"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("dhcpoptionsnetworklist[%d].key", i), k)
			u.Set(fmt.Sprintf("dhcpoptionsnetworklist[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["displayname"]; found {
		u.Set("displayname", v.(string))
	}
	if v, found := p.p["displayvm"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvm", vv)
	}
	if v, found := p.p["extraconfig"]; found {
		u.Set("extraconfig", v.(string))
	}
	if v, found := p.p["group"]; found {
		u.Set("group", v.(string))
	}
	if v, found := p.p["haenable"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("haenable", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["instancename"]; found {
		u.Set("instancename", v.(string))
	}
	if v, found := p.p["isdynamicallyscalable"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("isdynamicallyscalable", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["ostypeid"]; found {
		u.Set("ostypeid", v.(string))
	}
	if v, found := p.p["securitygroupids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("securitygroupids", vv)
	}
	if v, found := p.p["securitygroupnames"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("securitygroupnames", vv)
	}
	if v, found := p.p["userdata"]; found {
		u.Set("userdata", v.(string))
	}
	return u
}

func (p *UpdateVirtualMachineParams) SetCleanupdetails(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["cleanupdetails"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetDetails(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["details"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetDhcpoptionsnetworklist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["dhcpoptionsnetworklist"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetDisplayname(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayname"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetDisplayvm(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvm"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetExtraconfig(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["extraconfig"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetGroup(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["group"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetHaenable(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["haenable"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetInstancename(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["instancename"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetIsdynamicallyscalable(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isdynamicallyscalable"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetOstypeid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ostypeid"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetSecuritygroupids(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["securitygroupids"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetSecuritygroupnames(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["securitygroupnames"] = v
	return
}

func (p *UpdateVirtualMachineParams) SetUserdata(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["userdata"] = v
	return
}

// You should always use this function to get a new UpdateVirtualMachineParams instance,
// as then you are sure you have configured all required params
func (s *VirtualMachineService) NewUpdateVirtualMachineParams(id string) *UpdateVirtualMachineParams {
	p := &UpdateVirtualMachineParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates properties of a virtual machine. The VM has to be stopped and restarted for the new properties to take effect. UpdateVirtualMachine does not first check whether the VM is stopped. Therefore, stop the VM manually before issuing this call.
func (s *VirtualMachineService) UpdateVirtualMachine(p *UpdateVirtualMachineParams) (*UpdateVirtualMachineResponse, error) {
	resp, err := s.cs.newRequest("updateVirtualMachine", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateVirtualMachineResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type UpdateVirtualMachineResponse struct {
	Account               string                                      `json:"account"`
	Affinitygroup         []UpdateVirtualMachineResponseAffinitygroup `json:"affinitygroup"`
	Cpunumber             int                                         `json:"cpunumber"`
	Cpuspeed              int                                         `json:"cpuspeed"`
	Cpuused               string                                      `json:"cpuused"`
	Created               string                                      `json:"created"`
	Details               map[string]string                           `json:"details"`
	Diskioread            int64                                       `json:"diskioread"`
	Diskiowrite           int64                                       `json:"diskiowrite"`
	Diskkbsread           int64                                       `json:"diskkbsread"`
	Diskkbswrite          int64                                       `json:"diskkbswrite"`
	Diskofferingid        string                                      `json:"diskofferingid"`
	Diskofferingname      string                                      `json:"diskofferingname"`
	Displayname           string                                      `json:"displayname"`
	Displayvm             bool                                        `json:"displayvm"`
	Domain                string                                      `json:"domain"`
	Domainid              string                                      `json:"domainid"`
	Forvirtualnetwork     bool                                        `json:"forvirtualnetwork"`
	Group                 string                                      `json:"group"`
	Groupid               string                                      `json:"groupid"`
	Guestosid             string                                      `json:"guestosid"`
	Haenable              bool                                        `json:"haenable"`
	Hostid                string                                      `json:"hostid"`
	Hostname              string                                      `json:"hostname"`
	Hypervisor            string                                      `json:"hypervisor"`
	Id                    string                                      `json:"id"`
	Instancename          string                                      `json:"instancename"`
	Isdynamicallyscalable bool                                        `json:"isdynamicallyscalable"`
	Isodisplaytext        string                                      `json:"isodisplaytext"`
	Isoid                 string                                      `json:"isoid"`
	Isoname               string                                      `json:"isoname"`
	JobID                 string                                      `json:"jobid"`
	Jobstatus             int                                         `json:"jobstatus"`
	Keypair               string                                      `json:"keypair"`
	Memory                int                                         `json:"memory"`
	Memoryintfreekbs      int64                                       `json:"memoryintfreekbs"`
	Memorykbs             int64                                       `json:"memorykbs"`
	Memorytargetkbs       int64                                       `json:"memorytargetkbs"`
	Name                  string                                      `json:"name"`
	Networkkbsread        int64                                       `json:"networkkbsread"`
	Networkkbswrite       int64                                       `json:"networkkbswrite"`
	Nic                   []Nic                                       `json:"nic"`
	Ostypeid              string                                      `json:"ostypeid"`
	Password              string                                      `json:"password"`
	Passwordenabled       bool                                        `json:"passwordenabled"`
	Project               string                                      `json:"project"`
	Projectid             string                                      `json:"projectid"`
	Publicip              string                                      `json:"publicip"`
	Publicipid            string                                      `json:"publicipid"`
	Rootdeviceid          int64                                       `json:"rootdeviceid"`
	Rootdevicetype        string                                      `json:"rootdevicetype"`
	Securitygroup         []UpdateVirtualMachineResponseSecuritygroup `json:"securitygroup"`
	Serviceofferingid     string                                      `json:"serviceofferingid"`
	Serviceofferingname   string                                      `json:"serviceofferingname"`
	Servicestate          string                                      `json:"servicestate"`
	State                 string                                      `json:"state"`
	Tags                  []Tags                                      `json:"tags"`
	Templatedisplaytext   string                                      `json:"templatedisplaytext"`
	Templateid            string                                      `json:"templateid"`
	Templatename          string                                      `json:"templatename"`
	Userid                string                                      `json:"userid"`
	Username              string                                      `json:"username"`
	Vgpu                  string                                      `json:"vgpu"`
	Zoneid                string                                      `json:"zoneid"`
	Zonename              string                                      `json:"zonename"`
}

type UpdateVirtualMachineResponseSecuritygroup struct {
	Account             string                                          `json:"account"`
	Description         string                                          `json:"description"`
	Domain              string                                          `json:"domain"`
	Domainid            string                                          `json:"domainid"`
	Egressrule          []UpdateVirtualMachineResponseSecuritygroupRule `json:"egressrule"`
	Id                  string                                          `json:"id"`
	Ingressrule         []UpdateVirtualMachineResponseSecuritygroupRule `json:"ingressrule"`
	Name                string                                          `json:"name"`
	Project             string                                          `json:"project"`
	Projectid           string                                          `json:"projectid"`
	Tags                []Tags                                          `json:"tags"`
	Virtualmachinecount int                                             `json:"virtualmachinecount"`
	Virtualmachineids   []interface{}                                   `json:"virtualmachineids"`
}

type UpdateVirtualMachineResponseSecuritygroupRule struct {
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

type UpdateVirtualMachineResponseAffinitygroup struct {
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

func (r *UpdateVirtualMachineResponse) UnmarshalJSON(b []byte) error {
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

	type alias UpdateVirtualMachineResponse
	return json.Unmarshal(b, (*alias)(r))
}

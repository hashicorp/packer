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

type AddNetworkServiceProviderParams struct {
	p map[string]interface{}
}

func (p *AddNetworkServiceProviderParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["destinationphysicalnetworkid"]; found {
		u.Set("destinationphysicalnetworkid", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["servicelist"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("servicelist", vv)
	}
	return u
}

func (p *AddNetworkServiceProviderParams) SetDestinationphysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["destinationphysicalnetworkid"] = v
	return
}

func (p *AddNetworkServiceProviderParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *AddNetworkServiceProviderParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *AddNetworkServiceProviderParams) SetServicelist(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["servicelist"] = v
	return
}

// You should always use this function to get a new AddNetworkServiceProviderParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewAddNetworkServiceProviderParams(name string, physicalnetworkid string) *AddNetworkServiceProviderParams {
	p := &AddNetworkServiceProviderParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	p.p["physicalnetworkid"] = physicalnetworkid
	return p
}

// Adds a network serviceProvider to a physical network
func (s *NetworkService) AddNetworkServiceProvider(p *AddNetworkServiceProviderParams) (*AddNetworkServiceProviderResponse, error) {
	resp, err := s.cs.newRequest("addNetworkServiceProvider", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddNetworkServiceProviderResponse
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

type AddNetworkServiceProviderResponse struct {
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

type AddOpenDaylightControllerParams struct {
	p map[string]interface{}
}

func (p *AddOpenDaylightControllerParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["password"]; found {
		u.Set("password", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["username"]; found {
		u.Set("username", v.(string))
	}
	return u
}

func (p *AddOpenDaylightControllerParams) SetPassword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["password"] = v
	return
}

func (p *AddOpenDaylightControllerParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *AddOpenDaylightControllerParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *AddOpenDaylightControllerParams) SetUsername(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["username"] = v
	return
}

// You should always use this function to get a new AddOpenDaylightControllerParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewAddOpenDaylightControllerParams(password string, physicalnetworkid string, url string, username string) *AddOpenDaylightControllerParams {
	p := &AddOpenDaylightControllerParams{}
	p.p = make(map[string]interface{})
	p.p["password"] = password
	p.p["physicalnetworkid"] = physicalnetworkid
	p.p["url"] = url
	p.p["username"] = username
	return p
}

// Adds an OpenDyalight controler
func (s *NetworkService) AddOpenDaylightController(p *AddOpenDaylightControllerParams) (*AddOpenDaylightControllerResponse, error) {
	resp, err := s.cs.newRequest("addOpenDaylightController", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AddOpenDaylightControllerResponse
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

type AddOpenDaylightControllerResponse struct {
	Id                string `json:"id"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
	Name              string `json:"name"`
	Physicalnetworkid string `json:"physicalnetworkid"`
	Url               string `json:"url"`
	Username          string `json:"username"`
}

type CreateNetworkParams struct {
	p map[string]interface{}
}

func (p *CreateNetworkParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["aclid"]; found {
		u.Set("aclid", v.(string))
	}
	if v, found := p.p["acltype"]; found {
		u.Set("acltype", v.(string))
	}
	if v, found := p.p["bypassvlanoverlapcheck"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("bypassvlanoverlapcheck", vv)
	}
	if v, found := p.p["displaynetwork"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displaynetwork", vv)
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["endip"]; found {
		u.Set("endip", v.(string))
	}
	if v, found := p.p["endipv6"]; found {
		u.Set("endipv6", v.(string))
	}
	if v, found := p.p["externalid"]; found {
		u.Set("externalid", v.(string))
	}
	if v, found := p.p["gateway"]; found {
		u.Set("gateway", v.(string))
	}
	if v, found := p.p["ip6cidr"]; found {
		u.Set("ip6cidr", v.(string))
	}
	if v, found := p.p["ip6gateway"]; found {
		u.Set("ip6gateway", v.(string))
	}
	if v, found := p.p["isolatedpvlan"]; found {
		u.Set("isolatedpvlan", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["netmask"]; found {
		u.Set("netmask", v.(string))
	}
	if v, found := p.p["networkdomain"]; found {
		u.Set("networkdomain", v.(string))
	}
	if v, found := p.p["networkofferingid"]; found {
		u.Set("networkofferingid", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["startip"]; found {
		u.Set("startip", v.(string))
	}
	if v, found := p.p["startipv6"]; found {
		u.Set("startipv6", v.(string))
	}
	if v, found := p.p["subdomainaccess"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("subdomainaccess", vv)
	}
	if v, found := p.p["vlan"]; found {
		u.Set("vlan", v.(string))
	}
	if v, found := p.p["vpcid"]; found {
		u.Set("vpcid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *CreateNetworkParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *CreateNetworkParams) SetAclid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["aclid"] = v
	return
}

func (p *CreateNetworkParams) SetAcltype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["acltype"] = v
	return
}

func (p *CreateNetworkParams) SetBypassvlanoverlapcheck(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["bypassvlanoverlapcheck"] = v
	return
}

func (p *CreateNetworkParams) SetDisplaynetwork(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaynetwork"] = v
	return
}

func (p *CreateNetworkParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *CreateNetworkParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreateNetworkParams) SetEndip(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["endip"] = v
	return
}

func (p *CreateNetworkParams) SetEndipv6(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["endipv6"] = v
	return
}

func (p *CreateNetworkParams) SetExternalid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["externalid"] = v
	return
}

func (p *CreateNetworkParams) SetGateway(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["gateway"] = v
	return
}

func (p *CreateNetworkParams) SetIp6cidr(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ip6cidr"] = v
	return
}

func (p *CreateNetworkParams) SetIp6gateway(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ip6gateway"] = v
	return
}

func (p *CreateNetworkParams) SetIsolatedpvlan(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isolatedpvlan"] = v
	return
}

func (p *CreateNetworkParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateNetworkParams) SetNetmask(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["netmask"] = v
	return
}

func (p *CreateNetworkParams) SetNetworkdomain(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkdomain"] = v
	return
}

func (p *CreateNetworkParams) SetNetworkofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkofferingid"] = v
	return
}

func (p *CreateNetworkParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *CreateNetworkParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *CreateNetworkParams) SetStartip(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["startip"] = v
	return
}

func (p *CreateNetworkParams) SetStartipv6(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["startipv6"] = v
	return
}

func (p *CreateNetworkParams) SetSubdomainaccess(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["subdomainaccess"] = v
	return
}

func (p *CreateNetworkParams) SetVlan(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

func (p *CreateNetworkParams) SetVpcid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcid"] = v
	return
}

func (p *CreateNetworkParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new CreateNetworkParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewCreateNetworkParams(displaytext string, name string, networkofferingid string, zoneid string) *CreateNetworkParams {
	p := &CreateNetworkParams{}
	p.p = make(map[string]interface{})
	p.p["displaytext"] = displaytext
	p.p["name"] = name
	p.p["networkofferingid"] = networkofferingid
	p.p["zoneid"] = zoneid
	return p
}

// Creates a network
func (s *NetworkService) CreateNetwork(p *CreateNetworkParams) (*CreateNetworkResponse, error) {
	resp, err := s.cs.newRequest("createNetwork", p.toURLValues())
	if err != nil {
		return nil, err
	}

	if resp, err = getRawValue(resp); err != nil {
		return nil, err
	}

	var r CreateNetworkResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type CreateNetworkResponse struct {
	Account                     string                         `json:"account"`
	Aclid                       string                         `json:"aclid"`
	Acltype                     string                         `json:"acltype"`
	Broadcastdomaintype         string                         `json:"broadcastdomaintype"`
	Broadcasturi                string                         `json:"broadcasturi"`
	Canusefordeploy             bool                           `json:"canusefordeploy"`
	Cidr                        string                         `json:"cidr"`
	Displaynetwork              bool                           `json:"displaynetwork"`
	Displaytext                 string                         `json:"displaytext"`
	Dns1                        string                         `json:"dns1"`
	Dns2                        string                         `json:"dns2"`
	Domain                      string                         `json:"domain"`
	Domainid                    string                         `json:"domainid"`
	Externalid                  string                         `json:"externalid"`
	Gateway                     string                         `json:"gateway"`
	Id                          string                         `json:"id"`
	Ip6cidr                     string                         `json:"ip6cidr"`
	Ip6gateway                  string                         `json:"ip6gateway"`
	Isdefault                   bool                           `json:"isdefault"`
	Ispersistent                bool                           `json:"ispersistent"`
	Issystem                    bool                           `json:"issystem"`
	JobID                       string                         `json:"jobid"`
	Jobstatus                   int                            `json:"jobstatus"`
	Name                        string                         `json:"name"`
	Netmask                     string                         `json:"netmask"`
	Networkcidr                 string                         `json:"networkcidr"`
	Networkdomain               string                         `json:"networkdomain"`
	Networkofferingavailability string                         `json:"networkofferingavailability"`
	Networkofferingconservemode bool                           `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                         `json:"networkofferingdisplaytext"`
	Networkofferingid           string                         `json:"networkofferingid"`
	Networkofferingname         string                         `json:"networkofferingname"`
	Physicalnetworkid           string                         `json:"physicalnetworkid"`
	Project                     string                         `json:"project"`
	Projectid                   string                         `json:"projectid"`
	Redundantrouter             bool                           `json:"redundantrouter"`
	Related                     string                         `json:"related"`
	Reservediprange             string                         `json:"reservediprange"`
	Restartrequired             bool                           `json:"restartrequired"`
	Service                     []CreateNetworkResponseService `json:"service"`
	Specifyipranges             bool                           `json:"specifyipranges"`
	State                       string                         `json:"state"`
	Strechedl2subnet            bool                           `json:"strechedl2subnet"`
	Subdomainaccess             bool                           `json:"subdomainaccess"`
	Tags                        []Tags                         `json:"tags"`
	Traffictype                 string                         `json:"traffictype"`
	Type                        string                         `json:"type"`
	Vlan                        string                         `json:"vlan"`
	Vpcid                       string                         `json:"vpcid"`
	Zoneid                      string                         `json:"zoneid"`
	Zonename                    string                         `json:"zonename"`
	Zonesnetworkspans           []interface{}                  `json:"zonesnetworkspans"`
}

type CreateNetworkResponseService struct {
	Capability []CreateNetworkResponseServiceCapability `json:"capability"`
	Name       string                                   `json:"name"`
	Provider   []CreateNetworkResponseServiceProvider   `json:"provider"`
}

type CreateNetworkResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type CreateNetworkResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type CreatePhysicalNetworkParams struct {
	p map[string]interface{}
}

func (p *CreatePhysicalNetworkParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["broadcastdomainrange"]; found {
		u.Set("broadcastdomainrange", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["isolationmethods"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("isolationmethods", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["networkspeed"]; found {
		u.Set("networkspeed", v.(string))
	}
	if v, found := p.p["tags"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("tags", vv)
	}
	if v, found := p.p["vlan"]; found {
		u.Set("vlan", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *CreatePhysicalNetworkParams) SetBroadcastdomainrange(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["broadcastdomainrange"] = v
	return
}

func (p *CreatePhysicalNetworkParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreatePhysicalNetworkParams) SetIsolationmethods(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isolationmethods"] = v
	return
}

func (p *CreatePhysicalNetworkParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreatePhysicalNetworkParams) SetNetworkspeed(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkspeed"] = v
	return
}

func (p *CreatePhysicalNetworkParams) SetTags(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *CreatePhysicalNetworkParams) SetVlan(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

func (p *CreatePhysicalNetworkParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new CreatePhysicalNetworkParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewCreatePhysicalNetworkParams(name string, zoneid string) *CreatePhysicalNetworkParams {
	p := &CreatePhysicalNetworkParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	p.p["zoneid"] = zoneid
	return p
}

// Creates a physical network
func (s *NetworkService) CreatePhysicalNetwork(p *CreatePhysicalNetworkParams) (*CreatePhysicalNetworkResponse, error) {
	resp, err := s.cs.newRequest("createPhysicalNetwork", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreatePhysicalNetworkResponse
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

type CreatePhysicalNetworkResponse struct {
	Broadcastdomainrange string `json:"broadcastdomainrange"`
	Domainid             string `json:"domainid"`
	Id                   string `json:"id"`
	Isolationmethods     string `json:"isolationmethods"`
	JobID                string `json:"jobid"`
	Jobstatus            int    `json:"jobstatus"`
	Name                 string `json:"name"`
	Networkspeed         string `json:"networkspeed"`
	State                string `json:"state"`
	Tags                 string `json:"tags"`
	Vlan                 string `json:"vlan"`
	Zoneid               string `json:"zoneid"`
}

type CreateServiceInstanceParams struct {
	p map[string]interface{}
}

func (p *CreateServiceInstanceParams) toURLValues() url.Values {
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
	if v, found := p.p["leftnetworkid"]; found {
		u.Set("leftnetworkid", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["rightnetworkid"]; found {
		u.Set("rightnetworkid", v.(string))
	}
	if v, found := p.p["serviceofferingid"]; found {
		u.Set("serviceofferingid", v.(string))
	}
	if v, found := p.p["templateid"]; found {
		u.Set("templateid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *CreateServiceInstanceParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *CreateServiceInstanceParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreateServiceInstanceParams) SetLeftnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["leftnetworkid"] = v
	return
}

func (p *CreateServiceInstanceParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateServiceInstanceParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *CreateServiceInstanceParams) SetRightnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["rightnetworkid"] = v
	return
}

func (p *CreateServiceInstanceParams) SetServiceofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceofferingid"] = v
	return
}

func (p *CreateServiceInstanceParams) SetTemplateid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["templateid"] = v
	return
}

func (p *CreateServiceInstanceParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new CreateServiceInstanceParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewCreateServiceInstanceParams(leftnetworkid string, name string, rightnetworkid string, serviceofferingid string, templateid string, zoneid string) *CreateServiceInstanceParams {
	p := &CreateServiceInstanceParams{}
	p.p = make(map[string]interface{})
	p.p["leftnetworkid"] = leftnetworkid
	p.p["name"] = name
	p.p["rightnetworkid"] = rightnetworkid
	p.p["serviceofferingid"] = serviceofferingid
	p.p["templateid"] = templateid
	p.p["zoneid"] = zoneid
	return p
}

// Creates a system virtual-machine that implements network services
func (s *NetworkService) CreateServiceInstance(p *CreateServiceInstanceParams) (*CreateServiceInstanceResponse, error) {
	resp, err := s.cs.newRequest("createServiceInstance", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateServiceInstanceResponse
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

type CreateServiceInstanceResponse struct {
	Account     string `json:"account"`
	Displayname string `json:"displayname"`
	Domain      string `json:"domain"`
	Domainid    string `json:"domainid"`
	Id          string `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
	Project     string `json:"project"`
	Projectid   string `json:"projectid"`
}

type CreateStorageNetworkIpRangeParams struct {
	p map[string]interface{}
}

func (p *CreateStorageNetworkIpRangeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["endip"]; found {
		u.Set("endip", v.(string))
	}
	if v, found := p.p["gateway"]; found {
		u.Set("gateway", v.(string))
	}
	if v, found := p.p["netmask"]; found {
		u.Set("netmask", v.(string))
	}
	if v, found := p.p["podid"]; found {
		u.Set("podid", v.(string))
	}
	if v, found := p.p["startip"]; found {
		u.Set("startip", v.(string))
	}
	if v, found := p.p["vlan"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("vlan", vv)
	}
	return u
}

func (p *CreateStorageNetworkIpRangeParams) SetEndip(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["endip"] = v
	return
}

func (p *CreateStorageNetworkIpRangeParams) SetGateway(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["gateway"] = v
	return
}

func (p *CreateStorageNetworkIpRangeParams) SetNetmask(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["netmask"] = v
	return
}

func (p *CreateStorageNetworkIpRangeParams) SetPodid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["podid"] = v
	return
}

func (p *CreateStorageNetworkIpRangeParams) SetStartip(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["startip"] = v
	return
}

func (p *CreateStorageNetworkIpRangeParams) SetVlan(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

// You should always use this function to get a new CreateStorageNetworkIpRangeParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewCreateStorageNetworkIpRangeParams(gateway string, netmask string, podid string, startip string) *CreateStorageNetworkIpRangeParams {
	p := &CreateStorageNetworkIpRangeParams{}
	p.p = make(map[string]interface{})
	p.p["gateway"] = gateway
	p.p["netmask"] = netmask
	p.p["podid"] = podid
	p.p["startip"] = startip
	return p
}

// Creates a Storage network IP range.
func (s *NetworkService) CreateStorageNetworkIpRange(p *CreateStorageNetworkIpRangeParams) (*CreateStorageNetworkIpRangeResponse, error) {
	resp, err := s.cs.newRequest("createStorageNetworkIpRange", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateStorageNetworkIpRangeResponse
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

type CreateStorageNetworkIpRangeResponse struct {
	Endip     string `json:"endip"`
	Gateway   string `json:"gateway"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Netmask   string `json:"netmask"`
	Networkid string `json:"networkid"`
	Podid     string `json:"podid"`
	Startip   string `json:"startip"`
	Vlan      int    `json:"vlan"`
	Zoneid    string `json:"zoneid"`
}

type DedicatePublicIpRangeParams struct {
	p map[string]interface{}
}

func (p *DedicatePublicIpRangeParams) toURLValues() url.Values {
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
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	return u
}

func (p *DedicatePublicIpRangeParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *DedicatePublicIpRangeParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *DedicatePublicIpRangeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *DedicatePublicIpRangeParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

// You should always use this function to get a new DedicatePublicIpRangeParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewDedicatePublicIpRangeParams(domainid string, id string) *DedicatePublicIpRangeParams {
	p := &DedicatePublicIpRangeParams{}
	p.p = make(map[string]interface{})
	p.p["domainid"] = domainid
	p.p["id"] = id
	return p
}

// Dedicates a Public IP range to an account
func (s *NetworkService) DedicatePublicIpRange(p *DedicatePublicIpRangeParams) (*DedicatePublicIpRangeResponse, error) {
	resp, err := s.cs.newRequest("dedicatePublicIpRange", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DedicatePublicIpRangeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DedicatePublicIpRangeResponse struct {
	Account           string `json:"account"`
	Description       string `json:"description"`
	Domain            string `json:"domain"`
	Domainid          string `json:"domainid"`
	Endip             string `json:"endip"`
	Endipv6           string `json:"endipv6"`
	Forsystemvms      bool   `json:"forsystemvms"`
	Forvirtualnetwork bool   `json:"forvirtualnetwork"`
	Gateway           string `json:"gateway"`
	Id                string `json:"id"`
	Ip6cidr           string `json:"ip6cidr"`
	Ip6gateway        string `json:"ip6gateway"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
	Netmask           string `json:"netmask"`
	Networkid         string `json:"networkid"`
	Physicalnetworkid string `json:"physicalnetworkid"`
	Podid             string `json:"podid"`
	Podname           string `json:"podname"`
	Project           string `json:"project"`
	Projectid         string `json:"projectid"`
	Startip           string `json:"startip"`
	Startipv6         string `json:"startipv6"`
	Vlan              string `json:"vlan"`
	Zoneid            string `json:"zoneid"`
}

type DeleteNetworkParams struct {
	p map[string]interface{}
}

func (p *DeleteNetworkParams) toURLValues() url.Values {
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

func (p *DeleteNetworkParams) SetForced(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forced"] = v
	return
}

func (p *DeleteNetworkParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteNetworkParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewDeleteNetworkParams(id string) *DeleteNetworkParams {
	p := &DeleteNetworkParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a network
func (s *NetworkService) DeleteNetwork(p *DeleteNetworkParams) (*DeleteNetworkResponse, error) {
	resp, err := s.cs.newRequest("deleteNetwork", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteNetworkResponse
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

type DeleteNetworkResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type DeleteNetworkServiceProviderParams struct {
	p map[string]interface{}
}

func (p *DeleteNetworkServiceProviderParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteNetworkServiceProviderParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteNetworkServiceProviderParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewDeleteNetworkServiceProviderParams(id string) *DeleteNetworkServiceProviderParams {
	p := &DeleteNetworkServiceProviderParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a Network Service Provider.
func (s *NetworkService) DeleteNetworkServiceProvider(p *DeleteNetworkServiceProviderParams) (*DeleteNetworkServiceProviderResponse, error) {
	resp, err := s.cs.newRequest("deleteNetworkServiceProvider", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteNetworkServiceProviderResponse
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

type DeleteNetworkServiceProviderResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type DeleteOpenDaylightControllerParams struct {
	p map[string]interface{}
}

func (p *DeleteOpenDaylightControllerParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteOpenDaylightControllerParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteOpenDaylightControllerParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewDeleteOpenDaylightControllerParams(id string) *DeleteOpenDaylightControllerParams {
	p := &DeleteOpenDaylightControllerParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Removes an OpenDyalight controler
func (s *NetworkService) DeleteOpenDaylightController(p *DeleteOpenDaylightControllerParams) (*DeleteOpenDaylightControllerResponse, error) {
	resp, err := s.cs.newRequest("deleteOpenDaylightController", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteOpenDaylightControllerResponse
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

type DeleteOpenDaylightControllerResponse struct {
	Id                string `json:"id"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
	Name              string `json:"name"`
	Physicalnetworkid string `json:"physicalnetworkid"`
	Url               string `json:"url"`
	Username          string `json:"username"`
}

type DeletePhysicalNetworkParams struct {
	p map[string]interface{}
}

func (p *DeletePhysicalNetworkParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeletePhysicalNetworkParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeletePhysicalNetworkParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewDeletePhysicalNetworkParams(id string) *DeletePhysicalNetworkParams {
	p := &DeletePhysicalNetworkParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a Physical Network.
func (s *NetworkService) DeletePhysicalNetwork(p *DeletePhysicalNetworkParams) (*DeletePhysicalNetworkResponse, error) {
	resp, err := s.cs.newRequest("deletePhysicalNetwork", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeletePhysicalNetworkResponse
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

type DeletePhysicalNetworkResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type DeleteStorageNetworkIpRangeParams struct {
	p map[string]interface{}
}

func (p *DeleteStorageNetworkIpRangeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteStorageNetworkIpRangeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteStorageNetworkIpRangeParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewDeleteStorageNetworkIpRangeParams(id string) *DeleteStorageNetworkIpRangeParams {
	p := &DeleteStorageNetworkIpRangeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a storage network IP Range.
func (s *NetworkService) DeleteStorageNetworkIpRange(p *DeleteStorageNetworkIpRangeParams) (*DeleteStorageNetworkIpRangeResponse, error) {
	resp, err := s.cs.newRequest("deleteStorageNetworkIpRange", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteStorageNetworkIpRangeResponse
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

type DeleteStorageNetworkIpRangeResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type ListNetscalerLoadBalancerNetworksParams struct {
	p map[string]interface{}
}

func (p *ListNetscalerLoadBalancerNetworksParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["lbdeviceid"]; found {
		u.Set("lbdeviceid", v.(string))
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

func (p *ListNetscalerLoadBalancerNetworksParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListNetscalerLoadBalancerNetworksParams) SetLbdeviceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["lbdeviceid"] = v
	return
}

func (p *ListNetscalerLoadBalancerNetworksParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListNetscalerLoadBalancerNetworksParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

// You should always use this function to get a new ListNetscalerLoadBalancerNetworksParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListNetscalerLoadBalancerNetworksParams(lbdeviceid string) *ListNetscalerLoadBalancerNetworksParams {
	p := &ListNetscalerLoadBalancerNetworksParams{}
	p.p = make(map[string]interface{})
	p.p["lbdeviceid"] = lbdeviceid
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetNetscalerLoadBalancerNetworkID(keyword string, lbdeviceid string, opts ...OptionFunc) (string, int, error) {
	p := &ListNetscalerLoadBalancerNetworksParams{}
	p.p = make(map[string]interface{})

	p.p["keyword"] = keyword
	p.p["lbdeviceid"] = lbdeviceid

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListNetscalerLoadBalancerNetworks(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", keyword, l)
	}

	if l.Count == 1 {
		return l.NetscalerLoadBalancerNetworks[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.NetscalerLoadBalancerNetworks {
			if v.Name == keyword {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", keyword, l)
}

// lists network that are using a netscaler load balancer device
func (s *NetworkService) ListNetscalerLoadBalancerNetworks(p *ListNetscalerLoadBalancerNetworksParams) (*ListNetscalerLoadBalancerNetworksResponse, error) {
	resp, err := s.cs.newRequest("listNetscalerLoadBalancerNetworks", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListNetscalerLoadBalancerNetworksResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListNetscalerLoadBalancerNetworksResponse struct {
	Count                         int                             `json:"count"`
	NetscalerLoadBalancerNetworks []*NetscalerLoadBalancerNetwork `json:"netscalerloadbalancernetwork"`
}

type NetscalerLoadBalancerNetwork struct {
	Account                     string                                `json:"account"`
	Aclid                       string                                `json:"aclid"`
	Acltype                     string                                `json:"acltype"`
	Broadcastdomaintype         string                                `json:"broadcastdomaintype"`
	Broadcasturi                string                                `json:"broadcasturi"`
	Canusefordeploy             bool                                  `json:"canusefordeploy"`
	Cidr                        string                                `json:"cidr"`
	Displaynetwork              bool                                  `json:"displaynetwork"`
	Displaytext                 string                                `json:"displaytext"`
	Dns1                        string                                `json:"dns1"`
	Dns2                        string                                `json:"dns2"`
	Domain                      string                                `json:"domain"`
	Domainid                    string                                `json:"domainid"`
	Externalid                  string                                `json:"externalid"`
	Gateway                     string                                `json:"gateway"`
	Id                          string                                `json:"id"`
	Ip6cidr                     string                                `json:"ip6cidr"`
	Ip6gateway                  string                                `json:"ip6gateway"`
	Isdefault                   bool                                  `json:"isdefault"`
	Ispersistent                bool                                  `json:"ispersistent"`
	Issystem                    bool                                  `json:"issystem"`
	JobID                       string                                `json:"jobid"`
	Jobstatus                   int                                   `json:"jobstatus"`
	Name                        string                                `json:"name"`
	Netmask                     string                                `json:"netmask"`
	Networkcidr                 string                                `json:"networkcidr"`
	Networkdomain               string                                `json:"networkdomain"`
	Networkofferingavailability string                                `json:"networkofferingavailability"`
	Networkofferingconservemode bool                                  `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                                `json:"networkofferingdisplaytext"`
	Networkofferingid           string                                `json:"networkofferingid"`
	Networkofferingname         string                                `json:"networkofferingname"`
	Physicalnetworkid           string                                `json:"physicalnetworkid"`
	Project                     string                                `json:"project"`
	Projectid                   string                                `json:"projectid"`
	Redundantrouter             bool                                  `json:"redundantrouter"`
	Related                     string                                `json:"related"`
	Reservediprange             string                                `json:"reservediprange"`
	Restartrequired             bool                                  `json:"restartrequired"`
	Service                     []NetscalerLoadBalancerNetworkService `json:"service"`
	Specifyipranges             bool                                  `json:"specifyipranges"`
	State                       string                                `json:"state"`
	Strechedl2subnet            bool                                  `json:"strechedl2subnet"`
	Subdomainaccess             bool                                  `json:"subdomainaccess"`
	Tags                        []Tags                                `json:"tags"`
	Traffictype                 string                                `json:"traffictype"`
	Type                        string                                `json:"type"`
	Vlan                        string                                `json:"vlan"`
	Vpcid                       string                                `json:"vpcid"`
	Zoneid                      string                                `json:"zoneid"`
	Zonename                    string                                `json:"zonename"`
	Zonesnetworkspans           []interface{}                         `json:"zonesnetworkspans"`
}

type NetscalerLoadBalancerNetworkService struct {
	Capability []NetscalerLoadBalancerNetworkServiceCapability `json:"capability"`
	Name       string                                          `json:"name"`
	Provider   []NetscalerLoadBalancerNetworkServiceProvider   `json:"provider"`
}

type NetscalerLoadBalancerNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type NetscalerLoadBalancerNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type ListNetworkIsolationMethodsParams struct {
	p map[string]interface{}
}

func (p *ListNetworkIsolationMethodsParams) toURLValues() url.Values {
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

func (p *ListNetworkIsolationMethodsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListNetworkIsolationMethodsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListNetworkIsolationMethodsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

// You should always use this function to get a new ListNetworkIsolationMethodsParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListNetworkIsolationMethodsParams() *ListNetworkIsolationMethodsParams {
	p := &ListNetworkIsolationMethodsParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists supported methods of network isolation
func (s *NetworkService) ListNetworkIsolationMethods(p *ListNetworkIsolationMethodsParams) (*ListNetworkIsolationMethodsResponse, error) {
	resp, err := s.cs.newRequest("listNetworkIsolationMethods", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListNetworkIsolationMethodsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListNetworkIsolationMethodsResponse struct {
	Count                   int                       `json:"count"`
	NetworkIsolationMethods []*NetworkIsolationMethod `json:"networkisolationmethod"`
}

type NetworkIsolationMethod struct {
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Name      string `json:"name"`
}

type ListNetworkServiceProvidersParams struct {
	p map[string]interface{}
}

func (p *ListNetworkServiceProvidersParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
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
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	return u
}

func (p *ListNetworkServiceProvidersParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListNetworkServiceProvidersParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListNetworkServiceProvidersParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListNetworkServiceProvidersParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListNetworkServiceProvidersParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *ListNetworkServiceProvidersParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

// You should always use this function to get a new ListNetworkServiceProvidersParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListNetworkServiceProvidersParams() *ListNetworkServiceProvidersParams {
	p := &ListNetworkServiceProvidersParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetNetworkServiceProviderID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListNetworkServiceProvidersParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListNetworkServiceProviders(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.NetworkServiceProviders[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.NetworkServiceProviders {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// Lists network serviceproviders for a given physical network.
func (s *NetworkService) ListNetworkServiceProviders(p *ListNetworkServiceProvidersParams) (*ListNetworkServiceProvidersResponse, error) {
	resp, err := s.cs.newRequest("listNetworkServiceProviders", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListNetworkServiceProvidersResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListNetworkServiceProvidersResponse struct {
	Count                   int                       `json:"count"`
	NetworkServiceProviders []*NetworkServiceProvider `json:"networkserviceprovider"`
}

type NetworkServiceProvider struct {
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

type ListNetworksParams struct {
	p map[string]interface{}
}

func (p *ListNetworksParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["acltype"]; found {
		u.Set("acltype", v.(string))
	}
	if v, found := p.p["canusefordeploy"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("canusefordeploy", vv)
	}
	if v, found := p.p["displaynetwork"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displaynetwork", vv)
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["forvpc"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("forvpc", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["isrecursive"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("isrecursive", vv)
	}
	if v, found := p.p["issystem"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("issystem", vv)
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["listall"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("listall", vv)
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
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["restartrequired"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("restartrequired", vv)
	}
	if v, found := p.p["specifyipranges"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("specifyipranges", vv)
	}
	if v, found := p.p["supportedservices"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("supportedservices", vv)
	}
	if v, found := p.p["tags"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("tags[%d].key", i), k)
			u.Set(fmt.Sprintf("tags[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["traffictype"]; found {
		u.Set("traffictype", v.(string))
	}
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	if v, found := p.p["vpcid"]; found {
		u.Set("vpcid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListNetworksParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListNetworksParams) SetAcltype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["acltype"] = v
	return
}

func (p *ListNetworksParams) SetCanusefordeploy(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["canusefordeploy"] = v
	return
}

func (p *ListNetworksParams) SetDisplaynetwork(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaynetwork"] = v
	return
}

func (p *ListNetworksParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListNetworksParams) SetForvpc(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forvpc"] = v
	return
}

func (p *ListNetworksParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListNetworksParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListNetworksParams) SetIssystem(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["issystem"] = v
	return
}

func (p *ListNetworksParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListNetworksParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListNetworksParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListNetworksParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListNetworksParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *ListNetworksParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListNetworksParams) SetRestartrequired(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["restartrequired"] = v
	return
}

func (p *ListNetworksParams) SetSpecifyipranges(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["specifyipranges"] = v
	return
}

func (p *ListNetworksParams) SetSupportedservices(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["supportedservices"] = v
	return
}

func (p *ListNetworksParams) SetTags(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *ListNetworksParams) SetTraffictype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["traffictype"] = v
	return
}

func (p *ListNetworksParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

func (p *ListNetworksParams) SetVpcid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcid"] = v
	return
}

func (p *ListNetworksParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListNetworksParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListNetworksParams() *ListNetworksParams {
	p := &ListNetworksParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetNetworkID(keyword string, opts ...OptionFunc) (string, int, error) {
	p := &ListNetworksParams{}
	p.p = make(map[string]interface{})

	p.p["keyword"] = keyword

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListNetworks(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", keyword, l)
	}

	if l.Count == 1 {
		return l.Networks[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.Networks {
			if v.Name == keyword {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", keyword, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetNetworkByName(name string, opts ...OptionFunc) (*Network, int, error) {
	id, count, err := s.GetNetworkID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetNetworkByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetNetworkByID(id string, opts ...OptionFunc) (*Network, int, error) {
	p := &ListNetworksParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListNetworks(p)
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
		return l.Networks[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for Network UUID: %s!", id)
}

// Lists all available networks.
func (s *NetworkService) ListNetworks(p *ListNetworksParams) (*ListNetworksResponse, error) {
	resp, err := s.cs.newRequest("listNetworks", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListNetworksResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListNetworksResponse struct {
	Count    int        `json:"count"`
	Networks []*Network `json:"network"`
}

type Network struct {
	Account                     string                   `json:"account"`
	Aclid                       string                   `json:"aclid"`
	Acltype                     string                   `json:"acltype"`
	Broadcastdomaintype         string                   `json:"broadcastdomaintype"`
	Broadcasturi                string                   `json:"broadcasturi"`
	Canusefordeploy             bool                     `json:"canusefordeploy"`
	Cidr                        string                   `json:"cidr"`
	Displaynetwork              bool                     `json:"displaynetwork"`
	Displaytext                 string                   `json:"displaytext"`
	Dns1                        string                   `json:"dns1"`
	Dns2                        string                   `json:"dns2"`
	Domain                      string                   `json:"domain"`
	Domainid                    string                   `json:"domainid"`
	Externalid                  string                   `json:"externalid"`
	Gateway                     string                   `json:"gateway"`
	Id                          string                   `json:"id"`
	Ip6cidr                     string                   `json:"ip6cidr"`
	Ip6gateway                  string                   `json:"ip6gateway"`
	Isdefault                   bool                     `json:"isdefault"`
	Ispersistent                bool                     `json:"ispersistent"`
	Issystem                    bool                     `json:"issystem"`
	JobID                       string                   `json:"jobid"`
	Jobstatus                   int                      `json:"jobstatus"`
	Name                        string                   `json:"name"`
	Netmask                     string                   `json:"netmask"`
	Networkcidr                 string                   `json:"networkcidr"`
	Networkdomain               string                   `json:"networkdomain"`
	Networkofferingavailability string                   `json:"networkofferingavailability"`
	Networkofferingconservemode bool                     `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                   `json:"networkofferingdisplaytext"`
	Networkofferingid           string                   `json:"networkofferingid"`
	Networkofferingname         string                   `json:"networkofferingname"`
	Physicalnetworkid           string                   `json:"physicalnetworkid"`
	Project                     string                   `json:"project"`
	Projectid                   string                   `json:"projectid"`
	Redundantrouter             bool                     `json:"redundantrouter"`
	Related                     string                   `json:"related"`
	Reservediprange             string                   `json:"reservediprange"`
	Restartrequired             bool                     `json:"restartrequired"`
	Service                     []NetworkServiceInternal `json:"service"`
	Specifyipranges             bool                     `json:"specifyipranges"`
	State                       string                   `json:"state"`
	Strechedl2subnet            bool                     `json:"strechedl2subnet"`
	Subdomainaccess             bool                     `json:"subdomainaccess"`
	Tags                        []Tags                   `json:"tags"`
	Traffictype                 string                   `json:"traffictype"`
	Type                        string                   `json:"type"`
	Vlan                        string                   `json:"vlan"`
	Vpcid                       string                   `json:"vpcid"`
	Zoneid                      string                   `json:"zoneid"`
	Zonename                    string                   `json:"zonename"`
	Zonesnetworkspans           []interface{}            `json:"zonesnetworkspans"`
}

type NetworkServiceInternal struct {
	Capability []NetworkServiceInternalCapability `json:"capability"`
	Name       string                             `json:"name"`
	Provider   []NetworkServiceInternalProvider   `json:"provider"`
}

type NetworkServiceInternalProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type NetworkServiceInternalCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type ListNiciraNvpDeviceNetworksParams struct {
	p map[string]interface{}
}

func (p *ListNiciraNvpDeviceNetworksParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["nvpdeviceid"]; found {
		u.Set("nvpdeviceid", v.(string))
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

func (p *ListNiciraNvpDeviceNetworksParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListNiciraNvpDeviceNetworksParams) SetNvpdeviceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["nvpdeviceid"] = v
	return
}

func (p *ListNiciraNvpDeviceNetworksParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListNiciraNvpDeviceNetworksParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

// You should always use this function to get a new ListNiciraNvpDeviceNetworksParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListNiciraNvpDeviceNetworksParams(nvpdeviceid string) *ListNiciraNvpDeviceNetworksParams {
	p := &ListNiciraNvpDeviceNetworksParams{}
	p.p = make(map[string]interface{})
	p.p["nvpdeviceid"] = nvpdeviceid
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetNiciraNvpDeviceNetworkID(keyword string, nvpdeviceid string, opts ...OptionFunc) (string, int, error) {
	p := &ListNiciraNvpDeviceNetworksParams{}
	p.p = make(map[string]interface{})

	p.p["keyword"] = keyword
	p.p["nvpdeviceid"] = nvpdeviceid

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListNiciraNvpDeviceNetworks(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", keyword, l)
	}

	if l.Count == 1 {
		return l.NiciraNvpDeviceNetworks[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.NiciraNvpDeviceNetworks {
			if v.Name == keyword {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", keyword, l)
}

// lists network that are using a nicira nvp device
func (s *NetworkService) ListNiciraNvpDeviceNetworks(p *ListNiciraNvpDeviceNetworksParams) (*ListNiciraNvpDeviceNetworksResponse, error) {
	resp, err := s.cs.newRequest("listNiciraNvpDeviceNetworks", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListNiciraNvpDeviceNetworksResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListNiciraNvpDeviceNetworksResponse struct {
	Count                   int                       `json:"count"`
	NiciraNvpDeviceNetworks []*NiciraNvpDeviceNetwork `json:"niciranvpdevicenetwork"`
}

type NiciraNvpDeviceNetwork struct {
	Account                     string                          `json:"account"`
	Aclid                       string                          `json:"aclid"`
	Acltype                     string                          `json:"acltype"`
	Broadcastdomaintype         string                          `json:"broadcastdomaintype"`
	Broadcasturi                string                          `json:"broadcasturi"`
	Canusefordeploy             bool                            `json:"canusefordeploy"`
	Cidr                        string                          `json:"cidr"`
	Displaynetwork              bool                            `json:"displaynetwork"`
	Displaytext                 string                          `json:"displaytext"`
	Dns1                        string                          `json:"dns1"`
	Dns2                        string                          `json:"dns2"`
	Domain                      string                          `json:"domain"`
	Domainid                    string                          `json:"domainid"`
	Externalid                  string                          `json:"externalid"`
	Gateway                     string                          `json:"gateway"`
	Id                          string                          `json:"id"`
	Ip6cidr                     string                          `json:"ip6cidr"`
	Ip6gateway                  string                          `json:"ip6gateway"`
	Isdefault                   bool                            `json:"isdefault"`
	Ispersistent                bool                            `json:"ispersistent"`
	Issystem                    bool                            `json:"issystem"`
	JobID                       string                          `json:"jobid"`
	Jobstatus                   int                             `json:"jobstatus"`
	Name                        string                          `json:"name"`
	Netmask                     string                          `json:"netmask"`
	Networkcidr                 string                          `json:"networkcidr"`
	Networkdomain               string                          `json:"networkdomain"`
	Networkofferingavailability string                          `json:"networkofferingavailability"`
	Networkofferingconservemode bool                            `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                          `json:"networkofferingdisplaytext"`
	Networkofferingid           string                          `json:"networkofferingid"`
	Networkofferingname         string                          `json:"networkofferingname"`
	Physicalnetworkid           string                          `json:"physicalnetworkid"`
	Project                     string                          `json:"project"`
	Projectid                   string                          `json:"projectid"`
	Redundantrouter             bool                            `json:"redundantrouter"`
	Related                     string                          `json:"related"`
	Reservediprange             string                          `json:"reservediprange"`
	Restartrequired             bool                            `json:"restartrequired"`
	Service                     []NiciraNvpDeviceNetworkService `json:"service"`
	Specifyipranges             bool                            `json:"specifyipranges"`
	State                       string                          `json:"state"`
	Strechedl2subnet            bool                            `json:"strechedl2subnet"`
	Subdomainaccess             bool                            `json:"subdomainaccess"`
	Tags                        []Tags                          `json:"tags"`
	Traffictype                 string                          `json:"traffictype"`
	Type                        string                          `json:"type"`
	Vlan                        string                          `json:"vlan"`
	Vpcid                       string                          `json:"vpcid"`
	Zoneid                      string                          `json:"zoneid"`
	Zonename                    string                          `json:"zonename"`
	Zonesnetworkspans           []interface{}                   `json:"zonesnetworkspans"`
}

type NiciraNvpDeviceNetworkService struct {
	Capability []NiciraNvpDeviceNetworkServiceCapability `json:"capability"`
	Name       string                                    `json:"name"`
	Provider   []NiciraNvpDeviceNetworkServiceProvider   `json:"provider"`
}

type NiciraNvpDeviceNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type NiciraNvpDeviceNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type ListOpenDaylightControllersParams struct {
	p map[string]interface{}
}

func (p *ListOpenDaylightControllersParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	return u
}

func (p *ListOpenDaylightControllersParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListOpenDaylightControllersParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

// You should always use this function to get a new ListOpenDaylightControllersParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListOpenDaylightControllersParams() *ListOpenDaylightControllersParams {
	p := &ListOpenDaylightControllersParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetOpenDaylightControllerByID(id string, opts ...OptionFunc) (*OpenDaylightController, int, error) {
	p := &ListOpenDaylightControllersParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListOpenDaylightControllers(p)
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
		return l.OpenDaylightControllers[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for OpenDaylightController UUID: %s!", id)
}

// Lists OpenDyalight controllers
func (s *NetworkService) ListOpenDaylightControllers(p *ListOpenDaylightControllersParams) (*ListOpenDaylightControllersResponse, error) {
	resp, err := s.cs.newRequest("listOpenDaylightControllers", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListOpenDaylightControllersResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListOpenDaylightControllersResponse struct {
	Count                   int                       `json:"count"`
	OpenDaylightControllers []*OpenDaylightController `json:"opendaylightcontroller"`
}

type OpenDaylightController struct {
	Id                string `json:"id"`
	JobID             string `json:"jobid"`
	Jobstatus         int    `json:"jobstatus"`
	Name              string `json:"name"`
	Physicalnetworkid string `json:"physicalnetworkid"`
	Url               string `json:"url"`
	Username          string `json:"username"`
}

type ListPaloAltoFirewallNetworksParams struct {
	p map[string]interface{}
}

func (p *ListPaloAltoFirewallNetworksParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
	}
	if v, found := p.p["lbdeviceid"]; found {
		u.Set("lbdeviceid", v.(string))
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

func (p *ListPaloAltoFirewallNetworksParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListPaloAltoFirewallNetworksParams) SetLbdeviceid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["lbdeviceid"] = v
	return
}

func (p *ListPaloAltoFirewallNetworksParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListPaloAltoFirewallNetworksParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

// You should always use this function to get a new ListPaloAltoFirewallNetworksParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListPaloAltoFirewallNetworksParams(lbdeviceid string) *ListPaloAltoFirewallNetworksParams {
	p := &ListPaloAltoFirewallNetworksParams{}
	p.p = make(map[string]interface{})
	p.p["lbdeviceid"] = lbdeviceid
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetPaloAltoFirewallNetworkID(keyword string, lbdeviceid string, opts ...OptionFunc) (string, int, error) {
	p := &ListPaloAltoFirewallNetworksParams{}
	p.p = make(map[string]interface{})

	p.p["keyword"] = keyword
	p.p["lbdeviceid"] = lbdeviceid

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListPaloAltoFirewallNetworks(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", keyword, l)
	}

	if l.Count == 1 {
		return l.PaloAltoFirewallNetworks[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.PaloAltoFirewallNetworks {
			if v.Name == keyword {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", keyword, l)
}

// lists network that are using Palo Alto firewall device
func (s *NetworkService) ListPaloAltoFirewallNetworks(p *ListPaloAltoFirewallNetworksParams) (*ListPaloAltoFirewallNetworksResponse, error) {
	resp, err := s.cs.newRequest("listPaloAltoFirewallNetworks", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListPaloAltoFirewallNetworksResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListPaloAltoFirewallNetworksResponse struct {
	Count                    int                        `json:"count"`
	PaloAltoFirewallNetworks []*PaloAltoFirewallNetwork `json:"paloaltofirewallnetwork"`
}

type PaloAltoFirewallNetwork struct {
	Account                     string                           `json:"account"`
	Aclid                       string                           `json:"aclid"`
	Acltype                     string                           `json:"acltype"`
	Broadcastdomaintype         string                           `json:"broadcastdomaintype"`
	Broadcasturi                string                           `json:"broadcasturi"`
	Canusefordeploy             bool                             `json:"canusefordeploy"`
	Cidr                        string                           `json:"cidr"`
	Displaynetwork              bool                             `json:"displaynetwork"`
	Displaytext                 string                           `json:"displaytext"`
	Dns1                        string                           `json:"dns1"`
	Dns2                        string                           `json:"dns2"`
	Domain                      string                           `json:"domain"`
	Domainid                    string                           `json:"domainid"`
	Externalid                  string                           `json:"externalid"`
	Gateway                     string                           `json:"gateway"`
	Id                          string                           `json:"id"`
	Ip6cidr                     string                           `json:"ip6cidr"`
	Ip6gateway                  string                           `json:"ip6gateway"`
	Isdefault                   bool                             `json:"isdefault"`
	Ispersistent                bool                             `json:"ispersistent"`
	Issystem                    bool                             `json:"issystem"`
	JobID                       string                           `json:"jobid"`
	Jobstatus                   int                              `json:"jobstatus"`
	Name                        string                           `json:"name"`
	Netmask                     string                           `json:"netmask"`
	Networkcidr                 string                           `json:"networkcidr"`
	Networkdomain               string                           `json:"networkdomain"`
	Networkofferingavailability string                           `json:"networkofferingavailability"`
	Networkofferingconservemode bool                             `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                           `json:"networkofferingdisplaytext"`
	Networkofferingid           string                           `json:"networkofferingid"`
	Networkofferingname         string                           `json:"networkofferingname"`
	Physicalnetworkid           string                           `json:"physicalnetworkid"`
	Project                     string                           `json:"project"`
	Projectid                   string                           `json:"projectid"`
	Redundantrouter             bool                             `json:"redundantrouter"`
	Related                     string                           `json:"related"`
	Reservediprange             string                           `json:"reservediprange"`
	Restartrequired             bool                             `json:"restartrequired"`
	Service                     []PaloAltoFirewallNetworkService `json:"service"`
	Specifyipranges             bool                             `json:"specifyipranges"`
	State                       string                           `json:"state"`
	Strechedl2subnet            bool                             `json:"strechedl2subnet"`
	Subdomainaccess             bool                             `json:"subdomainaccess"`
	Tags                        []Tags                           `json:"tags"`
	Traffictype                 string                           `json:"traffictype"`
	Type                        string                           `json:"type"`
	Vlan                        string                           `json:"vlan"`
	Vpcid                       string                           `json:"vpcid"`
	Zoneid                      string                           `json:"zoneid"`
	Zonename                    string                           `json:"zonename"`
	Zonesnetworkspans           []interface{}                    `json:"zonesnetworkspans"`
}

type PaloAltoFirewallNetworkService struct {
	Capability []PaloAltoFirewallNetworkServiceCapability `json:"capability"`
	Name       string                                     `json:"name"`
	Provider   []PaloAltoFirewallNetworkServiceProvider   `json:"provider"`
}

type PaloAltoFirewallNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type PaloAltoFirewallNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type ListPhysicalNetworksParams struct {
	p map[string]interface{}
}

func (p *ListPhysicalNetworksParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
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
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListPhysicalNetworksParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListPhysicalNetworksParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListPhysicalNetworksParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListPhysicalNetworksParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListPhysicalNetworksParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListPhysicalNetworksParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListPhysicalNetworksParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListPhysicalNetworksParams() *ListPhysicalNetworksParams {
	p := &ListPhysicalNetworksParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetPhysicalNetworkID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListPhysicalNetworksParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListPhysicalNetworks(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.PhysicalNetworks[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.PhysicalNetworks {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetPhysicalNetworkByName(name string, opts ...OptionFunc) (*PhysicalNetwork, int, error) {
	id, count, err := s.GetPhysicalNetworkID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetPhysicalNetworkByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetPhysicalNetworkByID(id string, opts ...OptionFunc) (*PhysicalNetwork, int, error) {
	p := &ListPhysicalNetworksParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListPhysicalNetworks(p)
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
		return l.PhysicalNetworks[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for PhysicalNetwork UUID: %s!", id)
}

// Lists physical networks
func (s *NetworkService) ListPhysicalNetworks(p *ListPhysicalNetworksParams) (*ListPhysicalNetworksResponse, error) {
	resp, err := s.cs.newRequest("listPhysicalNetworks", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListPhysicalNetworksResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListPhysicalNetworksResponse struct {
	Count            int                `json:"count"`
	PhysicalNetworks []*PhysicalNetwork `json:"physicalnetwork"`
}

type PhysicalNetwork struct {
	Broadcastdomainrange string `json:"broadcastdomainrange"`
	Domainid             string `json:"domainid"`
	Id                   string `json:"id"`
	Isolationmethods     string `json:"isolationmethods"`
	JobID                string `json:"jobid"`
	Jobstatus            int    `json:"jobstatus"`
	Name                 string `json:"name"`
	Networkspeed         string `json:"networkspeed"`
	State                string `json:"state"`
	Tags                 string `json:"tags"`
	Vlan                 string `json:"vlan"`
	Zoneid               string `json:"zoneid"`
}

type ListStorageNetworkIpRangeParams struct {
	p map[string]interface{}
}

func (p *ListStorageNetworkIpRangeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
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
	if v, found := p.p["podid"]; found {
		u.Set("podid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListStorageNetworkIpRangeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListStorageNetworkIpRangeParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListStorageNetworkIpRangeParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListStorageNetworkIpRangeParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListStorageNetworkIpRangeParams) SetPodid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["podid"] = v
	return
}

func (p *ListStorageNetworkIpRangeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListStorageNetworkIpRangeParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListStorageNetworkIpRangeParams() *ListStorageNetworkIpRangeParams {
	p := &ListStorageNetworkIpRangeParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkService) GetStorageNetworkIpRangeByID(id string, opts ...OptionFunc) (*StorageNetworkIpRange, int, error) {
	p := &ListStorageNetworkIpRangeParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListStorageNetworkIpRange(p)
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
		return l.StorageNetworkIpRange[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for StorageNetworkIpRange UUID: %s!", id)
}

// List a storage network IP range.
func (s *NetworkService) ListStorageNetworkIpRange(p *ListStorageNetworkIpRangeParams) (*ListStorageNetworkIpRangeResponse, error) {
	resp, err := s.cs.newRequest("listStorageNetworkIpRange", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListStorageNetworkIpRangeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListStorageNetworkIpRangeResponse struct {
	Count                 int                      `json:"count"`
	StorageNetworkIpRange []*StorageNetworkIpRange `json:"storagenetworkiprange"`
}

type StorageNetworkIpRange struct {
	Endip     string `json:"endip"`
	Gateway   string `json:"gateway"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Netmask   string `json:"netmask"`
	Networkid string `json:"networkid"`
	Podid     string `json:"podid"`
	Startip   string `json:"startip"`
	Vlan      int    `json:"vlan"`
	Zoneid    string `json:"zoneid"`
}

type ListSupportedNetworkServicesParams struct {
	p map[string]interface{}
}

func (p *ListSupportedNetworkServicesParams) toURLValues() url.Values {
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
	if v, found := p.p["provider"]; found {
		u.Set("provider", v.(string))
	}
	if v, found := p.p["service"]; found {
		u.Set("service", v.(string))
	}
	return u
}

func (p *ListSupportedNetworkServicesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListSupportedNetworkServicesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListSupportedNetworkServicesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListSupportedNetworkServicesParams) SetProvider(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["provider"] = v
	return
}

func (p *ListSupportedNetworkServicesParams) SetService(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["service"] = v
	return
}

// You should always use this function to get a new ListSupportedNetworkServicesParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewListSupportedNetworkServicesParams() *ListSupportedNetworkServicesParams {
	p := &ListSupportedNetworkServicesParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists all network services provided by CloudStack or for the given Provider.
func (s *NetworkService) ListSupportedNetworkServices(p *ListSupportedNetworkServicesParams) (*ListSupportedNetworkServicesResponse, error) {
	resp, err := s.cs.newRequest("listSupportedNetworkServices", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListSupportedNetworkServicesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListSupportedNetworkServicesResponse struct {
	Count                    int                        `json:"count"`
	SupportedNetworkServices []*SupportedNetworkService `json:"supportednetworkservice"`
}

type SupportedNetworkService struct {
	Capability []SupportedNetworkServiceCapability `json:"capability"`
	JobID      string                              `json:"jobid"`
	Jobstatus  int                                 `json:"jobstatus"`
	Name       string                              `json:"name"`
	Provider   []SupportedNetworkServiceProvider   `json:"provider"`
}

type SupportedNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type SupportedNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type ReleasePublicIpRangeParams struct {
	p map[string]interface{}
}

func (p *ReleasePublicIpRangeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *ReleasePublicIpRangeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new ReleasePublicIpRangeParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewReleasePublicIpRangeParams(id string) *ReleasePublicIpRangeParams {
	p := &ReleasePublicIpRangeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Releases a Public IP range back to the system pool
func (s *NetworkService) ReleasePublicIpRange(p *ReleasePublicIpRangeParams) (*ReleasePublicIpRangeResponse, error) {
	resp, err := s.cs.newRequest("releasePublicIpRange", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ReleasePublicIpRangeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ReleasePublicIpRangeResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *ReleasePublicIpRangeResponse) UnmarshalJSON(b []byte) error {
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

	type alias ReleasePublicIpRangeResponse
	return json.Unmarshal(b, (*alias)(r))
}

type RestartNetworkParams struct {
	p map[string]interface{}
}

func (p *RestartNetworkParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["cleanup"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("cleanup", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["makeredundant"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("makeredundant", vv)
	}
	return u
}

func (p *RestartNetworkParams) SetCleanup(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["cleanup"] = v
	return
}

func (p *RestartNetworkParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *RestartNetworkParams) SetMakeredundant(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["makeredundant"] = v
	return
}

// You should always use this function to get a new RestartNetworkParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewRestartNetworkParams(id string) *RestartNetworkParams {
	p := &RestartNetworkParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Restarts the network; includes 1) restarting network elements - virtual routers, DHCP servers 2) reapplying all public IPs 3) reapplying loadBalancing/portForwarding rules
func (s *NetworkService) RestartNetwork(p *RestartNetworkParams) (*RestartNetworkResponse, error) {
	resp, err := s.cs.newRequest("restartNetwork", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r RestartNetworkResponse
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

type RestartNetworkResponse struct {
	Account                   string `json:"account"`
	Allocated                 string `json:"allocated"`
	Associatednetworkid       string `json:"associatednetworkid"`
	Associatednetworkname     string `json:"associatednetworkname"`
	Domain                    string `json:"domain"`
	Domainid                  string `json:"domainid"`
	Fordisplay                bool   `json:"fordisplay"`
	Forvirtualnetwork         bool   `json:"forvirtualnetwork"`
	Id                        string `json:"id"`
	Ipaddress                 string `json:"ipaddress"`
	Isportable                bool   `json:"isportable"`
	Issourcenat               bool   `json:"issourcenat"`
	Isstaticnat               bool   `json:"isstaticnat"`
	Issystem                  bool   `json:"issystem"`
	JobID                     string `json:"jobid"`
	Jobstatus                 int    `json:"jobstatus"`
	Networkid                 string `json:"networkid"`
	Physicalnetworkid         string `json:"physicalnetworkid"`
	Project                   string `json:"project"`
	Projectid                 string `json:"projectid"`
	Purpose                   string `json:"purpose"`
	State                     string `json:"state"`
	Tags                      []Tags `json:"tags"`
	Virtualmachinedisplayname string `json:"virtualmachinedisplayname"`
	Virtualmachineid          string `json:"virtualmachineid"`
	Virtualmachinename        string `json:"virtualmachinename"`
	Vlanid                    string `json:"vlanid"`
	Vlanname                  string `json:"vlanname"`
	Vmipaddress               string `json:"vmipaddress"`
	Vpcid                     string `json:"vpcid"`
	Zoneid                    string `json:"zoneid"`
	Zonename                  string `json:"zonename"`
}

type UpdateNetworkParams struct {
	p map[string]interface{}
}

func (p *UpdateNetworkParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["changecidr"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("changecidr", vv)
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["displaynetwork"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displaynetwork", vv)
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["forced"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("forced", vv)
	}
	if v, found := p.p["guestvmcidr"]; found {
		u.Set("guestvmcidr", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["networkdomain"]; found {
		u.Set("networkdomain", v.(string))
	}
	if v, found := p.p["networkofferingid"]; found {
		u.Set("networkofferingid", v.(string))
	}
	if v, found := p.p["updateinsequence"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("updateinsequence", vv)
	}
	return u
}

func (p *UpdateNetworkParams) SetChangecidr(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["changecidr"] = v
	return
}

func (p *UpdateNetworkParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *UpdateNetworkParams) SetDisplaynetwork(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaynetwork"] = v
	return
}

func (p *UpdateNetworkParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *UpdateNetworkParams) SetForced(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forced"] = v
	return
}

func (p *UpdateNetworkParams) SetGuestvmcidr(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["guestvmcidr"] = v
	return
}

func (p *UpdateNetworkParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateNetworkParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UpdateNetworkParams) SetNetworkdomain(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkdomain"] = v
	return
}

func (p *UpdateNetworkParams) SetNetworkofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkofferingid"] = v
	return
}

func (p *UpdateNetworkParams) SetUpdateinsequence(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["updateinsequence"] = v
	return
}

// You should always use this function to get a new UpdateNetworkParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewUpdateNetworkParams(id string) *UpdateNetworkParams {
	p := &UpdateNetworkParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates a network
func (s *NetworkService) UpdateNetwork(p *UpdateNetworkParams) (*UpdateNetworkResponse, error) {
	resp, err := s.cs.newRequest("updateNetwork", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateNetworkResponse
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

type UpdateNetworkResponse struct {
	Account                     string                         `json:"account"`
	Aclid                       string                         `json:"aclid"`
	Acltype                     string                         `json:"acltype"`
	Broadcastdomaintype         string                         `json:"broadcastdomaintype"`
	Broadcasturi                string                         `json:"broadcasturi"`
	Canusefordeploy             bool                           `json:"canusefordeploy"`
	Cidr                        string                         `json:"cidr"`
	Displaynetwork              bool                           `json:"displaynetwork"`
	Displaytext                 string                         `json:"displaytext"`
	Dns1                        string                         `json:"dns1"`
	Dns2                        string                         `json:"dns2"`
	Domain                      string                         `json:"domain"`
	Domainid                    string                         `json:"domainid"`
	Externalid                  string                         `json:"externalid"`
	Gateway                     string                         `json:"gateway"`
	Id                          string                         `json:"id"`
	Ip6cidr                     string                         `json:"ip6cidr"`
	Ip6gateway                  string                         `json:"ip6gateway"`
	Isdefault                   bool                           `json:"isdefault"`
	Ispersistent                bool                           `json:"ispersistent"`
	Issystem                    bool                           `json:"issystem"`
	JobID                       string                         `json:"jobid"`
	Jobstatus                   int                            `json:"jobstatus"`
	Name                        string                         `json:"name"`
	Netmask                     string                         `json:"netmask"`
	Networkcidr                 string                         `json:"networkcidr"`
	Networkdomain               string                         `json:"networkdomain"`
	Networkofferingavailability string                         `json:"networkofferingavailability"`
	Networkofferingconservemode bool                           `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                         `json:"networkofferingdisplaytext"`
	Networkofferingid           string                         `json:"networkofferingid"`
	Networkofferingname         string                         `json:"networkofferingname"`
	Physicalnetworkid           string                         `json:"physicalnetworkid"`
	Project                     string                         `json:"project"`
	Projectid                   string                         `json:"projectid"`
	Redundantrouter             bool                           `json:"redundantrouter"`
	Related                     string                         `json:"related"`
	Reservediprange             string                         `json:"reservediprange"`
	Restartrequired             bool                           `json:"restartrequired"`
	Service                     []UpdateNetworkResponseService `json:"service"`
	Specifyipranges             bool                           `json:"specifyipranges"`
	State                       string                         `json:"state"`
	Strechedl2subnet            bool                           `json:"strechedl2subnet"`
	Subdomainaccess             bool                           `json:"subdomainaccess"`
	Tags                        []Tags                         `json:"tags"`
	Traffictype                 string                         `json:"traffictype"`
	Type                        string                         `json:"type"`
	Vlan                        string                         `json:"vlan"`
	Vpcid                       string                         `json:"vpcid"`
	Zoneid                      string                         `json:"zoneid"`
	Zonename                    string                         `json:"zonename"`
	Zonesnetworkspans           []interface{}                  `json:"zonesnetworkspans"`
}

type UpdateNetworkResponseService struct {
	Capability []UpdateNetworkResponseServiceCapability `json:"capability"`
	Name       string                                   `json:"name"`
	Provider   []UpdateNetworkResponseServiceProvider   `json:"provider"`
}

type UpdateNetworkResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type UpdateNetworkResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type UpdateNetworkServiceProviderParams struct {
	p map[string]interface{}
}

func (p *UpdateNetworkServiceProviderParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["servicelist"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("servicelist", vv)
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	return u
}

func (p *UpdateNetworkServiceProviderParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateNetworkServiceProviderParams) SetServicelist(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["servicelist"] = v
	return
}

func (p *UpdateNetworkServiceProviderParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

// You should always use this function to get a new UpdateNetworkServiceProviderParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewUpdateNetworkServiceProviderParams(id string) *UpdateNetworkServiceProviderParams {
	p := &UpdateNetworkServiceProviderParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates a network serviceProvider of a physical network
func (s *NetworkService) UpdateNetworkServiceProvider(p *UpdateNetworkServiceProviderParams) (*UpdateNetworkServiceProviderResponse, error) {
	resp, err := s.cs.newRequest("updateNetworkServiceProvider", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateNetworkServiceProviderResponse
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

type UpdateNetworkServiceProviderResponse struct {
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

type UpdatePhysicalNetworkParams struct {
	p map[string]interface{}
}

func (p *UpdatePhysicalNetworkParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["networkspeed"]; found {
		u.Set("networkspeed", v.(string))
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["tags"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("tags", vv)
	}
	if v, found := p.p["vlan"]; found {
		u.Set("vlan", v.(string))
	}
	return u
}

func (p *UpdatePhysicalNetworkParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdatePhysicalNetworkParams) SetNetworkspeed(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkspeed"] = v
	return
}

func (p *UpdatePhysicalNetworkParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *UpdatePhysicalNetworkParams) SetTags(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *UpdatePhysicalNetworkParams) SetVlan(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

// You should always use this function to get a new UpdatePhysicalNetworkParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewUpdatePhysicalNetworkParams(id string) *UpdatePhysicalNetworkParams {
	p := &UpdatePhysicalNetworkParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates a physical network
func (s *NetworkService) UpdatePhysicalNetwork(p *UpdatePhysicalNetworkParams) (*UpdatePhysicalNetworkResponse, error) {
	resp, err := s.cs.newRequest("updatePhysicalNetwork", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdatePhysicalNetworkResponse
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

type UpdatePhysicalNetworkResponse struct {
	Broadcastdomainrange string `json:"broadcastdomainrange"`
	Domainid             string `json:"domainid"`
	Id                   string `json:"id"`
	Isolationmethods     string `json:"isolationmethods"`
	JobID                string `json:"jobid"`
	Jobstatus            int    `json:"jobstatus"`
	Name                 string `json:"name"`
	Networkspeed         string `json:"networkspeed"`
	State                string `json:"state"`
	Tags                 string `json:"tags"`
	Vlan                 string `json:"vlan"`
	Zoneid               string `json:"zoneid"`
}

type UpdateStorageNetworkIpRangeParams struct {
	p map[string]interface{}
}

func (p *UpdateStorageNetworkIpRangeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["endip"]; found {
		u.Set("endip", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["netmask"]; found {
		u.Set("netmask", v.(string))
	}
	if v, found := p.p["startip"]; found {
		u.Set("startip", v.(string))
	}
	if v, found := p.p["vlan"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("vlan", vv)
	}
	return u
}

func (p *UpdateStorageNetworkIpRangeParams) SetEndip(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["endip"] = v
	return
}

func (p *UpdateStorageNetworkIpRangeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateStorageNetworkIpRangeParams) SetNetmask(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["netmask"] = v
	return
}

func (p *UpdateStorageNetworkIpRangeParams) SetStartip(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["startip"] = v
	return
}

func (p *UpdateStorageNetworkIpRangeParams) SetVlan(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

// You should always use this function to get a new UpdateStorageNetworkIpRangeParams instance,
// as then you are sure you have configured all required params
func (s *NetworkService) NewUpdateStorageNetworkIpRangeParams(id string) *UpdateStorageNetworkIpRangeParams {
	p := &UpdateStorageNetworkIpRangeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Update a Storage network IP range, only allowed when no IPs in this range have been allocated.
func (s *NetworkService) UpdateStorageNetworkIpRange(p *UpdateStorageNetworkIpRangeParams) (*UpdateStorageNetworkIpRangeResponse, error) {
	resp, err := s.cs.newRequest("updateStorageNetworkIpRange", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateStorageNetworkIpRangeResponse
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

type UpdateStorageNetworkIpRangeResponse struct {
	Endip     string `json:"endip"`
	Gateway   string `json:"gateway"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Netmask   string `json:"netmask"`
	Networkid string `json:"networkid"`
	Podid     string `json:"podid"`
	Startip   string `json:"startip"`
	Vlan      int    `json:"vlan"`
	Zoneid    string `json:"zoneid"`
}

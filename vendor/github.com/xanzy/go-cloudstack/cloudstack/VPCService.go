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

type CreatePrivateGatewayParams struct {
	p map[string]interface{}
}

func (p *CreatePrivateGatewayParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["aclid"]; found {
		u.Set("aclid", v.(string))
	}
	if v, found := p.p["gateway"]; found {
		u.Set("gateway", v.(string))
	}
	if v, found := p.p["ipaddress"]; found {
		u.Set("ipaddress", v.(string))
	}
	if v, found := p.p["netmask"]; found {
		u.Set("netmask", v.(string))
	}
	if v, found := p.p["networkofferingid"]; found {
		u.Set("networkofferingid", v.(string))
	}
	if v, found := p.p["physicalnetworkid"]; found {
		u.Set("physicalnetworkid", v.(string))
	}
	if v, found := p.p["sourcenatsupported"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("sourcenatsupported", vv)
	}
	if v, found := p.p["vlan"]; found {
		u.Set("vlan", v.(string))
	}
	if v, found := p.p["vpcid"]; found {
		u.Set("vpcid", v.(string))
	}
	return u
}

func (p *CreatePrivateGatewayParams) SetAclid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["aclid"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetGateway(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["gateway"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetIpaddress(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ipaddress"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetNetmask(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["netmask"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetNetworkofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkofferingid"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetPhysicalnetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["physicalnetworkid"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetSourcenatsupported(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["sourcenatsupported"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetVlan(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

func (p *CreatePrivateGatewayParams) SetVpcid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcid"] = v
	return
}

// You should always use this function to get a new CreatePrivateGatewayParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewCreatePrivateGatewayParams(gateway string, ipaddress string, netmask string, vlan string, vpcid string) *CreatePrivateGatewayParams {
	p := &CreatePrivateGatewayParams{}
	p.p = make(map[string]interface{})
	p.p["gateway"] = gateway
	p.p["ipaddress"] = ipaddress
	p.p["netmask"] = netmask
	p.p["vlan"] = vlan
	p.p["vpcid"] = vpcid
	return p
}

// Creates a private gateway
func (s *VPCService) CreatePrivateGateway(p *CreatePrivateGatewayParams) (*CreatePrivateGatewayResponse, error) {
	resp, err := s.cs.newRequest("createPrivateGateway", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreatePrivateGatewayResponse
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

type CreatePrivateGatewayResponse struct {
	Account            string `json:"account"`
	Aclid              string `json:"aclid"`
	Domain             string `json:"domain"`
	Domainid           string `json:"domainid"`
	Gateway            string `json:"gateway"`
	Id                 string `json:"id"`
	Ipaddress          string `json:"ipaddress"`
	JobID              string `json:"jobid"`
	Jobstatus          int    `json:"jobstatus"`
	Netmask            string `json:"netmask"`
	Physicalnetworkid  string `json:"physicalnetworkid"`
	Project            string `json:"project"`
	Projectid          string `json:"projectid"`
	Sourcenatsupported bool   `json:"sourcenatsupported"`
	State              string `json:"state"`
	Vlan               string `json:"vlan"`
	Vpcid              string `json:"vpcid"`
	Zoneid             string `json:"zoneid"`
	Zonename           string `json:"zonename"`
}

type CreateStaticRouteParams struct {
	p map[string]interface{}
}

func (p *CreateStaticRouteParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["cidr"]; found {
		u.Set("cidr", v.(string))
	}
	if v, found := p.p["gatewayid"]; found {
		u.Set("gatewayid", v.(string))
	}
	return u
}

func (p *CreateStaticRouteParams) SetCidr(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["cidr"] = v
	return
}

func (p *CreateStaticRouteParams) SetGatewayid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["gatewayid"] = v
	return
}

// You should always use this function to get a new CreateStaticRouteParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewCreateStaticRouteParams(cidr string, gatewayid string) *CreateStaticRouteParams {
	p := &CreateStaticRouteParams{}
	p.p = make(map[string]interface{})
	p.p["cidr"] = cidr
	p.p["gatewayid"] = gatewayid
	return p
}

// Creates a static route
func (s *VPCService) CreateStaticRoute(p *CreateStaticRouteParams) (*CreateStaticRouteResponse, error) {
	resp, err := s.cs.newRequest("createStaticRoute", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateStaticRouteResponse
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

type CreateStaticRouteResponse struct {
	Account   string `json:"account"`
	Cidr      string `json:"cidr"`
	Domain    string `json:"domain"`
	Domainid  string `json:"domainid"`
	Gatewayid string `json:"gatewayid"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Project   string `json:"project"`
	Projectid string `json:"projectid"`
	State     string `json:"state"`
	Tags      []Tags `json:"tags"`
	Vpcid     string `json:"vpcid"`
}

type CreateVPCParams struct {
	p map[string]interface{}
}

func (p *CreateVPCParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["cidr"]; found {
		u.Set("cidr", v.(string))
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["fordisplay"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("fordisplay", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["networkdomain"]; found {
		u.Set("networkdomain", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["start"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("start", vv)
	}
	if v, found := p.p["vpcofferingid"]; found {
		u.Set("vpcofferingid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *CreateVPCParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *CreateVPCParams) SetCidr(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["cidr"] = v
	return
}

func (p *CreateVPCParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *CreateVPCParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreateVPCParams) SetFordisplay(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["fordisplay"] = v
	return
}

func (p *CreateVPCParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateVPCParams) SetNetworkdomain(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkdomain"] = v
	return
}

func (p *CreateVPCParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *CreateVPCParams) SetStart(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["start"] = v
	return
}

func (p *CreateVPCParams) SetVpcofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcofferingid"] = v
	return
}

func (p *CreateVPCParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new CreateVPCParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewCreateVPCParams(cidr string, displaytext string, name string, vpcofferingid string, zoneid string) *CreateVPCParams {
	p := &CreateVPCParams{}
	p.p = make(map[string]interface{})
	p.p["cidr"] = cidr
	p.p["displaytext"] = displaytext
	p.p["name"] = name
	p.p["vpcofferingid"] = vpcofferingid
	p.p["zoneid"] = zoneid
	return p
}

// Creates a VPC
func (s *VPCService) CreateVPC(p *CreateVPCParams) (*CreateVPCResponse, error) {
	resp, err := s.cs.newRequest("createVPC", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateVPCResponse
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

type CreateVPCResponse struct {
	Account              string                     `json:"account"`
	Cidr                 string                     `json:"cidr"`
	Created              string                     `json:"created"`
	Displaytext          string                     `json:"displaytext"`
	Distributedvpcrouter bool                       `json:"distributedvpcrouter"`
	Domain               string                     `json:"domain"`
	Domainid             string                     `json:"domainid"`
	Fordisplay           bool                       `json:"fordisplay"`
	Id                   string                     `json:"id"`
	JobID                string                     `json:"jobid"`
	Jobstatus            int                        `json:"jobstatus"`
	Name                 string                     `json:"name"`
	Network              []CreateVPCResponseNetwork `json:"network"`
	Networkdomain        string                     `json:"networkdomain"`
	Project              string                     `json:"project"`
	Projectid            string                     `json:"projectid"`
	Redundantvpcrouter   bool                       `json:"redundantvpcrouter"`
	Regionlevelvpc       bool                       `json:"regionlevelvpc"`
	Restartrequired      bool                       `json:"restartrequired"`
	Service              []CreateVPCResponseService `json:"service"`
	State                string                     `json:"state"`
	Tags                 []Tags                     `json:"tags"`
	Vpcofferingid        string                     `json:"vpcofferingid"`
	Zoneid               string                     `json:"zoneid"`
	Zonename             string                     `json:"zonename"`
}

type CreateVPCResponseService struct {
	Capability []CreateVPCResponseServiceCapability `json:"capability"`
	Name       string                               `json:"name"`
	Provider   []CreateVPCResponseServiceProvider   `json:"provider"`
}

type CreateVPCResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type CreateVPCResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type CreateVPCResponseNetwork struct {
	Account                     string                            `json:"account"`
	Aclid                       string                            `json:"aclid"`
	Acltype                     string                            `json:"acltype"`
	Broadcastdomaintype         string                            `json:"broadcastdomaintype"`
	Broadcasturi                string                            `json:"broadcasturi"`
	Canusefordeploy             bool                              `json:"canusefordeploy"`
	Cidr                        string                            `json:"cidr"`
	Displaynetwork              bool                              `json:"displaynetwork"`
	Displaytext                 string                            `json:"displaytext"`
	Dns1                        string                            `json:"dns1"`
	Dns2                        string                            `json:"dns2"`
	Domain                      string                            `json:"domain"`
	Domainid                    string                            `json:"domainid"`
	Externalid                  string                            `json:"externalid"`
	Gateway                     string                            `json:"gateway"`
	Id                          string                            `json:"id"`
	Ip6cidr                     string                            `json:"ip6cidr"`
	Ip6gateway                  string                            `json:"ip6gateway"`
	Isdefault                   bool                              `json:"isdefault"`
	Ispersistent                bool                              `json:"ispersistent"`
	Issystem                    bool                              `json:"issystem"`
	Name                        string                            `json:"name"`
	Netmask                     string                            `json:"netmask"`
	Networkcidr                 string                            `json:"networkcidr"`
	Networkdomain               string                            `json:"networkdomain"`
	Networkofferingavailability string                            `json:"networkofferingavailability"`
	Networkofferingconservemode bool                              `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                            `json:"networkofferingdisplaytext"`
	Networkofferingid           string                            `json:"networkofferingid"`
	Networkofferingname         string                            `json:"networkofferingname"`
	Physicalnetworkid           string                            `json:"physicalnetworkid"`
	Project                     string                            `json:"project"`
	Projectid                   string                            `json:"projectid"`
	Redundantrouter             bool                              `json:"redundantrouter"`
	Related                     string                            `json:"related"`
	Reservediprange             string                            `json:"reservediprange"`
	Restartrequired             bool                              `json:"restartrequired"`
	Service                     []CreateVPCResponseNetworkService `json:"service"`
	Specifyipranges             bool                              `json:"specifyipranges"`
	State                       string                            `json:"state"`
	Strechedl2subnet            bool                              `json:"strechedl2subnet"`
	Subdomainaccess             bool                              `json:"subdomainaccess"`
	Tags                        []Tags                            `json:"tags"`
	Traffictype                 string                            `json:"traffictype"`
	Type                        string                            `json:"type"`
	Vlan                        string                            `json:"vlan"`
	Vpcid                       string                            `json:"vpcid"`
	Zoneid                      string                            `json:"zoneid"`
	Zonename                    string                            `json:"zonename"`
	Zonesnetworkspans           []interface{}                     `json:"zonesnetworkspans"`
}

type CreateVPCResponseNetworkService struct {
	Capability []CreateVPCResponseNetworkServiceCapability `json:"capability"`
	Name       string                                      `json:"name"`
	Provider   []CreateVPCResponseNetworkServiceProvider   `json:"provider"`
}

type CreateVPCResponseNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type CreateVPCResponseNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type CreateVPCOfferingParams struct {
	p map[string]interface{}
}

func (p *CreateVPCOfferingParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["servicecapabilitylist"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("servicecapabilitylist[%d].key", i), k)
			u.Set(fmt.Sprintf("servicecapabilitylist[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["serviceofferingid"]; found {
		u.Set("serviceofferingid", v.(string))
	}
	if v, found := p.p["serviceproviderlist"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("serviceproviderlist[%d].service", i), k)
			u.Set(fmt.Sprintf("serviceproviderlist[%d].provider", i), vv)
			i++
		}
	}
	if v, found := p.p["supportedservices"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("supportedservices", vv)
	}
	return u
}

func (p *CreateVPCOfferingParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *CreateVPCOfferingParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateVPCOfferingParams) SetServicecapabilitylist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["servicecapabilitylist"] = v
	return
}

func (p *CreateVPCOfferingParams) SetServiceofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceofferingid"] = v
	return
}

func (p *CreateVPCOfferingParams) SetServiceproviderlist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceproviderlist"] = v
	return
}

func (p *CreateVPCOfferingParams) SetSupportedservices(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["supportedservices"] = v
	return
}

// You should always use this function to get a new CreateVPCOfferingParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewCreateVPCOfferingParams(displaytext string, name string, supportedservices []string) *CreateVPCOfferingParams {
	p := &CreateVPCOfferingParams{}
	p.p = make(map[string]interface{})
	p.p["displaytext"] = displaytext
	p.p["name"] = name
	p.p["supportedservices"] = supportedservices
	return p
}

// Creates VPC offering
func (s *VPCService) CreateVPCOffering(p *CreateVPCOfferingParams) (*CreateVPCOfferingResponse, error) {
	resp, err := s.cs.newRequest("createVPCOffering", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateVPCOfferingResponse
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

type CreateVPCOfferingResponse struct {
	Created                string                             `json:"created"`
	Displaytext            string                             `json:"displaytext"`
	Distributedvpcrouter   bool                               `json:"distributedvpcrouter"`
	Id                     string                             `json:"id"`
	Isdefault              bool                               `json:"isdefault"`
	JobID                  string                             `json:"jobid"`
	Jobstatus              int                                `json:"jobstatus"`
	Name                   string                             `json:"name"`
	Service                []CreateVPCOfferingResponseService `json:"service"`
	State                  string                             `json:"state"`
	SupportsregionLevelvpc bool                               `json:"supportsregionLevelvpc"`
}

type CreateVPCOfferingResponseService struct {
	Capability []CreateVPCOfferingResponseServiceCapability `json:"capability"`
	Name       string                                       `json:"name"`
	Provider   []CreateVPCOfferingResponseServiceProvider   `json:"provider"`
}

type CreateVPCOfferingResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type CreateVPCOfferingResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type DeletePrivateGatewayParams struct {
	p map[string]interface{}
}

func (p *DeletePrivateGatewayParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeletePrivateGatewayParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeletePrivateGatewayParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewDeletePrivateGatewayParams(id string) *DeletePrivateGatewayParams {
	p := &DeletePrivateGatewayParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a Private gateway
func (s *VPCService) DeletePrivateGateway(p *DeletePrivateGatewayParams) (*DeletePrivateGatewayResponse, error) {
	resp, err := s.cs.newRequest("deletePrivateGateway", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeletePrivateGatewayResponse
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

type DeletePrivateGatewayResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type DeleteStaticRouteParams struct {
	p map[string]interface{}
}

func (p *DeleteStaticRouteParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteStaticRouteParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteStaticRouteParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewDeleteStaticRouteParams(id string) *DeleteStaticRouteParams {
	p := &DeleteStaticRouteParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a static route
func (s *VPCService) DeleteStaticRoute(p *DeleteStaticRouteParams) (*DeleteStaticRouteResponse, error) {
	resp, err := s.cs.newRequest("deleteStaticRoute", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteStaticRouteResponse
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

type DeleteStaticRouteResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type DeleteVPCParams struct {
	p map[string]interface{}
}

func (p *DeleteVPCParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteVPCParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteVPCParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewDeleteVPCParams(id string) *DeleteVPCParams {
	p := &DeleteVPCParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a VPC
func (s *VPCService) DeleteVPC(p *DeleteVPCParams) (*DeleteVPCResponse, error) {
	resp, err := s.cs.newRequest("deleteVPC", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteVPCResponse
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

type DeleteVPCResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type DeleteVPCOfferingParams struct {
	p map[string]interface{}
}

func (p *DeleteVPCOfferingParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteVPCOfferingParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteVPCOfferingParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewDeleteVPCOfferingParams(id string) *DeleteVPCOfferingParams {
	p := &DeleteVPCOfferingParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes VPC offering
func (s *VPCService) DeleteVPCOffering(p *DeleteVPCOfferingParams) (*DeleteVPCOfferingResponse, error) {
	resp, err := s.cs.newRequest("deleteVPCOffering", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteVPCOfferingResponse
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

type DeleteVPCOfferingResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

type ListPrivateGatewaysParams struct {
	p map[string]interface{}
}

func (p *ListPrivateGatewaysParams) toURLValues() url.Values {
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
	if v, found := p.p["ipaddress"]; found {
		u.Set("ipaddress", v.(string))
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
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["vlan"]; found {
		u.Set("vlan", v.(string))
	}
	if v, found := p.p["vpcid"]; found {
		u.Set("vpcid", v.(string))
	}
	return u
}

func (p *ListPrivateGatewaysParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetIpaddress(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ipaddress"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetVlan(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vlan"] = v
	return
}

func (p *ListPrivateGatewaysParams) SetVpcid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcid"] = v
	return
}

// You should always use this function to get a new ListPrivateGatewaysParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewListPrivateGatewaysParams() *ListPrivateGatewaysParams {
	p := &ListPrivateGatewaysParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetPrivateGatewayByID(id string, opts ...OptionFunc) (*PrivateGateway, int, error) {
	p := &ListPrivateGatewaysParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListPrivateGateways(p)
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
		return l.PrivateGateways[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for PrivateGateway UUID: %s!", id)
}

// List private gateways
func (s *VPCService) ListPrivateGateways(p *ListPrivateGatewaysParams) (*ListPrivateGatewaysResponse, error) {
	resp, err := s.cs.newRequest("listPrivateGateways", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListPrivateGatewaysResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListPrivateGatewaysResponse struct {
	Count           int               `json:"count"`
	PrivateGateways []*PrivateGateway `json:"privategateway"`
}

type PrivateGateway struct {
	Account            string `json:"account"`
	Aclid              string `json:"aclid"`
	Domain             string `json:"domain"`
	Domainid           string `json:"domainid"`
	Gateway            string `json:"gateway"`
	Id                 string `json:"id"`
	Ipaddress          string `json:"ipaddress"`
	JobID              string `json:"jobid"`
	Jobstatus          int    `json:"jobstatus"`
	Netmask            string `json:"netmask"`
	Physicalnetworkid  string `json:"physicalnetworkid"`
	Project            string `json:"project"`
	Projectid          string `json:"projectid"`
	Sourcenatsupported bool   `json:"sourcenatsupported"`
	State              string `json:"state"`
	Vlan               string `json:"vlan"`
	Vpcid              string `json:"vpcid"`
	Zoneid             string `json:"zoneid"`
	Zonename           string `json:"zonename"`
}

type ListStaticRoutesParams struct {
	p map[string]interface{}
}

func (p *ListStaticRoutesParams) toURLValues() url.Values {
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
	if v, found := p.p["gatewayid"]; found {
		u.Set("gatewayid", v.(string))
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
	if v, found := p.p["tags"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("tags[%d].key", i), k)
			u.Set(fmt.Sprintf("tags[%d].value", i), vv)
			i++
		}
	}
	if v, found := p.p["vpcid"]; found {
		u.Set("vpcid", v.(string))
	}
	return u
}

func (p *ListStaticRoutesParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListStaticRoutesParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListStaticRoutesParams) SetGatewayid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["gatewayid"] = v
	return
}

func (p *ListStaticRoutesParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListStaticRoutesParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListStaticRoutesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListStaticRoutesParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListStaticRoutesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListStaticRoutesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListStaticRoutesParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListStaticRoutesParams) SetTags(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *ListStaticRoutesParams) SetVpcid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcid"] = v
	return
}

// You should always use this function to get a new ListStaticRoutesParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewListStaticRoutesParams() *ListStaticRoutesParams {
	p := &ListStaticRoutesParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetStaticRouteByID(id string, opts ...OptionFunc) (*StaticRoute, int, error) {
	p := &ListStaticRoutesParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListStaticRoutes(p)
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
		return l.StaticRoutes[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for StaticRoute UUID: %s!", id)
}

// Lists all static routes
func (s *VPCService) ListStaticRoutes(p *ListStaticRoutesParams) (*ListStaticRoutesResponse, error) {
	resp, err := s.cs.newRequest("listStaticRoutes", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListStaticRoutesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListStaticRoutesResponse struct {
	Count        int            `json:"count"`
	StaticRoutes []*StaticRoute `json:"staticroute"`
}

type StaticRoute struct {
	Account   string `json:"account"`
	Cidr      string `json:"cidr"`
	Domain    string `json:"domain"`
	Domainid  string `json:"domainid"`
	Gatewayid string `json:"gatewayid"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Project   string `json:"project"`
	Projectid string `json:"projectid"`
	State     string `json:"state"`
	Tags      []Tags `json:"tags"`
	Vpcid     string `json:"vpcid"`
}

type ListVPCOfferingsParams struct {
	p map[string]interface{}
}

func (p *ListVPCOfferingsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["isdefault"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("isdefault", vv)
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
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["supportedservices"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("supportedservices", vv)
	}
	return u
}

func (p *ListVPCOfferingsParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *ListVPCOfferingsParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListVPCOfferingsParams) SetIsdefault(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isdefault"] = v
	return
}

func (p *ListVPCOfferingsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListVPCOfferingsParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListVPCOfferingsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListVPCOfferingsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListVPCOfferingsParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *ListVPCOfferingsParams) SetSupportedservices(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["supportedservices"] = v
	return
}

// You should always use this function to get a new ListVPCOfferingsParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewListVPCOfferingsParams() *ListVPCOfferingsParams {
	p := &ListVPCOfferingsParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetVPCOfferingID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListVPCOfferingsParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListVPCOfferings(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.VPCOfferings[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.VPCOfferings {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetVPCOfferingByName(name string, opts ...OptionFunc) (*VPCOffering, int, error) {
	id, count, err := s.GetVPCOfferingID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetVPCOfferingByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetVPCOfferingByID(id string, opts ...OptionFunc) (*VPCOffering, int, error) {
	p := &ListVPCOfferingsParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListVPCOfferings(p)
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
		return l.VPCOfferings[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for VPCOffering UUID: %s!", id)
}

// Lists VPC offerings
func (s *VPCService) ListVPCOfferings(p *ListVPCOfferingsParams) (*ListVPCOfferingsResponse, error) {
	resp, err := s.cs.newRequest("listVPCOfferings", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListVPCOfferingsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListVPCOfferingsResponse struct {
	Count        int            `json:"count"`
	VPCOfferings []*VPCOffering `json:"vpcoffering"`
}

type VPCOffering struct {
	Created                string               `json:"created"`
	Displaytext            string               `json:"displaytext"`
	Distributedvpcrouter   bool                 `json:"distributedvpcrouter"`
	Id                     string               `json:"id"`
	Isdefault              bool                 `json:"isdefault"`
	JobID                  string               `json:"jobid"`
	Jobstatus              int                  `json:"jobstatus"`
	Name                   string               `json:"name"`
	Service                []VPCOfferingService `json:"service"`
	State                  string               `json:"state"`
	SupportsregionLevelvpc bool                 `json:"supportsregionLevelvpc"`
}

type VPCOfferingService struct {
	Capability []VPCOfferingServiceCapability `json:"capability"`
	Name       string                         `json:"name"`
	Provider   []VPCOfferingServiceProvider   `json:"provider"`
}

type VPCOfferingServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type VPCOfferingServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type ListVPCsParams struct {
	p map[string]interface{}
}

func (p *ListVPCsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["cidr"]; found {
		u.Set("cidr", v.(string))
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["fordisplay"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("fordisplay", vv)
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
	if v, found := p.p["restartrequired"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("restartrequired", vv)
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
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
	if v, found := p.p["vpcofferingid"]; found {
		u.Set("vpcofferingid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListVPCsParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListVPCsParams) SetCidr(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["cidr"] = v
	return
}

func (p *ListVPCsParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *ListVPCsParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListVPCsParams) SetFordisplay(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["fordisplay"] = v
	return
}

func (p *ListVPCsParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListVPCsParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListVPCsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListVPCsParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListVPCsParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListVPCsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListVPCsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListVPCsParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListVPCsParams) SetRestartrequired(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["restartrequired"] = v
	return
}

func (p *ListVPCsParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *ListVPCsParams) SetSupportedservices(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["supportedservices"] = v
	return
}

func (p *ListVPCsParams) SetTags(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *ListVPCsParams) SetVpcofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["vpcofferingid"] = v
	return
}

func (p *ListVPCsParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListVPCsParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewListVPCsParams() *ListVPCsParams {
	p := &ListVPCsParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetVPCID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListVPCsParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListVPCs(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.VPCs[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.VPCs {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetVPCByName(name string, opts ...OptionFunc) (*VPC, int, error) {
	id, count, err := s.GetVPCID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetVPCByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VPCService) GetVPCByID(id string, opts ...OptionFunc) (*VPC, int, error) {
	p := &ListVPCsParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListVPCs(p)
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
		return l.VPCs[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for VPC UUID: %s!", id)
}

// Lists VPCs
func (s *VPCService) ListVPCs(p *ListVPCsParams) (*ListVPCsResponse, error) {
	resp, err := s.cs.newRequest("listVPCs", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListVPCsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListVPCsResponse struct {
	Count int    `json:"count"`
	VPCs  []*VPC `json:"vpc"`
}

type VPC struct {
	Account              string               `json:"account"`
	Cidr                 string               `json:"cidr"`
	Created              string               `json:"created"`
	Displaytext          string               `json:"displaytext"`
	Distributedvpcrouter bool                 `json:"distributedvpcrouter"`
	Domain               string               `json:"domain"`
	Domainid             string               `json:"domainid"`
	Fordisplay           bool                 `json:"fordisplay"`
	Id                   string               `json:"id"`
	JobID                string               `json:"jobid"`
	Jobstatus            int                  `json:"jobstatus"`
	Name                 string               `json:"name"`
	Network              []VPCNetwork         `json:"network"`
	Networkdomain        string               `json:"networkdomain"`
	Project              string               `json:"project"`
	Projectid            string               `json:"projectid"`
	Redundantvpcrouter   bool                 `json:"redundantvpcrouter"`
	Regionlevelvpc       bool                 `json:"regionlevelvpc"`
	Restartrequired      bool                 `json:"restartrequired"`
	Service              []VPCServiceInternal `json:"service"`
	State                string               `json:"state"`
	Tags                 []Tags               `json:"tags"`
	Vpcofferingid        string               `json:"vpcofferingid"`
	Zoneid               string               `json:"zoneid"`
	Zonename             string               `json:"zonename"`
}

type VPCServiceInternal struct {
	Capability []VPCServiceInternalCapability `json:"capability"`
	Name       string                         `json:"name"`
	Provider   []VPCServiceInternalProvider   `json:"provider"`
}

type VPCServiceInternalProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type VPCServiceInternalCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type VPCNetwork struct {
	Account                     string              `json:"account"`
	Aclid                       string              `json:"aclid"`
	Acltype                     string              `json:"acltype"`
	Broadcastdomaintype         string              `json:"broadcastdomaintype"`
	Broadcasturi                string              `json:"broadcasturi"`
	Canusefordeploy             bool                `json:"canusefordeploy"`
	Cidr                        string              `json:"cidr"`
	Displaynetwork              bool                `json:"displaynetwork"`
	Displaytext                 string              `json:"displaytext"`
	Dns1                        string              `json:"dns1"`
	Dns2                        string              `json:"dns2"`
	Domain                      string              `json:"domain"`
	Domainid                    string              `json:"domainid"`
	Externalid                  string              `json:"externalid"`
	Gateway                     string              `json:"gateway"`
	Id                          string              `json:"id"`
	Ip6cidr                     string              `json:"ip6cidr"`
	Ip6gateway                  string              `json:"ip6gateway"`
	Isdefault                   bool                `json:"isdefault"`
	Ispersistent                bool                `json:"ispersistent"`
	Issystem                    bool                `json:"issystem"`
	Name                        string              `json:"name"`
	Netmask                     string              `json:"netmask"`
	Networkcidr                 string              `json:"networkcidr"`
	Networkdomain               string              `json:"networkdomain"`
	Networkofferingavailability string              `json:"networkofferingavailability"`
	Networkofferingconservemode bool                `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string              `json:"networkofferingdisplaytext"`
	Networkofferingid           string              `json:"networkofferingid"`
	Networkofferingname         string              `json:"networkofferingname"`
	Physicalnetworkid           string              `json:"physicalnetworkid"`
	Project                     string              `json:"project"`
	Projectid                   string              `json:"projectid"`
	Redundantrouter             bool                `json:"redundantrouter"`
	Related                     string              `json:"related"`
	Reservediprange             string              `json:"reservediprange"`
	Restartrequired             bool                `json:"restartrequired"`
	Service                     []VPCNetworkService `json:"service"`
	Specifyipranges             bool                `json:"specifyipranges"`
	State                       string              `json:"state"`
	Strechedl2subnet            bool                `json:"strechedl2subnet"`
	Subdomainaccess             bool                `json:"subdomainaccess"`
	Tags                        []Tags              `json:"tags"`
	Traffictype                 string              `json:"traffictype"`
	Type                        string              `json:"type"`
	Vlan                        string              `json:"vlan"`
	Vpcid                       string              `json:"vpcid"`
	Zoneid                      string              `json:"zoneid"`
	Zonename                    string              `json:"zonename"`
	Zonesnetworkspans           []interface{}       `json:"zonesnetworkspans"`
}

type VPCNetworkService struct {
	Capability []VPCNetworkServiceCapability `json:"capability"`
	Name       string                        `json:"name"`
	Provider   []VPCNetworkServiceProvider   `json:"provider"`
}

type VPCNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type VPCNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type RestartVPCParams struct {
	p map[string]interface{}
}

func (p *RestartVPCParams) toURLValues() url.Values {
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

func (p *RestartVPCParams) SetCleanup(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["cleanup"] = v
	return
}

func (p *RestartVPCParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *RestartVPCParams) SetMakeredundant(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["makeredundant"] = v
	return
}

// You should always use this function to get a new RestartVPCParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewRestartVPCParams(id string) *RestartVPCParams {
	p := &RestartVPCParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Restarts a VPC
func (s *VPCService) RestartVPC(p *RestartVPCParams) (*RestartVPCResponse, error) {
	resp, err := s.cs.newRequest("restartVPC", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r RestartVPCResponse
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

type RestartVPCResponse struct {
	Account              string                      `json:"account"`
	Cidr                 string                      `json:"cidr"`
	Created              string                      `json:"created"`
	Displaytext          string                      `json:"displaytext"`
	Distributedvpcrouter bool                        `json:"distributedvpcrouter"`
	Domain               string                      `json:"domain"`
	Domainid             string                      `json:"domainid"`
	Fordisplay           bool                        `json:"fordisplay"`
	Id                   string                      `json:"id"`
	JobID                string                      `json:"jobid"`
	Jobstatus            int                         `json:"jobstatus"`
	Name                 string                      `json:"name"`
	Network              []RestartVPCResponseNetwork `json:"network"`
	Networkdomain        string                      `json:"networkdomain"`
	Project              string                      `json:"project"`
	Projectid            string                      `json:"projectid"`
	Redundantvpcrouter   bool                        `json:"redundantvpcrouter"`
	Regionlevelvpc       bool                        `json:"regionlevelvpc"`
	Restartrequired      bool                        `json:"restartrequired"`
	Service              []RestartVPCResponseService `json:"service"`
	State                string                      `json:"state"`
	Tags                 []Tags                      `json:"tags"`
	Vpcofferingid        string                      `json:"vpcofferingid"`
	Zoneid               string                      `json:"zoneid"`
	Zonename             string                      `json:"zonename"`
}

type RestartVPCResponseService struct {
	Capability []RestartVPCResponseServiceCapability `json:"capability"`
	Name       string                                `json:"name"`
	Provider   []RestartVPCResponseServiceProvider   `json:"provider"`
}

type RestartVPCResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type RestartVPCResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type RestartVPCResponseNetwork struct {
	Account                     string                             `json:"account"`
	Aclid                       string                             `json:"aclid"`
	Acltype                     string                             `json:"acltype"`
	Broadcastdomaintype         string                             `json:"broadcastdomaintype"`
	Broadcasturi                string                             `json:"broadcasturi"`
	Canusefordeploy             bool                               `json:"canusefordeploy"`
	Cidr                        string                             `json:"cidr"`
	Displaynetwork              bool                               `json:"displaynetwork"`
	Displaytext                 string                             `json:"displaytext"`
	Dns1                        string                             `json:"dns1"`
	Dns2                        string                             `json:"dns2"`
	Domain                      string                             `json:"domain"`
	Domainid                    string                             `json:"domainid"`
	Externalid                  string                             `json:"externalid"`
	Gateway                     string                             `json:"gateway"`
	Id                          string                             `json:"id"`
	Ip6cidr                     string                             `json:"ip6cidr"`
	Ip6gateway                  string                             `json:"ip6gateway"`
	Isdefault                   bool                               `json:"isdefault"`
	Ispersistent                bool                               `json:"ispersistent"`
	Issystem                    bool                               `json:"issystem"`
	Name                        string                             `json:"name"`
	Netmask                     string                             `json:"netmask"`
	Networkcidr                 string                             `json:"networkcidr"`
	Networkdomain               string                             `json:"networkdomain"`
	Networkofferingavailability string                             `json:"networkofferingavailability"`
	Networkofferingconservemode bool                               `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                             `json:"networkofferingdisplaytext"`
	Networkofferingid           string                             `json:"networkofferingid"`
	Networkofferingname         string                             `json:"networkofferingname"`
	Physicalnetworkid           string                             `json:"physicalnetworkid"`
	Project                     string                             `json:"project"`
	Projectid                   string                             `json:"projectid"`
	Redundantrouter             bool                               `json:"redundantrouter"`
	Related                     string                             `json:"related"`
	Reservediprange             string                             `json:"reservediprange"`
	Restartrequired             bool                               `json:"restartrequired"`
	Service                     []RestartVPCResponseNetworkService `json:"service"`
	Specifyipranges             bool                               `json:"specifyipranges"`
	State                       string                             `json:"state"`
	Strechedl2subnet            bool                               `json:"strechedl2subnet"`
	Subdomainaccess             bool                               `json:"subdomainaccess"`
	Tags                        []Tags                             `json:"tags"`
	Traffictype                 string                             `json:"traffictype"`
	Type                        string                             `json:"type"`
	Vlan                        string                             `json:"vlan"`
	Vpcid                       string                             `json:"vpcid"`
	Zoneid                      string                             `json:"zoneid"`
	Zonename                    string                             `json:"zonename"`
	Zonesnetworkspans           []interface{}                      `json:"zonesnetworkspans"`
}

type RestartVPCResponseNetworkService struct {
	Capability []RestartVPCResponseNetworkServiceCapability `json:"capability"`
	Name       string                                       `json:"name"`
	Provider   []RestartVPCResponseNetworkServiceProvider   `json:"provider"`
}

type RestartVPCResponseNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type RestartVPCResponseNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type UpdateVPCParams struct {
	p map[string]interface{}
}

func (p *UpdateVPCParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["fordisplay"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("fordisplay", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	return u
}

func (p *UpdateVPCParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *UpdateVPCParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *UpdateVPCParams) SetFordisplay(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["fordisplay"] = v
	return
}

func (p *UpdateVPCParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateVPCParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

// You should always use this function to get a new UpdateVPCParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewUpdateVPCParams(id string) *UpdateVPCParams {
	p := &UpdateVPCParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates a VPC
func (s *VPCService) UpdateVPC(p *UpdateVPCParams) (*UpdateVPCResponse, error) {
	resp, err := s.cs.newRequest("updateVPC", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateVPCResponse
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

type UpdateVPCResponse struct {
	Account              string                     `json:"account"`
	Cidr                 string                     `json:"cidr"`
	Created              string                     `json:"created"`
	Displaytext          string                     `json:"displaytext"`
	Distributedvpcrouter bool                       `json:"distributedvpcrouter"`
	Domain               string                     `json:"domain"`
	Domainid             string                     `json:"domainid"`
	Fordisplay           bool                       `json:"fordisplay"`
	Id                   string                     `json:"id"`
	JobID                string                     `json:"jobid"`
	Jobstatus            int                        `json:"jobstatus"`
	Name                 string                     `json:"name"`
	Network              []UpdateVPCResponseNetwork `json:"network"`
	Networkdomain        string                     `json:"networkdomain"`
	Project              string                     `json:"project"`
	Projectid            string                     `json:"projectid"`
	Redundantvpcrouter   bool                       `json:"redundantvpcrouter"`
	Regionlevelvpc       bool                       `json:"regionlevelvpc"`
	Restartrequired      bool                       `json:"restartrequired"`
	Service              []UpdateVPCResponseService `json:"service"`
	State                string                     `json:"state"`
	Tags                 []Tags                     `json:"tags"`
	Vpcofferingid        string                     `json:"vpcofferingid"`
	Zoneid               string                     `json:"zoneid"`
	Zonename             string                     `json:"zonename"`
}

type UpdateVPCResponseService struct {
	Capability []UpdateVPCResponseServiceCapability `json:"capability"`
	Name       string                               `json:"name"`
	Provider   []UpdateVPCResponseServiceProvider   `json:"provider"`
}

type UpdateVPCResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type UpdateVPCResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type UpdateVPCResponseNetwork struct {
	Account                     string                            `json:"account"`
	Aclid                       string                            `json:"aclid"`
	Acltype                     string                            `json:"acltype"`
	Broadcastdomaintype         string                            `json:"broadcastdomaintype"`
	Broadcasturi                string                            `json:"broadcasturi"`
	Canusefordeploy             bool                              `json:"canusefordeploy"`
	Cidr                        string                            `json:"cidr"`
	Displaynetwork              bool                              `json:"displaynetwork"`
	Displaytext                 string                            `json:"displaytext"`
	Dns1                        string                            `json:"dns1"`
	Dns2                        string                            `json:"dns2"`
	Domain                      string                            `json:"domain"`
	Domainid                    string                            `json:"domainid"`
	Externalid                  string                            `json:"externalid"`
	Gateway                     string                            `json:"gateway"`
	Id                          string                            `json:"id"`
	Ip6cidr                     string                            `json:"ip6cidr"`
	Ip6gateway                  string                            `json:"ip6gateway"`
	Isdefault                   bool                              `json:"isdefault"`
	Ispersistent                bool                              `json:"ispersistent"`
	Issystem                    bool                              `json:"issystem"`
	Name                        string                            `json:"name"`
	Netmask                     string                            `json:"netmask"`
	Networkcidr                 string                            `json:"networkcidr"`
	Networkdomain               string                            `json:"networkdomain"`
	Networkofferingavailability string                            `json:"networkofferingavailability"`
	Networkofferingconservemode bool                              `json:"networkofferingconservemode"`
	Networkofferingdisplaytext  string                            `json:"networkofferingdisplaytext"`
	Networkofferingid           string                            `json:"networkofferingid"`
	Networkofferingname         string                            `json:"networkofferingname"`
	Physicalnetworkid           string                            `json:"physicalnetworkid"`
	Project                     string                            `json:"project"`
	Projectid                   string                            `json:"projectid"`
	Redundantrouter             bool                              `json:"redundantrouter"`
	Related                     string                            `json:"related"`
	Reservediprange             string                            `json:"reservediprange"`
	Restartrequired             bool                              `json:"restartrequired"`
	Service                     []UpdateVPCResponseNetworkService `json:"service"`
	Specifyipranges             bool                              `json:"specifyipranges"`
	State                       string                            `json:"state"`
	Strechedl2subnet            bool                              `json:"strechedl2subnet"`
	Subdomainaccess             bool                              `json:"subdomainaccess"`
	Tags                        []Tags                            `json:"tags"`
	Traffictype                 string                            `json:"traffictype"`
	Type                        string                            `json:"type"`
	Vlan                        string                            `json:"vlan"`
	Vpcid                       string                            `json:"vpcid"`
	Zoneid                      string                            `json:"zoneid"`
	Zonename                    string                            `json:"zonename"`
	Zonesnetworkspans           []interface{}                     `json:"zonesnetworkspans"`
}

type UpdateVPCResponseNetworkService struct {
	Capability []UpdateVPCResponseNetworkServiceCapability `json:"capability"`
	Name       string                                      `json:"name"`
	Provider   []UpdateVPCResponseNetworkServiceProvider   `json:"provider"`
}

type UpdateVPCResponseNetworkServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type UpdateVPCResponseNetworkServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type UpdateVPCOfferingParams struct {
	p map[string]interface{}
}

func (p *UpdateVPCOfferingParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	return u
}

func (p *UpdateVPCOfferingParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *UpdateVPCOfferingParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateVPCOfferingParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UpdateVPCOfferingParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

// You should always use this function to get a new UpdateVPCOfferingParams instance,
// as then you are sure you have configured all required params
func (s *VPCService) NewUpdateVPCOfferingParams(id string) *UpdateVPCOfferingParams {
	p := &UpdateVPCOfferingParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates VPC offering
func (s *VPCService) UpdateVPCOffering(p *UpdateVPCOfferingParams) (*UpdateVPCOfferingResponse, error) {
	resp, err := s.cs.newRequest("updateVPCOffering", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateVPCOfferingResponse
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

type UpdateVPCOfferingResponse struct {
	Created                string                             `json:"created"`
	Displaytext            string                             `json:"displaytext"`
	Distributedvpcrouter   bool                               `json:"distributedvpcrouter"`
	Id                     string                             `json:"id"`
	Isdefault              bool                               `json:"isdefault"`
	JobID                  string                             `json:"jobid"`
	Jobstatus              int                                `json:"jobstatus"`
	Name                   string                             `json:"name"`
	Service                []UpdateVPCOfferingResponseService `json:"service"`
	State                  string                             `json:"state"`
	SupportsregionLevelvpc bool                               `json:"supportsregionLevelvpc"`
}

type UpdateVPCOfferingResponseService struct {
	Capability []UpdateVPCOfferingResponseServiceCapability `json:"capability"`
	Name       string                                       `json:"name"`
	Provider   []UpdateVPCOfferingResponseServiceProvider   `json:"provider"`
}

type UpdateVPCOfferingResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type UpdateVPCOfferingResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

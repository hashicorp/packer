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

type CreateNetworkOfferingParams struct {
	p map[string]interface{}
}

func (p *CreateNetworkOfferingParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["availability"]; found {
		u.Set("availability", v.(string))
	}
	if v, found := p.p["conservemode"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("conservemode", vv)
	}
	if v, found := p.p["details"]; found {
		i := 0
		for k, vv := range v.(map[string]string) {
			u.Set(fmt.Sprintf("details[%d].%s", i, k), vv)
			i++
		}
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["egressdefaultpolicy"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("egressdefaultpolicy", vv)
	}
	if v, found := p.p["forvpc"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("forvpc", vv)
	}
	if v, found := p.p["guestiptype"]; found {
		u.Set("guestiptype", v.(string))
	}
	if v, found := p.p["ispersistent"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("ispersistent", vv)
	}
	if v, found := p.p["keepaliveenabled"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("keepaliveenabled", vv)
	}
	if v, found := p.p["maxconnections"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("maxconnections", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["networkrate"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("networkrate", vv)
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
	if v, found := p.p["specifyipranges"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("specifyipranges", vv)
	}
	if v, found := p.p["specifyvlan"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("specifyvlan", vv)
	}
	if v, found := p.p["supportedservices"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("supportedservices", vv)
	}
	if v, found := p.p["tags"]; found {
		u.Set("tags", v.(string))
	}
	if v, found := p.p["traffictype"]; found {
		u.Set("traffictype", v.(string))
	}
	return u
}

func (p *CreateNetworkOfferingParams) SetAvailability(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["availability"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetConservemode(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["conservemode"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetDetails(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["details"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetEgressdefaultpolicy(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["egressdefaultpolicy"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetForvpc(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forvpc"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetGuestiptype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["guestiptype"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetIspersistent(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ispersistent"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetKeepaliveenabled(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keepaliveenabled"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetMaxconnections(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["maxconnections"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetNetworkrate(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkrate"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetServicecapabilitylist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["servicecapabilitylist"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetServiceofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceofferingid"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetServiceproviderlist(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["serviceproviderlist"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetSpecifyipranges(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["specifyipranges"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetSpecifyvlan(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["specifyvlan"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetSupportedservices(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["supportedservices"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetTags(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *CreateNetworkOfferingParams) SetTraffictype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["traffictype"] = v
	return
}

// You should always use this function to get a new CreateNetworkOfferingParams instance,
// as then you are sure you have configured all required params
func (s *NetworkOfferingService) NewCreateNetworkOfferingParams(displaytext string, guestiptype string, name string, supportedservices []string, traffictype string) *CreateNetworkOfferingParams {
	p := &CreateNetworkOfferingParams{}
	p.p = make(map[string]interface{})
	p.p["displaytext"] = displaytext
	p.p["guestiptype"] = guestiptype
	p.p["name"] = name
	p.p["supportedservices"] = supportedservices
	p.p["traffictype"] = traffictype
	return p
}

// Creates a network offering.
func (s *NetworkOfferingService) CreateNetworkOffering(p *CreateNetworkOfferingParams) (*CreateNetworkOfferingResponse, error) {
	resp, err := s.cs.newRequest("createNetworkOffering", p.toURLValues())
	if err != nil {
		return nil, err
	}

	if resp, err = getRawValue(resp); err != nil {
		return nil, err
	}

	var r CreateNetworkOfferingResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type CreateNetworkOfferingResponse struct {
	Availability             string                                 `json:"availability"`
	Conservemode             bool                                   `json:"conservemode"`
	Created                  string                                 `json:"created"`
	Details                  map[string]string                      `json:"details"`
	Displaytext              string                                 `json:"displaytext"`
	Egressdefaultpolicy      bool                                   `json:"egressdefaultpolicy"`
	Forvpc                   bool                                   `json:"forvpc"`
	Guestiptype              string                                 `json:"guestiptype"`
	Id                       string                                 `json:"id"`
	Isdefault                bool                                   `json:"isdefault"`
	Ispersistent             bool                                   `json:"ispersistent"`
	JobID                    string                                 `json:"jobid"`
	Jobstatus                int                                    `json:"jobstatus"`
	Maxconnections           int                                    `json:"maxconnections"`
	Name                     string                                 `json:"name"`
	Networkrate              int                                    `json:"networkrate"`
	Service                  []CreateNetworkOfferingResponseService `json:"service"`
	Serviceofferingid        string                                 `json:"serviceofferingid"`
	Specifyipranges          bool                                   `json:"specifyipranges"`
	Specifyvlan              bool                                   `json:"specifyvlan"`
	State                    string                                 `json:"state"`
	Supportspublicaccess     bool                                   `json:"supportspublicaccess"`
	Supportsstrechedl2subnet bool                                   `json:"supportsstrechedl2subnet"`
	Tags                     string                                 `json:"tags"`
	Traffictype              string                                 `json:"traffictype"`
}

type CreateNetworkOfferingResponseService struct {
	Capability []CreateNetworkOfferingResponseServiceCapability `json:"capability"`
	Name       string                                           `json:"name"`
	Provider   []CreateNetworkOfferingResponseServiceProvider   `json:"provider"`
}

type CreateNetworkOfferingResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type CreateNetworkOfferingResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type DeleteNetworkOfferingParams struct {
	p map[string]interface{}
}

func (p *DeleteNetworkOfferingParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteNetworkOfferingParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteNetworkOfferingParams instance,
// as then you are sure you have configured all required params
func (s *NetworkOfferingService) NewDeleteNetworkOfferingParams(id string) *DeleteNetworkOfferingParams {
	p := &DeleteNetworkOfferingParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a network offering.
func (s *NetworkOfferingService) DeleteNetworkOffering(p *DeleteNetworkOfferingParams) (*DeleteNetworkOfferingResponse, error) {
	resp, err := s.cs.newRequest("deleteNetworkOffering", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteNetworkOfferingResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DeleteNetworkOfferingResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *DeleteNetworkOfferingResponse) UnmarshalJSON(b []byte) error {
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

	type alias DeleteNetworkOfferingResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ListNetworkOfferingsParams struct {
	p map[string]interface{}
}

func (p *ListNetworkOfferingsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["availability"]; found {
		u.Set("availability", v.(string))
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["forvpc"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("forvpc", vv)
	}
	if v, found := p.p["guestiptype"]; found {
		u.Set("guestiptype", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["isdefault"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("isdefault", vv)
	}
	if v, found := p.p["istagged"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("istagged", vv)
	}
	if v, found := p.p["keyword"]; found {
		u.Set("keyword", v.(string))
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
	if v, found := p.p["sourcenatsupported"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("sourcenatsupported", vv)
	}
	if v, found := p.p["specifyipranges"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("specifyipranges", vv)
	}
	if v, found := p.p["specifyvlan"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("specifyvlan", vv)
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["supportedservices"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("supportedservices", vv)
	}
	if v, found := p.p["tags"]; found {
		u.Set("tags", v.(string))
	}
	if v, found := p.p["traffictype"]; found {
		u.Set("traffictype", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListNetworkOfferingsParams) SetAvailability(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["availability"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetForvpc(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["forvpc"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetGuestiptype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["guestiptype"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetIsdefault(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isdefault"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetIstagged(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["istagged"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetNetworkid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["networkid"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetSourcenatsupported(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["sourcenatsupported"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetSpecifyipranges(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["specifyipranges"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetSpecifyvlan(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["specifyvlan"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetSupportedservices(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["supportedservices"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetTags(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetTraffictype(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["traffictype"] = v
	return
}

func (p *ListNetworkOfferingsParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListNetworkOfferingsParams instance,
// as then you are sure you have configured all required params
func (s *NetworkOfferingService) NewListNetworkOfferingsParams() *ListNetworkOfferingsParams {
	p := &ListNetworkOfferingsParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkOfferingService) GetNetworkOfferingID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListNetworkOfferingsParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListNetworkOfferings(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.NetworkOfferings[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.NetworkOfferings {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkOfferingService) GetNetworkOfferingByName(name string, opts ...OptionFunc) (*NetworkOffering, int, error) {
	id, count, err := s.GetNetworkOfferingID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetNetworkOfferingByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *NetworkOfferingService) GetNetworkOfferingByID(id string, opts ...OptionFunc) (*NetworkOffering, int, error) {
	p := &ListNetworkOfferingsParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListNetworkOfferings(p)
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
		return l.NetworkOfferings[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for NetworkOffering UUID: %s!", id)
}

// Lists all available network offerings.
func (s *NetworkOfferingService) ListNetworkOfferings(p *ListNetworkOfferingsParams) (*ListNetworkOfferingsResponse, error) {
	resp, err := s.cs.newRequest("listNetworkOfferings", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListNetworkOfferingsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListNetworkOfferingsResponse struct {
	Count            int                `json:"count"`
	NetworkOfferings []*NetworkOffering `json:"networkoffering"`
}

type NetworkOffering struct {
	Availability             string                           `json:"availability"`
	Conservemode             bool                             `json:"conservemode"`
	Created                  string                           `json:"created"`
	Details                  map[string]string                `json:"details"`
	Displaytext              string                           `json:"displaytext"`
	Egressdefaultpolicy      bool                             `json:"egressdefaultpolicy"`
	Forvpc                   bool                             `json:"forvpc"`
	Guestiptype              string                           `json:"guestiptype"`
	Id                       string                           `json:"id"`
	Isdefault                bool                             `json:"isdefault"`
	Ispersistent             bool                             `json:"ispersistent"`
	JobID                    string                           `json:"jobid"`
	Jobstatus                int                              `json:"jobstatus"`
	Maxconnections           int                              `json:"maxconnections"`
	Name                     string                           `json:"name"`
	Networkrate              int                              `json:"networkrate"`
	Service                  []NetworkOfferingServiceInternal `json:"service"`
	Serviceofferingid        string                           `json:"serviceofferingid"`
	Specifyipranges          bool                             `json:"specifyipranges"`
	Specifyvlan              bool                             `json:"specifyvlan"`
	State                    string                           `json:"state"`
	Supportspublicaccess     bool                             `json:"supportspublicaccess"`
	Supportsstrechedl2subnet bool                             `json:"supportsstrechedl2subnet"`
	Tags                     string                           `json:"tags"`
	Traffictype              string                           `json:"traffictype"`
}

type NetworkOfferingServiceInternal struct {
	Capability []NetworkOfferingServiceInternalCapability `json:"capability"`
	Name       string                                     `json:"name"`
	Provider   []NetworkOfferingServiceInternalProvider   `json:"provider"`
}

type NetworkOfferingServiceInternalProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type NetworkOfferingServiceInternalCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

type UpdateNetworkOfferingParams struct {
	p map[string]interface{}
}

func (p *UpdateNetworkOfferingParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["availability"]; found {
		u.Set("availability", v.(string))
	}
	if v, found := p.p["displaytext"]; found {
		u.Set("displaytext", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["keepaliveenabled"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("keepaliveenabled", vv)
	}
	if v, found := p.p["maxconnections"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("maxconnections", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["sortkey"]; found {
		vv := strconv.Itoa(v.(int))
		u.Set("sortkey", vv)
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["tags"]; found {
		u.Set("tags", v.(string))
	}
	return u
}

func (p *UpdateNetworkOfferingParams) SetAvailability(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["availability"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetDisplaytext(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displaytext"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetKeepaliveenabled(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keepaliveenabled"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetMaxconnections(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["maxconnections"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetSortkey(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["sortkey"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *UpdateNetworkOfferingParams) SetTags(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

// You should always use this function to get a new UpdateNetworkOfferingParams instance,
// as then you are sure you have configured all required params
func (s *NetworkOfferingService) NewUpdateNetworkOfferingParams() *UpdateNetworkOfferingParams {
	p := &UpdateNetworkOfferingParams{}
	p.p = make(map[string]interface{})
	return p
}

// Updates a network offering.
func (s *NetworkOfferingService) UpdateNetworkOffering(p *UpdateNetworkOfferingParams) (*UpdateNetworkOfferingResponse, error) {
	resp, err := s.cs.newRequest("updateNetworkOffering", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateNetworkOfferingResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type UpdateNetworkOfferingResponse struct {
	Availability             string                                 `json:"availability"`
	Conservemode             bool                                   `json:"conservemode"`
	Created                  string                                 `json:"created"`
	Details                  map[string]string                      `json:"details"`
	Displaytext              string                                 `json:"displaytext"`
	Egressdefaultpolicy      bool                                   `json:"egressdefaultpolicy"`
	Forvpc                   bool                                   `json:"forvpc"`
	Guestiptype              string                                 `json:"guestiptype"`
	Id                       string                                 `json:"id"`
	Isdefault                bool                                   `json:"isdefault"`
	Ispersistent             bool                                   `json:"ispersistent"`
	JobID                    string                                 `json:"jobid"`
	Jobstatus                int                                    `json:"jobstatus"`
	Maxconnections           int                                    `json:"maxconnections"`
	Name                     string                                 `json:"name"`
	Networkrate              int                                    `json:"networkrate"`
	Service                  []UpdateNetworkOfferingResponseService `json:"service"`
	Serviceofferingid        string                                 `json:"serviceofferingid"`
	Specifyipranges          bool                                   `json:"specifyipranges"`
	Specifyvlan              bool                                   `json:"specifyvlan"`
	State                    string                                 `json:"state"`
	Supportspublicaccess     bool                                   `json:"supportspublicaccess"`
	Supportsstrechedl2subnet bool                                   `json:"supportsstrechedl2subnet"`
	Tags                     string                                 `json:"tags"`
	Traffictype              string                                 `json:"traffictype"`
}

type UpdateNetworkOfferingResponseService struct {
	Capability []UpdateNetworkOfferingResponseServiceCapability `json:"capability"`
	Name       string                                           `json:"name"`
	Provider   []UpdateNetworkOfferingResponseServiceProvider   `json:"provider"`
}

type UpdateNetworkOfferingResponseServiceProvider struct {
	Canenableindividualservice   bool     `json:"canenableindividualservice"`
	Destinationphysicalnetworkid string   `json:"destinationphysicalnetworkid"`
	Id                           string   `json:"id"`
	Name                         string   `json:"name"`
	Physicalnetworkid            string   `json:"physicalnetworkid"`
	Servicelist                  []string `json:"servicelist"`
	State                        string   `json:"state"`
}

type UpdateNetworkOfferingResponseServiceCapability struct {
	Canchooseservicecapability bool   `json:"canchooseservicecapability"`
	Name                       string `json:"name"`
	Value                      string `json:"value"`
}

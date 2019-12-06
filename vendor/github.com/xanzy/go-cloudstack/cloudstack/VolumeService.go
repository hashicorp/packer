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

type AttachVolumeParams struct {
	p map[string]interface{}
}

func (p *AttachVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["deviceid"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("deviceid", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *AttachVolumeParams) SetDeviceid(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["deviceid"] = v
	return
}

func (p *AttachVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *AttachVolumeParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new AttachVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewAttachVolumeParams(id string, virtualmachineid string) *AttachVolumeParams {
	p := &AttachVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	p.p["virtualmachineid"] = virtualmachineid
	return p
}

// Attaches a disk volume to a virtual machine.
func (s *VolumeService) AttachVolume(p *AttachVolumeParams) (*AttachVolumeResponse, error) {
	resp, err := s.cs.newRequest("attachVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r AttachVolumeResponse
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

type AttachVolumeResponse struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

type CreateVolumeParams struct {
	p map[string]interface{}
}

func (p *CreateVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["displayvolume"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvolume", vv)
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["maxiops"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("maxiops", vv)
	}
	if v, found := p.p["miniops"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("miniops", vv)
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["size"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("size", vv)
	}
	if v, found := p.p["snapshotid"]; found {
		u.Set("snapshotid", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *CreateVolumeParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *CreateVolumeParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *CreateVolumeParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *CreateVolumeParams) SetDisplayvolume(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvolume"] = v
	return
}

func (p *CreateVolumeParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *CreateVolumeParams) SetMaxiops(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["maxiops"] = v
	return
}

func (p *CreateVolumeParams) SetMiniops(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["miniops"] = v
	return
}

func (p *CreateVolumeParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateVolumeParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *CreateVolumeParams) SetSize(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["size"] = v
	return
}

func (p *CreateVolumeParams) SetSnapshotid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["snapshotid"] = v
	return
}

func (p *CreateVolumeParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

func (p *CreateVolumeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new CreateVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewCreateVolumeParams() *CreateVolumeParams {
	p := &CreateVolumeParams{}
	p.p = make(map[string]interface{})
	return p
}

// Creates a disk volume from a disk offering. This disk volume must still be attached to a virtual machine to make use of it.
func (s *VolumeService) CreateVolume(p *CreateVolumeParams) (*CreateVolumeResponse, error) {
	resp, err := s.cs.newRequest("createVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateVolumeResponse
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

type CreateVolumeResponse struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

type DeleteVolumeParams struct {
	p map[string]interface{}
}

func (p *DeleteVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewDeleteVolumeParams(id string) *DeleteVolumeParams {
	p := &DeleteVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a detached disk volume.
func (s *VolumeService) DeleteVolume(p *DeleteVolumeParams) (*DeleteVolumeResponse, error) {
	resp, err := s.cs.newRequest("deleteVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteVolumeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DeleteVolumeResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *DeleteVolumeResponse) UnmarshalJSON(b []byte) error {
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

	type alias DeleteVolumeResponse
	return json.Unmarshal(b, (*alias)(r))
}

type DetachVolumeParams struct {
	p map[string]interface{}
}

func (p *DetachVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["deviceid"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("deviceid", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	return u
}

func (p *DetachVolumeParams) SetDeviceid(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["deviceid"] = v
	return
}

func (p *DetachVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *DetachVolumeParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

// You should always use this function to get a new DetachVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewDetachVolumeParams() *DetachVolumeParams {
	p := &DetachVolumeParams{}
	p.p = make(map[string]interface{})
	return p
}

// Detaches a disk volume from a virtual machine.
func (s *VolumeService) DetachVolume(p *DetachVolumeParams) (*DetachVolumeResponse, error) {
	resp, err := s.cs.newRequest("detachVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DetachVolumeResponse
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

type DetachVolumeResponse struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

type ExtractVolumeParams struct {
	p map[string]interface{}
}

func (p *ExtractVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["mode"]; found {
		u.Set("mode", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ExtractVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ExtractVolumeParams) SetMode(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["mode"] = v
	return
}

func (p *ExtractVolumeParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *ExtractVolumeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ExtractVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewExtractVolumeParams(id string, mode string, zoneid string) *ExtractVolumeParams {
	p := &ExtractVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	p.p["mode"] = mode
	p.p["zoneid"] = zoneid
	return p
}

// Extracts volume
func (s *VolumeService) ExtractVolume(p *ExtractVolumeParams) (*ExtractVolumeResponse, error) {
	resp, err := s.cs.newRequest("extractVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ExtractVolumeResponse
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

type ExtractVolumeResponse struct {
	Accountid        string `json:"accountid"`
	Created          string `json:"created"`
	ExtractId        string `json:"extractId"`
	ExtractMode      string `json:"extractMode"`
	Id               string `json:"id"`
	JobID            string `json:"jobid"`
	Jobstatus        int    `json:"jobstatus"`
	Name             string `json:"name"`
	Resultstring     string `json:"resultstring"`
	State            string `json:"state"`
	Status           string `json:"status"`
	Storagetype      string `json:"storagetype"`
	Uploadpercentage int    `json:"uploadpercentage"`
	Url              string `json:"url"`
	Zoneid           string `json:"zoneid"`
	Zonename         string `json:"zonename"`
}

type GetPathForVolumeParams struct {
	p map[string]interface{}
}

func (p *GetPathForVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["volumeid"]; found {
		u.Set("volumeid", v.(string))
	}
	return u
}

func (p *GetPathForVolumeParams) SetVolumeid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["volumeid"] = v
	return
}

// You should always use this function to get a new GetPathForVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewGetPathForVolumeParams(volumeid string) *GetPathForVolumeParams {
	p := &GetPathForVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["volumeid"] = volumeid
	return p
}

// Get the path associated with the provided volume UUID
func (s *VolumeService) GetPathForVolume(p *GetPathForVolumeParams) (*GetPathForVolumeResponse, error) {
	resp, err := s.cs.newRequest("getPathForVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r GetPathForVolumeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type GetPathForVolumeResponse struct {
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Path      string `json:"path"`
}

type GetSolidFireVolumeSizeParams struct {
	p map[string]interface{}
}

func (p *GetSolidFireVolumeSizeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["volumeid"]; found {
		u.Set("volumeid", v.(string))
	}
	return u
}

func (p *GetSolidFireVolumeSizeParams) SetVolumeid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["volumeid"] = v
	return
}

// You should always use this function to get a new GetSolidFireVolumeSizeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewGetSolidFireVolumeSizeParams(volumeid string) *GetSolidFireVolumeSizeParams {
	p := &GetSolidFireVolumeSizeParams{}
	p.p = make(map[string]interface{})
	p.p["volumeid"] = volumeid
	return p
}

// Get the SF volume size including Hypervisor Snapshot Reserve
func (s *VolumeService) GetSolidFireVolumeSize(p *GetSolidFireVolumeSizeParams) (*GetSolidFireVolumeSizeResponse, error) {
	resp, err := s.cs.newRequest("getSolidFireVolumeSize", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r GetSolidFireVolumeSizeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type GetSolidFireVolumeSizeResponse struct {
	JobID               string `json:"jobid"`
	Jobstatus           int    `json:"jobstatus"`
	SolidFireVolumeSize int64  `json:"solidFireVolumeSize"`
}

type GetUploadParamsForVolumeParams struct {
	p map[string]interface{}
}

func (p *GetUploadParamsForVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["checksum"]; found {
		u.Set("checksum", v.(string))
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["format"]; found {
		u.Set("format", v.(string))
	}
	if v, found := p.p["imagestoreuuid"]; found {
		u.Set("imagestoreuuid", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *GetUploadParamsForVolumeParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetChecksum(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["checksum"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetFormat(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["format"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetImagestoreuuid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["imagestoreuuid"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *GetUploadParamsForVolumeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new GetUploadParamsForVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewGetUploadParamsForVolumeParams(format string, name string, zoneid string) *GetUploadParamsForVolumeParams {
	p := &GetUploadParamsForVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["format"] = format
	p.p["name"] = name
	p.p["zoneid"] = zoneid
	return p
}

// Upload a data disk to the cloudstack cloud.
func (s *VolumeService) GetUploadParamsForVolume(p *GetUploadParamsForVolumeParams) (*GetUploadParamsForVolumeResponse, error) {
	resp, err := s.cs.newRequest("getUploadParamsForVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r GetUploadParamsForVolumeResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type GetUploadParamsForVolumeResponse struct {
	Expires   string `json:"expires"`
	Id        string `json:"id"`
	JobID     string `json:"jobid"`
	Jobstatus int    `json:"jobstatus"`
	Metadata  string `json:"metadata"`
	PostURL   string `json:"postURL"`
	Signature string `json:"signature"`
}

type GetVolumeiScsiNameParams struct {
	p map[string]interface{}
}

func (p *GetVolumeiScsiNameParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["volumeid"]; found {
		u.Set("volumeid", v.(string))
	}
	return u
}

func (p *GetVolumeiScsiNameParams) SetVolumeid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["volumeid"] = v
	return
}

// You should always use this function to get a new GetVolumeiScsiNameParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewGetVolumeiScsiNameParams(volumeid string) *GetVolumeiScsiNameParams {
	p := &GetVolumeiScsiNameParams{}
	p.p = make(map[string]interface{})
	p.p["volumeid"] = volumeid
	return p
}

// Get Volume's iSCSI Name
func (s *VolumeService) GetVolumeiScsiName(p *GetVolumeiScsiNameParams) (*GetVolumeiScsiNameResponse, error) {
	resp, err := s.cs.newRequest("getVolumeiScsiName", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r GetVolumeiScsiNameResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type GetVolumeiScsiNameResponse struct {
	JobID           string `json:"jobid"`
	Jobstatus       int    `json:"jobstatus"`
	VolumeiScsiName string `json:"volumeiScsiName"`
}

type ListVolumesParams struct {
	p map[string]interface{}
}

func (p *ListVolumesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["clusterid"]; found {
		u.Set("clusterid", v.(string))
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["displayvolume"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvolume", vv)
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["hostid"]; found {
		u.Set("hostid", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["ids"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("ids", vv)
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
	if v, found := p.p["podid"]; found {
		u.Set("podid", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
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
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	if v, found := p.p["virtualmachineid"]; found {
		u.Set("virtualmachineid", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *ListVolumesParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *ListVolumesParams) SetClusterid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["clusterid"] = v
	return
}

func (p *ListVolumesParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *ListVolumesParams) SetDisplayvolume(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvolume"] = v
	return
}

func (p *ListVolumesParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *ListVolumesParams) SetHostid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["hostid"] = v
	return
}

func (p *ListVolumesParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListVolumesParams) SetIds(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ids"] = v
	return
}

func (p *ListVolumesParams) SetIsrecursive(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["isrecursive"] = v
	return
}

func (p *ListVolumesParams) SetKeyword(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["keyword"] = v
	return
}

func (p *ListVolumesParams) SetListall(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["listall"] = v
	return
}

func (p *ListVolumesParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListVolumesParams) SetPage(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["page"] = v
	return
}

func (p *ListVolumesParams) SetPagesize(v int) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["pagesize"] = v
	return
}

func (p *ListVolumesParams) SetPodid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["podid"] = v
	return
}

func (p *ListVolumesParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *ListVolumesParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *ListVolumesParams) SetTags(v map[string]string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["tags"] = v
	return
}

func (p *ListVolumesParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

func (p *ListVolumesParams) SetVirtualmachineid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["virtualmachineid"] = v
	return
}

func (p *ListVolumesParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new ListVolumesParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewListVolumesParams() *ListVolumesParams {
	p := &ListVolumesParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VolumeService) GetVolumeID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListVolumesParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListVolumes(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.Volumes[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.Volumes {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VolumeService) GetVolumeByName(name string, opts ...OptionFunc) (*Volume, int, error) {
	id, count, err := s.GetVolumeID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetVolumeByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *VolumeService) GetVolumeByID(id string, opts ...OptionFunc) (*Volume, int, error) {
	p := &ListVolumesParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListVolumes(p)
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
		return l.Volumes[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for Volume UUID: %s!", id)
}

// Lists all volumes.
func (s *VolumeService) ListVolumes(p *ListVolumesParams) (*ListVolumesResponse, error) {
	resp, err := s.cs.newRequest("listVolumes", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListVolumesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListVolumesResponse struct {
	Count   int       `json:"count"`
	Volumes []*Volume `json:"volume"`
}

type Volume struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

type MigrateVolumeParams struct {
	p map[string]interface{}
}

func (p *MigrateVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["livemigrate"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("livemigrate", vv)
	}
	if v, found := p.p["newdiskofferingid"]; found {
		u.Set("newdiskofferingid", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	if v, found := p.p["volumeid"]; found {
		u.Set("volumeid", v.(string))
	}
	return u
}

func (p *MigrateVolumeParams) SetLivemigrate(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["livemigrate"] = v
	return
}

func (p *MigrateVolumeParams) SetNewdiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["newdiskofferingid"] = v
	return
}

func (p *MigrateVolumeParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

func (p *MigrateVolumeParams) SetVolumeid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["volumeid"] = v
	return
}

// You should always use this function to get a new MigrateVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewMigrateVolumeParams(storageid string, volumeid string) *MigrateVolumeParams {
	p := &MigrateVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["storageid"] = storageid
	p.p["volumeid"] = volumeid
	return p
}

// Migrate volume
func (s *VolumeService) MigrateVolume(p *MigrateVolumeParams) (*MigrateVolumeResponse, error) {
	resp, err := s.cs.newRequest("migrateVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r MigrateVolumeResponse
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

type MigrateVolumeResponse struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

type ResizeVolumeParams struct {
	p map[string]interface{}
}

func (p *ResizeVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["maxiops"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("maxiops", vv)
	}
	if v, found := p.p["miniops"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("miniops", vv)
	}
	if v, found := p.p["shrinkok"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("shrinkok", vv)
	}
	if v, found := p.p["size"]; found {
		vv := strconv.FormatInt(v.(int64), 10)
		u.Set("size", vv)
	}
	return u
}

func (p *ResizeVolumeParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *ResizeVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ResizeVolumeParams) SetMaxiops(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["maxiops"] = v
	return
}

func (p *ResizeVolumeParams) SetMiniops(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["miniops"] = v
	return
}

func (p *ResizeVolumeParams) SetShrinkok(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["shrinkok"] = v
	return
}

func (p *ResizeVolumeParams) SetSize(v int64) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["size"] = v
	return
}

// You should always use this function to get a new ResizeVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewResizeVolumeParams(id string) *ResizeVolumeParams {
	p := &ResizeVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Resizes a volume
func (s *VolumeService) ResizeVolume(p *ResizeVolumeParams) (*ResizeVolumeResponse, error) {
	resp, err := s.cs.newRequest("resizeVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ResizeVolumeResponse
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

type ResizeVolumeResponse struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

type UpdateVolumeParams struct {
	p map[string]interface{}
}

func (p *UpdateVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["chaininfo"]; found {
		u.Set("chaininfo", v.(string))
	}
	if v, found := p.p["customid"]; found {
		u.Set("customid", v.(string))
	}
	if v, found := p.p["displayvolume"]; found {
		vv := strconv.FormatBool(v.(bool))
		u.Set("displayvolume", vv)
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["path"]; found {
		u.Set("path", v.(string))
	}
	if v, found := p.p["state"]; found {
		u.Set("state", v.(string))
	}
	if v, found := p.p["storageid"]; found {
		u.Set("storageid", v.(string))
	}
	return u
}

func (p *UpdateVolumeParams) SetChaininfo(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["chaininfo"] = v
	return
}

func (p *UpdateVolumeParams) SetCustomid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["customid"] = v
	return
}

func (p *UpdateVolumeParams) SetDisplayvolume(v bool) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["displayvolume"] = v
	return
}

func (p *UpdateVolumeParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateVolumeParams) SetPath(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["path"] = v
	return
}

func (p *UpdateVolumeParams) SetState(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["state"] = v
	return
}

func (p *UpdateVolumeParams) SetStorageid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["storageid"] = v
	return
}

// You should always use this function to get a new UpdateVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewUpdateVolumeParams() *UpdateVolumeParams {
	p := &UpdateVolumeParams{}
	p.p = make(map[string]interface{})
	return p
}

// Updates the volume.
func (s *VolumeService) UpdateVolume(p *UpdateVolumeParams) (*UpdateVolumeResponse, error) {
	resp, err := s.cs.newRequest("updateVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateVolumeResponse
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

type UpdateVolumeResponse struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

type UploadVolumeParams struct {
	p map[string]interface{}
}

func (p *UploadVolumeParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["account"]; found {
		u.Set("account", v.(string))
	}
	if v, found := p.p["checksum"]; found {
		u.Set("checksum", v.(string))
	}
	if v, found := p.p["diskofferingid"]; found {
		u.Set("diskofferingid", v.(string))
	}
	if v, found := p.p["domainid"]; found {
		u.Set("domainid", v.(string))
	}
	if v, found := p.p["format"]; found {
		u.Set("format", v.(string))
	}
	if v, found := p.p["imagestoreuuid"]; found {
		u.Set("imagestoreuuid", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["projectid"]; found {
		u.Set("projectid", v.(string))
	}
	if v, found := p.p["url"]; found {
		u.Set("url", v.(string))
	}
	if v, found := p.p["zoneid"]; found {
		u.Set("zoneid", v.(string))
	}
	return u
}

func (p *UploadVolumeParams) SetAccount(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["account"] = v
	return
}

func (p *UploadVolumeParams) SetChecksum(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["checksum"] = v
	return
}

func (p *UploadVolumeParams) SetDiskofferingid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["diskofferingid"] = v
	return
}

func (p *UploadVolumeParams) SetDomainid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["domainid"] = v
	return
}

func (p *UploadVolumeParams) SetFormat(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["format"] = v
	return
}

func (p *UploadVolumeParams) SetImagestoreuuid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["imagestoreuuid"] = v
	return
}

func (p *UploadVolumeParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UploadVolumeParams) SetProjectid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["projectid"] = v
	return
}

func (p *UploadVolumeParams) SetUrl(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["url"] = v
	return
}

func (p *UploadVolumeParams) SetZoneid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["zoneid"] = v
	return
}

// You should always use this function to get a new UploadVolumeParams instance,
// as then you are sure you have configured all required params
func (s *VolumeService) NewUploadVolumeParams(format string, name string, url string, zoneid string) *UploadVolumeParams {
	p := &UploadVolumeParams{}
	p.p = make(map[string]interface{})
	p.p["format"] = format
	p.p["name"] = name
	p.p["url"] = url
	p.p["zoneid"] = zoneid
	return p
}

// Uploads a data disk.
func (s *VolumeService) UploadVolume(p *UploadVolumeParams) (*UploadVolumeResponse, error) {
	resp, err := s.cs.newRequest("uploadVolume", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UploadVolumeResponse
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

type UploadVolumeResponse struct {
	Account                    string `json:"account"`
	Attached                   string `json:"attached"`
	Chaininfo                  string `json:"chaininfo"`
	Clusterid                  string `json:"clusterid"`
	Clustername                string `json:"clustername"`
	Created                    string `json:"created"`
	Destroyed                  bool   `json:"destroyed"`
	Deviceid                   int64  `json:"deviceid"`
	DiskBytesReadRate          int64  `json:"diskBytesReadRate"`
	DiskBytesWriteRate         int64  `json:"diskBytesWriteRate"`
	DiskIopsReadRate           int64  `json:"diskIopsReadRate"`
	DiskIopsWriteRate          int64  `json:"diskIopsWriteRate"`
	Diskofferingdisplaytext    string `json:"diskofferingdisplaytext"`
	Diskofferingid             string `json:"diskofferingid"`
	Diskofferingname           string `json:"diskofferingname"`
	Displayvolume              bool   `json:"displayvolume"`
	Domain                     string `json:"domain"`
	Domainid                   string `json:"domainid"`
	Hypervisor                 string `json:"hypervisor"`
	Id                         string `json:"id"`
	Isextractable              bool   `json:"isextractable"`
	Isodisplaytext             string `json:"isodisplaytext"`
	Isoid                      string `json:"isoid"`
	Isoname                    string `json:"isoname"`
	JobID                      string `json:"jobid"`
	Jobstatus                  int    `json:"jobstatus"`
	Maxiops                    int64  `json:"maxiops"`
	Miniops                    int64  `json:"miniops"`
	Name                       string `json:"name"`
	Path                       string `json:"path"`
	Physicalsize               int64  `json:"physicalsize"`
	Podid                      string `json:"podid"`
	Podname                    string `json:"podname"`
	Project                    string `json:"project"`
	Projectid                  string `json:"projectid"`
	Provisioningtype           string `json:"provisioningtype"`
	Quiescevm                  bool   `json:"quiescevm"`
	Serviceofferingdisplaytext string `json:"serviceofferingdisplaytext"`
	Serviceofferingid          string `json:"serviceofferingid"`
	Serviceofferingname        string `json:"serviceofferingname"`
	Size                       int64  `json:"size"`
	Snapshotid                 string `json:"snapshotid"`
	State                      string `json:"state"`
	Status                     string `json:"status"`
	Storage                    string `json:"storage"`
	Storageid                  string `json:"storageid"`
	Storagetype                string `json:"storagetype"`
	Tags                       []Tags `json:"tags"`
	Templatedisplaytext        string `json:"templatedisplaytext"`
	Templateid                 string `json:"templateid"`
	Templatename               string `json:"templatename"`
	Type                       string `json:"type"`
	Utilization                string `json:"utilization"`
	Virtualmachineid           string `json:"virtualmachineid"`
	Virtualsize                int64  `json:"virtualsize"`
	Vmdisplayname              string `json:"vmdisplayname"`
	Vmname                     string `json:"vmname"`
	Vmstate                    string `json:"vmstate"`
	Zoneid                     string `json:"zoneid"`
	Zonename                   string `json:"zonename"`
}

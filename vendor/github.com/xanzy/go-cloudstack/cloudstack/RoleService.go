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

type CreateRoleParams struct {
	p map[string]interface{}
}

func (p *CreateRoleParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["description"]; found {
		u.Set("description", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	return u
}

func (p *CreateRoleParams) SetDescription(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["description"] = v
	return
}

func (p *CreateRoleParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *CreateRoleParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

// You should always use this function to get a new CreateRoleParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewCreateRoleParams(name string, roleType string) *CreateRoleParams {
	p := &CreateRoleParams{}
	p.p = make(map[string]interface{})
	p.p["name"] = name
	p.p["type"] = roleType
	return p
}

// Creates a role
func (s *RoleService) CreateRole(p *CreateRoleParams) (*CreateRoleResponse, error) {
	resp, err := s.cs.newRequest("createRole", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateRoleResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type CreateRoleResponse struct {
	Description string `json:"description"`
	Id          string `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type CreateRolePermissionParams struct {
	p map[string]interface{}
}

func (p *CreateRolePermissionParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["description"]; found {
		u.Set("description", v.(string))
	}
	if v, found := p.p["permission"]; found {
		u.Set("permission", v.(string))
	}
	if v, found := p.p["roleid"]; found {
		u.Set("roleid", v.(string))
	}
	if v, found := p.p["rule"]; found {
		u.Set("rule", v.(string))
	}
	return u
}

func (p *CreateRolePermissionParams) SetDescription(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["description"] = v
	return
}

func (p *CreateRolePermissionParams) SetPermission(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["permission"] = v
	return
}

func (p *CreateRolePermissionParams) SetRoleid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["roleid"] = v
	return
}

func (p *CreateRolePermissionParams) SetRule(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["rule"] = v
	return
}

// You should always use this function to get a new CreateRolePermissionParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewCreateRolePermissionParams(permission string, roleid string, rule string) *CreateRolePermissionParams {
	p := &CreateRolePermissionParams{}
	p.p = make(map[string]interface{})
	p.p["permission"] = permission
	p.p["roleid"] = roleid
	p.p["rule"] = rule
	return p
}

// Adds a API permission to a role
func (s *RoleService) CreateRolePermission(p *CreateRolePermissionParams) (*CreateRolePermissionResponse, error) {
	resp, err := s.cs.newRequest("createRolePermission", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r CreateRolePermissionResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type CreateRolePermissionResponse struct {
	Description string `json:"description"`
	Id          string `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Permission  string `json:"permission"`
	Roleid      string `json:"roleid"`
	Rolename    string `json:"rolename"`
	Rule        string `json:"rule"`
}

type DeleteRoleParams struct {
	p map[string]interface{}
}

func (p *DeleteRoleParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteRoleParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteRoleParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewDeleteRoleParams(id string) *DeleteRoleParams {
	p := &DeleteRoleParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a role
func (s *RoleService) DeleteRole(p *DeleteRoleParams) (*DeleteRoleResponse, error) {
	resp, err := s.cs.newRequest("deleteRole", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteRoleResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DeleteRoleResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *DeleteRoleResponse) UnmarshalJSON(b []byte) error {
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

	type alias DeleteRoleResponse
	return json.Unmarshal(b, (*alias)(r))
}

type DeleteRolePermissionParams struct {
	p map[string]interface{}
}

func (p *DeleteRolePermissionParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	return u
}

func (p *DeleteRolePermissionParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

// You should always use this function to get a new DeleteRolePermissionParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewDeleteRolePermissionParams(id string) *DeleteRolePermissionParams {
	p := &DeleteRolePermissionParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Deletes a role permission
func (s *RoleService) DeleteRolePermission(p *DeleteRolePermissionParams) (*DeleteRolePermissionResponse, error) {
	resp, err := s.cs.newRequest("deleteRolePermission", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r DeleteRolePermissionResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type DeleteRolePermissionResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *DeleteRolePermissionResponse) UnmarshalJSON(b []byte) error {
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

	type alias DeleteRolePermissionResponse
	return json.Unmarshal(b, (*alias)(r))
}

type ListRolePermissionsParams struct {
	p map[string]interface{}
}

func (p *ListRolePermissionsParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["roleid"]; found {
		u.Set("roleid", v.(string))
	}
	return u
}

func (p *ListRolePermissionsParams) SetRoleid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["roleid"] = v
	return
}

// You should always use this function to get a new ListRolePermissionsParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewListRolePermissionsParams() *ListRolePermissionsParams {
	p := &ListRolePermissionsParams{}
	p.p = make(map[string]interface{})
	return p
}

// Lists role permissions
func (s *RoleService) ListRolePermissions(p *ListRolePermissionsParams) (*ListRolePermissionsResponse, error) {
	resp, err := s.cs.newRequest("listRolePermissions", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListRolePermissionsResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListRolePermissionsResponse struct {
	Count           int               `json:"count"`
	RolePermissions []*RolePermission `json:"rolepermission"`
}

type RolePermission struct {
	Description string `json:"description"`
	Id          string `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Permission  string `json:"permission"`
	Roleid      string `json:"roleid"`
	Rolename    string `json:"rolename"`
	Rule        string `json:"rule"`
}

type ListRolesParams struct {
	p map[string]interface{}
}

func (p *ListRolesParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	return u
}

func (p *ListRolesParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *ListRolesParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *ListRolesParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

// You should always use this function to get a new ListRolesParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewListRolesParams() *ListRolesParams {
	p := &ListRolesParams{}
	p.p = make(map[string]interface{})
	return p
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *RoleService) GetRoleID(name string, opts ...OptionFunc) (string, int, error) {
	p := &ListRolesParams{}
	p.p = make(map[string]interface{})

	p.p["name"] = name

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return "", -1, err
		}
	}

	l, err := s.ListRoles(p)
	if err != nil {
		return "", -1, err
	}

	if l.Count == 0 {
		return "", l.Count, fmt.Errorf("No match found for %s: %+v", name, l)
	}

	if l.Count == 1 {
		return l.Roles[0].Id, l.Count, nil
	}

	if l.Count > 1 {
		for _, v := range l.Roles {
			if v.Name == name {
				return v.Id, l.Count, nil
			}
		}
	}
	return "", l.Count, fmt.Errorf("Could not find an exact match for %s: %+v", name, l)
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *RoleService) GetRoleByName(name string, opts ...OptionFunc) (*Role, int, error) {
	id, count, err := s.GetRoleID(name, opts...)
	if err != nil {
		return nil, count, err
	}

	r, count, err := s.GetRoleByID(id, opts...)
	if err != nil {
		return nil, count, err
	}
	return r, count, nil
}

// This is a courtesy helper function, which in some cases may not work as expected!
func (s *RoleService) GetRoleByID(id string, opts ...OptionFunc) (*Role, int, error) {
	p := &ListRolesParams{}
	p.p = make(map[string]interface{})

	p.p["id"] = id

	for _, fn := range append(s.cs.options, opts...) {
		if err := fn(s.cs, p); err != nil {
			return nil, -1, err
		}
	}

	l, err := s.ListRoles(p)
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
		return l.Roles[0], l.Count, nil
	}
	return nil, l.Count, fmt.Errorf("There is more then one result for Role UUID: %s!", id)
}

// Lists dynamic roles in CloudStack
func (s *RoleService) ListRoles(p *ListRolesParams) (*ListRolesResponse, error) {
	resp, err := s.cs.newRequest("listRoles", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r ListRolesResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type ListRolesResponse struct {
	Count int     `json:"count"`
	Roles []*Role `json:"role"`
}

type Role struct {
	Description string `json:"description"`
	Id          string `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type UpdateRoleParams struct {
	p map[string]interface{}
}

func (p *UpdateRoleParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["description"]; found {
		u.Set("description", v.(string))
	}
	if v, found := p.p["id"]; found {
		u.Set("id", v.(string))
	}
	if v, found := p.p["name"]; found {
		u.Set("name", v.(string))
	}
	if v, found := p.p["type"]; found {
		u.Set("type", v.(string))
	}
	return u
}

func (p *UpdateRoleParams) SetDescription(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["description"] = v
	return
}

func (p *UpdateRoleParams) SetId(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["id"] = v
	return
}

func (p *UpdateRoleParams) SetName(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["name"] = v
	return
}

func (p *UpdateRoleParams) SetType(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["type"] = v
	return
}

// You should always use this function to get a new UpdateRoleParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewUpdateRoleParams(id string) *UpdateRoleParams {
	p := &UpdateRoleParams{}
	p.p = make(map[string]interface{})
	p.p["id"] = id
	return p
}

// Updates a role
func (s *RoleService) UpdateRole(p *UpdateRoleParams) (*UpdateRoleResponse, error) {
	resp, err := s.cs.newRequest("updateRole", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateRoleResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type UpdateRoleResponse struct {
	Description string `json:"description"`
	Id          string `json:"id"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Name        string `json:"name"`
	Type        string `json:"type"`
}

type UpdateRolePermissionParams struct {
	p map[string]interface{}
}

func (p *UpdateRolePermissionParams) toURLValues() url.Values {
	u := url.Values{}
	if p.p == nil {
		return u
	}
	if v, found := p.p["permission"]; found {
		u.Set("permission", v.(string))
	}
	if v, found := p.p["roleid"]; found {
		u.Set("roleid", v.(string))
	}
	if v, found := p.p["ruleid"]; found {
		u.Set("ruleid", v.(string))
	}
	if v, found := p.p["ruleorder"]; found {
		vv := strings.Join(v.([]string), ",")
		u.Set("ruleorder", vv)
	}
	return u
}

func (p *UpdateRolePermissionParams) SetPermission(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["permission"] = v
	return
}

func (p *UpdateRolePermissionParams) SetRoleid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["roleid"] = v
	return
}

func (p *UpdateRolePermissionParams) SetRuleid(v string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ruleid"] = v
	return
}

func (p *UpdateRolePermissionParams) SetRuleorder(v []string) {
	if p.p == nil {
		p.p = make(map[string]interface{})
	}
	p.p["ruleorder"] = v
	return
}

// You should always use this function to get a new UpdateRolePermissionParams instance,
// as then you are sure you have configured all required params
func (s *RoleService) NewUpdateRolePermissionParams(roleid string) *UpdateRolePermissionParams {
	p := &UpdateRolePermissionParams{}
	p.p = make(map[string]interface{})
	p.p["roleid"] = roleid
	return p
}

// Updates a role permission order
func (s *RoleService) UpdateRolePermission(p *UpdateRolePermissionParams) (*UpdateRolePermissionResponse, error) {
	resp, err := s.cs.newRequest("updateRolePermission", p.toURLValues())
	if err != nil {
		return nil, err
	}

	var r UpdateRolePermissionResponse
	if err := json.Unmarshal(resp, &r); err != nil {
		return nil, err
	}

	return &r, nil
}

type UpdateRolePermissionResponse struct {
	Displaytext string `json:"displaytext"`
	JobID       string `json:"jobid"`
	Jobstatus   int    `json:"jobstatus"`
	Success     bool   `json:"success"`
}

func (r *UpdateRolePermissionResponse) UnmarshalJSON(b []byte) error {
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

	type alias UpdateRolePermissionResponse
	return json.Unmarshal(b, (*alias)(r))
}

package profitbricks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
)

type Groups struct {
	Id         string       `json:"id,omitempty"`
	Type_      string       `json:"type,omitempty"`
	Href       string       `json:"href,omitempty"`
	Items      []Group      `json:"items,omitempty"`
	Response   string       `json:"Response,omitempty"`
	Headers    *http.Header `json:"headers,omitempty"`
	StatusCode int          `json:"headers,omitempty"`
}

type Group struct {
	Id         string          `json:"id,omitempty"`
	Type_      string          `json:"type,omitempty"`
	Href       string          `json:"href,omitempty"`
	Properties GroupProperties `json:"properties,omitempty"`
	Entities   *GroupEntities  `json:"entities,omitempty"`
	Response   string          `json:"Response,omitempty"`
	Headers    *http.Header    `json:"headers,omitempty"`
	StatusCode int             `json:"headers,omitempty"`
}

type GroupProperties struct {
	Name              string `json:"name,omitempty"`
	CreateDataCenter  *bool  `json:"createDataCenter,omitempty"`
	CreateSnapshot    *bool  `json:"createSnapshot,omitempty"`
	ReserveIp         *bool  `json:"reserveIp,omitempty"`
	AccessActivityLog *bool  `json:"accessActivityLog,omitempty"`
}

type GroupEntities struct {
	Users     Users     `json:"users,omitempty"`
	Resources Resources `json:"resources,omitempty"`
}

type Users struct {
	Id         string       `json:"id,omitempty"`
	Type_      string       `json:"type,omitempty"`
	Href       string       `json:"href,omitempty"`
	Items      []User       `json:"items,omitempty"`
	Response   string       `json:"Response,omitempty"`
	Headers    *http.Header `json:"headers,omitempty"`
	StatusCode int          `json:"headers,omitempty"`
}

type User struct {
	Id         string          `json:"id,omitempty"`
	Type_      string          `json:"type,omitempty"`
	Href       string          `json:"href,omitempty"`
	Metadata   *Metadata       `json:"metadata,omitempty"`
	Properties *UserProperties `json:"properties,omitempty"`
	Entities   *UserEntities   `json:"entities,omitempty"`
	Response   string          `json:"Response,omitempty"`
	Headers    *http.Header    `json:"headers,omitempty"`
	StatusCode int             `json:"headers,omitempty"`
}

type UserProperties struct {
	Firstname     string `json:"firstname,omitempty"`
	Lastname      string `json:"lastname,omitempty"`
	Email         string `json:"email,omitempty"`
	Password      string `json:"password,omitempty"`
	Administrator bool   `json:"administrator,omitempty"`
	ForceSecAuth  bool   `json:"forceSecAuth,omitempty"`
	SecAuthActive bool   `json:"secAuthActive,omitempty"`
}

type UserEntities struct {
	Groups Groups `json:"groups,omitempty"`
	Owns   Owns   `json:"owns,omitempty"`
}

type Resources struct {
	Id         string       `json:"id,omitempty"`
	Type_      string       `json:"type,omitempty"`
	Href       string       `json:"href,omitempty"`
	Items      []Resource   `json:"items,omitempty"`
	Response   string       `json:"Response,omitempty"`
	Headers    *http.Header `json:"headers,omitempty"`
	StatusCode int          `json:"headers,omitempty"`
}

type Resource struct {
	Id         string            `json:"id,omitempty"`
	Type_      string            `json:"type,omitempty"`
	Href       string            `json:"href,omitempty"`
	Metadata   *Metadata         `json:"metadata,omitempty"`
	Entities   *ResourceEntities `json:"entities,omitempty"`
	Response   string            `json:"Response,omitempty"`
	Headers    *http.Header      `json:"headers,omitempty"`
	StatusCode int               `json:"headers,omitempty"`
}

type ResourceEntities struct {
	Groups Groups `json:"groups,omitempty"`
}

type Owns struct {
	Id         string       `json:"id,omitempty"`
	Type_      string       `json:"type,omitempty"`
	Href       string       `json:"href,omitempty"`
	Items      []Entity     `json:"items,omitempty"`
	Response   string       `json:"Response,omitempty"`
	Headers    *http.Header `json:"headers,omitempty"`
	StatusCode int          `json:"headers,omitempty"`
}

type Entity struct {
	Id         string       `json:"id,omitempty"`
	Type_      string       `json:"type,omitempty"`
	Href       string       `json:"href,omitempty"`
	Metadata   *Metadata    `json:"metadata,omitempty"`
	Response   string       `json:"Response,omitempty"`
	Headers    *http.Header `json:"headers,omitempty"`
	StatusCode int          `json:"headers,omitempty"`
}

type Shares struct {
	Id         string       `json:"id,omitempty"`
	Type_      string       `json:"type,omitempty"`
	Href       string       `json:"href,omitempty"`
	Items      []Share      `json:"items,omitempty"`
	Response   string       `json:"Response,omitempty"`
	Headers    *http.Header `json:"headers,omitempty"`
	StatusCode int          `json:"headers,omitempty"`
}

type Share struct {
	Id         string          `json:"id,omitempty"`
	Type_      string          `json:"type,omitempty"`
	Href       string          `json:"href,omitempty"`
	Properties ShareProperties `json:"properties,omitempty"`
	Response   string          `json:"Response,omitempty"`
	Headers    *http.Header    `json:"headers,omitempty"`
	StatusCode int             `json:"headers,omitempty"`
}

type ShareProperties struct {
	EditPrivilege  *bool `json:"editPrivilege,omitempty"`
	SharePrivilege *bool `json:"sharePrivilege,omitempty"`
}

//Group fucntions
func ListGroups() Groups {
	path := um_groups()
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toGroups(resp)
}

func GetGroup(groupid string) Group {
	path := um_group_path(groupid)
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	return toGroup(do(req))
}

func CreateGroup(grp Group) Group {
	obj, _ := json.Marshal(grp)
	path := um_groups()
	url := mk_url(path)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(obj))
	req.Header.Add("Content-Type", FullHeader)

	return toGroup(do(req))
}

func UpdateGroup(groupid string, obj Group) Group {
	jason_patch := []byte(MkJson(obj))
	path := um_group_path(groupid)
	url := mk_url(path) + `?depth=` + Depth
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jason_patch))
	req.Header.Add("Content-Type", PatchHeader)
	return toGroup(do(req))
}

func DeleteGroup(groupid string) Resp {
	path := um_group_path(groupid)
	return is_delete(path)
}

func toGroup(resp Resp) Group {
	var grp Group
	json.Unmarshal(resp.Body, &grp)
	grp.Response = string(resp.Body)
	grp.Headers = &resp.Headers
	grp.StatusCode = resp.StatusCode
	return grp
}

func toGroups(resp Resp) Groups {
	var col Groups
	json.Unmarshal(resp.Body, &col)
	col.Response = string(resp.Body)
	col.Headers = &resp.Headers
	col.StatusCode = resp.StatusCode
	return col
}

//Shares functions
func ListShares(grpid string) Shares {
	path := um_group_shares(grpid)
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toShares(resp)
}

func GetShare(groupid string, resourceid string) Share {
	path := um_group_share_path(groupid, resourceid)
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	return toShare(do(req))
}

func AddShare(share Share, groupid string, resourceid string) Share {
	obj, _ := json.Marshal(share)
	path := um_group_share_path(groupid, resourceid)
	url := mk_url(path)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(obj))
	req.Header.Add("Content-Type", FullHeader)

	return toShare(do(req))
}

func UpdateShare(groupid string, resourceid string, obj Share) Share {
	jason_patch := []byte(MkJson(obj))
	path := um_group_share_path(groupid, resourceid)
	url := mk_url(path) + `?depth=` + Depth
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jason_patch))
	req.Header.Add("Content-Type", PatchHeader)
	return toShare(do(req))
}

func DeleteShare(groupid string, resourceid string) Resp {
	path := um_group_share_path(groupid, resourceid)
	return is_delete(path)
}

func toShare(resp Resp) Share {
	var shr Share
	json.Unmarshal(resp.Body, &shr)
	shr.Response = string(resp.Body)
	shr.Headers = &resp.Headers
	shr.StatusCode = resp.StatusCode
	return shr
}

func toShares(resp Resp) Shares {
	var col Shares
	json.Unmarshal(resp.Body, &col)
	col.Response = string(resp.Body)
	col.Headers = &resp.Headers
	col.StatusCode = resp.StatusCode
	return col
}

//Users in a group
func ListGroupUsers(groupid string) Users {
	path := um_group_users(groupid)
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toUsers(resp)
}

func AddUserToGroup(groupid string, userid string) User {
	var usr User
	usr.Id = userid
	obj, _ := json.Marshal(usr)
	path := um_group_users(groupid)
	url := mk_url(path)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(obj))
	req.Header.Add("Content-Type", FullHeader)

	return toUser(do(req))
}

func DeleteUserFromGroup(groupid string, userid string) Resp {
	path := um_group_users_path(groupid, userid)
	return is_delete(path)
}

func toUser(resp Resp) User {
	var usr User
	json.Unmarshal(resp.Body, &usr)
	usr.Response = string(resp.Body)
	usr.Headers = &resp.Headers
	usr.StatusCode = resp.StatusCode
	return usr
}

func toUsers(resp Resp) Users {
	var col Users
	json.Unmarshal(resp.Body, &col)
	col.Response = string(resp.Body)
	col.Headers = &resp.Headers
	col.StatusCode = resp.StatusCode
	return col
}

//Users
func ListUsers() Users {
	path := um_users()
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toUsers(resp)
}

func GetUser(usrid string) User {
	path := um_users_path(usrid)
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	return toUser(do(req))
}

func CreateUser(usr User) User {
	obj, _ := json.Marshal(usr)
	path := um_users()
	url := mk_url(path)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(obj))
	req.Header.Add("Content-Type", FullHeader)

	return toUser(do(req))
}

func UpdateUser(userid string, obj User) User {
	jason_patch := []byte(MkJson(obj))
	path := um_users_path(userid)
	url := mk_url(path)
	req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jason_patch))
	req.Header.Add("Content-Type", PatchHeader)
	return toUser(do(req))
}

func DeleteUser(groupid string) Resp {
	path := um_users_path(groupid)
	return is_delete(path)
}

//Resources
func ListResources() Resources {
	path := um_resources()
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toResources(resp)
}

func GetResourceByType(resourcetype string, resourceid string) Resource {
	path := um_resources_type_path(resourcetype, resourceid)
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toResource(resp)
}

func ListResourcesByType(resourcetype string) Resources {
	path := um_resources_type(resourcetype)
	url := mk_url(path) + `?depth=` + Depth + `&pretty=` + strconv.FormatBool(Pretty)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Content-Type", FullHeader)
	resp := do(req)
	return toResources(resp)
}

func toResources(resp Resp) Resources {
	var col Resources
	json.Unmarshal(resp.Body, &col)
	col.Response = string(resp.Body)
	col.Headers = &resp.Headers
	col.StatusCode = resp.StatusCode
	return col
}

func toResource(resp Resp) Resource {
	var col Resource
	json.Unmarshal(resp.Body, &col)
	col.Response = string(resp.Body)
	col.Headers = &resp.Headers
	col.StatusCode = resp.StatusCode
	return col
}

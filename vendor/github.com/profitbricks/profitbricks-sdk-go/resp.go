package profitbricks

import "net/http"
import "fmt"
import (
	"encoding/json"
	"github.com/profitbricks/profitbricks-sdk-go/model"
)

func MkJson(i interface{}) string {
	jason, err := json.MarshalIndent(&i, "", "    ")
	if err != nil {
		panic(err)
	}
	//	fmt.Println(string(jason))
	return string(jason)
}

// MetaData is a map for metadata returned in a Resp.Body
type StringMap map[string]string

type StringIfaceMap map[string]interface{}

type StringCollectionMap map[string]Collection

// Resp is the struct returned by all Rest request functions
type Resp struct {
	Req        *http.Request
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// PrintHeaders prints the http headers as k,v pairs
func (r *Resp) PrintHeaders() {
	for key, value := range r.Headers {
		fmt.Println(key, " : ", value[0])
	}

}

type Id_Type_Href struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Href string `json:"href"`
}

type MetaData StringIfaceMap

// toInstance converts a Resp struct into a Instance struct
func toInstance(resp Resp) Instance {
	var ins Instance
	json.Unmarshal(resp.Body, &ins)
	ins.Resp = resp
	return ins
}

func toDataCenter(resp Resp) model.Datacenter {
	var dc model.Datacenter
	json.Unmarshal(resp.Body, &dc)
	dc.Response = string(resp.Body)
	dc.Headers = &resp.Headers
	dc.StatusCode = resp.StatusCode
	return dc
}

type Instance struct {
	Id_Type_Href
	MetaData   StringMap           `json:"metaData,omitempty"`
	Properties StringIfaceMap      `json:"properties,omitempty"`
	Entities   StringCollectionMap `json:"entities,omitempty"`
	Resp       Resp                `json:"-"`
}

// Save converts the Instance struct's properties to json
// and "patch"es them to the rest server
func (ins *Instance) Save() {
	path := ins.Href
	jason, err := json.MarshalIndent(&ins.Properties, "", "    ")
	if err != nil {
		panic(err)
	}
	r := is_patch(path, jason).Resp
	fmt.Println("save status code is ", r.StatusCode)
}

// ShowProps prints the properties as k,v pairs
func (ins *Instance) ShowProps() {
	for key, value := range ins.Properties {
		fmt.Println(key, " : ", value)
	}
}
func (ins *Instance) SetProp(key, val string) {
	ins.Properties[key] = val
}

// ShowEnts prints the Entities  as k,v pairs
func (ins *Instance) ShowEnts() {
	for key, value := range ins.Entities {
		v := MkJson(value)
		fmt.Println(key, " : ", v)
	}
}

// toServers converts a Resp struct into a Collection struct
func toCollection(resp Resp) Collection {
	var col Collection
	json.Unmarshal(resp.Body, &col)
	col.Resp = resp
	return col
}

type Collection struct {
	Id_Type_Href
	Items []Instance `json:"items,omitempty"`
	Resp  Resp       `json:"-"`
}

package model

import "net/http"

type Datacenter struct {
	Id         string  `json:"id,omitempty"`
	Type_      string  `json:"type,omitempty"`
	Href       string  `json:"href,omitempty"`
	Properties DatacenterProperties  `json:"properties,omitempty"`
	Entities   DatacenterEntities  `json:"entities,omitempty"`
	Response   string `json:"Response,omitempty"`
	Headers    *http.Header `json:"headers,omitempty"`
	StatusCode int `json:"headers,omitempty"`
}

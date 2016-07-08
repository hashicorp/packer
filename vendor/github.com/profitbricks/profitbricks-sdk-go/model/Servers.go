package model

type Servers struct {
	Id    string  `json:"id,omitempty"`
	Type_ string  `json:"type,omitempty"`
	Href  string  `json:"href,omitempty"`
	Items []Server  `json:"items,omitempty"`
}

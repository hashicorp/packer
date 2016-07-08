package model

type Volumes struct {
	Id    string  `json:"id,omitempty"`
	Type_ string  `json:"type,omitempty"`
	Href  string  `json:"href,omitempty"`
	Items []Volume  `json:"items,omitempty"`
}

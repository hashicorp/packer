package model



type LanNics struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Items  []Nic  `json:"items,omitempty"`
    
}

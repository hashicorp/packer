package model



type Cdroms struct {
    Id  string  `json:"id,omitempty"`
    Type_  string  `json:"type,omitempty"`
    Href  string  `json:"href,omitempty"`
    Items  []Image  `json:"items,omitempty"`
    
}

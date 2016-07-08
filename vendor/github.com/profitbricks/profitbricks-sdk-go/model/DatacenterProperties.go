package model



type DatacenterProperties struct {
    Name  string  `json:"name,omitempty"`
    Description  string  `json:"description,omitempty"`
    Location  string  `json:"location,omitempty"`
    Version  int32  `json:"version,omitempty"`
    
}

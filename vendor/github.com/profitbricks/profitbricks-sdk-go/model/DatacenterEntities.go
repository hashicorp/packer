package model



type DatacenterEntities struct {
    Servers  *Servers  `json:"servers,omitempty"`
    Volumes  *Volumes  `json:"volumes,omitempty"`
    Loadbalancers  *Loadbalancers  `json:"loadbalancers,omitempty"`
    Lans  *Lans  `json:"lans,omitempty"`
    
}

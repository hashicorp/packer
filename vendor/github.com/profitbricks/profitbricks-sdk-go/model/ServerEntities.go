package model

type ServerEntities struct {
	Cdroms  *Cdroms  `json:"cdroms,omitempty"`
	Volumes *AttachedVolumes  `json:"volumes,omitempty"`
	Nics    *Nics  `json:"nics,omitempty"`
}

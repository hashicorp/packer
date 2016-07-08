package model

import (
    "time"
)

type DatacenterElementMetadata struct {
    CreatedDate  time.Time  `json:"createdDate,omitempty"`
    CreatedBy  string  `json:"createdBy,omitempty"`
    Etag  string  `json:"etag,omitempty"`
    LastModifiedDate  time.Time  `json:"lastModifiedDate,omitempty"`
    LastModifiedBy  string  `json:"lastModifiedBy,omitempty"`
    State  string  `json:"state,omitempty"`
    
}

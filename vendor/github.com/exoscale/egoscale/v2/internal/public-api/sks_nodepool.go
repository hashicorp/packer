package publicapi

import (
	"encoding/json"
	"time"
)

// UnmarshalJSON unmarshals a SksNodepool structure into a temporary structure whose "CreatedAt" field of type
// string to be able to parse the original timestamp (ISO 8601) into a time.Time object, since json.Unmarshal()
// only supports RFC 3339 format.
func (n *SksNodepool) UnmarshalJSON(data []byte) error {
	raw := struct {
		AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`
		CreatedAt          *string              `json:"created-at,omitempty"`
		Description        *string              `json:"description,omitempty"`
		DiskSize           *int64               `json:"disk-size,omitempty"`
		Id                 *string              `json:"id,omitempty"` // nolint:golint
		InstancePool       *InstancePool        `json:"instance-pool,omitempty"`
		InstanceType       *InstanceType        `json:"instance-type,omitempty"`
		Name               *string              `json:"name,omitempty"`
		SecurityGroups     *[]SecurityGroup     `json:"security-groups,omitempty"`
		Size               *int64               `json:"size,omitempty"`
		State              *string              `json:"state,omitempty"`
		Template           *Template            `json:"template,omitempty"`
		Version            *string              `json:"version,omitempty"`
	}{}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.CreatedAt != nil {
		createdAt, err := time.Parse(iso8601Format, *raw.CreatedAt)
		if err != nil {
			return err
		}
		n.CreatedAt = &createdAt
	}

	n.AntiAffinityGroups = raw.AntiAffinityGroups
	n.Description = raw.Description
	n.DiskSize = raw.DiskSize
	n.Id = raw.Id
	n.InstancePool = raw.InstancePool
	n.InstanceType = raw.InstanceType
	n.Name = raw.Name
	n.SecurityGroups = raw.SecurityGroups
	n.Size = raw.Size
	n.State = raw.State
	n.Template = raw.Template
	n.Version = raw.Version

	return nil
}

// MarshalJSON returns the JSON encoding of a SksNodepool structure after having formatted the CreatedAt field
// in the original timestamp (ISO 8601), since time.MarshalJSON() only supports RFC 3339 format.
func (n *SksNodepool) MarshalJSON() ([]byte, error) {
	raw := struct {
		AntiAffinityGroups *[]AntiAffinityGroup `json:"anti-affinity-groups,omitempty"`
		CreatedAt          *string              `json:"created-at,omitempty"`
		Description        *string              `json:"description,omitempty"`
		DiskSize           *int64               `json:"disk-size,omitempty"`
		Id                 *string              `json:"id,omitempty"` // nolint:golint
		InstancePool       *InstancePool        `json:"instance-pool,omitempty"`
		InstanceType       *InstanceType        `json:"instance-type,omitempty"`
		Name               *string              `json:"name,omitempty"`
		SecurityGroups     *[]SecurityGroup     `json:"security-groups,omitempty"`
		Size               *int64               `json:"size,omitempty"`
		State              *string              `json:"state,omitempty"`
		Template           *Template            `json:"template,omitempty"`
		Version            *string              `json:"version,omitempty"`
	}{}

	if n.CreatedAt != nil {
		createdAt := n.CreatedAt.Format(iso8601Format)
		raw.CreatedAt = &createdAt
	}

	raw.AntiAffinityGroups = n.AntiAffinityGroups
	raw.Description = n.Description
	raw.DiskSize = n.DiskSize
	raw.Id = n.Id
	raw.InstancePool = n.InstancePool
	raw.InstanceType = n.InstanceType
	raw.Name = n.Name
	raw.SecurityGroups = n.SecurityGroups
	raw.Size = n.Size
	raw.State = n.State
	raw.Template = n.Template
	raw.Version = n.Version

	return json.Marshal(raw)
}

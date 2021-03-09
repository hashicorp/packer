package publicapi

import (
	"encoding/json"
	"time"
)

// UnmarshalJSON unmarshals a SksCluster structure into a temporary structure whose "CreatedAt" field of type
// string to be able to parse the original timestamp (ISO 8601) into a time.Time object, since json.Unmarshal()
// only supports RFC 3339 format.
func (c *SksCluster) UnmarshalJSON(data []byte) error {
	raw := struct {
		Addons      *[]string      `json:"addons,omitempty"`
		Cni         *string        `json:"cni,omitempty"`
		CreatedAt   *string        `json:"created-at,omitempty"`
		Description *string        `json:"description,omitempty"`
		Endpoint    *string        `json:"endpoint,omitempty"`
		Id          *string        `json:"id,omitempty"` // nolint:golint
		Level       *string        `json:"level,omitempty"`
		Name        *string        `json:"name,omitempty"`
		Nodepools   *[]SksNodepool `json:"nodepools,omitempty"`
		State       *string        `json:"state,omitempty"`
		Version     *string        `json:"version,omitempty"`
	}{}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.CreatedAt != nil {
		createdAt, err := time.Parse(iso8601Format, *raw.CreatedAt)
		if err != nil {
			return err
		}
		c.CreatedAt = &createdAt
	}

	c.Addons = raw.Addons
	c.Cni = raw.Cni
	c.Description = raw.Description
	c.Endpoint = raw.Endpoint
	c.Id = raw.Id
	c.Level = raw.Level
	c.Name = raw.Name
	c.Nodepools = raw.Nodepools
	c.State = raw.State
	c.Version = raw.Version

	return nil
}

// MarshalJSON returns the JSON encoding of a SksCluster structure after having formatted the CreatedAt field
// in the original timestamp (ISO 8601), since time.MarshalJSON() only supports RFC 3339 format.
func (c *SksCluster) MarshalJSON() ([]byte, error) {
	raw := struct {
		Addons      *[]string      `json:"addons,omitempty"`
		Cni         *string        `json:"cni,omitempty"`
		CreatedAt   *string        `json:"created-at,omitempty"`
		Description *string        `json:"description,omitempty"`
		Endpoint    *string        `json:"endpoint,omitempty"`
		Id          *string        `json:"id,omitempty"` // nolint:golint
		Level       *string        `json:"level,omitempty"`
		Name        *string        `json:"name,omitempty"`
		Nodepools   *[]SksNodepool `json:"nodepools,omitempty"`
		State       *string        `json:"state,omitempty"`
		Version     *string        `json:"version,omitempty"`
	}{}

	if c.CreatedAt != nil {
		createdAt := c.CreatedAt.Format(iso8601Format)
		raw.CreatedAt = &createdAt
	}

	raw.Addons = c.Addons
	raw.Cni = c.Cni
	raw.Description = c.Description
	raw.Endpoint = c.Endpoint
	raw.Id = c.Id
	raw.Level = c.Level
	raw.Name = c.Name
	raw.Nodepools = c.Nodepools
	raw.State = c.State
	raw.Version = c.Version

	return json.Marshal(raw)
}

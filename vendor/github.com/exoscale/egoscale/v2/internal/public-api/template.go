package publicapi

import (
	"encoding/json"
	"time"
)

// UnmarshalJSON unmarshals a Template structure into a temporary structure whose "CreatedAt" field of type
// string to be able to parse the original timestamp (ISO 8601) into a time.Time object, since json.Unmarshal()
// only supports RFC 3339 format.
func (t *Template) UnmarshalJSON(data []byte) error {
	raw := struct {
		CreatedAt       *string `json:"created-at,omitempty"`
		BootMode        *string `json:"boot-mode,omitempty"`
		Build           *string `json:"build,omitempty"`
		Checksum        *string `json:"checksum,omitempty"`
		DefaultUser     *string `json:"default-user,omitempty"`
		Description     *string `json:"description,omitempty"`
		Family          *string `json:"family,omitempty"`
		Id              *string `json:"id,omitempty"` // nolint:golint
		Name            *string `json:"name,omitempty"`
		PasswordEnabled *bool   `json:"password-enabled,omitempty"`
		Size            *int64  `json:"size,omitempty"`
		SshKeyEnabled   *bool   `json:"ssh-key-enabled,omitempty"` // nolint:golint
		Url             *string `json:"url,omitempty"`             // nolint:golint
		Version         *string `json:"version,omitempty"`
		Visibility      *string `json:"visibility,omitempty"`
	}{}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.CreatedAt != nil {
		createdAt, err := time.Parse(iso8601Format, *raw.CreatedAt)
		if err != nil {
			return err
		}
		t.CreatedAt = &createdAt
	}

	t.BootMode = raw.BootMode
	t.Build = raw.Build
	t.Checksum = raw.Checksum
	t.DefaultUser = raw.DefaultUser
	t.Description = raw.Description
	t.Family = raw.Family
	t.Id = raw.Id
	t.Name = raw.Name
	t.PasswordEnabled = raw.PasswordEnabled
	t.Size = raw.Size
	t.SshKeyEnabled = raw.SshKeyEnabled
	t.Url = raw.Url
	t.Version = raw.Version
	t.Visibility = raw.Visibility

	return nil
}

// MarshalJSON returns the JSON encoding of a Template structure after having formatted the CreatedAt field
// in the original timestamp (ISO 8601), since time.MarshalJSON() only supports RFC 3339 format.
func (t *Template) MarshalJSON() ([]byte, error) {
	raw := struct {
		CreatedAt       *string `json:"created-at,omitempty"`
		BootMode        *string `json:"boot-mode,omitempty"`
		Build           *string `json:"build,omitempty"`
		Checksum        *string `json:"checksum,omitempty"`
		DefaultUser     *string `json:"default-user,omitempty"`
		Description     *string `json:"description,omitempty"`
		Family          *string `json:"family,omitempty"`
		Id              *string `json:"id,omitempty"` // nolint:golint
		Name            *string `json:"name,omitempty"`
		PasswordEnabled *bool   `json:"password-enabled,omitempty"`
		Size            *int64  `json:"size,omitempty"`
		SshKeyEnabled   *bool   `json:"ssh-key-enabled,omitempty"` // nolint:golint
		Url             *string `json:"url,omitempty"`             // nolint:golint
		Version         *string `json:"version,omitempty"`
		Visibility      *string `json:"visibility,omitempty"`
	}{}

	if t.CreatedAt != nil {
		createdAt := t.CreatedAt.Format(iso8601Format)
		raw.CreatedAt = &createdAt
	}

	raw.BootMode = t.BootMode
	raw.Build = t.Build
	raw.Checksum = t.Checksum
	raw.DefaultUser = t.DefaultUser
	raw.Description = t.Description
	raw.Family = t.Family
	raw.Id = t.Id
	raw.Name = t.Name
	raw.PasswordEnabled = t.PasswordEnabled
	raw.Size = t.Size
	raw.SshKeyEnabled = t.SshKeyEnabled
	raw.Url = t.Url
	raw.Version = t.Version
	raw.Visibility = t.Visibility

	return json.Marshal(raw)
}

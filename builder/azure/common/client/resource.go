package client

import (
	"errors"
	"fmt"
	"strings"
)

// ParseResourceID parses an Azure resource ID
func ParseResourceID(resourceID string) (Resource, error) {
	resourceID = strings.Trim(resourceID, "/")
	segments := strings.Split(resourceID, "/")
	if len(segments)%2 != 0 {
		return Resource{}, errors.New("Expected even number of segments")
	}

	npairs := len(segments) / 2

	keys := make([]string, npairs)
	values := make([]string, npairs)
	for i := 0; i < len(segments); i += 2 {
		keys[i/2] = segments[i]
		values[i/2] = segments[i+1]

		if keys[i/2] == "" {
			return Resource{}, fmt.Errorf("Found empty segment (%d)", i)
		}
		if values[i/2] == "" {
			return Resource{}, fmt.Errorf("Found empty segment (%d)", i+1)
		}
	}

	if !strings.EqualFold(keys[0], "subscriptions") {
		return Resource{}, fmt.Errorf("Expected first segment to be 'subscriptions', but found %q", keys[0])
	}

	if !strings.EqualFold(keys[1], "resourceGroups") {
		return Resource{}, fmt.Errorf("Expected second segment to be 'resourceGroups', but found %q", keys[1])
	}

	if !strings.EqualFold(keys[2], "providers") {
		return Resource{}, fmt.Errorf("Expected third segment to be 'providers', but found %q", keys[1])
	}

	r := Resource{
		values[0],
		values[1],
		values[2],
		CompoundName(keys[3:]),
		CompoundName(values[3:]),
	}
	if err := r.Validate(); err != nil {
		return Resource{}, fmt.Errorf("Error validating resource: %w", err)
	}

	return r, nil
}

type Resource struct {
	Subscription  string
	ResourceGroup string
	Provider      string
	ResourceType  CompoundName
	ResourceName  CompoundName
}

func (r Resource) String() string {
	return fmt.Sprintf(
		"/subscriptions/%s"+
			"/resourceGroups/%s"+
			"/providers/%s"+
			"/%s",
		r.Subscription,
		r.ResourceGroup,
		r.Provider,
		strings.Join(zipstrings(r.ResourceType, r.ResourceName), "/"))
}

func (r Resource) Validate() error {
	if r.Subscription == "" {
		return errors.New("subscription is not set")
	}
	if r.ResourceGroup == "" {
		return errors.New("resource group is not set")
	}
	if r.Provider == "" {
		return errors.New("provider is not set")
	}
	if len(r.ResourceType) > len(r.ResourceName) {
		return errors.New("not enough values in resource name")
	}
	if len(r.ResourceType) < len(r.ResourceName) {
		return errors.New("too many values in resource name")
	}
	return nil
}

// Parent produces a resource ID representing the parent resource if this is a child resource
func (r Resource) Parent() (Resource, error) {
	newLen := len(r.ResourceType) - 1
	if newLen == 0 {
		return Resource{}, errors.New("Top-level resource has no parent")
	}
	return Resource{
		Subscription:  r.Subscription,
		ResourceGroup: r.ResourceGroup,
		Provider:      r.Provider,
		ResourceType:  r.ResourceType[:newLen],
		ResourceName:  r.ResourceName[:newLen],
	}, nil
}

type CompoundName []string

func (n CompoundName) String() string {
	return strings.Join(n, "/")
}

func zipstrings(a []string, b []string) []string {
	c := make([]string, 0, len(a)+len(b))
	for i := 0; i < len(a) && i < len(b); i++ {
		c = append(c, a[i], b[i])
	}
	return c
}

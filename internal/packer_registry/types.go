package packer_registry

type Bucket struct {
	Slug        string
	Description string
	Labels      map[string]string
}

type Build struct {
	ComponentType string
	RunUUID       string
	PARtifacts    []PARtifact
}

type PARtifact struct {
	ID                           string
	ProviderName, ProviderRegion string
	Metadata                     map[string]string
}

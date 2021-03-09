package egoscale

// APIKeyType holds the type of the API key
type APIKeyType string

const (
	// APIKeyTypeUnrestricted is unrestricted
	APIKeyTypeUnrestricted APIKeyType = "unrestricted"
	// APIKeyTypeRestricted is restricted
	APIKeyTypeRestricted APIKeyType = "restricted"
)

// APIKey represents an API key
type APIKey struct {
	Name       string     `json:"name"`
	Key        string     `json:"key"`
	Secret     string     `json:"secret,omitempty"`
	Operations []string   `json:"operations,omitempty"`
	Resources  []string   `json:"resources,omitempty"`
	Type       APIKeyType `json:"type"`
}

// CreateAPIKey represents an API key creation
type CreateAPIKey struct {
	Name       string `json:"name"`
	Operations string `json:"operations,omitempty"`
	Resources  string `json:"resources,omitempty"`
	_          bool   `name:"createApiKey" description:"Create an API key."`
}

// Response returns the struct to unmarshal
func (CreateAPIKey) Response() interface{} {
	return new(APIKey)
}

// ListAPIKeys represents a search for API keys
type ListAPIKeys struct {
	_ bool `name:"listApiKeys" description:"List API keys."`
}

// ListAPIKeysResponse represents a list of API keys
type ListAPIKeysResponse struct {
	Count   int      `json:"count"`
	APIKeys []APIKey `json:"apikey"`
}

// Response returns the struct to unmarshal
func (ListAPIKeys) Response() interface{} {
	return new(ListAPIKeysResponse)
}

// ListAPIKeyOperations represents a search for operations for the current API key
type ListAPIKeyOperations struct {
	_ bool `name:"listApiKeyOperations" description:"List operations allowed for the current API key."`
}

// ListAPIKeyOperationsResponse represents a list of operations for the current API key
type ListAPIKeyOperationsResponse struct {
	Operations []string `json:"operations"`
}

// Response returns the struct to unmarshal
func (ListAPIKeyOperations) Response() interface{} {
	return new(ListAPIKeyOperationsResponse)
}

// GetAPIKey get an API key
type GetAPIKey struct {
	Key string `json:"key"`
	_   bool   `name:"getApiKey" description:"Get an API key."`
}

// Response returns the struct to unmarshal
func (GetAPIKey) Response() interface{} {
	return new(APIKey)
}

// RevokeAPIKey represents a revocation of an API key
type RevokeAPIKey struct {
	Key string `json:"key"`
	_   bool   `name:"revokeApiKey" description:"Revoke an API key."`
}

// RevokeAPIKeyResponse represents the response to an API key revocation
type RevokeAPIKeyResponse struct {
	Success bool `json:"success"`
}

// Response returns the struct to unmarshal
func (RevokeAPIKey) Response() interface{} {
	return new(RevokeAPIKeyResponse)
}

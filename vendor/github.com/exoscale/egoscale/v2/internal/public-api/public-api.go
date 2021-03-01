// Package publicapi is an internal package containing code generated from the
// Exoscale API OpenAPI specs, as well as helpers and transition types exposed
// in the public-facing package.
package publicapi

//go:generate oapi-codegen -generate types,client -package publicapi -o public-api.gen.go ../../../public-api.json

// OptionalString returns the dereferenced string value of v if not nil, otherwise an empty string.
func OptionalString(v *string) string {
	if v != nil {
		return *v
	}

	return ""
}

// OptionalInt64 returns the dereferenced int64 value of v if not nil, otherwise 0.
func OptionalInt64(v *int64) int64 {
	if v != nil {
		return *v
	}

	return 0
}

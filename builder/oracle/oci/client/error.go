package oci

import "fmt"

// APIError encapsulates an error returned from the API
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("OCI: [%s] '%s'", e.Code, e.Message)
}

// firstError is a helper function to work out which error to return from calls
// to the API.
func firstError(err error, apiError *APIError) error {
	if err != nil {
		return err
	}

	if apiError != nil && len(apiError.Code) > 0 {
		return apiError
	}

	return nil
}

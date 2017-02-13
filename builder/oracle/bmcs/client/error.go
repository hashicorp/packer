// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

import "fmt"

// APIError encapsulates an error returned from the API
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("BMCS: [%s] '%s'", e.Code, e.Message)
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

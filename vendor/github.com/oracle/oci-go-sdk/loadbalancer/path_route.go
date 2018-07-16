// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Load Balancing Service API
//
// API for the Load Balancing Service
//

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
)

// PathRoute A "path route rule" to evaluate an incoming URI path, and then route a matching request to the specified backend set.
// Path route rules apply only to HTTP and HTTPS requests. They have no effect on TCP requests.
type PathRoute struct {

	// The name of the target backend set for requests where the incoming URI matches the specified path.
	// Example: `example_backend_set`
	BackendSetName *string `mandatory:"true" json:"backendSetName"`

	// The path string to match against the incoming URI path.
	// *  Path strings are case-insensitive.
	// *  Asterisk (*) wildcards are not supported.
	// *  Regular expressions are not supported.
	// Example: `/example/video/123`
	Path *string `mandatory:"true" json:"path"`

	// The type of matching to apply to incoming URIs.
	PathMatchType *PathMatchType `mandatory:"true" json:"pathMatchType"`
}

func (m PathRoute) String() string {
	return common.PointerString(m)
}

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

// UpdateHostnameDetails The configuration details for updating a virtual hostname.
// For more information on virtual hostnames, see
// Managing Request Routing (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/managingrequest.htm).
type UpdateHostnameDetails struct {

	// The virtual hostname to update. For more information about virtual hostname string construction, see
	// Managing Request Routing (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/managingrequest.htm#routing).
	// Example: `app.example.com`
	Hostname *string `mandatory:"false" json:"hostname"`
}

func (m UpdateHostnameDetails) String() string {
	return common.PointerString(m)
}

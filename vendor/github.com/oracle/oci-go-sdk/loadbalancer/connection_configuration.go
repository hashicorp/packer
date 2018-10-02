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

// ConnectionConfiguration Configuration details for the connection between the client and backend servers.
type ConnectionConfiguration struct {

	// The maximum idle time, in seconds, allowed between two successive receive or two successive send operations
	// between the client and backend servers. A send operation does not reset the timer for receive operations. A
	// receive operation does not reset the timer for send operations.
	// For more information, see Connection Configuration (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Reference/connectionreuse.htm#ConnectionConfiguration).
	// Example: `1200`
	IdleTimeout *int `mandatory:"true" json:"idleTimeout"`
}

func (m ConnectionConfiguration) String() string {
	return common.PointerString(m)
}

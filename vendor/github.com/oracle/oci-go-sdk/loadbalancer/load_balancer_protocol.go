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

// LoadBalancerProtocol The protocol that defines the type of traffic accepted by a listener.
type LoadBalancerProtocol struct {

	// The name of the protocol.
	Name *string `mandatory:"true" json:"name"`
}

func (m LoadBalancerProtocol) String() string {
	return common.PointerString(m)
}

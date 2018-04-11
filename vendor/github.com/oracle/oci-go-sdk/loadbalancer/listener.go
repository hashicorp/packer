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

// Listener The listener's configuration.
// For more information on backend set configuration, see
// Managing Load Balancer Listeners (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/tasks/managinglisteners.htm).
type Listener struct {

	// The name of the associated backend set.
	DefaultBackendSetName *string `mandatory:"true" json:"defaultBackendSetName"`

	// A friendly name for the listener. It must be unique and it cannot be changed.
	// Example: `My listener`
	Name *string `mandatory:"true" json:"name"`

	// The communication port for the listener.
	// Example: `80`
	Port *int `mandatory:"true" json:"port"`

	// The protocol on which the listener accepts connection requests.
	// To get a list of valid protocols, use the ListProtocols
	// operation.
	// Example: `HTTP`
	Protocol *string `mandatory:"true" json:"protocol"`

	SslConfiguration *SslConfiguration `mandatory:"false" json:"sslConfiguration"`
}

func (m Listener) String() string {
	return common.PointerString(m)
}

// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Public DNS Service
//
// API for managing DNS zones, records, and policies.
//

package dns

import (
	"github.com/oracle/oci-go-sdk/common"
)

// ExternalMaster An external master name server used as the source of zone data.
type ExternalMaster struct {

	// The server's IP address (IPv4 or IPv6).
	Address *string `mandatory:"true" json:"address"`

	// The server's port.
	Port *int `mandatory:"false" json:"port"`

	Tsig *Tsig `mandatory:"false" json:"tsig"`
}

func (m ExternalMaster) String() string {
	return common.PointerString(m)
}

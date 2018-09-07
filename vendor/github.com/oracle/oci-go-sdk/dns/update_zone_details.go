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

// UpdateZoneDetails The body for updating a zone.
type UpdateZoneDetails struct {

	// External master servers for the zone.
	ExternalMasters []ExternalMaster `mandatory:"false" json:"externalMasters"`
}

func (m UpdateZoneDetails) String() string {
	return common.PointerString(m)
}

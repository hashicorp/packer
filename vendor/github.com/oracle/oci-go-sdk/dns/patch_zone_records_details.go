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

// PatchZoneRecordsDetails The representation of PatchZoneRecordsDetails
type PatchZoneRecordsDetails struct {
	Items []RecordOperation `mandatory:"false" json:"items"`
}

func (m PatchZoneRecordsDetails) String() string {
	return common.PointerString(m)
}

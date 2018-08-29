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

// RecordCollection A collection of DNS resource records.
type RecordCollection struct {
	Items []Record `mandatory:"false" json:"items"`
}

func (m RecordCollection) String() string {
	return common.PointerString(m)
}

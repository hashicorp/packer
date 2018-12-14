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

// RrSet A collection of DNS records of the same domain and type. For more
// information about record types, see Resource Record (RR) TYPEs (https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4).
type RrSet struct {
	Items []Record `mandatory:"false" json:"items"`
}

func (m RrSet) String() string {
	return common.PointerString(m)
}

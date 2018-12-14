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

// Tsig A TSIG (https://tools.ietf.org/html/rfc2845) key.
type Tsig struct {

	// A domain name identifying the key for a given pair of hosts.
	Name *string `mandatory:"true" json:"name"`

	// A base64 string encoding the binary shared secret.
	Secret *string `mandatory:"true" json:"secret"`

	// TSIG Algorithms are encoded as domain names, but most consist of only one
	// non-empty label, which is not required to be explicitly absolute. For a
	// full list of TSIG algorithms, see Secret Key Transaction Authentication for DNS (TSIG) Algorithm Names (http://www.iana.org/assignments/tsig-algorithm-names/tsig-algorithm-names.xhtml#tsig-algorithm-names-1)
	Algorithm *string `mandatory:"true" json:"algorithm"`
}

func (m Tsig) String() string {
	return common.PointerString(m)
}

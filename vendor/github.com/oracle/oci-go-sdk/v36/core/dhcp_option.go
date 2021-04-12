// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Core Services API
//
// API covering the Networking (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/overview.htm),
// Compute (https://docs.cloud.oracle.com/iaas/Content/Compute/Concepts/computeoverview.htm), and
// Block Volume (https://docs.cloud.oracle.com/iaas/Content/Block/Concepts/overview.htm) services. Use this API
// to manage resources such as virtual cloud networks (VCNs), compute instances, and
// block storage volumes.
//

package core

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/v36/common"
)

// DhcpOption A single DHCP option according to RFC 1533 (https://tools.ietf.org/html/rfc1533).
// The two options available to use are DhcpDnsOption
// and DhcpSearchDomainOption. For more
// information, see DNS in Your Virtual Cloud Network (https://docs.cloud.oracle.com/iaas/Content/Network/Concepts/dns.htm)
// and DHCP Options (https://docs.cloud.oracle.com/iaas/Content/Network/Tasks/managingDHCP.htm).
type DhcpOption interface {
}

type dhcpoption struct {
	JsonData []byte
	Type     string `json:"type"`
}

// UnmarshalJSON unmarshals json
func (m *dhcpoption) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerdhcpoption dhcpoption
	s := struct {
		Model Unmarshalerdhcpoption
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.Type = s.Model.Type

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *dhcpoption) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {

	if data == nil || string(data) == "null" {
		return nil, nil
	}

	var err error
	switch m.Type {
	case "DomainNameServer":
		mm := DhcpDnsOption{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "SearchDomain":
		mm := DhcpSearchDomainOption{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return *m, nil
	}
}

func (m dhcpoption) String() string {
	return common.PointerString(m)
}

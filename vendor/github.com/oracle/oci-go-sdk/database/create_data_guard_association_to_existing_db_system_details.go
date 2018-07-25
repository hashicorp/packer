// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// CreateDataGuardAssociationToExistingDbSystemDetails The configuration details for creating a Data Guard association to an existing database.
type CreateDataGuardAssociationToExistingDbSystemDetails struct {

	// A strong password for the `SYS`, `SYSTEM`, and `PDB Admin` users to apply during standby creation.
	// The password must contain no fewer than nine characters and include:
	// * At least two uppercase characters.
	// * At least two lowercase characters.
	// * At least two numeric characters.
	// * At least two special characters. Valid special characters include "_", "#", and "-" only.
	// **The password MUST be the same as the primary admin password.**
	DatabaseAdminPassword *string `mandatory:"true" json:"databaseAdminPassword"`

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm) of the DB System to create the standby database on.
	PeerDbSystemId *string `mandatory:"false" json:"peerDbSystemId"`

	// The protection mode to set up between the primary and standby databases. For more information, see
	// Oracle Data Guard Protection Modes (http://docs.oracle.com/database/122/SBYDB/oracle-data-guard-protection-modes.htm#SBYDB02000)
	// in the Oracle Data Guard documentation.
	// **IMPORTANT** - The only protection mode currently supported by the Database Service is MAXIMUM_PERFORMANCE.
	ProtectionMode CreateDataGuardAssociationDetailsProtectionModeEnum `mandatory:"true" json:"protectionMode"`

	// The redo transport type to use for this Data Guard association.  Valid values depend on the specified `protectionMode`:
	// * MAXIMUM_AVAILABILITY - SYNC or FASTSYNC
	// * MAXIMUM_PERFORMANCE - ASYNC
	// * MAXIMUM_PROTECTION - SYNC
	// For more information, see
	// Redo Transport Services (http://docs.oracle.com/database/122/SBYDB/oracle-data-guard-redo-transport-services.htm#SBYDB00400)
	// in the Oracle Data Guard documentation.
	// **IMPORTANT** - The only transport type currently supported by the Database Service is ASYNC.
	TransportType CreateDataGuardAssociationDetailsTransportTypeEnum `mandatory:"true" json:"transportType"`
}

//GetDatabaseAdminPassword returns DatabaseAdminPassword
func (m CreateDataGuardAssociationToExistingDbSystemDetails) GetDatabaseAdminPassword() *string {
	return m.DatabaseAdminPassword
}

//GetProtectionMode returns ProtectionMode
func (m CreateDataGuardAssociationToExistingDbSystemDetails) GetProtectionMode() CreateDataGuardAssociationDetailsProtectionModeEnum {
	return m.ProtectionMode
}

//GetTransportType returns TransportType
func (m CreateDataGuardAssociationToExistingDbSystemDetails) GetTransportType() CreateDataGuardAssociationDetailsTransportTypeEnum {
	return m.TransportType
}

func (m CreateDataGuardAssociationToExistingDbSystemDetails) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m CreateDataGuardAssociationToExistingDbSystemDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeCreateDataGuardAssociationToExistingDbSystemDetails CreateDataGuardAssociationToExistingDbSystemDetails
	s := struct {
		DiscriminatorParam string `json:"creationType"`
		MarshalTypeCreateDataGuardAssociationToExistingDbSystemDetails
	}{
		"ExistingDbSystem",
		(MarshalTypeCreateDataGuardAssociationToExistingDbSystemDetails)(m),
	}

	return json.Marshal(&s)
}

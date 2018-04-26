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

// CreateDbHomeWithDbSystemIdDetails The representation of CreateDbHomeWithDbSystemIdDetails
type CreateDbHomeWithDbSystemIdDetails struct {

	// The OCID of the DB System.
	DbSystemId *string `mandatory:"true" json:"dbSystemId"`

	Database *CreateDatabaseDetails `mandatory:"true" json:"database"`

	// A valid Oracle database version. To get a list of supported versions, use the ListDbVersions operation.
	DbVersion *string `mandatory:"true" json:"dbVersion"`

	// The user-provided name of the database home.
	DisplayName *string `mandatory:"false" json:"displayName"`
}

//GetDbSystemId returns DbSystemId
func (m CreateDbHomeWithDbSystemIdDetails) GetDbSystemId() *string {
	return m.DbSystemId
}

//GetDisplayName returns DisplayName
func (m CreateDbHomeWithDbSystemIdDetails) GetDisplayName() *string {
	return m.DisplayName
}

func (m CreateDbHomeWithDbSystemIdDetails) String() string {
	return common.PointerString(m)
}

// MarshalJSON marshals to json representation
func (m CreateDbHomeWithDbSystemIdDetails) MarshalJSON() (buff []byte, e error) {
	type MarshalTypeCreateDbHomeWithDbSystemIdDetails CreateDbHomeWithDbSystemIdDetails
	s := struct {
		DiscriminatorParam string `json:"source"`
		MarshalTypeCreateDbHomeWithDbSystemIdDetails
	}{
		"NONE",
		(MarshalTypeCreateDbHomeWithDbSystemIdDetails)(m),
	}

	return json.Marshal(&s)
}

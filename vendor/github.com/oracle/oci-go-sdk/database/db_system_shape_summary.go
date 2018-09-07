// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"github.com/oracle/oci-go-sdk/common"
)

// DbSystemShapeSummary The shape of the DB System. The shape determines resources to allocate to the DB system - CPU cores and memory for VM shapes; CPU cores, memory and storage for non-VM (or bare metal) shapes.
// For a description of shapes, see DB System Launch Options (https://docs.us-phoenix-1.oraclecloud.com/Content/Database/References/launchoptions.htm).
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized, talk to an administrator.
// If you're an administrator who needs to write policies to give users access,
// see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type DbSystemShapeSummary struct {

	// The maximum number of CPU cores that can be enabled on the DB System for this shape.
	AvailableCoreCount *int `mandatory:"true" json:"availableCoreCount"`

	// The name of the shape used for the DB System.
	Name *string `mandatory:"true" json:"name"`

	// The discrete number by which the CPU core count for this shape can be increased or decreased.
	CoreCountIncrement *int `mandatory:"false" json:"coreCountIncrement"`

	// The maximum number of database nodes available for this shape.
	MaximumNodeCount *int `mandatory:"false" json:"maximumNodeCount"`

	// The minimum number of CPU cores that can be enabled on the DB System for this shape.
	MinimumCoreCount *int `mandatory:"false" json:"minimumCoreCount"`

	// The minimum number of database nodes available for this shape.
	MinimumNodeCount *int `mandatory:"false" json:"minimumNodeCount"`

	// Deprecated. Use `name` instead of `shape`.
	Shape *string `mandatory:"false" json:"shape"`
}

func (m DbSystemShapeSummary) String() string {
	return common.PointerString(m)
}

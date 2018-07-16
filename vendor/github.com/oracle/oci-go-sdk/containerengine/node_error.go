// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Container Engine for Kubernetes API
//
// Container Engine for Kubernetes API
//

package containerengine

import (
	"github.com/oracle/oci-go-sdk/common"
)

// NodeError The properties that define an upstream error while managing a node.
type NodeError struct {

	// A short error code that defines the upstream error, meant for programmatic parsing. See API Errors (https://docs.us-phoenix-1.oraclecloud.com/Content/API/References/apierrors.htm).
	Code *string `mandatory:"true" json:"code"`

	// A human-readable error string of the upstream error.
	Message *string `mandatory:"true" json:"message"`

	// The status of the HTTP response encountered in the upstream error.
	Status *string `mandatory:"false" json:"status"`

	// Unique Oracle-assigned identifier for the upstream request. If you need to contact Oracle about a particular upstream request, please provide the request ID.
	OpcRequestId *string `mandatory:"false" json:"opc-request-id"`
}

func (m NodeError) String() string {
	return common.PointerString(m)
}

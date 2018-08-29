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

// ClusterEndpoints The properties that define endpoints for a cluster.
type ClusterEndpoints struct {

	// The Kubernetes API server endpoint.
	Kubernetes *string `mandatory:"false" json:"kubernetes"`
}

func (m ClusterEndpoints) String() string {
	return common.PointerString(m)
}

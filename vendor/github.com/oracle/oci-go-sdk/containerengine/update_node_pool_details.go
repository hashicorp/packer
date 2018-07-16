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

// UpdateNodePoolDetails The properties that define a request to update a node pool.
type UpdateNodePoolDetails struct {

	// The new name for the cluster. Avoid entering confidential information.
	Name *string `mandatory:"false" json:"name"`

	// The version of Kubernetes to which the nodes in the node pool should be upgraded.
	KubernetesVersion *string `mandatory:"false" json:"kubernetesVersion"`

	// The number of nodes to ensure in each subnet.
	QuantityPerSubnet *int `mandatory:"false" json:"quantityPerSubnet"`

	// A list of key/value pairs to add to nodes after they join the Kubernetes cluster.
	InitialNodeLabels []KeyValue `mandatory:"false" json:"initialNodeLabels"`

	// The OCIDs of the subnets in which to place nodes for this node pool.
	SubnetIds []string `mandatory:"false" json:"subnetIds"`
}

func (m UpdateNodePoolDetails) String() string {
	return common.PointerString(m)
}

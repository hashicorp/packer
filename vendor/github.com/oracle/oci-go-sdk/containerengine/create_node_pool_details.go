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

// CreateNodePoolDetails The properties that define a request to create a node pool.
type CreateNodePoolDetails struct {

	// The OCID of the compartment in which the node pool exists.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the cluster to which this node pool is attached.
	ClusterId *string `mandatory:"true" json:"clusterId"`

	// The name of the node pool. Avoid entering confidential information.
	Name *string `mandatory:"true" json:"name"`

	// The version of Kubernetes to install on the nodes in the node pool.
	KubernetesVersion *string `mandatory:"true" json:"kubernetesVersion"`

	// The name of the image running on the nodes in the node pool.
	NodeImageName *string `mandatory:"true" json:"nodeImageName"`

	// The name of the node shape of the nodes in the node pool.
	NodeShape *string `mandatory:"true" json:"nodeShape"`

	// The OCIDs of the subnets in which to place nodes for this node pool.
	SubnetIds []string `mandatory:"true" json:"subnetIds"`

	// A list of key/value pairs to add to nodes after they join the Kubernetes cluster.
	InitialNodeLabels []KeyValue `mandatory:"false" json:"initialNodeLabels"`

	// The SSH public key to add to each node in the node pool.
	SshPublicKey *string `mandatory:"false" json:"sshPublicKey"`

	// The number of nodes to create in each subnet.
	QuantityPerSubnet *int `mandatory:"false" json:"quantityPerSubnet"`
}

func (m CreateNodePoolDetails) String() string {
	return common.PointerString(m)
}

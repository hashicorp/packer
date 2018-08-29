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

// ClusterCreateOptions The properties that define extra options for a cluster.
type ClusterCreateOptions struct {

	// The OCIDs of the subnets used for Kubernetes services load balancers.
	ServiceLbSubnetIds []string `mandatory:"false" json:"serviceLbSubnetIds"`

	// Network configuration for Kubernetes.
	KubernetesNetworkConfig *KubernetesNetworkConfig `mandatory:"false" json:"kubernetesNetworkConfig"`

	// Configurable cluster add-ons
	AddOns *AddOnOptions `mandatory:"false" json:"addOns"`
}

func (m ClusterCreateOptions) String() string {
	return common.PointerString(m)
}

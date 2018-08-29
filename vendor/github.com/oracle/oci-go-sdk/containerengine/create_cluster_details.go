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

// CreateClusterDetails The properties that define a request to create a cluster.
type CreateClusterDetails struct {

	// The name of the cluster. Avoid entering confidential information.
	Name *string `mandatory:"true" json:"name"`

	// The OCID of the compartment in which to create the cluster.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The OCID of the virtual cloud network (VCN) in which to create the cluster.
	VcnId *string `mandatory:"true" json:"vcnId"`

	// The version of Kubernetes to install into the cluster masters.
	KubernetesVersion *string `mandatory:"true" json:"kubernetesVersion"`

	// Optional attributes for the cluster.
	Options *ClusterCreateOptions `mandatory:"false" json:"options"`
}

func (m CreateClusterDetails) String() string {
	return common.PointerString(m)
}

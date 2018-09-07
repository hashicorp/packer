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

// CreateClusterKubeconfigContentDetails The properties that define a request to create a cluster kubeconfig.
type CreateClusterKubeconfigContentDetails struct {

	// The version of the kubeconfig token.
	TokenVersion *string `mandatory:"false" json:"tokenVersion"`

	// The desired expiration, in seconds, to use for the kubeconfig token.
	Expiration *int `mandatory:"false" json:"expiration"`
}

func (m CreateClusterKubeconfigContentDetails) String() string {
	return common.PointerString(m)
}

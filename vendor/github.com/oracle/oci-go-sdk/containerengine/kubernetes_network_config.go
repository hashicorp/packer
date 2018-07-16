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

// KubernetesNetworkConfig The properties that define the network configuration for Kubernetes.
type KubernetesNetworkConfig struct {

	// The CIDR block for Kubernetes pods.
	PodsCidr *string `mandatory:"false" json:"podsCidr"`

	// The CIDR block for Kubernetes services.
	ServicesCidr *string `mandatory:"false" json:"servicesCidr"`
}

func (m KubernetesNetworkConfig) String() string {
	return common.PointerString(m)
}

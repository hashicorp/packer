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

// NodePoolOptions Options for creating or updating node pools.
type NodePoolOptions struct {

	// Available Kubernetes versions.
	KubernetesVersions []string `mandatory:"false" json:"kubernetesVersions"`

	// Available Kubernetes versions.
	Images []string `mandatory:"false" json:"images"`

	// Available shapes for nodes.
	Shapes []string `mandatory:"false" json:"shapes"`
}

func (m NodePoolOptions) String() string {
	return common.PointerString(m)
}

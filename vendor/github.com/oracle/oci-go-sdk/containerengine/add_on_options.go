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

// AddOnOptions The properties that define options for supported add-ons.
type AddOnOptions struct {

	// Whether or not to enable the Kubernetes Dashboard add-on.
	IsKubernetesDashboardEnabled *bool `mandatory:"false" json:"isKubernetesDashboardEnabled"`

	// Whether or not to enable the Tiller add-on.
	IsTillerEnabled *bool `mandatory:"false" json:"isTillerEnabled"`
}

func (m AddOnOptions) String() string {
	return common.PointerString(m)
}

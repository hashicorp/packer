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

// KeyValue The properties that define a key value pair.
type KeyValue struct {

	// The key of the pair.
	Key *string `mandatory:"false" json:"key"`

	// The value of the pair.
	Value *string `mandatory:"false" json:"value"`
}

func (m KeyValue) String() string {
	return common.PointerString(m)
}

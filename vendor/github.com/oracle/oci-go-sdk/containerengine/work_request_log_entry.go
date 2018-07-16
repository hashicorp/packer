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

// WorkRequestLogEntry Log entries related to a specific work request.
type WorkRequestLogEntry struct {

	// The description of an action that occurred.
	Message *string `mandatory:"false" json:"message"`

	// The date and time the log entry occurred.
	Timestamp *string `mandatory:"false" json:"timestamp"`
}

func (m WorkRequestLogEntry) String() string {
	return common.PointerString(m)
}

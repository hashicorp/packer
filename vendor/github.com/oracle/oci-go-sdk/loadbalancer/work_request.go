// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Load Balancing Service API
//
// API for the Load Balancing Service
//

package loadbalancer

import (
	"github.com/oracle/oci-go-sdk/common"
)

// WorkRequest Many of the API requests you use to create and configure load balancing do not take effect immediately.
// In these cases, the request spawns an asynchronous work flow to fulfill the request. WorkRequest objects provide visibility
// for in-progress work flows.
// For more information about work requests, see Viewing the State of a Work Request (https://docs.us-phoenix-1.oraclecloud.com/Content/Balance/Tasks/viewingworkrequest.htm).
type WorkRequest struct {
	ErrorDetails []WorkRequestError `mandatory:"true" json:"errorDetails"`

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm) of the work request.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the work request.
	LifecycleState WorkRequestLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/identifiers.htm) of the load balancer with which the work request
	// is associated.
	LoadBalancerId *string `mandatory:"true" json:"loadBalancerId"`

	// A collection of data, related to the load balancer provisioning process, that helps with debugging in the event of failure.
	// Possible data elements include:
	// - workflow name
	// - event ID
	// - work request ID
	// - load balancer ID
	// - workflow completion message
	Message *string `mandatory:"true" json:"message"`

	// The date and time the work request was created, in the format defined by RFC3339.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeAccepted *common.SDKTime `mandatory:"true" json:"timeAccepted"`

	// The type of action the work request represents.
	// Example: `CreateListener`
	Type *string `mandatory:"true" json:"type"`

	// The date and time the work request was completed, in the format defined by RFC3339.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeFinished *common.SDKTime `mandatory:"false" json:"timeFinished"`
}

func (m WorkRequest) String() string {
	return common.PointerString(m)
}

// WorkRequestLifecycleStateEnum Enum with underlying type: string
type WorkRequestLifecycleStateEnum string

// Set of constants representing the allowable values for WorkRequestLifecycleState
const (
	WorkRequestLifecycleStateAccepted   WorkRequestLifecycleStateEnum = "ACCEPTED"
	WorkRequestLifecycleStateInProgress WorkRequestLifecycleStateEnum = "IN_PROGRESS"
	WorkRequestLifecycleStateFailed     WorkRequestLifecycleStateEnum = "FAILED"
	WorkRequestLifecycleStateSucceeded  WorkRequestLifecycleStateEnum = "SUCCEEDED"
)

var mappingWorkRequestLifecycleState = map[string]WorkRequestLifecycleStateEnum{
	"ACCEPTED":    WorkRequestLifecycleStateAccepted,
	"IN_PROGRESS": WorkRequestLifecycleStateInProgress,
	"FAILED":      WorkRequestLifecycleStateFailed,
	"SUCCEEDED":   WorkRequestLifecycleStateSucceeded,
}

// GetWorkRequestLifecycleStateEnumValues Enumerates the set of values for WorkRequestLifecycleState
func GetWorkRequestLifecycleStateEnumValues() []WorkRequestLifecycleStateEnum {
	values := make([]WorkRequestLifecycleStateEnum, 0)
	for _, v := range mappingWorkRequestLifecycleState {
		values = append(values, v)
	}
	return values
}

// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"github.com/oracle/oci-go-sdk/common"
)

// PatchSummary A Patch for a DB System or DB Home.
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized,
// talk to an administrator. If you're an administrator who needs to write policies to give users access,
// see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
type PatchSummary struct {

	// The text describing this patch package.
	Description *string `mandatory:"true" json:"description"`

	// The OCID of the patch.
	Id *string `mandatory:"true" json:"id"`

	// The date and time that the patch was released.
	TimeReleased *common.SDKTime `mandatory:"true" json:"timeReleased"`

	// The version of this patch package.
	Version *string `mandatory:"true" json:"version"`

	// Actions that can possibly be performed using this patch.
	AvailableActions []PatchSummaryAvailableActionsEnum `mandatory:"false" json:"availableActions,omitempty"`

	// Action that is currently being performed or was completed last.
	LastAction PatchSummaryLastActionEnum `mandatory:"false" json:"lastAction,omitempty"`

	// A descriptive text associated with the lifecycleState.
	// Typically can contain additional displayable text.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The current state of the patch as a result of lastAction.
	LifecycleState PatchSummaryLifecycleStateEnum `mandatory:"false" json:"lifecycleState,omitempty"`
}

func (m PatchSummary) String() string {
	return common.PointerString(m)
}

// PatchSummaryAvailableActionsEnum Enum with underlying type: string
type PatchSummaryAvailableActionsEnum string

// Set of constants representing the allowable values for PatchSummaryAvailableActions
const (
	PatchSummaryAvailableActionsApply    PatchSummaryAvailableActionsEnum = "APPLY"
	PatchSummaryAvailableActionsPrecheck PatchSummaryAvailableActionsEnum = "PRECHECK"
)

var mappingPatchSummaryAvailableActions = map[string]PatchSummaryAvailableActionsEnum{
	"APPLY":    PatchSummaryAvailableActionsApply,
	"PRECHECK": PatchSummaryAvailableActionsPrecheck,
}

// GetPatchSummaryAvailableActionsEnumValues Enumerates the set of values for PatchSummaryAvailableActions
func GetPatchSummaryAvailableActionsEnumValues() []PatchSummaryAvailableActionsEnum {
	values := make([]PatchSummaryAvailableActionsEnum, 0)
	for _, v := range mappingPatchSummaryAvailableActions {
		values = append(values, v)
	}
	return values
}

// PatchSummaryLastActionEnum Enum with underlying type: string
type PatchSummaryLastActionEnum string

// Set of constants representing the allowable values for PatchSummaryLastAction
const (
	PatchSummaryLastActionApply    PatchSummaryLastActionEnum = "APPLY"
	PatchSummaryLastActionPrecheck PatchSummaryLastActionEnum = "PRECHECK"
)

var mappingPatchSummaryLastAction = map[string]PatchSummaryLastActionEnum{
	"APPLY":    PatchSummaryLastActionApply,
	"PRECHECK": PatchSummaryLastActionPrecheck,
}

// GetPatchSummaryLastActionEnumValues Enumerates the set of values for PatchSummaryLastAction
func GetPatchSummaryLastActionEnumValues() []PatchSummaryLastActionEnum {
	values := make([]PatchSummaryLastActionEnum, 0)
	for _, v := range mappingPatchSummaryLastAction {
		values = append(values, v)
	}
	return values
}

// PatchSummaryLifecycleStateEnum Enum with underlying type: string
type PatchSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for PatchSummaryLifecycleState
const (
	PatchSummaryLifecycleStateAvailable  PatchSummaryLifecycleStateEnum = "AVAILABLE"
	PatchSummaryLifecycleStateSuccess    PatchSummaryLifecycleStateEnum = "SUCCESS"
	PatchSummaryLifecycleStateInProgress PatchSummaryLifecycleStateEnum = "IN_PROGRESS"
	PatchSummaryLifecycleStateFailed     PatchSummaryLifecycleStateEnum = "FAILED"
)

var mappingPatchSummaryLifecycleState = map[string]PatchSummaryLifecycleStateEnum{
	"AVAILABLE":   PatchSummaryLifecycleStateAvailable,
	"SUCCESS":     PatchSummaryLifecycleStateSuccess,
	"IN_PROGRESS": PatchSummaryLifecycleStateInProgress,
	"FAILED":      PatchSummaryLifecycleStateFailed,
}

// GetPatchSummaryLifecycleStateEnumValues Enumerates the set of values for PatchSummaryLifecycleState
func GetPatchSummaryLifecycleStateEnumValues() []PatchSummaryLifecycleStateEnum {
	values := make([]PatchSummaryLifecycleStateEnum, 0)
	for _, v := range mappingPatchSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}

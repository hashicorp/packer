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

// PatchHistoryEntry The record of a patch action on a specified target.
type PatchHistoryEntry struct {

	// The OCID of the patch history entry.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the action.
	LifecycleState PatchHistoryEntryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The OCID of the patch.
	PatchId *string `mandatory:"true" json:"patchId"`

	// The date and time when the patch action started.
	TimeStarted *common.SDKTime `mandatory:"true" json:"timeStarted"`

	// The action being performed or was completed.
	Action PatchHistoryEntryActionEnum `mandatory:"false" json:"action,omitempty"`

	// A descriptive text associated with the lifecycleState.
	// Typically contains additional displayable text.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The date and time when the patch action completed.
	TimeEnded *common.SDKTime `mandatory:"false" json:"timeEnded"`
}

func (m PatchHistoryEntry) String() string {
	return common.PointerString(m)
}

// PatchHistoryEntryActionEnum Enum with underlying type: string
type PatchHistoryEntryActionEnum string

// Set of constants representing the allowable values for PatchHistoryEntryAction
const (
	PatchHistoryEntryActionApply    PatchHistoryEntryActionEnum = "APPLY"
	PatchHistoryEntryActionPrecheck PatchHistoryEntryActionEnum = "PRECHECK"
)

var mappingPatchHistoryEntryAction = map[string]PatchHistoryEntryActionEnum{
	"APPLY":    PatchHistoryEntryActionApply,
	"PRECHECK": PatchHistoryEntryActionPrecheck,
}

// GetPatchHistoryEntryActionEnumValues Enumerates the set of values for PatchHistoryEntryAction
func GetPatchHistoryEntryActionEnumValues() []PatchHistoryEntryActionEnum {
	values := make([]PatchHistoryEntryActionEnum, 0)
	for _, v := range mappingPatchHistoryEntryAction {
		values = append(values, v)
	}
	return values
}

// PatchHistoryEntryLifecycleStateEnum Enum with underlying type: string
type PatchHistoryEntryLifecycleStateEnum string

// Set of constants representing the allowable values for PatchHistoryEntryLifecycleState
const (
	PatchHistoryEntryLifecycleStateInProgress PatchHistoryEntryLifecycleStateEnum = "IN_PROGRESS"
	PatchHistoryEntryLifecycleStateSucceeded  PatchHistoryEntryLifecycleStateEnum = "SUCCEEDED"
	PatchHistoryEntryLifecycleStateFailed     PatchHistoryEntryLifecycleStateEnum = "FAILED"
)

var mappingPatchHistoryEntryLifecycleState = map[string]PatchHistoryEntryLifecycleStateEnum{
	"IN_PROGRESS": PatchHistoryEntryLifecycleStateInProgress,
	"SUCCEEDED":   PatchHistoryEntryLifecycleStateSucceeded,
	"FAILED":      PatchHistoryEntryLifecycleStateFailed,
}

// GetPatchHistoryEntryLifecycleStateEnumValues Enumerates the set of values for PatchHistoryEntryLifecycleState
func GetPatchHistoryEntryLifecycleStateEnumValues() []PatchHistoryEntryLifecycleStateEnum {
	values := make([]PatchHistoryEntryLifecycleStateEnum, 0)
	for _, v := range mappingPatchHistoryEntryLifecycleState {
		values = append(values, v)
	}
	return values
}

// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// File Storage Service API
//
// The API for the File Storage Service.
//

package filestorage

import (
	"github.com/oracle/oci-go-sdk/common"
)

// Export A file system and the path that you can use to mount it. Each export
// resource belongs to exactly one export set.
// The export's path attribute is not a path in the
// referenced file system, but the value used by clients for the path
// component of the remotetarget argument when mounting the file
// system.
// The path must start with a slash (/) followed by a sequence of zero or more
// slash-separated path elements. For any two export resources associated with
// the same export set, except those in a 'DELETED' state, the path element
// sequence for the first export resource can't contain the
// complete path element sequence of the second export resource.
//
// For example, the following are acceptable:
//   * /example and /path
//   * /example1 and /example2
//   * /example and /example1
// The following examples are not acceptable:
//   * /example and /example/path
//   * / and /example
// Paths may not end in a slash (/). No path element can be a period (.)
// or two periods in sequence (..). All path elements must be 255 bytes or less.
// No two non-'DELETED' export resources in the same export set can
// reference the same file system.
type Export struct {

	// The OCID of this export's export set.
	ExportSetId *string `mandatory:"true" json:"exportSetId"`

	// The OCID of this export's file system.
	FileSystemId *string `mandatory:"true" json:"fileSystemId"`

	// The OCID of this export.
	Id *string `mandatory:"true" json:"id"`

	// The current state of this export.
	LifecycleState ExportLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// Path used to access the associated file system.
	// Avoid entering confidential information.
	// Example: `/accounting`
	Path *string `mandatory:"true" json:"path"`

	// The date and time the export was created, expressed
	// in RFC 3339 (https://tools.ietf.org/rfc/rfc3339) timestamp format.
	// Example: `2016-08-25T21:10:29.600Z`
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`
}

func (m Export) String() string {
	return common.PointerString(m)
}

// ExportLifecycleStateEnum Enum with underlying type: string
type ExportLifecycleStateEnum string

// Set of constants representing the allowable values for ExportLifecycleState
const (
	ExportLifecycleStateCreating ExportLifecycleStateEnum = "CREATING"
	ExportLifecycleStateActive   ExportLifecycleStateEnum = "ACTIVE"
	ExportLifecycleStateDeleting ExportLifecycleStateEnum = "DELETING"
	ExportLifecycleStateDeleted  ExportLifecycleStateEnum = "DELETED"
)

var mappingExportLifecycleState = map[string]ExportLifecycleStateEnum{
	"CREATING": ExportLifecycleStateCreating,
	"ACTIVE":   ExportLifecycleStateActive,
	"DELETING": ExportLifecycleStateDeleting,
	"DELETED":  ExportLifecycleStateDeleted,
}

// GetExportLifecycleStateEnumValues Enumerates the set of values for ExportLifecycleState
func GetExportLifecycleStateEnumValues() []ExportLifecycleStateEnum {
	values := make([]ExportLifecycleStateEnum, 0)
	for _, v := range mappingExportLifecycleState {
		values = append(values, v)
	}
	return values
}

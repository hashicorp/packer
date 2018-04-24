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

// UpdateDbSystemDetails Describes the modification parameters for the DB System.
type UpdateDbSystemDetails struct {

	// The number of CPU Cores to be set on the DB System. Applicable only for non-VM based DB systems.
	CpuCoreCount *int `mandatory:"false" json:"cpuCoreCount"`

	// Size, in GBs, to which the currently attached storage needs to be scaled up to for VM based DB system. This must be greater than current storage size. Note that the total storage size attached will be more than what is requested, to account for REDO/RECO space and software volume.
	DataStorageSizeInGBs *int `mandatory:"false" json:"dataStorageSizeInGBs"`

	// The public key portion of the key pair to use for SSH access to the DB System. Multiple public keys can be provided. The length of the combined keys cannot exceed 10,000 characters.
	SshPublicKeys []string `mandatory:"false" json:"sshPublicKeys"`

	Version *PatchDetails `mandatory:"false" json:"version"`
}

func (m UpdateDbSystemDetails) String() string {
	return common.PointerString(m)
}

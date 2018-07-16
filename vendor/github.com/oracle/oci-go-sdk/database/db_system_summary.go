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

// DbSystemSummary The Database Service supports several types of DB Systems, ranging in size, price, and performance. For details about each type of system, see:
// - Exadata DB Systems (https://docs.us-phoenix-1.oraclecloud.com/Content/Database/Concepts/exaoverview.htm)
// - Bare Metal and Virtual Machine DB Systems (https://docs.us-phoenix-1.oraclecloud.com/Content/Database/Concepts/overview.htm)
// To use any of the API operations, you must be authorized in an IAM policy. If you're not authorized, talk to an administrator. If you're an administrator who needs to write policies to give users access, see Getting Started with Policies (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/policygetstarted.htm).
//
// For information about access control and compartments, see
// Overview of the Identity Service (https://docs.us-phoenix-1.oraclecloud.com/Content/Identity/Concepts/overview.htm).
// For information about Availability Domains, see
// Regions and Availability Domains (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/regions.htm).
// To get a list of Availability Domains, use the `ListAvailabilityDomains` operation
// in the Identity Service API.
type DbSystemSummary struct {

	// The name of the Availability Domain that the DB System is located in.
	AvailabilityDomain *string `mandatory:"true" json:"availabilityDomain"`

	// The OCID of the compartment.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The number of CPU cores enabled on the DB System.
	CpuCoreCount *int `mandatory:"true" json:"cpuCoreCount"`

	// The Oracle Database Edition that applies to all the databases on the DB System.
	DatabaseEdition DbSystemSummaryDatabaseEditionEnum `mandatory:"true" json:"databaseEdition"`

	// The user-friendly name for the DB System. It does not have to be unique.
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The domain name for the DB System.
	Domain *string `mandatory:"true" json:"domain"`

	// The host name for the DB Node.
	Hostname *string `mandatory:"true" json:"hostname"`

	// The OCID of the DB System.
	Id *string `mandatory:"true" json:"id"`

	// The current state of the DB System.
	LifecycleState DbSystemSummaryLifecycleStateEnum `mandatory:"true" json:"lifecycleState"`

	// The shape of the DB System. The shape determines resources to allocate to the DB system - CPU cores and memory for VM shapes; CPU cores, memory and storage for non-VM (or bare metal) shapes.
	Shape *string `mandatory:"true" json:"shape"`

	// The public key portion of one or more key pairs used for SSH access to the DB System.
	SshPublicKeys []string `mandatory:"true" json:"sshPublicKeys"`

	// The OCID of the subnet the DB System is associated with.
	// **Subnet Restrictions:**
	// - For single node and 2-node (RAC) DB Systems, do not use a subnet that overlaps with 192.168.16.16/28
	// - For Exadata and VM-based RAC DB Systems, do not use a subnet that overlaps with 192.168.128.0/20
	// These subnets are used by the Oracle Clusterware private interconnect on the database instance.
	// Specifying an overlapping subnet will cause the private interconnect to malfunction.
	// This restriction applies to both the client subnet and backup subnet.
	SubnetId *string `mandatory:"true" json:"subnetId"`

	// The OCID of the backup network subnet the DB System is associated with. Applicable only to Exadata.
	// **Subnet Restriction:** See above subnetId's 'Subnet Restriction'.
	// to malfunction.
	BackupSubnetId *string `mandatory:"false" json:"backupSubnetId"`

	// Cluster name for Exadata and 2-node RAC DB Systems. The cluster name must begin with an an alphabetic character, and may contain hyphens (-). Underscores (_) are not permitted. The cluster name can be no longer than 11 characters and is not case sensitive.
	ClusterName *string `mandatory:"false" json:"clusterName"`

	// The percentage assigned to DATA storage (user data and database files).
	// The remaining percentage is assigned to RECO storage (database redo logs, archive logs, and recovery manager backups). Accepted values are 40 and 80.
	DataStoragePercentage *int `mandatory:"false" json:"dataStoragePercentage"`

	// Data storage size, in GBs, that is currently available to the DB system. This is applicable only for VM-based DBs.
	DataStorageSizeInGBs *int `mandatory:"false" json:"dataStorageSizeInGBs"`

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	DefinedTags map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`

	// The type of redundancy configured for the DB System.
	// Normal is 2-way redundancy.
	// High is 3-way redundancy.
	DiskRedundancy DbSystemSummaryDiskRedundancyEnum `mandatory:"false" json:"diskRedundancy,omitempty"`

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	FreeformTags map[string]string `mandatory:"false" json:"freeformTags"`

	// The OCID of the last patch history. This is updated as soon as a patch operation is started.
	LastPatchHistoryEntryId *string `mandatory:"false" json:"lastPatchHistoryEntryId"`

	// The Oracle license model that applies to all the databases on the DB System. The default is LICENSE_INCLUDED.
	LicenseModel DbSystemSummaryLicenseModelEnum `mandatory:"false" json:"licenseModel,omitempty"`

	// Additional information about the current lifecycleState.
	LifecycleDetails *string `mandatory:"false" json:"lifecycleDetails"`

	// The port number configured for the listener on the DB System.
	ListenerPort *int `mandatory:"false" json:"listenerPort"`

	// Number of nodes in this DB system. For RAC DBs, this will be greater than 1.
	NodeCount *int `mandatory:"false" json:"nodeCount"`

	// RECO/REDO storage size, in GBs, that is currently allocated to the DB system. This is applicable only for VM-based DBs.
	RecoStorageSizeInGB *int `mandatory:"false" json:"recoStorageSizeInGB"`

	// The OCID of the DNS record for the SCAN IP addresses that are associated with the DB System.
	ScanDnsRecordId *string `mandatory:"false" json:"scanDnsRecordId"`

	// The OCID of the Single Client Access Name (SCAN) IP addresses associated with the DB System.
	// SCAN IP addresses are typically used for load balancing and are not assigned to any interface.
	// Clusterware directs the requests to the appropriate nodes in the cluster.
	// - For a single-node DB System, this list is empty.
	ScanIpIds []string `mandatory:"false" json:"scanIpIds"`

	// The date and time the DB System was created.
	TimeCreated *common.SDKTime `mandatory:"false" json:"timeCreated"`

	// The version of the DB System.
	Version *string `mandatory:"false" json:"version"`

	// The OCID of the virtual IP (VIP) addresses associated with the DB System.
	// The Cluster Ready Services (CRS) creates and maintains one VIP address for each node in the DB System to
	// enable failover. If one node fails, the VIP is reassigned to another active node in the cluster.
	// - For a single-node DB System, this list is empty.
	VipIds []string `mandatory:"false" json:"vipIds"`
}

func (m DbSystemSummary) String() string {
	return common.PointerString(m)
}

// DbSystemSummaryDatabaseEditionEnum Enum with underlying type: string
type DbSystemSummaryDatabaseEditionEnum string

// Set of constants representing the allowable values for DbSystemSummaryDatabaseEdition
const (
	DbSystemSummaryDatabaseEditionStandardEdition                     DbSystemSummaryDatabaseEditionEnum = "STANDARD_EDITION"
	DbSystemSummaryDatabaseEditionEnterpriseEdition                   DbSystemSummaryDatabaseEditionEnum = "ENTERPRISE_EDITION"
	DbSystemSummaryDatabaseEditionEnterpriseEditionExtremePerformance DbSystemSummaryDatabaseEditionEnum = "ENTERPRISE_EDITION_EXTREME_PERFORMANCE"
	DbSystemSummaryDatabaseEditionEnterpriseEditionHighPerformance    DbSystemSummaryDatabaseEditionEnum = "ENTERPRISE_EDITION_HIGH_PERFORMANCE"
)

var mappingDbSystemSummaryDatabaseEdition = map[string]DbSystemSummaryDatabaseEditionEnum{
	"STANDARD_EDITION":                       DbSystemSummaryDatabaseEditionStandardEdition,
	"ENTERPRISE_EDITION":                     DbSystemSummaryDatabaseEditionEnterpriseEdition,
	"ENTERPRISE_EDITION_EXTREME_PERFORMANCE": DbSystemSummaryDatabaseEditionEnterpriseEditionExtremePerformance,
	"ENTERPRISE_EDITION_HIGH_PERFORMANCE":    DbSystemSummaryDatabaseEditionEnterpriseEditionHighPerformance,
}

// GetDbSystemSummaryDatabaseEditionEnumValues Enumerates the set of values for DbSystemSummaryDatabaseEdition
func GetDbSystemSummaryDatabaseEditionEnumValues() []DbSystemSummaryDatabaseEditionEnum {
	values := make([]DbSystemSummaryDatabaseEditionEnum, 0)
	for _, v := range mappingDbSystemSummaryDatabaseEdition {
		values = append(values, v)
	}
	return values
}

// DbSystemSummaryDiskRedundancyEnum Enum with underlying type: string
type DbSystemSummaryDiskRedundancyEnum string

// Set of constants representing the allowable values for DbSystemSummaryDiskRedundancy
const (
	DbSystemSummaryDiskRedundancyHigh   DbSystemSummaryDiskRedundancyEnum = "HIGH"
	DbSystemSummaryDiskRedundancyNormal DbSystemSummaryDiskRedundancyEnum = "NORMAL"
)

var mappingDbSystemSummaryDiskRedundancy = map[string]DbSystemSummaryDiskRedundancyEnum{
	"HIGH":   DbSystemSummaryDiskRedundancyHigh,
	"NORMAL": DbSystemSummaryDiskRedundancyNormal,
}

// GetDbSystemSummaryDiskRedundancyEnumValues Enumerates the set of values for DbSystemSummaryDiskRedundancy
func GetDbSystemSummaryDiskRedundancyEnumValues() []DbSystemSummaryDiskRedundancyEnum {
	values := make([]DbSystemSummaryDiskRedundancyEnum, 0)
	for _, v := range mappingDbSystemSummaryDiskRedundancy {
		values = append(values, v)
	}
	return values
}

// DbSystemSummaryLicenseModelEnum Enum with underlying type: string
type DbSystemSummaryLicenseModelEnum string

// Set of constants representing the allowable values for DbSystemSummaryLicenseModel
const (
	DbSystemSummaryLicenseModelLicenseIncluded     DbSystemSummaryLicenseModelEnum = "LICENSE_INCLUDED"
	DbSystemSummaryLicenseModelBringYourOwnLicense DbSystemSummaryLicenseModelEnum = "BRING_YOUR_OWN_LICENSE"
)

var mappingDbSystemSummaryLicenseModel = map[string]DbSystemSummaryLicenseModelEnum{
	"LICENSE_INCLUDED":       DbSystemSummaryLicenseModelLicenseIncluded,
	"BRING_YOUR_OWN_LICENSE": DbSystemSummaryLicenseModelBringYourOwnLicense,
}

// GetDbSystemSummaryLicenseModelEnumValues Enumerates the set of values for DbSystemSummaryLicenseModel
func GetDbSystemSummaryLicenseModelEnumValues() []DbSystemSummaryLicenseModelEnum {
	values := make([]DbSystemSummaryLicenseModelEnum, 0)
	for _, v := range mappingDbSystemSummaryLicenseModel {
		values = append(values, v)
	}
	return values
}

// DbSystemSummaryLifecycleStateEnum Enum with underlying type: string
type DbSystemSummaryLifecycleStateEnum string

// Set of constants representing the allowable values for DbSystemSummaryLifecycleState
const (
	DbSystemSummaryLifecycleStateProvisioning DbSystemSummaryLifecycleStateEnum = "PROVISIONING"
	DbSystemSummaryLifecycleStateAvailable    DbSystemSummaryLifecycleStateEnum = "AVAILABLE"
	DbSystemSummaryLifecycleStateUpdating     DbSystemSummaryLifecycleStateEnum = "UPDATING"
	DbSystemSummaryLifecycleStateTerminating  DbSystemSummaryLifecycleStateEnum = "TERMINATING"
	DbSystemSummaryLifecycleStateTerminated   DbSystemSummaryLifecycleStateEnum = "TERMINATED"
	DbSystemSummaryLifecycleStateFailed       DbSystemSummaryLifecycleStateEnum = "FAILED"
)

var mappingDbSystemSummaryLifecycleState = map[string]DbSystemSummaryLifecycleStateEnum{
	"PROVISIONING": DbSystemSummaryLifecycleStateProvisioning,
	"AVAILABLE":    DbSystemSummaryLifecycleStateAvailable,
	"UPDATING":     DbSystemSummaryLifecycleStateUpdating,
	"TERMINATING":  DbSystemSummaryLifecycleStateTerminating,
	"TERMINATED":   DbSystemSummaryLifecycleStateTerminated,
	"FAILED":       DbSystemSummaryLifecycleStateFailed,
}

// GetDbSystemSummaryLifecycleStateEnumValues Enumerates the set of values for DbSystemSummaryLifecycleState
func GetDbSystemSummaryLifecycleStateEnumValues() []DbSystemSummaryLifecycleStateEnum {
	values := make([]DbSystemSummaryLifecycleStateEnum, 0)
	for _, v := range mappingDbSystemSummaryLifecycleState {
		values = append(values, v)
	}
	return values
}

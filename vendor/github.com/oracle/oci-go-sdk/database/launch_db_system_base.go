// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.
// Code generated. DO NOT EDIT.

// Database Service API
//
// The API for the Database Service.
//

package database

import (
	"encoding/json"
	"github.com/oracle/oci-go-sdk/common"
)

// LaunchDbSystemBase The representation of LaunchDbSystemBase
type LaunchDbSystemBase interface {

	// The Availability Domain where the DB System is located.
	GetAvailabilityDomain() *string

	// The Oracle Cloud ID (OCID) of the compartment the DB System  belongs in.
	GetCompartmentId() *string

	// The number of CPU cores to enable. The valid values depend on the specified shape:
	// - BM.DenseIO1.36 and BM.HighIO1.36 - Specify a multiple of 2, from 2 to 36.
	// - BM.RACLocalStorage1.72 - Specify a multiple of 4, from 4 to 72.
	// - Exadata.Quarter1.84 - Specify a multiple of 2, from 22 to 84.
	// - Exadata.Half1.168 - Specify a multiple of 4, from 44 to 168.
	// - Exadata.Full1.336 - Specify a multiple of 8, from 88 to 336.
	// For VM DB systems, the core count is inferred from the specific VM shape chosen, so this parameter is not used.
	GetCpuCoreCount() *int

	// The host name for the DB System. The host name must begin with an alphabetic character and
	// can contain a maximum of 30 alphanumeric characters, including hyphens (-).
	// The maximum length of the combined hostname and domain is 63 characters.
	// **Note:** The hostname must be unique within the subnet. If it is not unique,
	// the DB System will fail to provision.
	GetHostname() *string

	// The shape of the DB System. The shape determines resources allocated to the DB System - CPU cores and memory for VM shapes; CPU cores, memory and storage for non-VM (or bare metal) shapes. To get a list of shapes, use the ListDbSystemShapes operation.
	GetShape() *string

	// The public key portion of the key pair to use for SSH access to the DB System. Multiple public keys can be provided. The length of the combined keys cannot exceed 10,000 characters.
	GetSshPublicKeys() []string

	// The OCID of the subnet the DB System is associated with.
	// **Subnet Restrictions:**
	// - For single node and 2-node (RAC) DB Systems, do not use a subnet that overlaps with 192.168.16.16/28
	// - For Exadata and VM-based RAC DB Systems, do not use a subnet that overlaps with 192.168.128.0/20
	// These subnets are used by the Oracle Clusterware private interconnect on the database instance.
	// Specifying an overlapping subnet will cause the private interconnect to malfunction.
	// This restriction applies to both the client subnet and backup subnet.
	GetSubnetId() *string

	// The OCID of the backup network subnet the DB System is associated with. Applicable only to Exadata.
	// **Subnet Restrictions:** See above subnetId's **Subnet Restriction**.
	GetBackupSubnetId() *string

	// Cluster name for Exadata and 2-node RAC DB Systems. The cluster name must begin with an an alphabetic character, and may contain hyphens (-). Underscores (_) are not permitted. The cluster name can be no longer than 11 characters and is not case sensitive.
	GetClusterName() *string

	// The percentage assigned to DATA storage (user data and database files).
	// The remaining percentage is assigned to RECO storage (database redo logs, archive logs, and recovery manager backups).
	// Specify 80 or 40. The default is 80 percent assigned to DATA storage. This is not applicable for VM based DB systems.
	GetDataStoragePercentage() *int

	// Defined tags for this resource. Each key is predefined and scoped to a namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Operations": {"CostCenter": "42"}}`
	GetDefinedTags() map[string]map[string]interface{}

	// The user-friendly name for the DB System. It does not have to be unique.
	GetDisplayName() *string

	// A domain name used for the DB System. If the Oracle-provided Internet and VCN
	// Resolver is enabled for the specified subnet, the domain name for the subnet is used
	// (don't provide one). Otherwise, provide a valid DNS domain name. Hyphens (-) are not permitted.
	GetDomain() *string

	// Free-form tags for this resource. Each tag is a simple key-value pair with no predefined name, type, or namespace.
	// For more information, see Resource Tags (https://docs.us-phoenix-1.oraclecloud.com/Content/General/Concepts/resourcetags.htm).
	// Example: `{"Department": "Finance"}`
	GetFreeformTags() map[string]string

	// Size, in GBs, of the initial data volume that will be created and attached to VM-shape based DB system. This storage can later be scaled up if needed. Note that the total storage size attached will be more than what is requested, to account for REDO/RECO space and software volume.
	GetInitialDataStorageSizeInGB() *int

	// Number of nodes to launch for a VM-shape based RAC DB system.
	GetNodeCount() *int
}

type launchdbsystembase struct {
	JsonData                   []byte
	AvailabilityDomain         *string                           `mandatory:"true" json:"availabilityDomain"`
	CompartmentId              *string                           `mandatory:"true" json:"compartmentId"`
	CpuCoreCount               *int                              `mandatory:"true" json:"cpuCoreCount"`
	Hostname                   *string                           `mandatory:"true" json:"hostname"`
	Shape                      *string                           `mandatory:"true" json:"shape"`
	SshPublicKeys              []string                          `mandatory:"true" json:"sshPublicKeys"`
	SubnetId                   *string                           `mandatory:"true" json:"subnetId"`
	BackupSubnetId             *string                           `mandatory:"false" json:"backupSubnetId"`
	ClusterName                *string                           `mandatory:"false" json:"clusterName"`
	DataStoragePercentage      *int                              `mandatory:"false" json:"dataStoragePercentage"`
	DefinedTags                map[string]map[string]interface{} `mandatory:"false" json:"definedTags"`
	DisplayName                *string                           `mandatory:"false" json:"displayName"`
	Domain                     *string                           `mandatory:"false" json:"domain"`
	FreeformTags               map[string]string                 `mandatory:"false" json:"freeformTags"`
	InitialDataStorageSizeInGB *int                              `mandatory:"false" json:"initialDataStorageSizeInGB"`
	NodeCount                  *int                              `mandatory:"false" json:"nodeCount"`
	Source                     string                            `json:"source"`
}

// UnmarshalJSON unmarshals json
func (m *launchdbsystembase) UnmarshalJSON(data []byte) error {
	m.JsonData = data
	type Unmarshalerlaunchdbsystembase launchdbsystembase
	s := struct {
		Model Unmarshalerlaunchdbsystembase
	}{}
	err := json.Unmarshal(data, &s.Model)
	if err != nil {
		return err
	}
	m.AvailabilityDomain = s.Model.AvailabilityDomain
	m.CompartmentId = s.Model.CompartmentId
	m.CpuCoreCount = s.Model.CpuCoreCount
	m.Hostname = s.Model.Hostname
	m.Shape = s.Model.Shape
	m.SshPublicKeys = s.Model.SshPublicKeys
	m.SubnetId = s.Model.SubnetId
	m.BackupSubnetId = s.Model.BackupSubnetId
	m.ClusterName = s.Model.ClusterName
	m.DataStoragePercentage = s.Model.DataStoragePercentage
	m.DefinedTags = s.Model.DefinedTags
	m.DisplayName = s.Model.DisplayName
	m.Domain = s.Model.Domain
	m.FreeformTags = s.Model.FreeformTags
	m.InitialDataStorageSizeInGB = s.Model.InitialDataStorageSizeInGB
	m.NodeCount = s.Model.NodeCount
	m.Source = s.Model.Source

	return err
}

// UnmarshalPolymorphicJSON unmarshals polymorphic json
func (m *launchdbsystembase) UnmarshalPolymorphicJSON(data []byte) (interface{}, error) {
	var err error
	switch m.Source {
	case "NONE":
		mm := LaunchDbSystemDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	case "DB_BACKUP":
		mm := LaunchDbSystemFromBackupDetails{}
		err = json.Unmarshal(data, &mm)
		return mm, err
	default:
		return m, nil
	}
}

//GetAvailabilityDomain returns AvailabilityDomain
func (m launchdbsystembase) GetAvailabilityDomain() *string {
	return m.AvailabilityDomain
}

//GetCompartmentId returns CompartmentId
func (m launchdbsystembase) GetCompartmentId() *string {
	return m.CompartmentId
}

//GetCpuCoreCount returns CpuCoreCount
func (m launchdbsystembase) GetCpuCoreCount() *int {
	return m.CpuCoreCount
}

//GetHostname returns Hostname
func (m launchdbsystembase) GetHostname() *string {
	return m.Hostname
}

//GetShape returns Shape
func (m launchdbsystembase) GetShape() *string {
	return m.Shape
}

//GetSshPublicKeys returns SshPublicKeys
func (m launchdbsystembase) GetSshPublicKeys() []string {
	return m.SshPublicKeys
}

//GetSubnetId returns SubnetId
func (m launchdbsystembase) GetSubnetId() *string {
	return m.SubnetId
}

//GetBackupSubnetId returns BackupSubnetId
func (m launchdbsystembase) GetBackupSubnetId() *string {
	return m.BackupSubnetId
}

//GetClusterName returns ClusterName
func (m launchdbsystembase) GetClusterName() *string {
	return m.ClusterName
}

//GetDataStoragePercentage returns DataStoragePercentage
func (m launchdbsystembase) GetDataStoragePercentage() *int {
	return m.DataStoragePercentage
}

//GetDefinedTags returns DefinedTags
func (m launchdbsystembase) GetDefinedTags() map[string]map[string]interface{} {
	return m.DefinedTags
}

//GetDisplayName returns DisplayName
func (m launchdbsystembase) GetDisplayName() *string {
	return m.DisplayName
}

//GetDomain returns Domain
func (m launchdbsystembase) GetDomain() *string {
	return m.Domain
}

//GetFreeformTags returns FreeformTags
func (m launchdbsystembase) GetFreeformTags() map[string]string {
	return m.FreeformTags
}

//GetInitialDataStorageSizeInGB returns InitialDataStorageSizeInGB
func (m launchdbsystembase) GetInitialDataStorageSizeInGB() *int {
	return m.InitialDataStorageSizeInGB
}

//GetNodeCount returns NodeCount
func (m launchdbsystembase) GetNodeCount() *int {
	return m.NodeCount
}

func (m launchdbsystembase) String() string {
	return common.PointerString(m)
}

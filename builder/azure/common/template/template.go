package template

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-01-01/network"
)

/////////////////////////////////////////////////
// Template
type Template struct {
	Schema         *string                `json:"$schema"`
	ContentVersion *string                `json:"contentVersion"`
	Parameters     *map[string]Parameters `json:"parameters"`
	Variables      *map[string]string     `json:"variables"`
	Resources      *[]Resource            `json:"resources"`
}

/////////////////////////////////////////////////
// Template > Parameters
type Parameters struct {
	Type         *string `json:"type"`
	DefaultValue *string `json:"defaultValue,omitempty"`
}

/////////////////////////////////////////////////
// Template > Resource
type Resource struct {
	ApiVersion *string             `json:"apiVersion"`
	Name       *string             `json:"name"`
	Type       *string             `json:"type"`
	Location   *string             `json:"location,omitempty"`
	DependsOn  *[]string           `json:"dependsOn,omitempty"`
	Plan       *Plan               `json:"plan,omitempty"`
	Properties *Properties         `json:"properties,omitempty"`
	Tags       *map[string]*string `json:"tags,omitempty"`
	Resources  *[]Resource         `json:"resources,omitempty"`
}

type Plan struct {
	Name          *string `json:"name"`
	Product       *string `json:"product"`
	Publisher     *string `json:"publisher"`
	PromotionCode *string `json:"promotionCode,omitempty"`
}

type OSDiskUnion struct {
	OsType       compute.OperatingSystemTypes      `json:"osType,omitempty"`
	OsState      compute.OperatingSystemStateTypes `json:"osState,omitempty"`
	BlobURI      *string                           `json:"blobUri,omitempty"`
	Name         *string                           `json:"name,omitempty"`
	Vhd          *compute.VirtualHardDisk          `json:"vhd,omitempty"`
	Image        *compute.VirtualHardDisk          `json:"image,omitempty"`
	Caching      compute.CachingTypes              `json:"caching,omitempty"`
	CreateOption compute.DiskCreateOptionTypes     `json:"createOption,omitempty"`
	DiskSizeGB   *int32                            `json:"diskSizeGB,omitempty"`
	ManagedDisk  *compute.ManagedDiskParameters    `json:"managedDisk,omitempty"`
}

type DataDiskUnion struct {
	Lun          *int                           `json:"lun,omitempty"`
	BlobURI      *string                        `json:"blobUri,omitempty"`
	Name         *string                        `json:"name,omitempty"`
	Vhd          *compute.VirtualHardDisk       `json:"vhd,omitempty"`
	Image        *compute.VirtualHardDisk       `json:"image,omitempty"`
	Caching      compute.CachingTypes           `json:"caching,omitempty"`
	CreateOption compute.DiskCreateOptionTypes  `json:"createOption,omitempty"`
	DiskSizeGB   *int32                         `json:"diskSizeGB,omitempty"`
	ManagedDisk  *compute.ManagedDiskParameters `json:"managedDisk,omitempty"`
}

// Union of the StorageProfile and ImageStorageProfile types.
type StorageProfileUnion struct {
	ImageReference *compute.ImageReference `json:"imageReference,omitempty"`
	OsDisk         *OSDiskUnion            `json:"osDisk,omitempty"`
	DataDisks      *[]DataDiskUnion        `json:"dataDisks,omitempty"`
}

/////////////////////////////////////////////////
// Template > Resource > Properties
type Properties struct {
	AccessPolicies               *[]AccessPolicies                   `json:"accessPolicies,omitempty"`
	AddressSpace                 *network.AddressSpace               `json:"addressSpace,omitempty"`
	DiagnosticsProfile           *compute.DiagnosticsProfile         `json:"diagnosticsProfile,omitempty"`
	DNSSettings                  *network.PublicIPAddressDNSSettings `json:"dnsSettings,omitempty"`
	EnabledForDeployment         *string                             `json:"enabledForDeployment,omitempty"`
	EnabledForTemplateDeployment *string                             `json:"enabledForTemplateDeployment,omitempty"`
	HardwareProfile              *compute.HardwareProfile            `json:"hardwareProfile,omitempty"`
	IPConfigurations             *[]network.IPConfiguration          `json:"ipConfigurations,omitempty"`
	NetworkProfile               *compute.NetworkProfile             `json:"networkProfile,omitempty"`
	OsProfile                    *compute.OSProfile                  `json:"osProfile,omitempty"`
	PublicIPAllocatedMethod      *network.IPAllocationMethod         `json:"publicIPAllocationMethod,omitempty"`
	Sku                          *Sku                                `json:"sku,omitempty"`
	//StorageProfile3              *compute.StorageProfile             `json:"storageProfile,omitempty"`
	StorageProfile *StorageProfileUnion    `json:"storageProfile,omitempty"`
	Subnets        *[]network.Subnet       `json:"subnets,omitempty"`
	SecurityRules  *[]network.SecurityRule `json:"securityRules,omitempty"`
	TenantId       *string                 `json:"tenantId,omitempty"`
	Value          *string                 `json:"value,omitempty"`
}

type AccessPolicies struct {
	ObjectId    *string      `json:"objectId,omitempty"`
	TenantId    *string      `json:"tenantId,omitempty"`
	Permissions *Permissions `json:"permissions,omitempty"`
}

type Permissions struct {
	Keys    *[]string `json:"keys,omitempty"`
	Secrets *[]string `json:"secrets,omitempty"`
}

type Sku struct {
	Family *string `json:"family,omitempty"`
	Name   *string `json:"name,omitempty"`
}

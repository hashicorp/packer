package v3

import "time"

// Reference ...
type Reference struct {
	Kind *string `mapstructure:"kind" json:"kind"`
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`
	UUID *string `mapstructure:"uuid" json:"uuid"`
}

// VMVnumaConfig Indicates how VM vNUMA should be configured
type VMVnumaConfig struct {

	// Number of vNUMA nodes. 0 means vNUMA is disabled.
	NumVnumaNodes *int64 `mapstructure:"num_vnuma_nodes,omitempty" json:"num_vnuma_nodes,omitempty"`
}

// IPAddress An IP address.
type IPAddress struct {

	// Address *string.
	IP *string `mapstructure:"ip,omitempty" json:"ip,omitempty"`

	// Address type. It can only be \"ASSIGNED\" in the spec. If no type is specified in the spec, the default type is
	// set to \"ASSIGNED\".
	Type *string `mapstructure:"type,omitempty" json:"type,omitempty"`
}

// VMNic Virtual Machine NIC.
type VMNic struct {

	// IP endpoints for the adapter. Currently, IPv4 addresses are supported.
	IPEndpointList []*IPAddress `mapstructure:"ip_endpoint_list,omitempty" json:"ip_endpoint_list,omitempty"`

	// The MAC address for the adapter.
	MacAddress *string `mapstructure:"mac_address,omitempty" json:"mac_address,omitempty"`

	// The model of this NIC.
	Model *string `mapstructure:"model,omitempty" json:"model,omitempty"`

	NetworkFunctionChainReference *Reference `mapstructure:"network_function_chain_reference,omitempty" json:"network_function_chain_reference,omitempty"`

	// The type of this Network function NIC. Defaults to INGRESS.
	NetworkFunctionNicType *string `mapstructure:"network_function_nic_type,omitempty" json:"network_function_nic_type,omitempty"`

	// The type of this NIC. Defaults to NORMAL_NIC.
	NicType *string `mapstructure:"nic_type,omitempty" json:"nic_type,omitempty"`

	SubnetReference *Reference `mapstructure:"subnet_reference,omitempty" json:"subnet_reference,omitempty"`

	// The NIC's UUID, which is used to uniquely identify this particular NIC. This UUID may be used to refer to the NIC
	// outside the context of the particular VM it is attached to.
	UUID *string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
}

// DiskAddress Disk Address.
type DiskAddress struct {
	AdapterType *string `mapstructure:"adapter_type,omitempty" json:"adapter_type,omitempty"`
	DeviceIndex *int64  `mapstructure:"device_index,omitempty" json:"device_index,omitempty"`
}

// VMBootDevice Indicates which device a VM should boot from. One of disk_address or mac_address should be provided.
type VMBootDevice struct {

	// Address of disk to boot from.
	DiskAddress *DiskAddress `mapstructure:"disk_address,omitempty" json:"disk_address,omitempty"`

	// MAC address of nic to boot from.
	MacAddress *string `mapstructure:"mac_address,omitempty" json:"mac_address,omitempty"`
}

// VMBootConfig Indicates which device a VM should boot from.
type VMBootConfig struct {

	// Indicates which device a VM should boot from. Boot device takes precdence over boot device order. If both are
	// given then specified boot device will be primary boot device and remaining devices will be assigned boot order
	// according to boot device order field.
	BootDevice *VMBootDevice `mapstructure:"boot_device,omitempty" json:"boot_device,omitempty"`

	// Indicates the order of device types in which VM should try to boot from. If boot device order is not provided the
	// system will decide appropriate boot device order.
	BootDeviceOrderList []*string `mapstructure:"boot_device_order_list,omitempty" json:"boot_device_order_list,omitempty"`
}

// NutanixGuestToolsSpec Information regarding Nutanix Guest Tools.
type NutanixGuestToolsSpec struct {

	// Application names that are enabled.
	EnabledCapabilityList []*string `mapstructure:"enabled_capability_list,omitempty" json:"enabled_capability_list,omitempty"`

	// Desired mount state of Nutanix Guest Tools ISO.
	IsoMountState *string `mapstructure:"iso_mount_state,omitempty" json:"iso_mount_state,omitempty"`

	// Nutanix Guest Tools is enabled or not.
	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`
}

// GuestToolsSpec Information regarding guest tools.
type GuestToolsSpec struct {

	// Nutanix Guest Tools information
	NutanixGuestTools *NutanixGuestToolsSpec `mapstructure:"nutanix_guest_tools,omitempty" json:"nutanix_guest_tools,omitempty"`
}

// VMGpu Graphics resource information for the Virtual Machine.
type VMGpu struct {

	// The device ID of the GPU.
	DeviceID *int64 `mapstructure:"device_id,omitempty" json:"device_id,omitempty"`

	// The mode of this GPU.
	Mode *string `mapstructure:"mode,omitempty" json:"mode,omitempty"`

	// The vendor of the GPU.
	Vendor *string `mapstructure:"vendor,omitempty" json:"vendor,omitempty"`
}

// GuestCustomizationCloudInit If this field is set, the guest will be customized using cloud-init. Either user_data or
// custom_key_values should be provided. If custom_key_ves are provided then the user data will be generated using these
// key-value pairs.
type GuestCustomizationCloudInit struct {

	// Generic key value pair used for custom attributes
	CustomKeyValues map[string]string `mapstructure:"custom_key_values,omitempty" json:"custom_key_values,omitempty"`

	// The contents of the meta_data configuration for cloud-init. This can be formatted as YAML or JSON. The value must
	// be base64 encoded.
	MetaData *string `mapstructure:"meta_data,omitempty" json:"meta_data,omitempty"`

	// The contents of the user_data configuration for cloud-init. This can be formatted as YAML, JSON, or could be a
	// shell script. The value must be base64 encoded.
	UserData *string `mapstructure:"user_data,omitempty" json:"user_data,omitempty"`
}

// GuestCustomizationSysprep If this field is set, the guest will be customized using Sysprep. Either unattend_xml or
// custom_key_values should be provided. If custom_key_values are provided then the unattended answer file will be
// generated using these key-value pairs.
type GuestCustomizationSysprep struct {

	// Generic key value pair used for custom attributes
	CustomKeyValues map[string]string `mapstructure:"custom_key_values,omitempty" json:"custom_key_values,omitempty"`

	// Whether the guest will be freshly installed using this unattend configuration, or whether this unattend
	// configuration will be applied to a pre-prepared image. Default is \"PREPARED\".
	InstallType *string `mapstructure:"install_type,omitempty" json:"install_type,omitempty"`

	// This field contains a Sysprep unattend xml definition, as a *string. The value must be base64 encoded.
	UnattendXML *string `mapstructure:"unattend_xml,omitempty" json:"unattend_xml,omitempty"`
}

// GuestCustomization VM guests may be customized at boot time using one of several different methods. Currently,
// cloud-init w/ ConfigDriveV2 (for Linux VMs) and Sysprep (for Windows VMs) are supported. Only ONE OF sysprep or
// cloud_init should be provided. Note that guest customization can currently only be set during VM creation. Attempting
// to change it after creation will result in an error. Additional properties can be specified. For example - in the
// context of VM template creation if \"override_script\" is set to \"True\" then the deployer can upload their own
// custom script.
type GuestCustomization struct {
	CloudInit *GuestCustomizationCloudInit `mapstructure:"cloud_init,omitempty" json:"cloud_init,omitempty"`

	// Flag to allow override of customization by deployer.
	IsOverridable *bool `mapstructure:"is_overridable,omitempty" json:"is_overridable,omitempty"`

	Sysprep *GuestCustomizationSysprep `mapstructure:"sysprep,omitempty" json:"sysprep,omitempty"`
}

// VMGuestPowerStateTransitionConfig Extra configs related to power state transition.
type VMGuestPowerStateTransitionConfig struct {

	// Indicates whether to execute set script before ngt shutdown/reboot.
	EnableScriptExec *bool `mapstructure:"enable_script_exec,omitempty" json:"enable_script_exec,omitempty"`

	// Indicates whether to abort ngt shutdown/reboot if script fails.
	ShouldFailOnScriptFailure *bool `mapstructure:"should_fail_on_script_failure,omitempty" json:"should_fail_on_script_failure,omitempty"`
}

// VMPowerStateMechanism Indicates the mechanism guiding the VM power state transition. Currently used for the transition
// to \"OFF\" state.
type VMPowerStateMechanism struct {
	GuestTransitionConfig *VMGuestPowerStateTransitionConfig `mapstructure:"guest_transition_config,omitempty" json:"guest_transition_config,omitempty"`

	// Power state mechanism (ACPI/GUEST/HARD).
	Mechanism *string `mapstructure:"mechanism,omitempty" json:"mechanism,omitempty"`
}

// VMDiskDeviceProperties ...
type VMDiskDeviceProperties struct {
	DeviceType  *string      `mapstructure:"device_type,omitempty" json:"device_type,omitempty"`
	DiskAddress *DiskAddress `mapstructure:"disk_address,omitempty" json:"disk_address,omitempty"`
}

// VMDisk VirtualMachine Disk (VM Disk).
type VMDisk struct {
	DataSourceReference *Reference `mapstructure:"data_source_reference,omitempty" json:"data_source_reference,omitempty"`

	DeviceProperties *VMDiskDeviceProperties `mapstructure:"device_properties,omitempty" json:"device_properties,omitempty"`

	// Size of the disk in Bytes.
	DiskSizeBytes *int64 `mapstructure:"disk_size_bytes,omitempty" json:"disk_size_bytes,omitempty"`

	// Size of the disk in MiB. Must match the size specified in 'disk_size_bytes' - rounded up to the nearest MiB -
	// when that field is present.
	DiskSizeMib *int64 `mapstructure:"disk_size_mib,omitempty" json:"disk_size_mib,omitempty"`

	// The device ID which is used to uniquely identify this particular disk.
	UUID *string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`

	VolumeGroupReference *Reference `mapstructure:"volume_group_reference,omitempty" json:"volume_group_reference,omitempty"`
}

// VMResources VM Resources Definition.
type VMResources struct {

	// Indicates which device the VM should boot from.
	BootConfig *VMBootConfig `mapstructure:"boot_config,omitempty" json:"boot_config,omitempty"`

	// Disks attached to the VM.
	DiskList []*VMDisk `mapstructure:"disk_list,omitempty" json:"disk_list,omitempty"`

	// GPUs attached to the VM.
	GpuList []*VMGpu `mapstructure:"gpu_list,omitempty" json:"gpu_list,omitempty"`

	GuestCustomization *GuestCustomization `json:"guest_customization,omitempty" mapstructure:"guest_customization,omitempty"`

	// Guest OS Identifier. For ESX, refer to VMware documentation link
	// https://www.vmware.com/support/orchestrator/doc/vro-vsphere65-api/html/VcVirtualMachineGuestOsIdentifier.html
	// for the list of guest OS identifiers.
	GuestOsID *string `mapstructure:"guest_os_id,omitempty" json:"guest_os_id,omitempty"`

	// Information regarding guest tools.
	GuestTools *GuestToolsSpec `mapstructure:"guest_tools,omitempty" json:"guest_tools,omitempty"`

	// VM's hardware clock timezone in IANA TZDB format (America/Los_Angeles).
	HardwareClockTimezone *string `mapstructure:"hardware_clock_timezone,omitempty" json:"hardware_clock_timezone,omitempty"`

	// Memory size in MiB.
	MemorySizeMib *int64 `mapstructure:"memory_size_mib,omitempty" json:"memory_size_mib,omitempty"`

	// NICs attached to the VM.
	NicList []*VMNic `mapstructure:"nic_list,omitempty" json:"nic_list,omitempty"`

	// Number of threads per core
	NumThreads *int64 `mapstructure:"num_threads_per_core,omitempty" json:"num_threads_per_core,omitempty"`

	// Number of vCPU sockets.
	NumSockets *int64 `mapstructure:"num_sockets,omitempty" json:"num_sockets,omitempty"`

	// Number of vCPUs per socket.
	NumVcpusPerSocket *int64 `mapstructure:"num_vcpus_per_socket,omitempty" json:"num_vcpus_per_socket,omitempty"`

	// *Reference to an entity that the VM should be cloned from.
	ParentReference *Reference `mapstructure:"parent_reference,omitempty" json:"parent_reference,omitempty"`

	// The current or desired power state of the VM.
	PowerState *string `mapstructure:"power_state,omitempty" json:"power_state,omitempty"`

	PowerStateMechanism *VMPowerStateMechanism `mapstructure:"power_state_mechanism,omitempty" json:"power_state_mechanism,omitempty"`

	// Indicates whether VGA console should be enabled or not.
	VgaConsoleEnabled *bool `mapstructure:"vga_console_enabled,omitempty" json:"vga_console_enabled,omitempty"`

	// Information regarding vNUMA configuration.
	VMVnumaConfig *VMVnumaConfig `mapstructure:"vnuma_config,omitempty" json:"vnuma_config,omitempty"`
}

// VM An intentful representation of a vm spec
type VM struct {
	AvailabilityZoneReference *Reference `mapstructure:"availability_zone_reference,omitempty" json:"availability_zone_reference,omitempty"`

	ClusterReference *Reference `mapstructure:"cluster_reference,omitempty" json:"cluster_reference,omitempty"`

	// A description for vm.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// vm Name.
	Name *string `mapstructure:"name" json:"name"`

	Resources *VMResources `mapstructure:"resources" mapstructure:"resources,omitempty" json:"resources" mapstructure:"resources,omitempty"`
}

// VMIntentInput ...
type VMIntentInput struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *Metadata `mapstructure:"metadata" json:"metadata"`

	Spec *VM `mapstructure:"spec" json:"spec"`
}

// MessageResource ...
type MessageResource struct {

	// Custom key-value details relevant to the status.
	Details map[string]string `mapstructure:"details,omitempty" json:"details,omitempty"`

	// If state is ERROR, a message describing the error.
	Message *string `mapstructure:"message" json:"message"`

	// If state is ERROR, a machine-readable snake-cased *string.
	Reason *string `mapstructure:"reason" json:"reason"`
}

// VMStatus The status of a REST API call. Only used when there is a failure to report.
type VMStatus struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// The HTTP error code.
	Code *int64 `mapstructure:"code,omitempty" json:"code,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`

	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`
}

// VMNicOutputStatus Virtual Machine NIC Status.
type VMNicOutputStatus struct {

	// The Floating IP associated with the vnic.
	FloatingIP *string `mapstructure:"floating_ip,omitempty" json:"floating_ip,omitempty"`

	// IP endpoints for the adapter. Currently, IPv4 addresses are supported.
	IPEndpointList []*IPAddress `mapstructure:"ip_endpoint_list,omitempty" json:"ip_endpoint_list,omitempty"`

	// The MAC address for the adapter.
	MacAddress *string `mapstructure:"mac_address,omitempty" json:"mac_address,omitempty"`

	// The model of this NIC.
	Model *string `mapstructure:"model,omitempty" json:"model,omitempty"`

	NetworkFunctionChainReference *Reference `mapstructure:"network_function_chain_reference,omitempty" json:"network_function_chain_reference,omitempty"`

	// The type of this Network function NIC. Defaults to INGRESS.
	NetworkFunctionNicType *string `mapstructure:"network_function_nic_type,omitempty" json:"network_function_nic_type,omitempty"`

	// The type of this NIC. Defaults to NORMAL_NIC.
	NicType *string `mapstructure:"nic_type,omitempty" json:"nic_type,omitempty"`

	SubnetReference *Reference `mapstructure:"subnet_reference,omitempty" json:"subnet_reference,omitempty"`

	// The NIC's UUID, which is used to uniquely identify this particular NIC. This UUID may be used to refer to the NIC
	// outside the context of the particular VM it is attached to.
	UUID *string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
}

// NutanixGuestToolsStatus Information regarding Nutanix Guest Tools.
type NutanixGuestToolsStatus struct {

	// Version of Nutanix Guest Tools available on the cluster.
	AvailableVersion *string `mapstructure:"available_version,omitempty" json:"available_version,omitempty"`

	// Application names that are enabled.
	EnabledCapabilityList []*string `mapstructure:"enabled_capability_list,omitempty" json:"enabled_capability_list,omitempty"`

	// Version of the operating system on the VM.
	GuestOsVersion *string `mapstructure:"guest_os_version,omitempty" json:"guest_os_version,omitempty"`

	// Communication from VM to CVM is active or not.
	IsReachable *bool `mapstructure:"is_reachable,omitempty" json:"is_reachable,omitempty"`

	// Desired mount state of Nutanix Guest Tools ISO.
	IsoMountState *string `mapstructure:"iso_mount_state,omitempty" json:"iso_mount_state,omitempty"`

	// Nutanix Guest Tools is enabled or not.
	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`

	// Version of Nutanix Guest Tools installed on the VM.
	Version *string `mapstructure:"version,omitempty" json:"version,omitempty"`

	// Whether VM mobility drivers are installed in the VM.
	VMMobilityDriversInstalled *bool `mapstructure:"vm_mobility_drivers_installed,omitempty" json:"vm_mobility_drivers_installed,omitempty"`

	// Whether the VM is configured to take VSS snapshots through NGT.
	VSSSnapshotCapable *bool `mapstructure:"vss_snapshot_capable,omitempty" json:"vss_snapshot_capable,omitempty"`
}

// GuestToolsStatus Information regarding guest tools.
type GuestToolsStatus struct {

	// Nutanix Guest Tools information
	NutanixGuestTools *NutanixGuestToolsStatus `mapstructure:"nutanix_guest_tools,omitempty" json:"nutanix_guest_tools,omitempty"`
}

// VMGpuOutputStatus Graphics resource status information for the Virtual Machine.
type VMGpuOutputStatus struct {

	// The device ID of the GPU.
	DeviceID *int64 `mapstructure:"device_id,omitempty" json:"device_id,omitempty"`

	// Fraction of the physical GPU assigned.
	Fraction *int64 `mapstructure:"fraction,omitempty" json:"fraction,omitempty"`

	// GPU frame buffer size in MiB.
	FrameBufferSizeMib *int64 `mapstructure:"frame_buffer_size_mib,omitempty" json:"frame_buffer_size_mib,omitempty"`

	// Last determined guest driver version.
	GuestDriverVersion *string `mapstructure:"guest_driver_version,omitempty" json:"guest_driver_version,omitempty"`

	// The mode of this GPU
	Mode *string `mapstructure:"mode,omitempty" json:"mode,omitempty"`

	// Name of the GPU resource.
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`

	// Number of supported virtual display heads.
	NumVirtualDisplayHeads *int64 `mapstructure:"num_virtual_display_heads,omitempty" json:"num_virtual_display_heads,omitempty"`

	// GPU {segment:bus:device:function} (sbdf) address if assigned.
	PCIAddress *string `mapstructure:"pci_address,omitempty" json:"pci_address,omitempty"`

	// UUID of the GPU.
	UUID *string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`

	// The vendor of the GPU.
	Vendor *string `mapstructure:"vendor,omitempty" json:"vendor,omitempty"`
}

// GuestCustomizationStatus VM guests may be customized at boot time using one of several different methods. Currently,
// cloud-init w/ ConfigDriveV2 (for Linux VMs) and Sysprep (for Windows VMs) are supported. Only ONE OF sysprep or
// cloud_init should be provided. Note that guest customization can currently only be set during VM creation. Attempting
// to change it after creation will result in an error. Additional properties can be specified. For example - in the
// context of VM template creation if \"override_script\" is set to \"True\" then the deployer can upload their own
// custom script.
type GuestCustomizationStatus struct {
	CloudInit *GuestCustomizationCloudInit `mapstructure:"cloud_init,omitempty" json:"cloud_init,omitempty"`

	// Flag to allow override of customization by deployer.
	IsOverridable *bool `mapstructure:"is_overridable,omitempty" json:"is_overridable,omitempty"`

	Sysprep *GuestCustomizationSysprep `mapstructure:"sysprep,omitempty" json:"sysprep,omitempty"`
}

// VMResourcesDefStatus VM Resources Status Definition.
type VMResourcesDefStatus struct {

	// Indicates which device the VM should boot from.
	BootConfig *VMBootConfig `mapstructure:"boot_config,omitempty" json:"boot_config,omitempty"`

	// Disks attached to the VM.
	DiskList []*VMDisk `mapstructure:"disk_list,omitempty" json:"disk_list,omitempty"`

	// GPUs attached to the VM.
	GpuList []*VMGpuOutputStatus `mapstructure:"gpu_list,omitempty" json:"gpu_list,omitempty"`

	GuestCustomization *GuestCustomizationStatus `mapstructure:"guest_customization,omitempty" json:"guest_customization,omitempty"`

	// Guest OS Identifier. For ESX, refer to VMware documentation link
	// https://www.vmware.com/support/orchestrator/doc/vro-vsphere65-api/html/VcVirtualMachineGuestOsIdentifier.html
	// for the list of guest OS identifiers.
	GuestOsID *string `mapstructure:"guest_os_id,omitempty" json:"guest_os_id,omitempty"`

	// Information regarding guest tools.
	GuestTools *GuestToolsStatus `mapstructure:"guest_tools,omitempty" json:"guest_tools,omitempty"`

	// VM's hardware clock timezone in IANA TZDB format (America/Los_Angeles).
	HardwareClockTimezone *string `mapstructure:"hardware_clock_timezone,omitempty" json:"hardware_clock_timezone,omitempty"`

	HostReference *Reference `mapstructure:"host_reference,omitempty" json:"host_reference,omitempty"`

	// The hypervisor type for the hypervisor the VM is hosted on.
	HypervisorType *string `mapstructure:"hypervisor_type,omitempty" json:"hypervisor_type,omitempty"`

	// Memory size in MiB.
	MemorySizeMib *int64 `mapstructure:"memory_size_mib,omitempty" json:"memory_size_mib,omitempty"`

	// NICs attached to the VM.
	NicList []*VMNicOutputStatus `mapstructure:"nic_list,omitempty" json:"nic_list,omitempty"`

	// Number of vCPU sockets.
	NumSockets *int64 `mapstructure:"num_sockets,omitempty" json:"num_sockets,omitempty"`

	// Number of vCPUs per socket.
	NumVcpusPerSocket *int64 `mapstructure:"num_vcpus_per_socket,omitempty" json:"num_vcpus_per_socket,omitempty"`

	// *Reference to an entity that the VM cloned from.
	ParentReference *Reference `mapstructure:"parent_reference,omitempty" json:"parent_reference,omitempty"`

	// Current power state of the VM.
	PowerState *string `mapstructure:"power_state,omitempty" json:"power_state,omitempty"`

	PowerStateMechanism *VMPowerStateMechanism `mapstructure:"power_state_mechanism,omitempty" json:"power_state_mechanism,omitempty"`

	// Indicates whether VGA console has been enabled or not.
	VgaConsoleEnabled *bool `mapstructure:"vga_console_enabled,omitempty" json:"vga_console_enabled,omitempty"`

	// Information regarding vNUMA configuration.
	VnumaConfig *VMVnumaConfig `mapstructure:"vnuma_config,omitempty" json:"vnuma_config,omitempty"`
}

// VMDefStatus An intentful representation of a vm status
type VMDefStatus struct {
	AvailabilityZoneReference *Reference `mapstructure:"availability_zone_reference,omitempty" json:"availability_zone_reference,omitempty"`

	ClusterReference *Reference `mapstructure:"cluster_reference,omitempty" json:"cluster_reference,omitempty"`

	// A description for vm.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// Any error messages for the vm, if in an error state.
	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`

	// vm Name.
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`

	Resources *VMResourcesDefStatus `mapstructure:"resources,omitempty" json:"resources,omitempty"`

	// The state of the vm.
	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`

	ExecutionContext *ExecutionContext `mapstructure:"execution_context,omitempty" json:"execution_context,omitempty"`
}

//ExecutionContext ...
type ExecutionContext struct {
	TaskUUID interface{} `mapstructure:"task_uuid,omitempty" json:"task_uuid,omitempty"`
}

// VMIntentResponse Response object for intentful operations on a vm
type VMIntentResponse struct {
	APIVersion *string `mapstructure:"api_version" json:"api_version"`

	Metadata *Metadata `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`

	Spec *VM `mapstructure:"spec,omitempty" json:"spec,omitempty"`

	Status *VMDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// DSMetadata All api calls that return a list will have this metadata block as input
type DSMetadata struct {

	// The filter in FIQL syntax used for the results.
	Filter *string `mapstructure:"filter,omitempty" json:"filter,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	// The number of records to retrieve relative to the offset
	Length *int64 `mapstructure:"length,omitempty" json:"length,omitempty"`

	// Offset from the start of the entity list
	Offset *int64 `mapstructure:"offset,omitempty" json:"offset,omitempty"`

	// The attribute to perform sort on
	SortAttribute *string `mapstructure:"sort_attribute,omitempty" json:"sort_attribute,omitempty"`

	// The sort order in which results are returned
	SortOrder *string `mapstructure:"sort_order,omitempty" json:"sort_order,omitempty"`
}

// VMIntentResource Response object for intentful operations on a vm
type VMIntentResource struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *Metadata `mapstructure:"metadata" json:"metadata"`

	Spec *VM `mapstructure:"spec,omitempty" json:"spec,omitempty"`

	Status *VMDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// VMListIntentResponse Response object for intentful operation of vms
type VMListIntentResponse struct {
	APIVersion *string `mapstructure:"api_version" json:"api_version"`

	Entities []*VMIntentResource `mapstructure:"entities,omitempty" json:"entities,omitempty"`

	Metadata *ListMetadataOutput `mapstructure:"metadata" json:"metadata"`
}

// SubnetMetadata The subnet kind metadata
type SubnetMetadata struct {

	// Categories for the subnet
	Categories map[string]string `mapstructure:"categories,omitempty" json:"categories,omitempty"`

	// UTC date and time in RFC-3339 format when subnet was created
	CreationTime *time.Time `mapstructure:"creation_time,omitempty" json:"creation_time,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind" json:"kind"`

	// UTC date and time in RFC-3339 format when subnet was last updated
	LastUpdateTime *time.Time `mapstructure:"last_update_time,omitempty" json:"last_update_time,omitempty"`

	// subnet name
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`

	OwnerReference *Reference `mapstructure:"owner_reference,omitempty" json:"owner_reference,omitempty"`

	// project reference
	ProjectReference *Reference `mapstructure:"project_reference,omitempty" json:"project_reference,omitempty"`

	// Hash of the spec. This will be returned from server.
	SpecHash *string `mapstructure:"spec_hash,omitempty" json:"spec_hash,omitempty"`

	// Version number of the latest spec.
	SpecVersion *int64 `mapstructure:"spec_version,omitempty" json:"spec_version,omitempty"`

	// subnet uuid
	UUID *string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
}

// Address represents the Host address.
type Address struct {

	// Fully qualified domain name.
	FQDN *string `mapstructure:"fqdn,omitempty" json:"fqdn,omitempty"`

	// IPV4 address.
	IP *string `mapstructure:"ip,omitempty" json:"ip,omitempty"`

	// IPV6 address.
	IPV6 *string `mapstructure:"ipv6,omitempty" json:"ipv6,omitempty"`

	// Port Number
	Port *int64 `mapstructure:"port,omitempty" json:"port,omitempty"`
}

// IPPool represents IP pool.
type IPPool struct {

	// Range of IPs (example: 10.0.0.9 10.0.0.19).
	Range *string `mapstructure:"range,omitempty" json:"range,omitempty"`
}

// DHCPOptions Spec for defining DHCP options.
type DHCPOptions struct {
	BootFileName *string `mapstructure:"boot_file_name,omitempty" json:"boot_file_name,omitempty"`

	DomainName *string `mapstructure:"domain_name,omitempty" json:"domain_name,omitempty"`

	DomainNameServerList []*string `mapstructure:"domain_name_server_list,omitempty" json:"domain_name_server_list,omitempty"`

	DomainSearchList []*string `mapstructure:"domain_search_list,omitempty" json:"domain_search_list,omitempty"`

	TFTPServerName *string `mapstructure:"tftp_server_name,omitempty" json:"tftp_server_name,omitempty"`
}

// IPConfig represents the configurtion of IP.
type IPConfig struct {

	// Default gateway IP address.
	DefaultGatewayIP *string `mapstructure:"default_gateway_ip,omitempty" json:"default_gateway_ip,omitempty"`

	DHCPOptions *DHCPOptions `mapstructure:"dhcp_options,omitempty" json:"dhcp_options,omitempty"`

	DHCPServerAddress *Address `mapstructure:"dhcp_server_address,omitempty" json:"dhcp_server_address,omitempty"`

	PoolList []*IPPool `mapstructure:"pool_list,omitempty" json:"pool_list,omitempty"`

	PrefixLength *int64 `mapstructure:"prefix_length,omitempty" json:"prefix_length,omitempty"`

	// Subnet IP address.
	SubnetIP *string `mapstructure:"subnet_ip,omitempty" json:"subnet_ip,omitempty"`
}

// SubnetResources represents Subnet creation/modification spec.
type SubnetResources struct {
	IPConfig *IPConfig `mapstructure:"ip_config,omitempty" json:"ip_config,omitempty"`

	NetworkFunctionChainReference *Reference `mapstructure:"network_function_chain_reference,omitempty" json:"network_function_chain_reference,omitempty"`

	SubnetType *string `mapstructure:"subnet_type" json:"subnet_type"`

	VlanID *int64 `mapstructure:"vlan_id,omitempty" json:"vlan_id,omitempty"`

	VswitchName *string `mapstructure:"vswitch_name,omitempty" json:"vswitch_name,omitempty"`
}

// Subnet An intentful representation of a subnet spec
type Subnet struct {
	AvailabilityZoneReference *Reference `mapstructure:"availability_zone_reference,omitempty" json:"availability_zone_reference,omitempty"`

	ClusterReference *Reference `mapstructure:"cluster_reference,omitempty" json:"cluster_reference,omitempty"`

	// A description for subnet.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// subnet Name.
	Name *string `mapstructure:"name" json:"name"`

	Resources *SubnetResources `mapstructure:"resources,omitempty" json:"resources,omitempty"`
}

// SubnetIntentInput An intentful representation of a subnet
type SubnetIntentInput struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *Metadata `mapstructure:"metadata" json:"metadata"`

	Spec *Subnet `mapstructure:"spec" json:"spec"`
}

// SubnetStatus represents The status of a REST API call. Only used when there is a failure to report.
type SubnetStatus struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// The HTTP error code.
	Code *int64 `mapstructure:"code,omitempty" json:"code,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`

	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`
}

// SubnetResourcesDefStatus represents a Subnet creation/modification status.
type SubnetResourcesDefStatus struct {
	IPConfig *IPConfig `mapstructure:"ip_config,omitempty" json:"ip_config,omitempty"`

	NetworkFunctionChainReference *Reference `mapstructure:"network_function_chain_reference,omitempty" json:"network_function_chain_reference,omitempty"`

	SubnetType *string `mapstructure:"subnet_type" json:"subnet_type"`

	VlanID *int64 `mapstructure:"vlan_id,omitempty" json:"vlan_id,omitempty"`

	VswitchName *string `mapstructure:"vswitch_name,omitempty" json:"vswitch_name,omitempty"`
}

// SubnetDefStatus An intentful representation of a subnet status
type SubnetDefStatus struct {
	AvailabilityZoneReference *Reference `mapstructure:"availability_zone_reference,omitempty" json:"availability_zone_reference,omitempty"`

	ClusterReference *Reference `mapstructure:"cluster_reference,omitempty" json:"cluster_reference,omitempty"`

	// A description for subnet.
	Description *string `mapstructure:"description" json:"description"`

	// Any error messages for the subnet, if in an error state.
	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`

	// subnet Name.
	Name *string `mapstructure:"name" json:"name"`

	Resources *SubnetResourcesDefStatus `mapstructure:"resources,omitempty" json:"resources,omitempty"`

	// The state of the subnet.
	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`

	ExecutionContext *ExecutionContext `mapstructure:"execution_context,omitempty" json:"execution_context,omitempty"`
}

// SubnetIntentResponse represents the response object for intentful operations on a subnet
type SubnetIntentResponse struct {
	APIVersion *string `mapstructure:"api_version" json:"api_version"`

	Metadata *Metadata `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`

	Spec *Subnet `mapstructure:"spec,omitempty" json:"spec,omitempty"`

	Status *SubnetDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// SubnetIntentResource represents Response object for intentful operations on a subnet
type SubnetIntentResource struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *Metadata `mapstructure:"metadata" json:"metadata"`

	Spec *Subnet `mapstructure:"spec,omitempty" json:"spec,omitempty"`

	Status *SubnetDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// SubnetListIntentResponse represents the response object for intentful operation of subnets
type SubnetListIntentResponse struct {
	APIVersion *string `mapstructure:"api_version" json:"api_version"`

	Entities []*SubnetIntentResponse `mapstructure:"entities,omitempty" json:"entities,omitempty"`

	Metadata *ListMetadataOutput `mapstructure:"metadata" json:"metadata"`
}

// SubnetListMetadata ...
type SubnetListMetadata struct {

	// The filter in FIQL syntax used for the results.
	Filter *string `mapstructure:"filter,omitempty" json:"filter,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	// The number of records to retrieve relative to the offset
	Length *int64 `mapstructure:"length,omitempty" json:"length,omitempty"`

	// Offset from the start of the entity list
	Offset *int64 `mapstructure:"offset,omitempty" json:"offset,omitempty"`

	// The attribute to perform sort on
	SortAttribute *string `mapstructure:"sort_attribute,omitempty" json:"sort_attribute,omitempty"`

	// The sort order in which results are returned
	SortOrder *string `mapstructure:"sort_order,omitempty" json:"sort_order,omitempty"`
}

// Checksum represents the image checksum
type Checksum struct {
	ChecksumAlgorithm *string `mapstructure:"checksum_algorithm" json:"checksum_algorithm"`
	ChecksumValue     *string `mapstructure:"checksum_value" json:"checksum_value"`
}

// ImageVersionResources The image version, which is composed of a product name and product version.
type ImageVersionResources struct {

	// Name of the producer/distribution of the image. For example windows or red hat.
	ProductName *string `mapstructure:"product_name" json:"product_name"`

	// Version *string for the disk image.
	ProductVersion *string `mapstructure:"product_version" json:"product_version"`
}

// ImageResources describes the image spec resources object.
type ImageResources struct {

	// The supported CPU architecture for a disk image.
	Architecture *string `mapstructure:"architecture,omitempty" json:"architecture,omitempty"`

	// Checksum of the image. The checksum is used for image validation if the image has a source specified. For images
	// that do not have their source specified the checksum is generated by the image service.
	Checksum *Checksum `mapstructure:"checksum,omitempty" json:"checksum,omitempty"`

	// The type of image.
	ImageType *string `mapstructure:"image_type,omitempty" json:"image_type,omitempty"`

	// The source URI points at the location of a the source image which is used to create/update image.
	SourceURI *string `mapstructure:"source_uri,omitempty" json:"source_uri,omitempty"`

	// The image version
	Version *ImageVersionResources `mapstructure:"version,omitempty" json:"version,omitempty"`

	// Reference to the source image such as 'vm_disk
	DataSourceReference *Reference `mapstructure:"data_source_reference,omitempty" json:"data_source_reference,omitempty"`
}

// Image An intentful representation of a image spec
type Image struct {

	// A description for image.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// image Name.
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`

	Resources *ImageResources `mapstructure:"resources" json:"resources"`
}

// ImageMetadata Metadata The image kind metadata
type ImageMetadata struct {

	// Categories for the image
	Categories map[string]string `mapstructure:"categories,omitempty" json:"categories,omitempty"`

	// UTC date and time in RFC-3339 format when vm was created
	CreationTime *time.Time `mapstructure:"creation_time,omitempty" json:"creation_time,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind" json:"kind"`

	// UTC date and time in RFC-3339 format when image was last updated
	LastUpdateTime *time.Time `mapstructure:"last_update_time,omitempty" json:"last_update_time,omitempty"`

	// image name
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`

	// project reference
	ProjectReference *Reference `mapstructure:"project_reference,omitempty" json:"project_reference,omitempty"`

	OwnerReference *Reference `mapstructure:"owner_reference,omitempty" json:"owner_reference,omitempty"`

	// Hash of the spec. This will be returned from server.
	SpecHash *string `mapstructure:"spec_hash,omitempty" json:"spec_hash,omitempty"`

	// Version number of the latest spec.
	SpecVersion *int64 `mapstructure:"spec_version,omitempty" json:"spec_version,omitempty"`

	// image uuid
	UUID *string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
}

// ImageIntentInput An intentful representation of a image
type ImageIntentInput struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *Metadata `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`

	Spec *Image `mapstructure:"spec,omitempty" json:"spec,omitempty"`
}

// ImageStatus represents the status of a REST API call. Only used when there is a failure to report.
type ImageStatus struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// The HTTP error code.
	Code *int64 `mapstructure:"code,omitempty" json:"code,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`

	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`
}

// ImageVersionStatus represents the image version, which is composed of a product name and product version.
type ImageVersionStatus struct {

	// Name of the producer/distribution of the image. For example windows or red hat.
	ProductName *string `mapstructure:"product_name" json:"product_name"`

	// Version *string for the disk image.
	ProductVersion *string `mapstructure:"product_version" json:"product_version"`
}

// ImageResourcesDefStatus describes the image status resources object.
type ImageResourcesDefStatus struct {

	// The supported CPU architecture for a disk image.
	Architecture *string `mapstructure:"architecture,omitempty" json:"architecture,omitempty"`

	// Checksum of the image. The checksum is used for image validation if the image has a source specified. For images
	// that do not have their source specified the checksum is generated by the image service.
	Checksum *Checksum `mapstructure:"checksum,omitempty" json:"checksum,omitempty"`

	// The type of image.
	ImageType *string `mapstructure:"image_type,omitempty" json:"image_type,omitempty"`

	// List of URIs where the raw image data can be accessed.
	RetrievalURIList []*string `mapstructure:"retrieval_uri_list,omitempty" json:"retrieval_uri_list,omitempty"`

	// The size of the image in bytes.
	SizeBytes *int64 `mapstructure:"size_bytes,omitempty" json:"size_bytes,omitempty"`

	// The source URI points at the location of a the source image which is used to create/update image.
	SourceURI *string `mapstructure:"source_uri,omitempty" json:"source_uri,omitempty"`

	// The image version
	Version *ImageVersionStatus `mapstructure:"version,omitempty" json:"version,omitempty"`
}

// ImageDefStatus represents an intentful representation of a image status
type ImageDefStatus struct {
	AvailabilityZoneReference *Reference `mapstructure:"availability_zone_reference,omitempty" json:"availability_zone_reference,omitempty"`

	ClusterReference *Reference `mapstructure:"cluster_reference,omitempty" json:"cluster_reference,omitempty"`

	// A description for image.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// Any error messages for the image, if in an error state.
	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`

	// image Name.
	Name *string `mapstructure:"name" json:"name"`

	Resources ImageResourcesDefStatus `mapstructure:"resources" json:"resources"`

	// The state of the image.
	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`

	ExecutionContext *ExecutionContext `mapstructure:"execution_context,omitempty" json:"execution_context,omitempty"`
}

// ImageIntentResponse represents the response object for intentful operations on a image
type ImageIntentResponse struct {
	APIVersion *string `mapstructure:"api_version" json:"api_version"`

	Metadata *Metadata `mapstructure:"metadata" json:"metadata"`

	Spec *Image `mapstructure:"spec,omitempty" json:"spec,omitempty"`

	Status *ImageDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// ImageListMetadata represents metadata input
type ImageListMetadata struct {

	// The filter in FIQL syntax used for the results.
	Filter *string `mapstructure:"filter,omitempty" json:"filter,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	// The number of records to retrieve relative to the offset
	Length *int64 `mapstructure:"length,omitempty" json:"length,omitempty"`

	// Offset from the start of the entity list
	Offset *int64 `mapstructure:"offset,omitempty" json:"offset,omitempty"`

	// The attribute to perform sort on
	SortAttribute *string `mapstructure:"sort_attribute,omitempty" json:"sort_attribute,omitempty"`

	// The sort order in which results are returned
	SortOrder *string `mapstructure:"sort_order,omitempty" json:"sort_order,omitempty"`
}

// ImageIntentResource represents the response object for intentful operations on a image
type ImageIntentResource struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *Metadata `mapstructure:"metadata" json:"metadata"`

	Spec *Image `mapstructure:"spec,omitempty" json:"spec,omitempty"`

	Status *ImageDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// ImageListIntentResponse represents the response object for intentful operation of images
type ImageListIntentResponse struct {
	APIVersion *string `mapstructure:"api_version" json:"api_version"`

	Entities []*ImageIntentResponse `mapstructure:"entities,omitempty" json:"entities,omitempty"`

	Metadata *ListMetadataOutput `mapstructure:"metadata" json:"metadata"`
}

// ClusterListIntentResponse ...
type ClusterListIntentResponse struct {
	APIVersion *string                  `mapstructure:"api_version" json:"api_version"`
	Entities   []*ClusterIntentResource `mapstructure:"entities,omitempty" json:"entities,omitempty"`
	Metadata   *ListMetadataOutput      `mapstructure:"metadata" json:"metadata"`
}

// ClusterIntentResource ...
type ClusterIntentResource struct {
	APIVersion *string           `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`
	Metadata   *Metadata         `mapstructure:"metadata" json:"metadata"`
	Spec       *Cluster          `mapstructure:"spec,omitempty" json:"spec,omitempty"`
	Status     *ClusterDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// ClusterIntentResponse ...
type ClusterIntentResponse struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *Metadata `mapstructure:"metadata" json:"metadata"`

	Spec *Cluster `mapstructure:"spec,omitempty" json:"spec,omitempty"`

	Status *ClusterDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// Cluster ...
type Cluster struct {
	Name      *string          `mapstructure:"name,omitempty" json:"name,omitempty"`
	Resources *ClusterResource `mapstructure:"resources,omitempty" json:"resources,omitempty"`
}

// ClusterDefStatus ...
type ClusterDefStatus struct {
	State       *string            `mapstructure:"state,omitempty" json:"state,omitempty"`
	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`
	Name        *string            `mapstructure:"name,omitempty" json:"name,omitempty"`
	Resources   *ClusterObj        `mapstructure:"resources,omitempty" json:"resources,omitempty"`
}

// ClusterObj ...
type ClusterObj struct {
	Nodes             *ClusterNodes    `mapstructure:"nodes,omitempty" json:"nodes,omitempty"`
	Config            *ClusterConfig   `mapstructure:"config,omitempty" json:"config,omitempty"`
	Network           *ClusterNetwork  `mapstructure:"network,omitempty" json:"network,omitempty"`
	Analysis          *ClusterAnalysis `mapstructure:"analysis,omitempty" json:"analysis,omitempty"`
	RuntimeStatusList []*string        `mapstructure:"runtime_status_list,omitempty" json:"runtime_status_list,omitempty"`
}

// ClusterNodes ...
type ClusterNodes struct {
	HypervisorServerList []*HypervisorServer `mapstructure:"hypervisor_server_list,omitempty" json:"hypervisor_server_list,omitempty"`
}

// SoftwareMapValues ...
type SoftwareMapValues struct {
	SoftwareType *string `mapstructure:"software_type,omitempty" json:"software_type,omitempty"`
	Status       *string `mapstructure:"status,omitempty" json:"status,omitempty"`
	Version      *string `mapstructure:"version,omitempty" json:"version,omitempty"`
}

// SoftwareMap ...
type SoftwareMap struct {
	NCC *SoftwareMapValues `mapstructure:"ncc,omitempty" json:"ncc,omitempty"`
	NOS *SoftwareMapValues `mapstructure:"nos,omitempty" json:"nos,omitempty"`
}

// ClusterConfig ...
type ClusterConfig struct {
	GpuDriverVersion              *string                    `mapstructure:"gpu_driver_version,omitempty" json:"gpu_driver_version,omitempty"`
	ClientAuth                    *ClientAuth                `mapstructure:"client_auth,omitempty" json:"client_auth,omitempty"`
	AuthorizedPublicKeyList       []*PublicKey               `mapstructure:"authorized_public_key_list,omitempty" json:"authorized_public_key_list,omitempty"`
	SoftwareMap                   *SoftwareMap               `mapstructure:"software_map,omitempty" json:"software_map,omitempty"`
	EncryptionStatus              *string                    `mapstructure:"encryption_status,omitempty" json:"encryption_status,omitempty"`
	SslKey                        *SslKey                    `mapstructure:"ssl_key,omitempty" json:"ssl_key,omitempty"`
	ServiceList                   []*string                  `mapstructure:"service_list,omitempty" json:"service_list,omitempty"`
	SupportedInformationVerbosity *string                    `mapstructure:"supported_information_verbosity,omitempty" json:"supported_information_verbosity,omitempty"`
	CertificationSigningInfo      *CertificationSigningInfo  `mapstructure:"certification_signing_info,omitempty" json:"certification_signing_info,omitempty"`
	RedundancyFactor              *int64                     `mapstructure:"redundancy_factor,omitempty" json:"redundancy_factor,omitempty"`
	ExternalConfigurations        *ExternalConfigurations    `mapstructure:"external_configurations,omitempty" json:"external_configurations,omitempty"`
	OperationMode                 *string                    `mapstructure:"operation_mode,omitempty" json:"operation_mode,omitempty"`
	CaCertificateList             []*CaCert                  `mapstructure:"ca_certificate_list,omitempty" json:"ca_certificate_list,omitempty"`
	EnabledFeatureList            []*string                  `mapstructure:"enabled_feature_list,omitempty" json:"enabled_feature_list,omitempty"`
	IsAvailable                   *bool                      `mapstructure:"is_available,omitempty" json:"is_available,omitempty"`
	Build                         *BuildInfo                 `mapstructure:"build,omitempty" json:"build,omitempty"`
	Timezone                      *string                    `mapstructure:"timezone,omitempty" json:"timezone,omitempty"`
	ClusterArch                   *string                    `mapstructure:"cluster_arch,omitempty" json:"cluster_arch,omitempty"`
	ManagementServerList          []*ClusterManagementServer `mapstructure:"management_server_list,omitempty" json:"management_server_list,omitempty"`
}

// ClusterManagementServer ...
type ClusterManagementServer struct {
	IP         *string   `mapstructure:"ip,omitempty" json:"ip,omitempty"`
	DrsEnabled *bool     `mapstructure:"drs_enabled,omitempty" json:"drs_enabled,omitempty"`
	StatusList []*string `mapstructure:"status_list,omitempty" json:"status_list,omitempty"`
	Type       *string   `mapstructure:"type,omitempty" json:"type,omitempty"`
}

// BuildInfo ...
type BuildInfo struct {
	CommitID      *string `mapstructure:"commit_id,omitempty" json:"commit_id,omitempty"`
	FullVersion   *string `mapstructure:"full_version,omitempty" json:"full_version,omitempty"`
	CommitDate    *string `mapstructure:"commit_date,omitempty" json:"commit_date,omitempty"`
	Version       *string `mapstructure:"version,omitempty" json:"version,omitempty"`
	ShortCommitID *string `mapstructure:"short_commit_id,omitempty" json:"short_commit_id,omitempty"`
	BuildType     *string `mapstructure:"build_type,omitempty" json:"build_type,omitempty"`
}

// CaCert ...
type CaCert struct {
	CaName      *string `mapstructure:"ca_name,omitempty" json:"ca_name,omitempty"`
	Certificate *string `mapstructure:"certificate,omitempty" json:"certificate,omitempty"`
}

// ExternalConfigurations ...
type ExternalConfigurations struct {
	CitrixConnectorConfig *CitrixConnectorConfigDetails `mapstructure:"citrix_connector_config,omitempty" json:"citrix_connector_config,omitempty"`
}

// CitrixConnectorConfigDetails ...
type CitrixConnectorConfigDetails struct {
	CitrixVMReferenceList *[]Reference            `mapstructure:"citrix_vm_reference_list,omitempty" json:"citrix_vm_reference_list,omitempty"`
	ClientSecret          *string                 `mapstructure:"client_secret,omitempty" json:"client_secret,omitempty"`
	CustomerID            *string                 `mapstructure:"customer_id,omitempty" json:"customer_id,omitempty"`
	ClientID              *string                 `mapstructure:"client_id,omitempty" json:"client_id,omitempty"`
	ResourceLocation      *CitrixResourceLocation `mapstructure:"resource_location,omitempty" json:"resource_location,omitempty"`
}

// CitrixResourceLocation ...
type CitrixResourceLocation struct {
	ID   *string `mapstructure:"id,omitempty" json:"id,omitempty"`
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`
}

// SslKey ...
type SslKey struct {
	KeyType        *string                   `mapstructure:"key_type,omitempty" json:"key_type,omitempty"`
	KeyName        *string                   `mapstructure:"key_name,omitempty" json:"key_name,omitempty"`
	SigningInfo    *CertificationSigningInfo `mapstructure:"signing_info,omitempty" json:"signing_info,omitempty"`
	ExpireDatetime *string                   `mapstructure:"expire_datetime,omitempty" json:"expire_datetime,omitempty"`
}

// CertificationSigningInfo ...
type CertificationSigningInfo struct {
	City             *string `mapstructure:"city,omitempty" json:"city,omitempty"`
	CommonNameSuffix *string `mapstructure:"common_name_suffix,omitempty" json:"common_name_suffix,omitempty"`
	State            *string `mapstructure:"state,omitempty" json:"state,omitempty"`
	CountryCode      *string `mapstructure:"country_code,omitempty" json:"country_code,omitempty"`
	CommonName       *string `mapstructure:"common_name,omitempty" json:"common_name,omitempty"`
	Organization     *string `mapstructure:"organization,omitempty" json:"organization,omitempty"`
	EmailAddress     *string `mapstructure:"email_address,omitempty" json:"email_address,omitempty"`
}

// PublicKey ...
type PublicKey struct {
	Key  *string `mapstructure:"key,omitempty" json:"key,omitempty"`
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`
}

// ClientAuth ...
type ClientAuth struct {
	Status  *string `mapstructure:"status,omitempty" json:"status,omitempty"`
	CaChain *string `mapstructure:"ca_chain,omitempty" json:"ca_chain,omitempty"`
	Name    *string `mapstructure:"name,omitempty" json:"name,omitempty"`
}

// HypervisorServer ...
type HypervisorServer struct {
	IP      *string `mapstructure:"ip,omitempty" json:"ip,omitempty"`
	Version *string `mapstructure:"version,omitempty" json:"version,omitempty"`
	Type    *string `mapstructure:"type,omitempty" json:"type,omitempty"`
}

// ClusterResource ...
type ClusterResource struct {
	Config            *ConfigClusterSpec `mapstructure:"config,omitempty" json:"config,omitempty"`
	Network           *ClusterNetwork    `mapstructure:"network,omitempty" json:"network,omitempty"`
	RunTimeStatusList []*string          `mapstructure:"runtime_status_list,omitempty" json:"runtime_status_list,omitempty"`
}

// ConfigClusterSpec ...
type ConfigClusterSpec struct {
	GpuDriverVersion              *string                     `mapstructure:"gpu_driver_version,omitempty" json:"gpu_driver_version,omitempty"`
	ClientAuth                    *ClientAuth                 `mapstructure:"client_auth,omitempty" json:"client_auth,omitempty"`
	AuthorizedPublicKeyList       []*PublicKey                `mapstructure:"authorized_public_key_list,omitempty" json:"authorized_public_key_list,omitempty"`
	SoftwareMap                   map[string]interface{}      `mapstructure:"software_map,omitempty" json:"software_map,omitempty"`
	EncryptionStatus              string                      `mapstructure:"encryption_status,omitempty" json:"encryption_status,omitempty"`
	RedundancyFactor              *int64                      `mapstructure:"redundancy_factor,omitempty" json:"redundancy_factor,omitempty"`
	CertificationSigningInfo      *CertificationSigningInfo   `mapstructure:"certification_signing_info,omitempty" json:"certification_signing_info,omitempty"`
	SupportedInformationVerbosity *string                     `mapstructure:"supported_information_verbosity,omitempty" json:"supported_information_verbosity,omitempty"`
	ExternalConfigurations        *ExternalConfigurationsSpec `mapstructure:"external_configurations,omitempty" json:"external_configurations,omitempty"`
	EnabledFeatureList            []*string                   `mapstructure:"enabled_feature_list,omitempty" json:"enabled_feature_list,omitempty"`
	Timezone                      *string                     `mapstructure:"timezone,omitempty" json:"timezone,omitempty"`
	OperationMode                 *string                     `mapstructure:"operation_mode,omitempty" json:"operation_mode,omitempty"`
}

// ExternalConfigurationsSpec ...
type ExternalConfigurationsSpec struct {
	CitrixConnectorConfig *CitrixConnectorConfigDetailsSpec `mapstructure:"citrix_connector_config,omitempty" json:"citrix_connector_config,omitempty"`
}

// CitrixConnectorConfigDetailsSpec ...
type CitrixConnectorConfigDetailsSpec struct {
	CitrixVMReferenceList []*Reference                `mapstructure:"citrix_connector_config,omitempty" json:"citrix_connector_config,omitempty"`
	ClientSecret          *string                     `mapstructure:"client_secret,omitempty" json:"client_secret,omitempty"`
	CustomerID            *string                     `mapstructure:"customer_id,omitempty" json:"customer_id,omitempty"`
	ClientID              *string                     `mapstructure:"client_id,omitempty" json:"client_id,omitempty"`
	ResourceLocation      *CitrixResourceLocationSpec `mapstructure:"resource_location,omitempty" json:"resource_location,omitempty"`
}

// CitrixResourceLocationSpec ...
type CitrixResourceLocationSpec struct {
	ID   *string `mapstructure:"id,omitempty" json:"id,omitempty"`
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`
}

// ClusterNetwork ...
type ClusterNetwork struct {
	MasqueradingPort       *int64                  `mapstructure:"masquerading_port,omitempty" json:"masquerading_port,omitempty"`
	MasqueradingIP         *string                 `mapstructure:"masquerading_ip,omitempty" json:"masquerading_ip,omitempty"`
	ExternalIP             *string                 `mapstructure:"external_ip,omitempty" json:"external_ip,omitempty"`
	HTTPProxyList          []*ClusterNetworkEntity `mapstructure:"http_proxy_list,omitempty" json:"http_proxy_list,omitempty"`
	SMTPServer             *SMTPServer             `mapstructure:"smtp_server,omitempty" json:"smtp_server,omitempty"`
	NTPServerIPList        []*string               `mapstructure:"ntp_server_ip_list,omitempty" json:"ntp_server_ip_list,omitempty"`
	ExternalSubnet         *string                 `mapstructure:"external_subnet,omitempty" json:"external_subnet,omitempty"`
	NFSSubnetWhitelist     []*string               `mapstructure:"nfs_subnet_whitelist,omitempty" json:"nfs_subnet_whitelist,omitempty"`
	ExternalDataServicesIP *string                 `mapstructure:"external_data_services_ip,omitempty" json:"external_data_services_ip,omitempty"`
	DomainServer           *ClusterDomainServer    `mapstructure:"domain_server,omitempty" json:"domain_server,omitempty"`
	NameServerIPList       []*string               `mapstructure:"name_server_ip_list,omitempty" json:"name_server_ip_list,omitempty"`
	HTTPProxyWhitelist     []*HTTPProxyWhitelist   `mapstructure:"http_proxy_whitelist,omitempty" json:"http_proxy_whitelist,omitempty"`
	InternalSubnet         *string                 `mapstructure:"internal_subnet,omitempty" json:"internal_subnet,omitempty"`
}

// HTTPProxyWhitelist ...
type HTTPProxyWhitelist struct {
	Target     *string `mapstructure:"target,omitempty" json:"target,omitempty"`
	TargetType *string `mapstructure:"target_type,omitempty" json:"target_type,omitempty"`
}

// ClusterDomainServer ...
type ClusterDomainServer struct {
	Nameserver        *string      `mapstructure:"nameserver,omitempty" json:"nameserver,omitempty"`
	Name              *string      `mapstructure:"name,omitempty" json:"name,omitempty"`
	DomainCredentials *Credentials `mapstructure:"external_data_services_ip,omitempty" json:"external_data_services_ip,omitempty"`
}

// SMTPServer ...
type SMTPServer struct {
	Type         *string               `mapstructure:"type,omitempty" json:"type,omitempty"`
	EmailAddress *string               `mapstructure:"email_address,omitempty" json:"email_address,omitempty"`
	Server       *ClusterNetworkEntity `mapstructure:"server,omitempty" json:"server,omitempty"`
}

// ClusterNetworkEntity ...
type ClusterNetworkEntity struct {
	Credentials   *Credentials `mapstructure:"credentials,omitempty" json:"credentials,omitempty"`
	ProxyTypeList []*string    `mapstructure:"proxy_type_list,omitempty" json:"proxy_type_list,omitempty"`
	Address       *Address     `mapstructure:"address,omitempty" json:"address,omitempty"`
}

// Credentials ...
type Credentials struct {
	Username *string `mapstructure:"username,omitempty" json:"username,omitempty"`
	Password *string `mapstructure:"password,omitempty" json:"password,omitempty"`
}

// VMEfficiencyMap ...
type VMEfficiencyMap struct {
	BullyVMNum           *string `mapstructure:"bully_vm_num,omitempty" json:"bully_vm_num,omitempty"`
	ConstrainedVMNum     *string `mapstructure:"constrained_vm_num,omitempty" json:"constrained_vm_num,omitempty"`
	DeadVMNum            *string `mapstructure:"dead_vm_num,omitempty" json:"dead_vm_num,omitempty"`
	InefficientVMNum     *string `mapstructure:"inefficient_vm_num,omitempty" json:"inefficient_vm_num,omitempty"`
	OverprovisionedVMNum *string `mapstructure:"overprovisioned_vm_num,omitempty" json:"overprovisioned_vm_num,omitempty"`
}

// ClusterAnalysis ...
type ClusterAnalysis struct {
	VMEfficiencyMap *VMEfficiencyMap `mapstructure:"vm_efficiency_map,omitempty" json:"vm_efficiency_map,omitempty"`
}

// CategoryListMetadata All api calls that return a list will have this metadata block as input
type CategoryListMetadata struct {

	// The filter in FIQL syntax used for the results.
	Filter *string `mapstructure:"filter,omitempty" json:"filter,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	// The number of records to retrieve relative to the offset
	Length *int64 `mapstructure:"length,omitempty" json:"length,omitempty"`

	// Offset from the start of the entity list
	Offset *int64 `mapstructure:"offset,omitempty" json:"offset,omitempty"`

	// The attribute to perform sort on
	SortAttribute *string `mapstructure:"sort_attribute,omitempty" json:"sort_attribute,omitempty"`

	// The sort order in which results are returned
	SortOrder *string `mapstructure:"sort_order,omitempty" json:"sort_order,omitempty"`
}

// CategoryKeyStatus represents Category Key Definition.
type CategoryKeyStatus struct {

	// API version.
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// Description of the category.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// Name of the category.
	Name *string `mapstructure:"name" json:"name"`

	// Specifying whether its a system defined category.
	SystemDefined *bool `mapstructure:"system_defined,omitempty" json:"system_defined,omitempty"`
}

// CategoryKeyListResponse represents the category key list response.
type CategoryKeyListResponse struct {

	// API Version.
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Entities []*CategoryKeyStatus `mapstructure:"entities,omitempty" json:"entities,omitempty"`

	Metadata *CategoryListMetadata `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`
}

// CategoryKey represents category key definition.
type CategoryKey struct {

	// API version.
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// Description of the category.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// Name of the category.
	Name *string `mapstructure:"name" json:"name"`
}

// CategoryStatus represents The status of a REST API call. Only used when there is a failure to report.
type CategoryStatus struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// The HTTP error code.
	Code *int64 `mapstructure:"code,omitempty" json:"code,omitempty"`

	// The kind name
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`

	State *string `mapstructure:"state,omitempty" json:"state,omitempty"`
}

// CategoryValueListResponse represents Category Value list response.
type CategoryValueListResponse struct {
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Entities []*CategoryValueStatus `mapstructure:"entities,omitempty" json:"entities,omitempty"`

	Metadata *CategoryListMetadata `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`
}

// CategoryValueStatus represents Category value definition.
type CategoryValueStatus struct {

	// API version.
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// Description of the category value.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty"`

	// The name of the category.
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`

	// Specifying whether its a system defined category.
	SystemDefined *bool `mapstructure:"system_defined,omitempty" json:"system_defined,omitempty"`

	// The value of the category.
	Value *string `mapstructure:"value,omitempty" json:"value,omitempty"`
}

// CategoryFilter represents A category filter.
type CategoryFilter struct {

	// List of kinds associated with this filter.
	KindList []*string `mapstructure:"kind_list,omitempty" json:"kind_list,omitempty"`

	// A list of category key and list of values.
	Params map[string][]string `mapstructure:"params,omitempty" json:"params,omitempty"`

	// The type of the filter being used.
	Type *string `mapstructure:"type,omitempty" json:"type,omitempty"`
}

// CategoryQueryInput represents Categories query input object.
type CategoryQueryInput struct {

	// API version.
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	CategoryFilter *CategoryFilter `mapstructure:"category_filter,omitempty" json:"category_filter,omitempty"`

	// The maximum number of members to return per group.
	GroupMemberCount *int64 `mapstructure:"group_member_count,omitempty" json:"group_member_count,omitempty"`

	// The offset into the total member set to return per group.
	GroupMemberOffset *int64 `mapstructure:"group_member_offset,omitempty" json:"group_member_offset,omitempty"`

	// TBD: USED_IN - to get policies in which specified categories are used. APPLIED_TO - to get entities attached to
	// specified categories.
	UsageType *string `mapstructure:"usage_type,omitempty" json:"usage_type,omitempty"`
}

// CategoryQueryResponseMetadata represents Response metadata.
type CategoryQueryResponseMetadata struct {

	// The maximum number of records to return per group.
	GroupMemberCount *int64 `mapstructure:"group_member_count,omitempty" json:"group_member_count,omitempty"`

	// The offset into the total records set to return per group.
	GroupMemberOffset *int64 `mapstructure:"group_member_offset,omitempty" json:"group_member_offset,omitempty"`

	// Total number of matched results.
	TotalMatches *int64 `mapstructure:"total_matches,omitempty" json:"total_matches,omitempty"`

	// TBD: USED_IN - to get policies in which specified categories are used. APPLIED_TO - to get entities attached to specified categories.
	UsageType *string `mapstructure:"usage_type,omitempty" json:"usage_type,omitempty"`
}

// EntityReference Reference to an entity.
type EntityReference struct {

	// Categories for the entity.
	Categories map[string]string `mapstructure:"categories,omitempty" json:"categories,omitempty"`

	// Kind of the reference.
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	// Name of the entity.
	Name *string `mapstructure:"name,omitempty" json:"name,omitempty"`

	// The type of filter being used. (Options : CATEGORIES_MATCH_ALL , CATEGORIES_MATCH_ANY)
	Type *string `mapstructure:"type,omitempty" json:"type,omitempty"`

	// UUID of the entity.
	UUID *string `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
}

// CategoryQueryResponseResults ...
type CategoryQueryResponseResults struct {

	// List of entity references.
	EntityAnyReferenceList []*EntityReference `mapstructure:"entity_any_reference_list,omitempty" json:"entity_any_reference_list,omitempty"`

	// Total number of filtered results.
	FilteredEntityCount *int64 `mapstructure:"filtered_entity_count,omitempty" json:"filtered_entity_count,omitempty"`

	// The entity kind.
	Kind *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`

	// Total number of the matched results.
	TotalEntityCount *int64 `mapstructure:"total_entity_count,omitempty" json:"total_entity_count,omitempty"`
}

// CategoryQueryResponse represents Categories query response object.
type CategoryQueryResponse struct {

	// API version.
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	Metadata *CategoryQueryResponseMetadata `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`

	Results []*CategoryQueryResponseResults `mapstructure:"results,omitempty" json:"results,omitempty"`
}

// CategoryValue represents Category value definition.
type CategoryValue struct {

	// API version.
	APIVersion *string `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`

	// Description of the category value.
	Description *string `mapstructure:"description,omitempty" json:"description,omitempty" `

	// Value for the category.
	Value *string `mapstructure:"value,omitempty" json:"value,omitempty"`
}

// PortRange represents Range of TCP/UDP ports.
type PortRange struct {
	EndPort *int64 `mapstructure:"end_port,omitempty" json:"end_port,omitempty"`

	StartPort *int64 `mapstructure:"start_port,omitempty" json:"start_port,omitempty"`
}

// IPSubnet IP subnet provided as an address and prefix length.
type IPSubnet struct {

	// IPV4 address.
	IP *string `mapstructure:"ip,omitempty" json:"ip,omitempty"`

	PrefixLength *int64 `mapstructure:"prefix_length,omitempty" json:"prefix_length,omitempty"`
}

// NetworkRuleIcmpTypeCodeList ..
type NetworkRuleIcmpTypeCodeList struct {
	Code *int64 `mapstructure:"code,omitempty" json:"code,omitempty"`

	Type *int64 `mapstructure:"type,omitempty" json:"type,omitempty"`
}

// NetworkRule ...
type NetworkRule struct {

	// Timestamp of expiration time.
	ExpirationTime *string `mapstructure:"expiration_time,omitempty" json:"expiration_time,omitempty"`

	// The set of categories that matching VMs need to have.
	Filter *CategoryFilter `mapstructure:"filter,omitempty" json:"filter,omitempty"`

	// List of ICMP types and codes allowed by this rule.
	IcmpTypeCodeList []*NetworkRuleIcmpTypeCodeList `mapstructure:"icmp_type_code_list,omitempty" json:"icmp_type_code_list,omitempty"`

	IPSubnet *IPSubnet `mapstructure:"ip_subnet,omitempty" json:"ip_subnet,omitempty"`

	NetworkFunctionChainReference *Reference `mapstructure:"network_function_chain_reference,omitempty" json:"network_function_chain_reference,omitempty"`

	// The set of categories that matching VMs need to have.
	PeerSpecificationType *string `mapstructure:"peer_specification_type,omitempty" json:"peer_specification_type,omitempty"`

	// Select a protocol to allow.  Multiple protocols can be allowed by repeating network_rule object.  If a protocol
	// is not configured in the network_rule object then it is allowed.
	Protocol *string `mapstructure:"protocol,omitempty" json:"protocol,omitempty"`

	// List of TCP ports that are allowed by this rule.
	TCPPortRangeList []*PortRange `mapstructure:"tcp_port_range_list,omitempty" json:"tcp_port_range_list,omitempty"`

	// List of UDP ports that are allowed by this rule.
	UDPPortRangeList []*PortRange `mapstructure:"udp_port_range_list,omitempty" json:"udp_port_range_list,omitempty"`
}

// TargetGroup ...
type TargetGroup struct {

	// Default policy for communication within target group.
	DefaultInternalPolicy *string `mapstructure:"default_internal_policy,omitempty" json:"default_internal_policy,omitempty"`

	// The set of categories that matching VMs need to have.
	Filter *CategoryFilter `mapstructure:"filter,omitempty" json:"filter,omitempty"`

	// Way to identify the object for which rule is applied.
	PeerSpecificationType *string `mapstructure:"peer_specification_type,omitempty" json:"peer_specification_type,omitempty"`
}

// NetworkSecurityRuleResourcesRule These rules are used for quarantining suspected VMs. Target group is a required
// attribute.  Empty inbound_allow_list will not allow anything into target group. Empty outbound_allow_list will allow
// everything from target group.
type NetworkSecurityRuleResourcesRule struct {
	Action            *string        `mapstructure:"action,omitempty" json:"action,omitempty"`                         // Type of action.
	InboundAllowList  []*NetworkRule `mapstructure:"inbound_allow_list,omitempty" json:"inbound_allow_list,omitempty"` //
	OutboundAllowList []*NetworkRule `mapstructure:"outbound_allow_list,omitempty" json:"outbound_allow_list,omitempty"`
	TargetGroup       *TargetGroup   `mapstructure:"target_group,omitempty" json:"target_group,omitempty"`
}

// NetworkSecurityRuleIsolationRule These rules are used for environmental isolation.
type NetworkSecurityRuleIsolationRule struct {
	Action             *string         `mapstructure:"action,omitempty" json:"action,omitempty"`                             // Type of action.
	FirstEntityFilter  *CategoryFilter `mapstructure:"first_entity_filter,omitempty" json:"first_entity_filter,omitempty"`   // The set of categories that matching VMs need to have.
	SecondEntityFilter *CategoryFilter `mapstructure:"second_entity_filter,omitempty" json:"second_entity_filter,omitempty"` // The set of categories that matching VMs need to have.
}

// NetworkSecurityRuleResources ...
type NetworkSecurityRuleResources struct {
	AppRule        *NetworkSecurityRuleResourcesRule `mapstructure:"app_rule,omitempty" json:"app_rule,omitempty"`
	IsolationRule  *NetworkSecurityRuleIsolationRule `mapstructure:"isolation_rule,omitempty" json:"isolation_rule,omitempty"`
	QuarantineRule *NetworkSecurityRuleResourcesRule `mapstructure:"quarantine_rule,omitempty" json:"quarantine_rule,omitempty"`
}

// NetworkSecurityRule ...
type NetworkSecurityRule struct {
	Description *string                       `mapstructure:"description" json:"description"`
	Name        *string                       `mapstructure:"name,omitempty" json:"name,omitempty"`
	Resources   *NetworkSecurityRuleResources `mapstructure:"resources,omitempty" json:"resources,omitempty" `
}

// Metadata Metadata The kind metadata
type Metadata struct {
	LastUpdateTime       *time.Time        `mapstructure:"last_update_time,omitempty" json:"last_update_time,omitempty"`   //
	Kind                 *string           `mapstructure:"kind" json:"kind"`                                               //
	UUID                 *string           `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`                           //
	ProjectReference     *Reference        `mapstructure:"project_reference,omitempty" json:"project_reference,omitempty"` // project reference
	CreationTime         *time.Time        `mapstructure:"creation_time,omitempty" json:"creation_time,omitempty"`
	SpecVersion          *int64            `mapstructure:"spec_version,omitempty" json:"spec_version,omitempty"`
	SpecHash             *string           `mapstructure:"spec_hash,omitempty" json:"spec_hash,omitempty"`
	OwnerReference       *Reference        `mapstructure:"owner_reference,omitempty" json:"owner_reference,omitempty"`
	Categories           map[string]string `mapstructure:"categories,omitempty" json:"categories,omitempty"`
	Name                 *string           `mapstructure:"name,omitempty" json:"name,omitempty"`
	ShouldForceTranslate *bool             `mapstructure:"should_force_translate,omitempty" json:"should_force_translate,omitempty"` // Applied on Prism Central only. Indicate whether force to translate the spec of the fanout request to fit the target cluster API schema.
}

// NetworkSecurityRuleIntentInput An intentful representation of a network_security_rule
type NetworkSecurityRuleIntentInput struct {
	APIVersion *string              `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`
	Metadata   *Metadata            `mapstructure:"metadata" json:"metadata"`
	Spec       *NetworkSecurityRule `mapstructure:"spec" json:"spec"`
}

// NetworkSecurityRuleDefStatus ... Network security rule status
type NetworkSecurityRuleDefStatus struct {
	AppRule          *NetworkSecurityRuleResourcesRule `mapstructure:"app_rule,omitempty" json:"app_rule,omitempty"`
	IsolationRule    *NetworkSecurityRuleIsolationRule `mapstructure:"isolation_rule,omitempty" json:"isolation_rule,omitempty"`
	QuarantineRule   *NetworkSecurityRuleResourcesRule `mapstructure:"quarantine_rule,omitempty" json:"quarantine_rule,omitempty"`
	State            *string                           `mapstructure:"state,omitmepty" json:"state,omitmepty"`
	ExecutionContext *ExecutionContext                 `mapstructure:"execution_context,omitempty" json:"execution_context,omitempty"`
}

// NetworkSecurityRuleIntentResponse Response object for intentful operations on a network_security_rule
type NetworkSecurityRuleIntentResponse struct {
	APIVersion *string                      `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`
	Metadata   *Metadata                    `mapstructure:"metadata" json:"metadata"`
	Spec       *NetworkSecurityRule         `mapstructure:"spec,omitempty" json:"spec,omitempty"`
	Status     NetworkSecurityRuleDefStatus `mapstructure:"status,omitempty" bson:"status,omitempty" json:"status,omitempty" bson:"status,omitempty"`
}

// NetworkSecurityRuleStatus The status of a REST API call. Only used when there is a failure to report.
type NetworkSecurityRuleStatus struct {
	APIVersion  *string            `mapstructure:"api_version,omitempty" json:"api_version,omitempty"` //
	Code        *int64             `mapstructure:"code,omitempty" json:"code,omitempty"`               // The HTTP error code.
	Kind        *string            `mapstructure:"kind,omitempty" json:"kind,omitempty"`               // The kind name
	MessageList []*MessageResource `mapstructure:"message_list,omitempty" json:"message_list,omitempty"`
	State       *string            `mapstructure:"state,omitempty" json:"state,omitempty"`
}

// ListMetadata All api calls that return a list will have this metadata block as input
type ListMetadata struct {
	Filter        *string `mapstructure:"filter,omitempty" json:"filter,omitempty"`                 // The filter in FIQL syntax used for the results.
	Kind          *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`                     // The kind name
	Length        *int64  `mapstructure:"length,omitempty" json:"length,omitempty"`                 // The number of records to retrieve relative to the offset
	Offset        *int64  `mapstructure:"offset,omitempty" json:"offset,omitempty"`                 // Offset from the start of the entity list
	SortAttribute *string `mapstructure:"sort_attribute,omitempty" json:"sort_attribute,omitempty"` // The attribute to perform sort on
	SortOrder     *string `mapstructure:"sort_order,omitempty" json:"sort_order,omitempty"`         // The sort order in which results are returned
}

// ListMetadataOutput All api calls that return a list will have this metadata block
type ListMetadataOutput struct {
	Filter        *string `mapstructure:"filter,omitempty" json:"filter,omitempty"`                 // The filter used for the results
	Kind          *string `mapstructure:"kind,omitempty" json:"kind,omitempty"`                     // The kind name
	Length        *int64  `mapstructure:"length,omitempty" json:"length,omitempty"`                 // The number of records retrieved relative to the offset
	Offset        *int64  `mapstructure:"offset,omitempty" json:"offset,omitempty"`                 // Offset from the start of the entity list
	SortAttribute *string `mapstructure:"sort_attribute,omitempty" json:"sort_attribute,omitempty"` // The attribute to perform sort on
	SortOrder     *string `mapstructure:"sort_order,omitempty" json:"sort_order,omitempty"`         // The sort order in which results are returned
	TotalMatches  *int64  `mapstructure:"total_matches,omitempty" json:"total_matches,omitempty"`   // Total matches found
}

// NetworkSecurityRuleIntentResource ... Response object for intentful operations on a network_security_rule
type NetworkSecurityRuleIntentResource struct {
	APIVersion *string                       `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`
	Metadata   *Metadata                     `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`
	Spec       *NetworkSecurityRule          `mapstructure:"spec,omitempty" json:"spec,omitempty"`
	Status     *NetworkSecurityRuleDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"`
}

// NetworkSecurityRuleListIntentResponse Response object for intentful operation of network_security_rules
type NetworkSecurityRuleListIntentResponse struct {
	APIVersion string                               `mapstructure:"api_version" json:"api_version"`
	Entities   []*NetworkSecurityRuleIntentResource `mapstructure:"entities,omitempty" bson:"entities,omitempty" json:"entities,omitempty" bson:"entities,omitempty"`
	Metadata   *ListMetadataOutput                  `mapstructure:"metadata" json:"metadata"`
}

// VolumeGroupInput Represents the request body for create volume_grop request
type VolumeGroupInput struct {
	APIVersion *string      `mapstructure:"api_version,omitempty" json:"api_version,omitempty"` // default 3.1.0
	Metadata   *Metadata    `mapstructure:"metadata,omitempty" json:"metadata,omitempty"`       // The volume_group kind metadata.
	Spec       *VolumeGroup `mapstructure:"spec,omitempty" json:"spec,omitempty"`               // Volume group input spec.
}

// VolumeGroup Represents volume group input spec.
type VolumeGroup struct {
	Name        *string               `mapstructure:"name" json:"name"`                                   // Volume Group name (required)
	Description *string               `mapstructure:"description,omitempty" json:"description,omitempty"` // Volume Group description.
	Resources   *VolumeGroupResources `mapstructure:"resources" json:"resources"`                         // Volume Group resources.
}

// VolumeGroupResources Represents the volume group resources
type VolumeGroupResources struct {
	FlashMode         *string         `mapstructure:"flash_mode,omitempty" json:"flash_mode,omitempty"`                   // Flash Mode, if enabled all disks of the VG are pinned to SSD
	FileSystemType    *string         `mapstructure:"file_system_type,omitempty" json:"file_system_type,omitempty"`       // File system to be used for volume
	SharingStatus     *string         `mapstructure:"sharing_status,omitempty" json:"sharing_status,omitempty"`           // Whether the VG can be shared across multiple iSCSI initiators
	AttachmentList    []*VMAttachment `mapstructure:"attachment_list,omitempty" json:"attachment_list,omitempty"`         // VMs attached to volume group.
	DiskList          []*VGDisk       `mapstructure:"disk_list,omitempty" json:"disk_list,omitempty"`                     // VGDisk Volume group disk specification.
	IscsiTargetPrefix *string         `mapstructure:"iscsi_target_prefix,omitempty" json:"iscsi_target_prefix,omitempty"` // iSCSI target prefix-name.
}

// VMAttachment VMs attached to volume group.
type VMAttachment struct {
	VMReference        *Reference `mapstructure:"vm_reference" json:"vm_reference"`                 // Reference to a kind
	IscsiInitiatorName *string    `mapstructure:"iscsi_initiator_name" json:"iscsi_initiator_name"` // Name of the iSCSI initiator of the workload outside Nutanix cluster.
}

// VGDisk Volume group disk specification.
type VGDisk struct {
	VmdiskUUID           *string    `mapstructure:"vmdisk_uuid" json:"vmdisk_uuid"`                       // The UUID of this volume disk
	Index                *int64     `mapstructure:"index" json:"index"`                                   // Index of the volume disk in the group.
	DataSourceReference  *Reference `mapstructure:"data_source_reference" json:"data_source_reference"`   // Reference to a kind
	DiskSizeMib          *int64     `mapstructure:"disk_size_mib" json:"disk_size_mib"`                   // Size of the disk in MiB.
	StorageContainerUUID *string    `mapstructure:"storage_container_uuid" json:"storage_container_uuid"` // Container UUID on which to create the disk.
}

// VolumeGroupResponse Response object for intentful operations on a volume_group
type VolumeGroupResponse struct {
	APIVersion *string               `mapstructure:"api_version" json:"api_version"`           //
	Metadata   *Metadata             `mapstructure:"metadata" json:"metadata"`                 // The volume_group kind metadata
	Spec       *VolumeGroup          `mapstructure:"spec,omitempty" json:"spec,omitempty"`     // Volume group input spec.
	Status     *VolumeGroupDefStatus `mapstructure:"status,omitempty" json:"status,omitempty"` // Volume group configuration.
}

// VolumeGroupDefStatus  Volume group configuration.
type VolumeGroupDefStatus struct {
	State       *string               `mapstructure:"state" json:"state"`               // The state of the volume group entity.
	MessageList []*MessageResource    `mapstructure:"message_list" json:"message_list"` // Volume group message list.
	Name        *string               `mapstructure:"name" json:"name"`                 // Volume group name.
	Resources   *VolumeGroupResources `mapstructure:"resources" json:"resources"`       // Volume group resources.
	Description *string               `mapstructure:"description" json:"description"`   // Volume group description.
}

// VolumeGroupListResponse Response object for intentful operation of volume_groups
type VolumeGroupListResponse struct {
	APIVersion *string                `mapstructure:"api_version" json:"api_version"`
	Entities   []*VolumeGroupResponse `mapstructure:"entities,omitempty" json:"entities,omitempty"`
	Metadata   *ListMetadataOutput    `mapstructure:"metadata" json:"metadata"`
}

// TasksResponse ...
type TasksResponse struct {
	Status               *string      `mapstructure:"status,omitempty" json:"status,omitempty"`
	LastUpdateTime       *time.Time   `mapstructure:"last_update_time,omitempty" json:"last_update_time,omitempty"`
	LogicalTimestamp     *int64       `mapstructure:"logical_timestamp,omitempty" json:"logical_timestamp,omitempty"`
	EntityReferenceList  []*Reference `mapstructure:"entity_reference_list,omitempty" json:"entity_reference_list,omitempty"`
	StartTime            *time.Time   `mapstructure:"start_time,omitempty" json:"start_time,omitempty"`
	CreationTime         *time.Time   `mapstructure:"creation_time,omitempty" json:"creation_time,omitempty"`
	ClusterReference     *Reference   `mapstructure:"cluster_reference,omitempty" json:"cluster_reference,omitempty"`
	SubtaskReferenceList []*Reference `mapstructure:"subtask_reference_list,omitempty" json:"subtask_reference_list,omitempty"`
	CompletionTime       *time.Time   `mapstructure:"completion_timev" json:"completion_timev"`
	ProgressMessage      *string      `mapstructure:"progress_message,omitempty" json:"progress_message,omitempty"`
	OperationType        *string      `mapstructure:"operation_type,omitempty" json:"operation_type,omitempty"`
	PercentageComplete   *int64       `mapstructure:"percentage_complete,omitempty" json:"percentage_complete,omitempty"`
	APIVersion           *string      `mapstructure:"api_version,omitempty" json:"api_version,omitempty"`
	UUID                 *string      `mapstructure:"uuid,omitempty" json:"uuid,omitempty"`
	ErrorDetail          *string      `mapstructure:"error_detail,omitempty" json:"error_detail,omitempty"`
}

// DeleteResponse ...
type DeleteResponse struct {
	Status     *DeleteStatus `mapstructure:"status" json:"status"`
	Spec       string        `mapstructure:"spec" json:"spec"`
	APIVersion string        `mapstructure:"api_version" json:"api_version"`
	Metadata   *Metadata     `mapstructure:"metadata" json:"metadata"`
}

// DeleteStatus ...
type DeleteStatus struct {
	State            string            `mapstructure:"state" json:"state"`
	ExecutionContext *ExecutionContext `mapstructure:"execution_context" json:"execution_context"`
}

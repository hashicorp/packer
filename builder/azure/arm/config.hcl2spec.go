// Code generated by "mapstructure-to-hcl2 -type Config,SharedImageGallery,SharedImageGalleryDestination,PlanInformation"; DO NOT EDIT.
package arm

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer/hcl2template"
	"github.com/zclconf/go-cty/cty"
)

// FlatConfig is an auto-generated flat version of Config.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatConfig struct {
	PackerBuildName                            *string                            `mapstructure:"packer_build_name" cty:"packer_build_name"`
	PackerBuilderType                          *string                            `mapstructure:"packer_builder_type" cty:"packer_builder_type"`
	PackerDebug                                *bool                              `mapstructure:"packer_debug" cty:"packer_debug"`
	PackerForce                                *bool                              `mapstructure:"packer_force" cty:"packer_force"`
	PackerOnError                              *string                            `mapstructure:"packer_on_error" cty:"packer_on_error"`
	PackerUserVars                             map[string]string                  `mapstructure:"packer_user_variables" cty:"packer_user_variables"`
	PackerSensitiveVars                        []string                           `mapstructure:"packer_sensitive_variables" cty:"packer_sensitive_variables"`
	CloudEnvironmentName                       *string                            `mapstructure:"cloud_environment_name" required:"false" cty:"cloud_environment_name"`
	ClientID                                   *string                            `mapstructure:"client_id" cty:"client_id"`
	ClientSecret                               *string                            `mapstructure:"client_secret" cty:"client_secret"`
	ClientCertPath                             *string                            `mapstructure:"client_cert_path" cty:"client_cert_path"`
	ClientJWT                                  *string                            `mapstructure:"client_jwt" cty:"client_jwt"`
	ObjectID                                   *string                            `mapstructure:"object_id" cty:"object_id"`
	TenantID                                   *string                            `mapstructure:"tenant_id" required:"false" cty:"tenant_id"`
	SubscriptionID                             *string                            `mapstructure:"subscription_id" cty:"subscription_id"`
	CaptureNamePrefix                          *string                            `mapstructure:"capture_name_prefix" cty:"capture_name_prefix"`
	CaptureContainerName                       *string                            `mapstructure:"capture_container_name" cty:"capture_container_name"`
	SharedGallery                              *FlatSharedImageGallery            `mapstructure:"shared_image_gallery" required:"false" cty:"shared_image_gallery"`
	SharedGalleryDestination                   *FlatSharedImageGalleryDestination `mapstructure:"shared_image_gallery_destination" cty:"shared_image_gallery_destination"`
	SharedGalleryTimeout                       *string                            `mapstructure:"shared_image_gallery_timeout" cty:"shared_image_gallery_timeout"`
	SharedGalleryImageVersionEndOfLifeDate     *string                            `mapstructure:"shared_gallery_image_version_end_of_life_date" required:"false" cty:"shared_gallery_image_version_end_of_life_date"`
	SharedGalleryImageVersionReplicaCount      *int32                             `mapstructure:"shared_image_gallery_replica_count" required:"false" cty:"shared_image_gallery_replica_count"`
	SharedGalleryImageVersionExcludeFromLatest *bool                              `mapstructure:"shared_gallery_image_version_exclude_from_latest" required:"false" cty:"shared_gallery_image_version_exclude_from_latest"`
	ImagePublisher                             *string                            `mapstructure:"image_publisher" required:"true" cty:"image_publisher"`
	ImageOffer                                 *string                            `mapstructure:"image_offer" required:"true" cty:"image_offer"`
	ImageSku                                   *string                            `mapstructure:"image_sku" required:"true" cty:"image_sku"`
	ImageVersion                               *string                            `mapstructure:"image_version" required:"false" cty:"image_version"`
	ImageUrl                                   *string                            `mapstructure:"image_url" required:"true" cty:"image_url"`
	CustomManagedImageName                     *string                            `mapstructure:"custom_managed_image_name" required:"true" cty:"custom_managed_image_name"`
	CustomManagedImageResourceGroupName        *string                            `mapstructure:"custom_managed_image_resource_group_name" required:"true" cty:"custom_managed_image_resource_group_name"`
	Location                                   *string                            `mapstructure:"location" cty:"location"`
	VMSize                                     *string                            `mapstructure:"vm_size" required:"false" cty:"vm_size"`
	ManagedImageResourceGroupName              *string                            `mapstructure:"managed_image_resource_group_name" cty:"managed_image_resource_group_name"`
	ManagedImageName                           *string                            `mapstructure:"managed_image_name" cty:"managed_image_name"`
	ManagedImageStorageAccountType             *string                            `mapstructure:"managed_image_storage_account_type" required:"false" cty:"managed_image_storage_account_type"`
	ManagedImageOSDiskSnapshotName             *string                            `mapstructure:"managed_image_os_disk_snapshot_name" required:"false" cty:"managed_image_os_disk_snapshot_name"`
	ManagedImageDataDiskSnapshotPrefix         *string                            `mapstructure:"managed_image_data_disk_snapshot_prefix" required:"false" cty:"managed_image_data_disk_snapshot_prefix"`
	ManagedImageZoneResilient                  *bool                              `mapstructure:"managed_image_zone_resilient" required:"false" cty:"managed_image_zone_resilient"`
	AzureTags                                  map[string]*string                 `mapstructure:"azure_tags" required:"false" cty:"azure_tags"`
	AzureTag                                   []hcl2template.FlatNameValue       `mapstructure:"azure_tag" required:"false" cty:"azure_tag"`
	ResourceGroupName                          *string                            `mapstructure:"resource_group_name" cty:"resource_group_name"`
	StorageAccount                             *string                            `mapstructure:"storage_account" cty:"storage_account"`
	TempComputeName                            *string                            `mapstructure:"temp_compute_name" required:"false" cty:"temp_compute_name"`
	TempResourceGroupName                      *string                            `mapstructure:"temp_resource_group_name" cty:"temp_resource_group_name"`
	BuildResourceGroupName                     *string                            `mapstructure:"build_resource_group_name" cty:"build_resource_group_name"`
	BuildKeyVaultName                          *string                            `mapstructure:"build_key_vault_name" cty:"build_key_vault_name"`
	BuildKeyVaultSKU                           *string                            `mapstructure:"build_key_vault_sku" cty:"build_key_vault_sku"`
	PrivateVirtualNetworkWithPublicIp          *bool                              `mapstructure:"private_virtual_network_with_public_ip" required:"false" cty:"private_virtual_network_with_public_ip"`
	VirtualNetworkName                         *string                            `mapstructure:"virtual_network_name" required:"false" cty:"virtual_network_name"`
	VirtualNetworkSubnetName                   *string                            `mapstructure:"virtual_network_subnet_name" required:"false" cty:"virtual_network_subnet_name"`
	VirtualNetworkResourceGroupName            *string                            `mapstructure:"virtual_network_resource_group_name" required:"false" cty:"virtual_network_resource_group_name"`
	CustomDataFile                             *string                            `mapstructure:"custom_data_file" required:"false" cty:"custom_data_file"`
	PlanInfo                                   *FlatPlanInformation               `mapstructure:"plan_info" required:"false" cty:"plan_info"`
	PollingDurationTimeout                     *string                            `mapstructure:"polling_duration_timeout" required:"false" cty:"polling_duration_timeout"`
	OSType                                     *string                            `mapstructure:"os_type" required:"false" cty:"os_type"`
	OSDiskSizeGB                               *int32                             `mapstructure:"os_disk_size_gb" required:"false" cty:"os_disk_size_gb"`
	AdditionalDiskSize                         []int32                            `mapstructure:"disk_additional_size" required:"false" cty:"disk_additional_size"`
	DiskCachingType                            *string                            `mapstructure:"disk_caching_type" required:"false" cty:"disk_caching_type"`
	AllowedInboundIpAddresses                  []string                           `mapstructure:"allowed_inbound_ip_addresses" cty:"allowed_inbound_ip_addresses"`
	BootDiagSTGAccount                         *string                            `mapstructure:"boot_diag_storage_account" required:"false" cty:"boot_diag_storage_account"`
	Type                                       *string                            `mapstructure:"communicator" cty:"communicator"`
	PauseBeforeConnect                         *string                            `mapstructure:"pause_before_connecting" cty:"pause_before_connecting"`
	SSHHost                                    *string                            `mapstructure:"ssh_host" cty:"ssh_host"`
	SSHPort                                    *int                               `mapstructure:"ssh_port" cty:"ssh_port"`
	SSHUsername                                *string                            `mapstructure:"ssh_username" cty:"ssh_username"`
	SSHPassword                                *string                            `mapstructure:"ssh_password" cty:"ssh_password"`
	SSHKeyPairName                             *string                            `mapstructure:"ssh_keypair_name" cty:"ssh_keypair_name"`
	SSHTemporaryKeyPairName                    *string                            `mapstructure:"temporary_key_pair_name" cty:"temporary_key_pair_name"`
	SSHClearAuthorizedKeys                     *bool                              `mapstructure:"ssh_clear_authorized_keys" cty:"ssh_clear_authorized_keys"`
	SSHPrivateKeyFile                          *string                            `mapstructure:"ssh_private_key_file" cty:"ssh_private_key_file"`
	SSHPty                                     *bool                              `mapstructure:"ssh_pty" cty:"ssh_pty"`
	SSHTimeout                                 *string                            `mapstructure:"ssh_timeout" cty:"ssh_timeout"`
	SSHAgentAuth                               *bool                              `mapstructure:"ssh_agent_auth" cty:"ssh_agent_auth"`
	SSHDisableAgentForwarding                  *bool                              `mapstructure:"ssh_disable_agent_forwarding" cty:"ssh_disable_agent_forwarding"`
	SSHHandshakeAttempts                       *int                               `mapstructure:"ssh_handshake_attempts" cty:"ssh_handshake_attempts"`
	SSHBastionHost                             *string                            `mapstructure:"ssh_bastion_host" cty:"ssh_bastion_host"`
	SSHBastionPort                             *int                               `mapstructure:"ssh_bastion_port" cty:"ssh_bastion_port"`
	SSHBastionAgentAuth                        *bool                              `mapstructure:"ssh_bastion_agent_auth" cty:"ssh_bastion_agent_auth"`
	SSHBastionUsername                         *string                            `mapstructure:"ssh_bastion_username" cty:"ssh_bastion_username"`
	SSHBastionPassword                         *string                            `mapstructure:"ssh_bastion_password" cty:"ssh_bastion_password"`
	SSHBastionInteractive                      *bool                              `mapstructure:"ssh_bastion_interactive" cty:"ssh_bastion_interactive"`
	SSHBastionPrivateKeyFile                   *string                            `mapstructure:"ssh_bastion_private_key_file" cty:"ssh_bastion_private_key_file"`
	SSHFileTransferMethod                      *string                            `mapstructure:"ssh_file_transfer_method" cty:"ssh_file_transfer_method"`
	SSHProxyHost                               *string                            `mapstructure:"ssh_proxy_host" cty:"ssh_proxy_host"`
	SSHProxyPort                               *int                               `mapstructure:"ssh_proxy_port" cty:"ssh_proxy_port"`
	SSHProxyUsername                           *string                            `mapstructure:"ssh_proxy_username" cty:"ssh_proxy_username"`
	SSHProxyPassword                           *string                            `mapstructure:"ssh_proxy_password" cty:"ssh_proxy_password"`
	SSHKeepAliveInterval                       *string                            `mapstructure:"ssh_keep_alive_interval" cty:"ssh_keep_alive_interval"`
	SSHReadWriteTimeout                        *string                            `mapstructure:"ssh_read_write_timeout" cty:"ssh_read_write_timeout"`
	SSHRemoteTunnels                           []string                           `mapstructure:"ssh_remote_tunnels" cty:"ssh_remote_tunnels"`
	SSHLocalTunnels                            []string                           `mapstructure:"ssh_local_tunnels" cty:"ssh_local_tunnels"`
	SSHPublicKey                               []byte                             `mapstructure:"ssh_public_key" cty:"ssh_public_key"`
	SSHPrivateKey                              []byte                             `mapstructure:"ssh_private_key" cty:"ssh_private_key"`
	WinRMUser                                  *string                            `mapstructure:"winrm_username" cty:"winrm_username"`
	WinRMPassword                              *string                            `mapstructure:"winrm_password" cty:"winrm_password"`
	WinRMHost                                  *string                            `mapstructure:"winrm_host" cty:"winrm_host"`
	WinRMPort                                  *int                               `mapstructure:"winrm_port" cty:"winrm_port"`
	WinRMTimeout                               *string                            `mapstructure:"winrm_timeout" cty:"winrm_timeout"`
	WinRMUseSSL                                *bool                              `mapstructure:"winrm_use_ssl" cty:"winrm_use_ssl"`
	WinRMInsecure                              *bool                              `mapstructure:"winrm_insecure" cty:"winrm_insecure"`
	WinRMUseNTLM                               *bool                              `mapstructure:"winrm_use_ntlm" cty:"winrm_use_ntlm"`
	AsyncResourceGroupDelete                   *bool                              `mapstructure:"async_resourcegroup_delete" required:"false" cty:"async_resourcegroup_delete"`
}

// FlatMapstructure returns a new FlatConfig.
// FlatConfig is an auto-generated flat version of Config.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*Config) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatConfig)
}

// HCL2Spec returns the hcl spec of a Config.
// This spec is used by HCL to read the fields of Config.
// The decoded values from this spec will then be applied to a FlatConfig.
func (*FlatConfig) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"packer_build_name":                &hcldec.AttrSpec{Name: "packer_build_name", Type: cty.String, Required: false},
		"packer_builder_type":              &hcldec.AttrSpec{Name: "packer_builder_type", Type: cty.String, Required: false},
		"packer_debug":                     &hcldec.AttrSpec{Name: "packer_debug", Type: cty.Bool, Required: false},
		"packer_force":                     &hcldec.AttrSpec{Name: "packer_force", Type: cty.Bool, Required: false},
		"packer_on_error":                  &hcldec.AttrSpec{Name: "packer_on_error", Type: cty.String, Required: false},
		"packer_user_variables":            &hcldec.AttrSpec{Name: "packer_user_variables", Type: cty.Map(cty.String), Required: false},
		"packer_sensitive_variables":       &hcldec.AttrSpec{Name: "packer_sensitive_variables", Type: cty.List(cty.String), Required: false},
		"cloud_environment_name":           &hcldec.AttrSpec{Name: "cloud_environment_name", Type: cty.String, Required: false},
		"client_id":                        &hcldec.AttrSpec{Name: "client_id", Type: cty.String, Required: false},
		"client_secret":                    &hcldec.AttrSpec{Name: "client_secret", Type: cty.String, Required: false},
		"client_cert_path":                 &hcldec.AttrSpec{Name: "client_cert_path", Type: cty.String, Required: false},
		"client_jwt":                       &hcldec.AttrSpec{Name: "client_jwt", Type: cty.String, Required: false},
		"object_id":                        &hcldec.AttrSpec{Name: "object_id", Type: cty.String, Required: false},
		"tenant_id":                        &hcldec.AttrSpec{Name: "tenant_id", Type: cty.String, Required: false},
		"subscription_id":                  &hcldec.AttrSpec{Name: "subscription_id", Type: cty.String, Required: false},
		"capture_name_prefix":              &hcldec.AttrSpec{Name: "capture_name_prefix", Type: cty.String, Required: false},
		"capture_container_name":           &hcldec.AttrSpec{Name: "capture_container_name", Type: cty.String, Required: false},
		"shared_image_gallery":             &hcldec.BlockSpec{TypeName: "shared_image_gallery", Nested: hcldec.ObjectSpec((*FlatSharedImageGallery)(nil).HCL2Spec())},
		"shared_image_gallery_destination": &hcldec.BlockSpec{TypeName: "shared_image_gallery_destination", Nested: hcldec.ObjectSpec((*FlatSharedImageGalleryDestination)(nil).HCL2Spec())},
		"shared_image_gallery_timeout":     &hcldec.AttrSpec{Name: "shared_image_gallery_timeout", Type: cty.String, Required: false},
		"shared_gallery_image_version_end_of_life_date":    &hcldec.AttrSpec{Name: "shared_gallery_image_version_end_of_life_date", Type: cty.String, Required: false},
		"shared_image_gallery_replica_count":               &hcldec.AttrSpec{Name: "shared_image_gallery_replica_count", Type: cty.Number, Required: false},
		"shared_gallery_image_version_exclude_from_latest": &hcldec.AttrSpec{Name: "shared_gallery_image_version_exclude_from_latest", Type: cty.Bool, Required: false},
		"image_publisher":           &hcldec.AttrSpec{Name: "image_publisher", Type: cty.String, Required: false},
		"image_offer":               &hcldec.AttrSpec{Name: "image_offer", Type: cty.String, Required: false},
		"image_sku":                 &hcldec.AttrSpec{Name: "image_sku", Type: cty.String, Required: false},
		"image_version":             &hcldec.AttrSpec{Name: "image_version", Type: cty.String, Required: false},
		"image_url":                 &hcldec.AttrSpec{Name: "image_url", Type: cty.String, Required: false},
		"custom_managed_image_name": &hcldec.AttrSpec{Name: "custom_managed_image_name", Type: cty.String, Required: false},
		"custom_managed_image_resource_group_name": &hcldec.AttrSpec{Name: "custom_managed_image_resource_group_name", Type: cty.String, Required: false},
		"location":                                &hcldec.AttrSpec{Name: "location", Type: cty.String, Required: false},
		"vm_size":                                 &hcldec.AttrSpec{Name: "vm_size", Type: cty.String, Required: false},
		"managed_image_resource_group_name":       &hcldec.AttrSpec{Name: "managed_image_resource_group_name", Type: cty.String, Required: false},
		"managed_image_name":                      &hcldec.AttrSpec{Name: "managed_image_name", Type: cty.String, Required: false},
		"managed_image_storage_account_type":      &hcldec.AttrSpec{Name: "managed_image_storage_account_type", Type: cty.String, Required: false},
		"managed_image_os_disk_snapshot_name":     &hcldec.AttrSpec{Name: "managed_image_os_disk_snapshot_name", Type: cty.String, Required: false},
		"managed_image_data_disk_snapshot_prefix": &hcldec.AttrSpec{Name: "managed_image_data_disk_snapshot_prefix", Type: cty.String, Required: false},
		"managed_image_zone_resilient":            &hcldec.AttrSpec{Name: "managed_image_zone_resilient", Type: cty.Bool, Required: false},
		"azure_tags":                              &hcldec.AttrSpec{Name: "azure_tags", Type: cty.Map(cty.String), Required: false},
		"azure_tag":                               &hcldec.BlockListSpec{TypeName: "azure_tag", Nested: hcldec.ObjectSpec((*hcl2template.FlatNameValue)(nil).HCL2Spec())},
		"resource_group_name":                     &hcldec.AttrSpec{Name: "resource_group_name", Type: cty.String, Required: false},
		"storage_account":                         &hcldec.AttrSpec{Name: "storage_account", Type: cty.String, Required: false},
		"temp_compute_name":                       &hcldec.AttrSpec{Name: "temp_compute_name", Type: cty.String, Required: false},
		"temp_resource_group_name":                &hcldec.AttrSpec{Name: "temp_resource_group_name", Type: cty.String, Required: false},
		"build_resource_group_name":               &hcldec.AttrSpec{Name: "build_resource_group_name", Type: cty.String, Required: false},
		"build_key_vault_name":                    &hcldec.AttrSpec{Name: "build_key_vault_name", Type: cty.String, Required: false},
		"build_key_vault_sku":                     &hcldec.AttrSpec{Name: "build_key_vault_sku", Type: cty.String, Required: false},
		"private_virtual_network_with_public_ip":  &hcldec.AttrSpec{Name: "private_virtual_network_with_public_ip", Type: cty.Bool, Required: false},
		"virtual_network_name":                    &hcldec.AttrSpec{Name: "virtual_network_name", Type: cty.String, Required: false},
		"virtual_network_subnet_name":             &hcldec.AttrSpec{Name: "virtual_network_subnet_name", Type: cty.String, Required: false},
		"virtual_network_resource_group_name":     &hcldec.AttrSpec{Name: "virtual_network_resource_group_name", Type: cty.String, Required: false},
		"custom_data_file":                        &hcldec.AttrSpec{Name: "custom_data_file", Type: cty.String, Required: false},
		"plan_info":                               &hcldec.BlockSpec{TypeName: "plan_info", Nested: hcldec.ObjectSpec((*FlatPlanInformation)(nil).HCL2Spec())},
		"polling_duration_timeout":                &hcldec.AttrSpec{Name: "polling_duration_timeout", Type: cty.String, Required: false},
		"os_type":                                 &hcldec.AttrSpec{Name: "os_type", Type: cty.String, Required: false},
		"os_disk_size_gb":                         &hcldec.AttrSpec{Name: "os_disk_size_gb", Type: cty.Number, Required: false},
		"disk_additional_size":                    &hcldec.AttrSpec{Name: "disk_additional_size", Type: cty.List(cty.Number), Required: false},
		"disk_caching_type":                       &hcldec.AttrSpec{Name: "disk_caching_type", Type: cty.String, Required: false},
		"allowed_inbound_ip_addresses":            &hcldec.AttrSpec{Name: "allowed_inbound_ip_addresses", Type: cty.List(cty.String), Required: false},
		"boot_diag_storage_account":               &hcldec.AttrSpec{Name: "boot_diag_storage_account", Type: cty.String, Required: false},
		"communicator":                            &hcldec.AttrSpec{Name: "communicator", Type: cty.String, Required: false},
		"pause_before_connecting":                 &hcldec.AttrSpec{Name: "pause_before_connecting", Type: cty.String, Required: false},
		"ssh_host":                                &hcldec.AttrSpec{Name: "ssh_host", Type: cty.String, Required: false},
		"ssh_port":                                &hcldec.AttrSpec{Name: "ssh_port", Type: cty.Number, Required: false},
		"ssh_username":                            &hcldec.AttrSpec{Name: "ssh_username", Type: cty.String, Required: false},
		"ssh_password":                            &hcldec.AttrSpec{Name: "ssh_password", Type: cty.String, Required: false},
		"ssh_keypair_name":                        &hcldec.AttrSpec{Name: "ssh_keypair_name", Type: cty.String, Required: false},
		"temporary_key_pair_name":                 &hcldec.AttrSpec{Name: "temporary_key_pair_name", Type: cty.String, Required: false},
		"ssh_clear_authorized_keys":               &hcldec.AttrSpec{Name: "ssh_clear_authorized_keys", Type: cty.Bool, Required: false},
		"ssh_private_key_file":                    &hcldec.AttrSpec{Name: "ssh_private_key_file", Type: cty.String, Required: false},
		"ssh_pty":                                 &hcldec.AttrSpec{Name: "ssh_pty", Type: cty.Bool, Required: false},
		"ssh_timeout":                             &hcldec.AttrSpec{Name: "ssh_timeout", Type: cty.String, Required: false},
		"ssh_agent_auth":                          &hcldec.AttrSpec{Name: "ssh_agent_auth", Type: cty.Bool, Required: false},
		"ssh_disable_agent_forwarding":            &hcldec.AttrSpec{Name: "ssh_disable_agent_forwarding", Type: cty.Bool, Required: false},
		"ssh_handshake_attempts":                  &hcldec.AttrSpec{Name: "ssh_handshake_attempts", Type: cty.Number, Required: false},
		"ssh_bastion_host":                        &hcldec.AttrSpec{Name: "ssh_bastion_host", Type: cty.String, Required: false},
		"ssh_bastion_port":                        &hcldec.AttrSpec{Name: "ssh_bastion_port", Type: cty.Number, Required: false},
		"ssh_bastion_agent_auth":                  &hcldec.AttrSpec{Name: "ssh_bastion_agent_auth", Type: cty.Bool, Required: false},
		"ssh_bastion_username":                    &hcldec.AttrSpec{Name: "ssh_bastion_username", Type: cty.String, Required: false},
		"ssh_bastion_password":                    &hcldec.AttrSpec{Name: "ssh_bastion_password", Type: cty.String, Required: false},
		"ssh_bastion_interactive":                 &hcldec.AttrSpec{Name: "ssh_bastion_interactive", Type: cty.Bool, Required: false},
		"ssh_bastion_private_key_file":            &hcldec.AttrSpec{Name: "ssh_bastion_private_key_file", Type: cty.String, Required: false},
		"ssh_file_transfer_method":                &hcldec.AttrSpec{Name: "ssh_file_transfer_method", Type: cty.String, Required: false},
		"ssh_proxy_host":                          &hcldec.AttrSpec{Name: "ssh_proxy_host", Type: cty.String, Required: false},
		"ssh_proxy_port":                          &hcldec.AttrSpec{Name: "ssh_proxy_port", Type: cty.Number, Required: false},
		"ssh_proxy_username":                      &hcldec.AttrSpec{Name: "ssh_proxy_username", Type: cty.String, Required: false},
		"ssh_proxy_password":                      &hcldec.AttrSpec{Name: "ssh_proxy_password", Type: cty.String, Required: false},
		"ssh_keep_alive_interval":                 &hcldec.AttrSpec{Name: "ssh_keep_alive_interval", Type: cty.String, Required: false},
		"ssh_read_write_timeout":                  &hcldec.AttrSpec{Name: "ssh_read_write_timeout", Type: cty.String, Required: false},
		"ssh_remote_tunnels":                      &hcldec.AttrSpec{Name: "ssh_remote_tunnels", Type: cty.List(cty.String), Required: false},
		"ssh_local_tunnels":                       &hcldec.AttrSpec{Name: "ssh_local_tunnels", Type: cty.List(cty.String), Required: false},
		"ssh_public_key":                          &hcldec.AttrSpec{Name: "ssh_public_key", Type: cty.List(cty.Number), Required: false},
		"ssh_private_key":                         &hcldec.AttrSpec{Name: "ssh_private_key", Type: cty.List(cty.Number), Required: false},
		"winrm_username":                          &hcldec.AttrSpec{Name: "winrm_username", Type: cty.String, Required: false},
		"winrm_password":                          &hcldec.AttrSpec{Name: "winrm_password", Type: cty.String, Required: false},
		"winrm_host":                              &hcldec.AttrSpec{Name: "winrm_host", Type: cty.String, Required: false},
		"winrm_port":                              &hcldec.AttrSpec{Name: "winrm_port", Type: cty.Number, Required: false},
		"winrm_timeout":                           &hcldec.AttrSpec{Name: "winrm_timeout", Type: cty.String, Required: false},
		"winrm_use_ssl":                           &hcldec.AttrSpec{Name: "winrm_use_ssl", Type: cty.Bool, Required: false},
		"winrm_insecure":                          &hcldec.AttrSpec{Name: "winrm_insecure", Type: cty.Bool, Required: false},
		"winrm_use_ntlm":                          &hcldec.AttrSpec{Name: "winrm_use_ntlm", Type: cty.Bool, Required: false},
		"async_resourcegroup_delete":              &hcldec.AttrSpec{Name: "async_resourcegroup_delete", Type: cty.Bool, Required: false},
	}
	return s
}

// FlatPlanInformation is an auto-generated flat version of PlanInformation.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatPlanInformation struct {
	PlanName          *string `mapstructure:"plan_name" cty:"plan_name"`
	PlanProduct       *string `mapstructure:"plan_product" cty:"plan_product"`
	PlanPublisher     *string `mapstructure:"plan_publisher" cty:"plan_publisher"`
	PlanPromotionCode *string `mapstructure:"plan_promotion_code" cty:"plan_promotion_code"`
}

// FlatMapstructure returns a new FlatPlanInformation.
// FlatPlanInformation is an auto-generated flat version of PlanInformation.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*PlanInformation) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatPlanInformation)
}

// HCL2Spec returns the hcl spec of a PlanInformation.
// This spec is used by HCL to read the fields of PlanInformation.
// The decoded values from this spec will then be applied to a FlatPlanInformation.
func (*FlatPlanInformation) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"plan_name":           &hcldec.AttrSpec{Name: "plan_name", Type: cty.String, Required: false},
		"plan_product":        &hcldec.AttrSpec{Name: "plan_product", Type: cty.String, Required: false},
		"plan_publisher":      &hcldec.AttrSpec{Name: "plan_publisher", Type: cty.String, Required: false},
		"plan_promotion_code": &hcldec.AttrSpec{Name: "plan_promotion_code", Type: cty.String, Required: false},
	}
	return s
}

// FlatSharedImageGallery is an auto-generated flat version of SharedImageGallery.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatSharedImageGallery struct {
	Subscription  *string `mapstructure:"subscription" cty:"subscription"`
	ResourceGroup *string `mapstructure:"resource_group" cty:"resource_group"`
	GalleryName   *string `mapstructure:"gallery_name" cty:"gallery_name"`
	ImageName     *string `mapstructure:"image_name" cty:"image_name"`
	ImageVersion  *string `mapstructure:"image_version" required:"false" cty:"image_version"`
}

// FlatMapstructure returns a new FlatSharedImageGallery.
// FlatSharedImageGallery is an auto-generated flat version of SharedImageGallery.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*SharedImageGallery) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatSharedImageGallery)
}

// HCL2Spec returns the hcl spec of a SharedImageGallery.
// This spec is used by HCL to read the fields of SharedImageGallery.
// The decoded values from this spec will then be applied to a FlatSharedImageGallery.
func (*FlatSharedImageGallery) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"subscription":   &hcldec.AttrSpec{Name: "subscription", Type: cty.String, Required: false},
		"resource_group": &hcldec.AttrSpec{Name: "resource_group", Type: cty.String, Required: false},
		"gallery_name":   &hcldec.AttrSpec{Name: "gallery_name", Type: cty.String, Required: false},
		"image_name":     &hcldec.AttrSpec{Name: "image_name", Type: cty.String, Required: false},
		"image_version":  &hcldec.AttrSpec{Name: "image_version", Type: cty.String, Required: false},
	}
	return s
}

// FlatSharedImageGalleryDestination is an auto-generated flat version of SharedImageGalleryDestination.
// Where the contents of a field with a `mapstructure:,squash` tag are bubbled up.
type FlatSharedImageGalleryDestination struct {
	SigDestinationResourceGroup      *string  `mapstructure:"resource_group" cty:"resource_group"`
	SigDestinationGalleryName        *string  `mapstructure:"gallery_name" cty:"gallery_name"`
	SigDestinationImageName          *string  `mapstructure:"image_name" cty:"image_name"`
	SigDestinationImageVersion       *string  `mapstructure:"image_version" cty:"image_version"`
	SigDestinationReplicationRegions []string `mapstructure:"replication_regions" cty:"replication_regions"`
}

// FlatMapstructure returns a new FlatSharedImageGalleryDestination.
// FlatSharedImageGalleryDestination is an auto-generated flat version of SharedImageGalleryDestination.
// Where the contents a fields with a `mapstructure:,squash` tag are bubbled up.
func (*SharedImageGalleryDestination) FlatMapstructure() interface{ HCL2Spec() map[string]hcldec.Spec } {
	return new(FlatSharedImageGalleryDestination)
}

// HCL2Spec returns the hcl spec of a SharedImageGalleryDestination.
// This spec is used by HCL to read the fields of SharedImageGalleryDestination.
// The decoded values from this spec will then be applied to a FlatSharedImageGalleryDestination.
func (*FlatSharedImageGalleryDestination) HCL2Spec() map[string]hcldec.Spec {
	s := map[string]hcldec.Spec{
		"resource_group":      &hcldec.AttrSpec{Name: "resource_group", Type: cty.String, Required: false},
		"gallery_name":        &hcldec.AttrSpec{Name: "gallery_name", Type: cty.String, Required: false},
		"image_name":          &hcldec.AttrSpec{Name: "image_name", Type: cty.String, Required: false},
		"image_version":       &hcldec.AttrSpec{Name: "image_version", Type: cty.String, Required: false},
		"replication_regions": &hcldec.AttrSpec{Name: "replication_regions", Type: cty.List(cty.String), Required: false},
	}
	return s
}

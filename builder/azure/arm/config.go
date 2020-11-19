//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config,SharedImageGallery,SharedImageGalleryDestination,PlanInformation

package arm

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/random"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/masterzen/winrm"

	azcommon "github.com/hashicorp/packer/builder/azure/common"
	"github.com/hashicorp/packer/builder/azure/common/client"
	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/builder/azure/pkcs12"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/common"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer/packer-plugin-sdk/template/interpolate"

	"golang.org/x/crypto/ssh"
)

const (
	DefaultImageVersion                      = "latest"
	DefaultUserName                          = "packer"
	DefaultPrivateVirtualNetworkWithPublicIp = false
	DefaultVMSize                            = "Standard_A1"
	DefaultKeyVaultSKU                       = "standard"
)

const (
	// https://docs.microsoft.com/en-us/azure/architecture/best-practices/naming-conventions#naming-rules-and-restrictions
	// Regular expressions in Go are not expressive enough, such that the regular expression returned by Azure
	// can be used (no backtracking).
	//
	//  -> ^[^_\W][\w-._]{0,79}(?<![-.])$
	//
	// This is not an exhaustive match, but it should be extremely close.
	validResourceGroupNameRe = "^[^_\\W][\\w-._\\(\\)]{0,89}$"
	validManagedDiskName     = "^[^_\\W][\\w-._)]{0,79}$"
	validResourceNamePrefix  = "^[^_\\W][\\w-._)]{0,10}$"
)

var (
	reCaptureContainerName = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{2,62}$`)
	reCaptureNamePrefix    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-\.]{0,23}$`)
	reManagedDiskName      = regexp.MustCompile(validManagedDiskName)
	reResourceGroupName    = regexp.MustCompile(validResourceGroupNameRe)
	reSnapshotName         = regexp.MustCompile(`^[A-Za-z0-9_]{1,79}$`)
	reSnapshotPrefix       = regexp.MustCompile(`^[A-Za-z0-9_]{1,59}$`)
	reResourceNamePrefix   = regexp.MustCompile(validResourceNamePrefix)
)

type PlanInformation struct {
	PlanName          string `mapstructure:"plan_name"`
	PlanProduct       string `mapstructure:"plan_product"`
	PlanPublisher     string `mapstructure:"plan_publisher"`
	PlanPromotionCode string `mapstructure:"plan_promotion_code"`
}

type SharedImageGallery struct {
	Subscription  string `mapstructure:"subscription"`
	ResourceGroup string `mapstructure:"resource_group"`
	GalleryName   string `mapstructure:"gallery_name"`
	ImageName     string `mapstructure:"image_name"`
	// Specify a specific version of an OS to boot from.
	// Defaults to latest. There may be a difference in versions available
	// across regions due to image synchronization latency. To ensure a consistent
	// version across regions set this value to one that is available in all
	// regions where you are deploying.
	ImageVersion string `mapstructure:"image_version" required:"false"`
}

type SharedImageGalleryDestination struct {
	SigDestinationSubscription       string   `mapstructure:"subscription"`
	SigDestinationResourceGroup      string   `mapstructure:"resource_group"`
	SigDestinationGalleryName        string   `mapstructure:"gallery_name"`
	SigDestinationImageName          string   `mapstructure:"image_name"`
	SigDestinationImageVersion       string   `mapstructure:"image_version"`
	SigDestinationReplicationRegions []string `mapstructure:"replication_regions"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Authentication via OAUTH
	ClientConfig client.Config `mapstructure:",squash"`

	// A list of one or more fully-qualified resource IDs of user assigned
	// managed identities to be configured on the VM.
	// See [documentation](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/how-to-use-vm-token)
	// for how to acquire tokens within the VM.
	// To assign a user assigned managed identity to a VM, the provided account or service principal must have [Managed Identity Operator](https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#managed-identity-operator)
	// and [Virtual Machine Contributor](https://docs.microsoft.com/en-us/azure/role-based-access-control/built-in-roles#virtual-machine-contributor) role assignments.
	UserAssignedManagedIdentities []string `mapstructure:"user_assigned_managed_identities" required:"false"`

	// VHD prefix.
	CaptureNamePrefix string `mapstructure:"capture_name_prefix"`
	// Destination container name.
	CaptureContainerName string `mapstructure:"capture_container_name"`
	// Use a [Shared Gallery
	// image](https://azure.microsoft.com/en-us/blog/announcing-the-public-preview-of-shared-image-gallery/)
	// as the source for this build. *VHD targets are incompatible with this
	// build type* - the target must be a *Managed Image*. When using shared_image_gallery as a source, image_publisher,
	// image_offer, image_sku, image_version, and custom_managed_image_name should not be set.
	//
	// In JSON
	// ```json
	// "shared_image_gallery": {
	//     "subscription": "00000000-0000-0000-0000-00000000000",
	//     "resource_group": "ResourceGroup",
	//     "gallery_name": "GalleryName",
	//     "image_name": "ImageName",
	//     "image_version": "1.0.0"
	// }
	// "managed_image_name": "TargetImageName",
	// "managed_image_resource_group_name": "TargetResourceGroup"
	// ```
	// In HCL2
	// ```hcl
	// shared_image_gallery {
	//     subscription = "00000000-0000-0000-0000-00000000000"
	//     resource_group = "ResourceGroup"
	//     gallery_name = "GalleryName"
	//     image_name = "ImageName"
	//     image_version = "1.0.0"
	// }
	// managed_image_name = "TargetImageName"
	// managed_image_resource_group_name = "TargetResourceGroup"
	// ```
	SharedGallery SharedImageGallery `mapstructure:"shared_image_gallery" required:"false"`
	// The name of the Shared Image Gallery under which the managed image will be published as Shared Gallery Image version.
	//
	// Following is an example.
	//
	// In JSON
	// ```json
	// "shared_image_gallery_destination": {
	//     "subscription": "00000000-0000-0000-0000-00000000000",
	//     "resource_group": "ResourceGroup",
	//     "gallery_name": "GalleryName",
	//     "image_name": "ImageName",
	//     "image_version": "1.0.0",
	//     "replication_regions": ["regionA", "regionB", "regionC"]
	// }
	// "managed_image_name": "TargetImageName",
	// "managed_image_resource_group_name": "TargetResourceGroup"
	// ```
	// In HCL2
	// ```hcl
	// shared_image_gallery_destination {
	//     subscription = "00000000-0000-0000-0000-00000000000"
	//     resource_group = "ResourceGroup"
	//     gallery_name = "GalleryName"
	//     image_name = "ImageName"
	//     image_version = "1.0.0"
	//     replication_regions = ["regionA", "regionB", "regionC"]
	// }
	// managed_image_name = "TargetImageName"
	// managed_image_resource_group_name = "TargetResourceGroup"
	// ```
	SharedGalleryDestination SharedImageGalleryDestination `mapstructure:"shared_image_gallery_destination"`
	// How long to wait for an image to be published to the shared image
	// gallery before timing out. If your Packer build is failing on the
	// Publishing to Shared Image Gallery step with the error `Original Error:
	// context deadline exceeded`, but the image is present when you check your
	// Azure dashboard, then you probably need to increase this timeout from
	// its default of "60m" (valid time units include `s` for seconds, `m` for
	// minutes, and `h` for hours.)
	SharedGalleryTimeout time.Duration `mapstructure:"shared_image_gallery_timeout"`
	// The end of life date (2006-01-02T15:04:05.99Z) of the gallery Image Version. This property
	// can be used for decommissioning purposes.
	SharedGalleryImageVersionEndOfLifeDate string `mapstructure:"shared_gallery_image_version_end_of_life_date" required:"false"`
	// The number of replicas of the Image Version to be created per region. This
	// property would take effect for a region when regionalReplicaCount is not specified.
	// Replica count must be between 1 and 10.
	SharedGalleryImageVersionReplicaCount int32 `mapstructure:"shared_image_gallery_replica_count" required:"false"`
	// If set to true, Virtual Machines deployed from the latest version of the
	// Image Definition won't use this Image Version.
	SharedGalleryImageVersionExcludeFromLatest bool `mapstructure:"shared_gallery_image_version_exclude_from_latest" required:"false"`
	// Name of the publisher to use for your base image (Azure Marketplace Images only). See
	// [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
	// for details.
	//
	// CLI example `az vm image list-publishers --location westus`
	ImagePublisher string `mapstructure:"image_publisher" required:"true"`
	// Name of the publisher's offer to use for your base image (Azure Marketplace Images only). See
	// [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
	// for details.
	//
	// CLI example
	// `az vm image list-offers --location westus --publisher Canonical`
	ImageOffer string `mapstructure:"image_offer" required:"true"`
	// SKU of the image offer to use for your base image (Azure Marketplace Images only). See
	// [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
	// for details.
	//
	// CLI example
	// `az vm image list-skus --location westus --publisher Canonical --offer UbuntuServer`
	ImageSku string `mapstructure:"image_sku" required:"true"`
	// Specify a specific version of an OS to boot from.
	// Defaults to `latest`. There may be a difference in versions available
	// across regions due to image synchronization latency. To ensure a consistent
	// version across regions set this value to one that is available in all
	// regions where you are deploying.
	//
	// CLI example
	// `az vm image list --location westus --publisher Canonical --offer UbuntuServer --sku 16.04.0-LTS --all`
	ImageVersion string `mapstructure:"image_version" required:"false"`
	// URL to a custom VHD to use for your base image. If this value is set,
	// image_publisher, image_offer, image_sku, or image_version should not be set.
	ImageUrl string `mapstructure:"image_url" required:"true"`
	// Name of a custom managed image to use for your base image. If this value is set, do
	// not set image_publisher, image_offer, image_sku, or image_version.
	// If this value is set, the option
	// `custom_managed_image_resource_group_name` must also be set. See
	// [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
	// to learn more about managed images.
	CustomManagedImageName string `mapstructure:"custom_managed_image_name" required:"true"`

	// Name of a custom managed image's resource group to use for your base image. If this
	// value is set, image_publisher, image_offer, image_sku, or image_version should not be set.
	// If this value is set, the option
	// `custom_managed_image_name` must also be set. See
	// [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
	// to learn more about managed images.
	CustomManagedImageResourceGroupName string `mapstructure:"custom_managed_image_resource_group_name" required:"true"`
	customManagedImageID                string

	// Azure datacenter in which your VM will build.
	Location string `mapstructure:"location"`
	// Size of the VM used for building. This can be changed when you deploy a
	// VM from your VHD. See
	// [pricing](https://azure.microsoft.com/en-us/pricing/details/virtual-machines/)
	// information. Defaults to `Standard_A1`.
	//
	// CLI example `az vm list-sizes --location westus`
	VMSize string `mapstructure:"vm_size" required:"false"`

	// Specify the managed image resource group name where the result of the
	// Packer build will be saved. The resource group must already exist. If
	// this value is set, the value managed_image_name must also be set. See
	// documentation to learn more about managed images.
	ManagedImageResourceGroupName string `mapstructure:"managed_image_resource_group_name"`
	// Specify the managed image name where the result of the Packer build will
	// be saved. The image name must not exist ahead of time, and will not be
	// overwritten. If this value is set, the value
	// managed_image_resource_group_name must also be set. See documentation to
	// learn more about managed images.
	ManagedImageName string `mapstructure:"managed_image_name"`
	// Specify the storage account
	// type for a managed image. Valid values are Standard_LRS and Premium_LRS.
	// The default is Standard_LRS.
	ManagedImageStorageAccountType string `mapstructure:"managed_image_storage_account_type" required:"false"`
	managedImageStorageAccountType compute.StorageAccountTypes
	// If
	// managed_image_os_disk_snapshot_name is set, a snapshot of the OS disk
	// is created with the same name as this value before the VM is captured.
	ManagedImageOSDiskSnapshotName string `mapstructure:"managed_image_os_disk_snapshot_name" required:"false"`
	// If
	// managed_image_data_disk_snapshot_prefix is set, snapshot of the data
	// disk(s) is created with the same prefix as this value before the VM is
	// captured.
	ManagedImageDataDiskSnapshotPrefix string `mapstructure:"managed_image_data_disk_snapshot_prefix" required:"false"`
	// Store the image in zone-resilient storage. You need to create it in a
	// region that supports [availability
	// zones](https://docs.microsoft.com/en-us/azure/availability-zones/az-overview).
	ManagedImageZoneResilient bool `mapstructure:"managed_image_zone_resilient" required:"false"`
	// Name/value pair tags to apply to every resource deployed i.e. Resource
	// Group, VM, NIC, VNET, Public IP, KeyVault, etc. The user can define up
	// to 15 tags. Tag names cannot exceed 512 characters, and tag values
	// cannot exceed 256 characters.
	AzureTags map[string]string `mapstructure:"azure_tags" required:"false"`
	// Same as [`azure_tags`](#azure_tags) but defined as a singular repeatable block
	// containing a `name` and a `value` field. In HCL2 mode the
	// [`dynamic_block`](/docs/configuration/from-1.5/expressions#dynamic-blocks)
	// will allow you to create those programatically.
	AzureTag config.NameValues `mapstructure:"azure_tag" required:"false"`
	// Resource group under which the final artifact will be stored.
	ResourceGroupName string `mapstructure:"resource_group_name"`
	// Storage account under which the final artifact will be stored.
	StorageAccount string `mapstructure:"storage_account"`
	// temporary name assigned to the VM. If this
	// value is not set, a random value will be assigned. Knowing the resource
	// group and VM name allows one to execute commands to update the VM during a
	// Packer build, e.g. attach a resource disk to the VM.
	TempComputeName string `mapstructure:"temp_compute_name" required:"false"`
	// name assigned to the temporary resource group created during the build.
	// If this value is not set, a random value will be assigned. This resource
	// group is deleted at the end of the build.
	TempResourceGroupName string `mapstructure:"temp_resource_group_name"`
	// Specify an existing resource group to run the build in.
	BuildResourceGroupName string `mapstructure:"build_resource_group_name"`
	// Specify an existing key vault to use for uploading certificates to the
	// instance to connect.
	BuildKeyVaultName string `mapstructure:"build_key_vault_name"`
	// Specify the KeyVault SKU to create during the build. Valid values are
	// standard or premium. The default value is standard.
	BuildKeyVaultSKU           string `mapstructure:"build_key_vault_sku"`
	storageAccountBlobEndpoint string
	// This value allows you to
	// set a virtual_network_name and obtain a public IP. If this value is not
	// set and virtual_network_name is defined Packer is only allowed to be
	// executed from a host on the same subnet / virtual network.
	PrivateVirtualNetworkWithPublicIp bool `mapstructure:"private_virtual_network_with_public_ip" required:"false"`
	// Use a pre-existing virtual network for the
	// VM. This option enables private communication with the VM, no public IP
	// address is used or provisioned (unless you set
	// private_virtual_network_with_public_ip).
	VirtualNetworkName string `mapstructure:"virtual_network_name" required:"false"`
	// If virtual_network_name is set,
	// this value may also be set. If virtual_network_name is set, and this
	// value is not set the builder attempts to determine the subnet to use with
	// the virtual network. If the subnet cannot be found, or it cannot be
	// disambiguated, this value should be set.
	VirtualNetworkSubnetName string `mapstructure:"virtual_network_subnet_name" required:"false"`
	// If virtual_network_name is
	// set, this value may also be set. If virtual_network_name is set, and
	// this value is not set the builder attempts to determine the resource group
	// containing the virtual network. If the resource group cannot be found, or
	// it cannot be disambiguated, this value should be set.
	VirtualNetworkResourceGroupName string `mapstructure:"virtual_network_resource_group_name" required:"false"`
	// Specify a file containing custom data to inject into the cloud-init
	// process. The contents of the file are read and injected into the ARM
	// template. The custom data will be passed to cloud-init for processing at
	// the time of provisioning. See
	// [documentation](http://cloudinit.readthedocs.io/en/latest/topics/examples.html)
	// to learn more about custom data, and how it can be used to influence the
	// provisioning process.
	CustomDataFile string `mapstructure:"custom_data_file" required:"false"`
	customData     string
	// Used for creating images from Marketplace images. Please refer to
	// [Deploy an image with Marketplace
	// terms](https://aka.ms/azuremarketplaceapideployment) for more details.
	// Not all Marketplace images support programmatic deployment, and support
	// is controlled by the image publisher.
	//
	// An example plan\_info object is defined below.
	//
	// ```json
	// {
	//   "plan_info": {
	//       "plan_name": "rabbitmq",
	//       "plan_product": "rabbitmq",
	//       "plan_publisher": "bitnami"
	//   }
	// }
	// ```
	//
	// `plan_name` (string) - The plan name, required. `plan_product` (string) -
	// The plan product, required. `plan_publisher` (string) - The plan publisher,
	// required. `plan_promotion_code` (string) - Some images accept a promotion
	// code, optional.
	//
	// Images created from the Marketplace with `plan_info` **must** specify
	// `plan_info` whenever the image is deployed. The builder automatically adds
	// tags to the image to ensure this information is not lost. The following
	// tags are added.
	//
	// ```text
	// 1.  PlanName
	// 2.  PlanProduct
	// 3.  PlanPublisher
	// 4.  PlanPromotionCode
	// ```
	//
	PlanInfo PlanInformation `mapstructure:"plan_info" required:"false"`
	// The default PollingDuration for azure is 15mins, this property will override
	// that value. See [Azure DefaultPollingDuration](https://godoc.org/github.com/Azure/go-autorest/autorest#pkg-constants)
	// If your Packer build is failing on the
	// ARM deployment step with the error `Original Error:
	// context deadline exceeded`, then you probably need to increase this timeout from
	// its default of "15m" (valid time units include `s` for seconds, `m` for
	// minutes, and `h` for hours.)
	PollingDurationTimeout time.Duration `mapstructure:"polling_duration_timeout" required:"false"`
	// If either Linux or Windows is specified Packer will
	// automatically configure authentication credentials for the provisioned
	// machine. For Linux this configures an SSH authorized key. For Windows
	// this configures a WinRM certificate.
	OSType string `mapstructure:"os_type" required:"false"`
	// Specify the size of the OS disk in GB
	// (gigabytes). Values of zero or less than zero are ignored.
	OSDiskSizeGB int32 `mapstructure:"os_disk_size_gb" required:"false"`
	// The size(s) of any additional hard disks for the VM in gigabytes. If
	// this is not specified then the VM will only contain an OS disk. The
	// number of additional disks and maximum size of a disk depends on the
	// configuration of your VM. See
	// [Windows](https://docs.microsoft.com/en-us/azure/virtual-machines/windows/about-disks-and-vhds)
	// or
	// [Linux](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/about-disks-and-vhds)
	// for more information.
	//
	// For VHD builds the final artifacts will be named
	// `PREFIX-dataDisk-<n>.UUID.vhd` and stored in the specified capture
	// container along side the OS disk. The additional disks are included in
	// the deployment template `PREFIX-vmTemplate.UUID`.
	//
	// For Managed build the final artifacts are included in the managed image.
	// The additional disk will have the same storage account type as the OS
	// disk, as specified with the `managed_image_storage_account_type`
	// setting.
	AdditionalDiskSize []int32 `mapstructure:"disk_additional_size" required:"false"`
	// Specify the disk caching type. Valid values
	// are None, ReadOnly, and ReadWrite. The default value is ReadWrite.
	DiskCachingType string `mapstructure:"disk_caching_type" required:"false"`
	diskCachingType compute.CachingTypes
	// Specify the list of IP addresses and CIDR blocks that should be
	// allowed access to the VM. If provided, an Azure Network Security
	// Group will be created with corresponding rules and be bound to
	// the subnet of the VM.
	// Providing `allowed_inbound_ip_addresses` in combination with
	// `virtual_network_name` is not allowed.
	AllowedInboundIpAddresses []string `mapstructure:"allowed_inbound_ip_addresses"`

	// Specify storage to store Boot Diagnostics -- Enabling this option
	// will create 2 Files in the specified storage account. (serial console log & screehshot file)
	// once the build is completed, it has to be removed manually.
	// see [here](https://docs.microsoft.com/en-us/azure/virtual-machines/troubleshooting/boot-diagnostics) for more info
	BootDiagSTGAccount string `mapstructure:"boot_diag_storage_account" required:"false"`

	// specify custom azure resource names during build limited to max 10 characters
	// this will set the prefix for the resources. The actuall resource names will be
	// `custom_resource_build_prefix` + resourcetype + 5 character random alphanumeric string
	CustomResourcePrefix string `mapstructure:"custom_resource_build_prefix" required:"false"`

	// Runtime Values
	UserName               string `mapstructure-to-hcl2:",skip"`
	Password               string `mapstructure-to-hcl2:",skip"`
	tmpAdminPassword       string
	tmpCertificatePassword string
	tmpResourceGroupName   string
	tmpComputeName         string
	tmpNicName             string
	tmpPublicIPAddressName string
	tmpDeploymentName      string
	tmpKeyVaultName        string
	tmpOSDiskName          string
	tmpDataDiskName        string
	tmpSubnetName          string
	tmpVirtualNetworkName  string
	tmpNsgName             string
	tmpWinRMCertificateUrl string

	// Authentication with the VM via SSH
	sshAuthorizedKey string

	// Authentication with the VM via WinRM
	winrmCertificate string

	Comm communicator.Config `mapstructure:",squash"`
	ctx  interpolate.Context
	// If you want packer to delete the
	// temporary resource group asynchronously set this value. It's a boolean
	// value and defaults to false. Important Setting this true means that
	// your builds are faster, however any failed deletes are not reported.
	AsyncResourceGroupDelete bool `mapstructure:"async_resourcegroup_delete" required:"false"`
}

type keyVaultCertificate struct {
	Data     string `json:"data"`
	DataType string `json:"dataType"`
	Password string `json:"password,omitempty"`
}

func (c *Config) toVMID() string {
	var resourceGroupName string
	if c.tmpResourceGroupName != "" {
		resourceGroupName = c.tmpResourceGroupName
	} else {
		resourceGroupName = c.BuildResourceGroupName
	}
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", c.ClientConfig.SubscriptionID, resourceGroupName, c.tmpComputeName)
}

func (c *Config) isManagedImage() bool {
	return c.ManagedImageName != ""
}

func (c *Config) toVirtualMachineCaptureParameters() *compute.VirtualMachineCaptureParameters {
	return &compute.VirtualMachineCaptureParameters{
		DestinationContainerName: &c.CaptureContainerName,
		VhdPrefix:                &c.CaptureNamePrefix,
		OverwriteVhds:            to.BoolPtr(false),
	}
}

func (c *Config) toImageParameters() *compute.Image {
	return &compute.Image{
		ImageProperties: &compute.ImageProperties{
			SourceVirtualMachine: &compute.SubResource{
				ID: to.StringPtr(c.toVMID()),
			},
			StorageProfile: &compute.ImageStorageProfile{
				ZoneResilient: to.BoolPtr(c.ManagedImageZoneResilient),
			},
		},
		Location: to.StringPtr(c.Location),
		Tags:     azcommon.MapToAzureTags(c.AzureTags),
	}
}

func (c *Config) createCertificate() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("Failed to Generate Private Key: %s", err)
		return "", err
	}

	host := fmt.Sprintf("%s.cloudapp.net", c.tmpComputeName)
	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		err = fmt.Errorf("Failed to Generate Serial Number: %v", err)
		return "", err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Issuer: pkix.Name{
			CommonName: host,
		},
		Subject: pkix.Name{
			CommonName: host,
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		err = fmt.Errorf("Failed to Create Certificate: %s", err)
		return "", err
	}

	pfxBytes, err := pkcs12.Encode(derBytes, privateKey, c.tmpCertificatePassword)
	if err != nil {
		err = fmt.Errorf("Failed to encode certificate as PFX: %s", err)
		return "", err
	}

	keyVaultDescription := keyVaultCertificate{
		Data:     base64.StdEncoding.EncodeToString(pfxBytes),
		DataType: "pfx",
		Password: c.tmpCertificatePassword,
	}

	bytes, err := json.Marshal(keyVaultDescription)
	if err != nil {
		err = fmt.Errorf("Failed to marshal key vault description: %s", err)
		return "", err
	}

	return base64.StdEncoding.EncodeToString(bytes), nil
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	c.ctx.Funcs = azcommon.TemplateFuncs
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)

	if err != nil {
		return nil, err
	}

	provideDefaultValues(c)
	setRuntimeValues(c)
	err = setUserNamePassword(c)
	if err != nil {
		return nil, err
	}

	// copy singular blocks
	c.AzureTag.CopyOn(&c.AzureTags)

	err = c.ClientConfig.SetDefaultValues()
	if err != nil {
		return nil, err
	}

	err = setCustomData(c)
	if err != nil {
		return nil, err
	}

	// NOTE: if the user did not specify a communicator, then default to both
	// SSH and WinRM.  This is for backwards compatibility because the code did
	// not specifically force the user to set a communicator.
	if c.Comm.Type == "" || strings.EqualFold(c.Comm.Type, "ssh") {
		err = setSshValues(c)
		if err != nil {
			return nil, err
		}
	}

	if c.Comm.Type == "" || strings.EqualFold(c.Comm.Type, "winrm") {
		err = setWinRMCertificate(c)
		if err != nil {
			return nil, err
		}
	}

	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)

	assertRequiredParametersSet(c, errs)
	assertTagProperties(c, errs)
	if errs != nil && len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

func setSshValues(c *Config) error {
	if c.Comm.SSHTimeout == 0 {
		c.Comm.SSHTimeout = 20 * time.Minute
	}

	if c.Comm.SSHPrivateKeyFile != "" {
		privateKeyBytes, err := c.Comm.ReadSSHPrivateKeyFile()
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			return err
		}

		publicKey := signer.PublicKey()
		c.sshAuthorizedKey = fmt.Sprintf("%s %s packer Azure Deployment%s",
			publicKey.Type(),
			base64.StdEncoding.EncodeToString(publicKey.Marshal()),
			time.Now().Format(time.RFC3339))
		c.Comm.SSHPrivateKey = privateKeyBytes

	} else {
		sshKeyPair, err := NewOpenSshKeyPair()
		if err != nil {
			return err
		}

		c.sshAuthorizedKey = sshKeyPair.AuthorizedKey()
		c.Comm.SSHPrivateKey = sshKeyPair.PrivateKey()
	}

	return nil
}

func setWinRMCertificate(c *Config) error {
	c.Comm.WinRMTransportDecorator =
		func() winrm.Transporter {
			return &winrm.ClientNTLM{}
		}

	cert, err := c.createCertificate()
	c.winrmCertificate = cert

	return err
}

func setRuntimeValues(c *Config) {
	var tempName = NewTempName(c.CustomResourcePrefix)

	c.tmpAdminPassword = tempName.AdminPassword
	// store so that we can access this later during provisioning
	packersdk.LogSecretFilter.Set(c.tmpAdminPassword)

	c.tmpCertificatePassword = tempName.CertificatePassword
	if c.TempComputeName == "" {
		c.tmpComputeName = tempName.ComputeName
	} else {
		c.tmpComputeName = c.TempComputeName
	}
	c.tmpDeploymentName = tempName.DeploymentName
	// Only set tmpResourceGroupName if no name has been specified
	if c.TempResourceGroupName == "" && c.BuildResourceGroupName == "" {
		c.tmpResourceGroupName = tempName.ResourceGroupName
	} else if c.TempResourceGroupName != "" && c.BuildResourceGroupName == "" {
		c.tmpResourceGroupName = c.TempResourceGroupName
	}
	c.tmpNicName = tempName.NicName
	c.tmpPublicIPAddressName = tempName.PublicIPAddressName
	c.tmpOSDiskName = tempName.OSDiskName
	c.tmpDataDiskName = tempName.DataDiskName
	c.tmpSubnetName = tempName.SubnetName
	c.tmpVirtualNetworkName = tempName.VirtualNetworkName
	c.tmpNsgName = tempName.NsgName
	c.tmpKeyVaultName = tempName.KeyVaultName
}

func setUserNamePassword(c *Config) error {
	// Set default credentials generated by the builder
	c.UserName = DefaultUserName
	c.Password = c.tmpAdminPassword

	// Set communicator specific credentials and update defaults if different.
	// Communicator specific credentials need to be updated as the standard Packer
	// SSHConfigFunc and WinRMConfigFunc use communicator specific credentials, unless overwritten.

	// SSH comm
	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = c.UserName
	}
	c.UserName = c.Comm.SSHUsername

	// if user has an explicit wish to use an SSH password, we'll set it
	if c.Comm.SSHPassword != "" {
		c.Password = c.Comm.SSHPassword
	}

	if c.Comm.Type == "ssh" {
		return nil
	}

	// WinRM comm
	if c.Comm.WinRMUser == "" {
		c.Comm.WinRMUser = c.UserName
	}
	c.UserName = c.Comm.WinRMUser

	if c.Comm.WinRMPassword == "" {
		// Configure password settings using Azure generated credentials
		c.Comm.WinRMPassword = c.Password
	}

	if !isValidPassword(c.Comm.WinRMPassword) {
		return fmt.Errorf("The supplied \"winrm_password\" must be between 8-123 characters long and must satisfy at least 3 from the following: \n1) Contains an uppercase character \n2) Contains a lowercase character\n3) Contains a numeric digit\n4) Contains a special character\n5) Control characters are not allowed")
	}
	c.Password = c.Comm.WinRMPassword

	return nil
}

func setCustomData(c *Config) error {
	if c.CustomDataFile == "" {
		return nil
	}

	b, err := ioutil.ReadFile(c.CustomDataFile)
	if err != nil {
		return err
	}

	c.customData = base64.StdEncoding.EncodeToString(b)
	return nil
}

func provideDefaultValues(c *Config) {
	if c.VMSize == "" {
		c.VMSize = DefaultVMSize
	}

	if c.ManagedImageStorageAccountType == "" {
		c.managedImageStorageAccountType = compute.StorageAccountTypesStandardLRS
	}

	if c.DiskCachingType == "" {
		c.diskCachingType = compute.CachingTypesReadWrite
	}

	if c.ImagePublisher != "" && c.ImageVersion == "" {
		c.ImageVersion = DefaultImageVersion
	}

	if c.BuildKeyVaultSKU == "" {
		c.BuildKeyVaultSKU = DefaultKeyVaultSKU
	}

	c.ClientConfig.SetDefaultValues()
}

func assertTagProperties(c *Config, errs *packersdk.MultiError) {
	if len(c.AzureTags) > 15 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("a max of 15 tags are supported, but %d were provided", len(c.AzureTags)))
	}

	for k, v := range c.AzureTags {
		if len(k) > 512 {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the tag name %q exceeds (%d) the 512 character limit", k, len(k)))
		}
		if len(v) > 256 {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the tag name %q exceeds (%d) the 256 character limit", v, len(v)))
		}
	}
}

func assertRequiredParametersSet(c *Config, errs *packersdk.MultiError) {
	c.ClientConfig.Validate(errs)

	/////////////////////////////////////////////
	// Identity
	if len(c.UserAssignedManagedIdentities) != 0 {
		for _, rid := range c.UserAssignedManagedIdentities {
			r, err := client.ParseResourceID(rid)
			if err != nil {
				err := fmt.Errorf("Error parsing resource ID from `user_assigned_managed_identities`; please make sure"+
					" that this value follows the full resource id format: "+
					"/subscriptions/<SUBSCRIPTON_ID>/resourcegroups/<RESOURCE_GROUP>/providers/Microsoft.ManagedIdentity/userAssignedIdentities/<USER_ASSIGNED_IDENTITY_NAME>.\n"+
					" Original error: %s", err)
				errs = packersdk.MultiErrorAppend(errs, err)
			} else {
				if !strings.EqualFold(r.Provider, "Microsoft.ManagedIdentity") {
					errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A valid user assigned managed identity resource id must have a correct resource provider"))
				}
				if !strings.EqualFold(r.ResourceType.String(), "userAssignedIdentities") {
					errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A valid user assigned managed identity resource id must have a correct resource type"))
				}
			}
		}
	}

	/////////////////////////////////////////////
	// Capture
	if c.CaptureContainerName == "" && c.ManagedImageName == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name or managed_image_name must be specified"))
	}

	if c.CaptureNamePrefix == "" && c.ManagedImageResourceGroupName == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix or managed_image_resource_group_name must be specified"))
	}

	if (c.CaptureNamePrefix != "" || c.CaptureContainerName != "") && (c.ManagedImageResourceGroupName != "" || c.ManagedImageName != "") {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Either a VHD or a managed image can be built, but not both. Please specify either capture_container_name and capture_name_prefix or managed_image_resource_group_name and managed_image_name."))
	}

	if c.CaptureContainerName != "" {
		if !reCaptureContainerName.MatchString(c.CaptureContainerName) {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name must satisfy the regular expression %q.", reCaptureContainerName.String()))
		}

		if strings.HasSuffix(c.CaptureContainerName, "-") {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name must not end with a hyphen, e.g. '-'."))
		}

		if strings.Contains(c.CaptureContainerName, "--") {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name must not contain consecutive hyphens, e.g. '--'."))
		}

		if c.CaptureNamePrefix == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix must be specified"))
		}

		if !reCaptureNamePrefix.MatchString(c.CaptureNamePrefix) {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix must satisfy the regular expression %q.", reCaptureNamePrefix.String()))
		}

		if strings.HasSuffix(c.CaptureNamePrefix, "-") || strings.HasSuffix(c.CaptureNamePrefix, ".") {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix must not end with a hyphen or period."))
		}
	}

	if c.TempResourceGroupName != "" && c.BuildResourceGroupName != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The settings temp_resource_group_name and build_resource_group_name cannot both be defined.  Please define one or neither."))
	}

	/////////////////////////////////////////////
	// Compute
	toInt := func(b bool) int {
		if b {
			return 1
		} else {
			return 0
		}
	}

	isImageUrl := c.ImageUrl != ""
	isCustomManagedImage := c.CustomManagedImageName != "" || c.CustomManagedImageResourceGroupName != ""
	isSharedGallery := c.SharedGallery.GalleryName != ""
	isPlatformImage := c.ImagePublisher != "" || c.ImageOffer != "" || c.ImageSku != ""

	countSourceInputs := toInt(isImageUrl) + toInt(isCustomManagedImage) + toInt(isPlatformImage) + toInt(isSharedGallery)

	if countSourceInputs > 1 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Specify either a VHD (image_url), Image Reference (image_publisher, image_offer, image_sku), a Managed Disk (custom_managed_disk_image_name, custom_managed_disk_resource_group_name), or a Shared Gallery Image (shared_image_gallery)"))
	}

	if isImageUrl && c.ManagedImageResourceGroupName != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A managed image must be created from a managed image, it cannot be created from a VHD."))
	}

	if c.SharedGallery.GalleryName != "" {
		if c.SharedGallery.Subscription == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A shared_image_gallery.subscription must be specified"))
		}
		if c.SharedGallery.ResourceGroup == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A shared_image_gallery.resource_group must be specified"))
		}
		if c.SharedGallery.ImageName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A shared_image_gallery.image_name must be specified"))
		}
		if c.CaptureContainerName != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("VHD Target [capture_container_name] is not supported when using Shared Image Gallery as source. Use managed_image_resource_group_name instead."))
		}
		if c.CaptureNamePrefix != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("VHD Target [capture_name_prefix] is not supported when using Shared Image Gallery as source. Use managed_image_name instead."))
		}
	} else if c.ImageUrl == "" && c.CustomManagedImageName == "" {
		if c.ImagePublisher == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An image_publisher must be specified"))
		}
		if c.ImageOffer == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An image_offer must be specified"))
		}
		if c.ImageSku == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An image_sku must be specified"))
		}
	} else if c.ImageUrl == "" && c.ImagePublisher == "" {
		if c.CustomManagedImageResourceGroupName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A custom_managed_image_resource_group_name must be specified"))
		}
		if c.CustomManagedImageName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A custom_managed_image_name must be specified"))
		}
		if c.ManagedImageResourceGroupName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A managed_image_resource_group_name must be specified"))
		}
		if c.ManagedImageName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A managed_image_name must be specified"))
		}
	} else {
		if c.ImagePublisher != "" || c.ImageOffer != "" || c.ImageSku != "" || c.ImageVersion != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An image_url must not be specified if image_publisher, image_offer, image_sku, or image_version is specified"))
		}
	}

	/////////////////////////////////////////////
	// Deployment
	xor := func(a, b bool) bool {
		return (a || b) && !(a && b)
	}

	if !xor((c.StorageAccount != "" || c.ResourceGroupName != ""), (c.ManagedImageName != "" || c.ManagedImageResourceGroupName != "")) {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Specify either a VHD (storage_account and resource_group_name) or Managed Image (managed_image_resource_group_name and managed_image_name) output"))
	}

	if !xor(c.Location != "", c.BuildResourceGroupName != "") {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("Specify either a location to create the resource group in or an existing build_resource_group_name, but not both."))
	}

	if c.ManagedImageName == "" && c.ManagedImageResourceGroupName == "" {
		if c.StorageAccount == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A storage_account must be specified"))
		}
		if c.ResourceGroupName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A resource_group_name must be specified"))
		}
	}

	if c.TempResourceGroupName != "" {
		if ok, err := assertResourceGroupName(c.TempResourceGroupName, "temp_resource_group_name"); !ok {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.BuildResourceGroupName != "" {
		if ok, err := assertResourceGroupName(c.BuildResourceGroupName, "build_resource_group_name"); !ok {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageResourceGroupName != "" {
		if ok, err := assertResourceGroupName(c.ManagedImageResourceGroupName, "managed_image_resource_group_name"); !ok {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageName != "" {
		if ok, err := assertManagedImageName(c.ManagedImageName, "managed_image_name"); !ok {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageName != "" && c.ManagedImageResourceGroupName != "" && c.SharedGalleryDestination.SigDestinationGalleryName != "" {
		if c.SharedGalleryDestination.SigDestinationResourceGroup == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A resource_group must be specified for shared_image_gallery_destination"))
		}
		if c.SharedGalleryDestination.SigDestinationImageName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An image_name must be specified for shared_image_gallery_destination"))
		}
		if c.SharedGalleryDestination.SigDestinationImageVersion == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An image_version must be specified for shared_image_gallery_destination"))
		}
		if len(c.SharedGalleryDestination.SigDestinationReplicationRegions) == 0 {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A list of replication_regions must be specified for shared_image_gallery_destination"))
		}
		if c.SharedGalleryDestination.SigDestinationSubscription == "" {
			c.SharedGalleryDestination.SigDestinationSubscription = c.ClientConfig.SubscriptionID
		}
	}
	if c.SharedGalleryTimeout == 0 {
		// default to a one-hour timeout. In the sdk, the default is 15 m.
		c.SharedGalleryTimeout = 60 * time.Minute
	}

	if c.ManagedImageOSDiskSnapshotName != "" {
		if ok, err := assertManagedImageOSDiskSnapshotName(c.ManagedImageOSDiskSnapshotName, "managed_image_os_disk_snapshot_name"); !ok {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageDataDiskSnapshotPrefix != "" {
		if ok, err := assertManagedImageDataDiskSnapshotName(c.ManagedImageDataDiskSnapshotPrefix, "managed_image_data_disk_snapshot_prefix"); !ok {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.CustomResourcePrefix != "" {
		if ok, err := assertResourceNamePrefix(c.CustomResourcePrefix, "custom_resource_build_prefix"); !ok {
			errs = packersdk.MultiErrorAppend(errs, err)
		}
	}

	if c.VirtualNetworkName == "" && c.VirtualNetworkResourceGroupName != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_resource_group_name is specified, so must virtual_network_name"))
	}
	if c.VirtualNetworkName == "" && c.VirtualNetworkSubnetName != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_subnet_name is specified, so must virtual_network_name"))
	}

	if c.AllowedInboundIpAddresses != nil && len(c.AllowedInboundIpAddresses) >= 1 {
		if c.VirtualNetworkName != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_name is specified, allowed_inbound_ip_addresses cannot be specified"))
		} else {
			if ok, err := assertAllowedInboundIpAddresses(c.AllowedInboundIpAddresses, "allowed_inbound_ip_addresses"); !ok {
				errs = packersdk.MultiErrorAppend(errs, err)
			}
		}
	}

	/////////////////////////////////////////////
	// Plan Info
	if c.PlanInfo.PlanName != "" || c.PlanInfo.PlanProduct != "" || c.PlanInfo.PlanPublisher != "" || c.PlanInfo.PlanPromotionCode != "" {
		if c.PlanInfo.PlanName == "" || c.PlanInfo.PlanProduct == "" || c.PlanInfo.PlanPublisher == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("if either plan_name, plan_product, plan_publisher, or plan_promotion_code are defined then plan_name, plan_product, and plan_publisher must be defined"))
		} else {
			if c.AzureTags == nil {
				c.AzureTags = make(map[string]string)
			}

			c.AzureTags["PlanInfo"] = c.PlanInfo.PlanName
			c.AzureTags["PlanProduct"] = c.PlanInfo.PlanProduct
			c.AzureTags["PlanPublisher"] = c.PlanInfo.PlanPublisher
			c.AzureTags["PlanPromotionCode"] = c.PlanInfo.PlanPromotionCode
		}
	}

	/////////////////////////////////////////////
	// Polling Duration Timeout
	if c.PollingDurationTimeout == 0 {
		// In the sdk, the default is 15 m.
		c.PollingDurationTimeout = 15 * time.Minute
	}

	/////////////////////////////////////////////
	// OS
	if strings.EqualFold(c.OSType, constants.Target_Linux) {
		c.OSType = constants.Target_Linux
	} else if strings.EqualFold(c.OSType, constants.Target_Windows) {
		c.OSType = constants.Target_Windows
	} else if c.OSType == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An os_type must be specified"))
	} else {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The os_type %q is invalid", c.OSType))
	}

	switch c.ManagedImageStorageAccountType {
	case "", string(compute.StorageAccountTypesStandardLRS):
		c.managedImageStorageAccountType = compute.StorageAccountTypesStandardLRS
	case string(compute.StorageAccountTypesPremiumLRS):
		c.managedImageStorageAccountType = compute.StorageAccountTypesPremiumLRS
	default:
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The managed_image_storage_account_type %q is invalid", c.ManagedImageStorageAccountType))
	}

	switch c.DiskCachingType {
	case string(compute.CachingTypesNone):
		c.diskCachingType = compute.CachingTypesNone
	case string(compute.CachingTypesReadOnly):
		c.diskCachingType = compute.CachingTypesReadOnly
	case "", string(compute.CachingTypesReadWrite):
		c.diskCachingType = compute.CachingTypesReadWrite
	default:
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The disk_caching_type %q is invalid", c.DiskCachingType))
	}
}

func assertManagedImageName(name, setting string) (bool, error) {
	if !isValidAzureName(reManagedDiskName, name) {
		return false, fmt.Errorf("The setting %s must match the regular expression %q, and not end with a '-' or '.'.", setting, validManagedDiskName)
	}
	return true, nil
}

func assertManagedImageOSDiskSnapshotName(name, setting string) (bool, error) {
	if !isValidAzureName(reSnapshotName, name) {
		return false, fmt.Errorf("The setting %s must only contain characters from a-z, A-Z, 0-9 and _ and the maximum length is 80 characters", setting)
	}
	return true, nil
}

func assertManagedImageDataDiskSnapshotName(name, setting string) (bool, error) {
	if !isValidAzureName(reSnapshotPrefix, name) {
		return false, fmt.Errorf("The setting %s must only contain characters from a-z, A-Z, 0-9 and _ and the maximum length (excluding the prefix) is 60 characters", setting)
	}
	return true, nil
}

func assertResourceNamePrefix(name, setting string) (bool, error) {
	if !isValidAzureName(reResourceNamePrefix, name) {
		return false, fmt.Errorf("The setting %s must only contain characters from a-z, A-Z, 0-9 and _ and the maximum length is 10 characters", setting)
	}
	return true, nil
}

func assertAllowedInboundIpAddresses(ipAddresses []string, setting string) (bool, error) {
	for _, ipAddress := range ipAddresses {
		if net.ParseIP(ipAddress) == nil {
			if _, _, err := net.ParseCIDR(ipAddress); err != nil {
				return false, fmt.Errorf("The setting %s must only contain valid IP addresses or CIDR blocks", setting)
			}
		}
	}
	return true, nil
}

func assertResourceGroupName(rgn, setting string) (bool, error) {
	if !isValidAzureName(reResourceGroupName, rgn) {
		return false, fmt.Errorf("The setting %s must match the regular expression %q, and not end with a '-' or '.'.", setting, validResourceGroupNameRe)
	}
	return true, nil
}

func isValidAzureName(re *regexp.Regexp, rgn string) bool {
	return re.Match([]byte(rgn)) &&
		!strings.HasSuffix(rgn, ".") &&
		!strings.HasSuffix(rgn, "-")
}

// The supplied password must be between 8-123 characters long and must satisfy at least 3 of password complexity requirements from the following:
// 1) Contains an uppercase character
// 2) Contains a lowercase character
// 3) Contains a numeric digit
// 4) Contains a special character
// 5) Control characters are not allowed (a very specific case - not included in this validation)
func isValidPassword(password string) bool {
	if !(len(password) >= 8 && len(password) <= 123) {
		return false
	}

	requirements := 0
	if strings.ContainsAny(password, random.PossibleNumbers) {
		requirements++
	}
	if strings.ContainsAny(password, random.PossibleLowerCase) {
		requirements++
	}
	if strings.ContainsAny(password, random.PossibleUpperCase) {
		requirements++
	}
	if strings.ContainsAny(password, random.PossibleSpecialCharacter) {
		requirements++
	}

	return requirements >= 3
}

func (c *Config) validateLocationZoneResiliency(say func(s string)) {
	// Docs on regions that support Availibility Zones:
	//   https://docs.microsoft.com/en-us/azure/availability-zones/az-overview#regions-that-support-availability-zones
	// Query technical names for locations:
	//   az account list-locations --query '[].name' -o tsv

	var zones = make(map[string]struct{})
	zones["westeurope"] = struct{}{}
	zones["centralus"] = struct{}{}
	zones["eastus2"] = struct{}{}
	zones["francecentral"] = struct{}{}
	zones["northeurope"] = struct{}{}
	zones["southeastasia"] = struct{}{}
	zones["westus2"] = struct{}{}

	if _, ok := zones[c.Location]; !ok {
		say(fmt.Sprintf("WARNING: Zone resiliency may not be supported in %s, checkout the docs at https://docs.microsoft.com/en-us/azure/availability-zones/", c.Location))
	}
}

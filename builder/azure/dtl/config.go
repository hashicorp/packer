//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type Config,SharedImageGallery,SharedImageGalleryDestination,DtlArtifact,ArtifactParameter

package dtl

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/masterzen/winrm"

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
)

const (
	// https://docs.microsoft.com/en-us/azure/architecture/best-practices/naming-conventions#naming-rules-and-restrictions
	// Regular expressions in Go are not expressive enough, such that the regular expression returned by Azure
	// can be used (no backtracking).
	//
	//  -> ^[^_\W][\w-._]{0,79}(?<![-.])$
	//
	// This is not an exhaustive match, but it should be extremely close.
	validResourceGroupNameRe = `^[^_\W][\w-._\(\)]{0,89}$`
	validManagedDiskName     = `^[^_\W][\w-._)]{0,79}$`
)

var (
	reCaptureContainerName = regexp.MustCompile(`^[a-z0-9][a-z0-9\-]{2,62}`)
	reCaptureNamePrefix    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_\-\.]{0,23}$`)
	reManagedDiskName      = regexp.MustCompile(validManagedDiskName)
	reResourceGroupName    = regexp.MustCompile(validResourceGroupNameRe)
)

type SharedImageGallery struct {
	Subscription  string `mapstructure:"subscription"`
	ResourceGroup string `mapstructure:"resource_group"`
	GalleryName   string `mapstructure:"gallery_name"`
	ImageName     string `mapstructure:"image_name"`
	ImageVersion  string `mapstructure:"image_version"`
}

type SharedImageGalleryDestination struct {
	SigDestinationResourceGroup      string   `mapstructure:"resource_group"`
	SigDestinationGalleryName        string   `mapstructure:"gallery_name"`
	SigDestinationImageName          string   `mapstructure:"image_name"`
	SigDestinationImageVersion       string   `mapstructure:"image_version"`
	SigDestinationReplicationRegions []string `mapstructure:"replication_regions"`
}

type DtlArtifact struct {
	ArtifactName   string              `mapstructure:"artifact_name"`
	RepositoryName string              `mapstructure:"repository_name"`
	ArtifactId     string              `mapstructure:"artifact_id"`
	Parameters     []ArtifactParameter `mapstructure:"parameters"`
}

type ArtifactParameter struct {
	Name  string `mapstructure:"name"`
	Value string `mapstructure:"value"`
	Type  string `mapstructure:"type"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Authentication via OAUTH
	ClientConfig client.Config `mapstructure:",squash"`

	// Capture
	CaptureNamePrefix    string `mapstructure:"capture_name_prefix"`
	CaptureContainerName string `mapstructure:"capture_container_name"`

	// Use a [Shared Gallery
	// image](https://azure.microsoft.com/en-us/blog/announcing-the-public-preview-of-shared-image-gallery/)
	// as the source for this build. *VHD targets are incompatible with this
	// build type* - the target must be a *Managed Image*.
	//
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
	SharedGallery SharedImageGallery `mapstructure:"shared_image_gallery"`

	// The name of the Shared Image Gallery under which the managed image will be published as Shared Gallery Image version.
	//
	// Following is an example.
	//
	// ```json
	// "shared_image_gallery_destination": {
	//     "resource_group": "ResourceGroup",
	//     "gallery_name": "GalleryName",
	//     "image_name": "ImageName",
	//     "image_version": "1.0.0",
	//     "replication_regions": ["regionA", "regionB", "regionC"]
	// }
	// "managed_image_name": "TargetImageName",
	// "managed_image_resource_group_name": "TargetResourceGroup"
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

	// PublisherName for your base image. See
	// [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
	// for details.
	//
	// CLI example `az vm image list-publishers --location westus`
	ImagePublisher string `mapstructure:"image_publisher"`
	// Offer for your base image. See
	// [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
	// for details.
	//
	// CLI example
	// `az vm image list-offers --location westus --publisher Canonical`
	ImageOffer string `mapstructure:"image_offer"`
	// SKU for your base image. See
	// [documentation](https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/)
	// for details.
	//
	// CLI example
	// `az vm image list-skus --location westus --publisher Canonical --offer UbuntuServer`
	ImageSku string `mapstructure:"image_sku"`
	// Specify a specific version of an OS to boot from.
	// Defaults to `latest`. There may be a difference in versions available
	// across regions due to image synchronization latency. To ensure a consistent
	// version across regions set this value to one that is available in all
	// regions where you are deploying.
	//
	// CLI example
	// `az vm image list --location westus --publisher Canonical --offer UbuntuServer --sku 16.04.0-LTS --all`
	ImageVersion string `mapstructure:"image_version"`
	// Specify a custom VHD to use. If this value is set, do
	// not set image_publisher, image_offer, image_sku, or image_version.
	ImageUrl string `mapstructure:"image_url"`

	// Specify the source managed image's resource group used to use. If this
	// value is set, do not set image\_publisher, image\_offer, image\_sku, or
	// image\_version. If this value is set, the value
	// `custom_managed_image_name` must also be set. See
	// [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
	// to learn more about managed images.
	CustomManagedImageResourceGroupName string `mapstructure:"custom_managed_image_resource_group_name"`
	// Specify the source managed image's name to use. If this value is set, do
	// not set image\_publisher, image\_offer, image\_sku, or image\_version.
	// If this value is set, the value
	// `custom_managed_image_resource_group_name` must also be set. See
	// [documentation](https://docs.microsoft.com/en-us/azure/storage/storage-managed-disks-overview#images)
	// to learn more about managed images.
	CustomManagedImageName string `mapstructure:"custom_managed_image_name"`
	customManagedImageID   string

	Location string `mapstructure:"location"`
	// Size of the VM used for building. This can be changed when you deploy a
	// VM from your VHD. See
	// [pricing](https://azure.microsoft.com/en-us/pricing/details/virtual-machines/)
	// information. Defaults to `Standard_A1`.
	//
	// CLI example `az vm list-sizes --location westus`
	VMSize string `mapstructure:"vm_size"`
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

	// the user can define up to 15
	// tags. Tag names cannot exceed 512 characters, and tag values cannot exceed
	// 256 characters. Tags are applied to every resource deployed by a Packer
	// build, i.e. Resource Group, VM, NIC, VNET, Public IP, KeyVault, etc.
	AzureTags map[string]*string `mapstructure:"azure_tags" required:"false"`

	// Used for creating images from Marketplace images. Please refer to
	// [Deploy an image with Marketplace
	// terms](https://aka.ms/azuremarketplaceapideployment) for more details.
	// Not all Marketplace images support programmatic deployment, and support
	// is controlled by the image publisher.
	// Plan_id is a string with unique identifier for the plan associated with images.
	// Ex plan_id="1-12ab"
	PlanID string `mapstructure:"plan_id" required:"false"`

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

	// For Managed build the final artifacts are included in the managed image.
	// The additional disk will have the same storage account type as the OS
	// disk, as specified with the `managed_image_storage_account_type`
	// setting.
	AdditionalDiskSize []int32 `mapstructure:"disk_additional_size" required:"false"`
	// Specify the disk caching type. Valid values
	// are None, ReadOnly, and ReadWrite. The default value is ReadWrite.
	DiskCachingType string `mapstructure:"disk_caching_type" required:"false"`
	diskCachingType compute.CachingTypes

	// DTL values
	StorageType           string `mapstructure:"storage_type"`
	LabVirtualNetworkName string `mapstructure:"lab_virtual_network_name"`
	LabName               string `mapstructure:"lab_name"`
	LabSubnetName         string `mapstructure:"lab_subnet_name"`
	LabResourceGroupName  string `mapstructure:"lab_resource_group_name"`

	DtlArtifacts []DtlArtifact `mapstructure:"dtl_artifacts"`
	VMName       string        `mapstructure:"vm_name"`

	// Runtime Values
	UserName                string
	Password                string
	tmpAdminPassword        string
	tmpCertificatePassword  string
	tmpResourceGroupName    string
	tmpComputeName          string
	tmpNicName              string
	tmpPublicIPAddressName  string
	tmpDeploymentName       string
	tmpKeyVaultName         string
	tmpOSDiskName           string
	tmpSubnetName           string
	tmpVirtualNetworkName   string
	VMCreationResourceGroup string
	tmpFQDN                 string

	// Authentication with the VM via SSH
	sshAuthorizedKey string

	// Authentication with the VM via WinRM
	winrmCertificate string
	winrmPassword    string

	Comm communicator.Config `mapstructure:",squash"`
	ctx  interpolate.Context
}

type keyVaultCertificate struct {
	Data     string `json:"data"`
	DataType string `json:"dataType"`
	Password string `json:"password,omitempty"`
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

func (c *Config) createCertificate() (string, string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		err = fmt.Errorf("Failed to Generate Private Key: %s", err)
		return "", "", err
	}

	host := fmt.Sprintf("%s.centralus.cloudapp.azure.com", c.tmpComputeName)
	notBefore := time.Now()
	notAfter := notBefore.Add(24 * time.Hour)

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		err = fmt.Errorf("Failed to Generate Serial Number: %v", err)
		return "", "", err
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
		return "", "", err
	}

	pfxBytes, err := pkcs12.Encode(derBytes, privateKey, c.tmpCertificatePassword)
	if err != nil {
		err = fmt.Errorf("Failed to encode certificate as PFX: %s", err)
		return "", "", err
	}

	keyVaultDescription := keyVaultCertificate{
		Data:     base64.StdEncoding.EncodeToString(pfxBytes),
		DataType: "pfx",
		Password: c.tmpCertificatePassword,
	}

	bytes, err := json.Marshal(keyVaultDescription)
	if err != nil {
		err = fmt.Errorf("Failed to marshal key vault description: %s", err)
		return "", "", err
	}

	certifcatePassowrd := base64.StdEncoding.EncodeToString([]byte(c.tmpCertificatePassword))
	return base64.StdEncoding.EncodeToString(bytes), certifcatePassowrd, nil
}

func newConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config
	c.ctx.Funcs = TemplateFuncs
	err := config.Decode(&c, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)

	if err != nil {
		return nil, nil, err
	}

	provideDefaultValues(&c)
	setRuntimeValues(&c)
	setUserNamePassword(&c)
	err = c.ClientConfig.SetDefaultValues()
	if err != nil {
		return nil, nil, err
	}

	// NOTE: if the user did not specify a communicator, then default to both
	// SSH and WinRM.  This is for backwards compatibility because the code did
	// not specifically force the user to set a communicator.
	if c.Comm.Type == "" || strings.EqualFold(c.Comm.Type, "ssh") {
		err = setSshValues(&c)
		if err != nil {
			return nil, nil, err
		}
	}

	//For DTL, communication is done by installing Mandatory public artifact "windows-winrm"
	if c.Comm.Type == "" || strings.EqualFold(c.Comm.Type, "winrm") {
		err = setWinRMCertificate(&c)
		if err != nil {
			return nil, nil, err
		}
	}

	var errs *packersdk.MultiError
	errs = packersdk.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)

	c.ClientConfig.Validate(errs)

	assertRequiredParametersSet(&c, errs)
	assertTagProperties(&c, errs)
	if errs != nil && len(errs.Errors) > 0 {
		return nil, nil, errs
	}

	return &c, nil, nil
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

	cert, password, err := c.createCertificate()

	c.winrmCertificate = cert
	c.winrmPassword = password

	return err
}

func setRuntimeValues(c *Config) {
	var tempName = NewTempName(c)

	c.tmpAdminPassword = tempName.AdminPassword
	packersdk.LogSecretFilter.Set(c.tmpAdminPassword)

	c.tmpCertificatePassword = tempName.CertificatePassword
	c.tmpComputeName = tempName.ComputeName

	c.tmpDeploymentName = tempName.DeploymentName
	if c.LabResourceGroupName == "" {
		c.tmpResourceGroupName = tempName.ResourceGroupName
	} else {
		c.tmpResourceGroupName = c.LabResourceGroupName
	}
	c.tmpNicName = tempName.NicName
	c.tmpPublicIPAddressName = tempName.PublicIPAddressName
	c.tmpOSDiskName = tempName.OSDiskName
	c.tmpSubnetName = tempName.SubnetName
	c.tmpVirtualNetworkName = tempName.VirtualNetworkName
	c.tmpKeyVaultName = tempName.KeyVaultName
}

func setUserNamePassword(c *Config) {
	if c.Comm.SSHUsername == "" {
		c.Comm.SSHUsername = DefaultUserName
	}

	c.UserName = c.Comm.SSHUsername

	if c.Comm.SSHPassword != "" {
		c.Password = c.Comm.SSHPassword
	} else {
		c.Password = c.tmpAdminPassword
	}
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
}

func assertTagProperties(c *Config, errs *packersdk.MultiError) {
	if len(c.AzureTags) > 15 {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("a max of 15 tags are supported, but %d were provided", len(c.AzureTags)))
	}

	for k, v := range c.AzureTags {
		if len(k) > 512 {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the tag name %q exceeds (%d) the 512 character limit", k, len(k)))
		}
		if len(*v) > 256 {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("the tag name %q exceeds (%d) the 256 character limit", *v, len(*v)))
		}
	}
}

func assertRequiredParametersSet(c *Config, errs *packersdk.MultiError) {
	c.ClientConfig.Validate(errs)

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

	if c.LabResourceGroupName == "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("The settings lab_resource_group_name needs to be defined."))
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
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An custom_managed_image_resource_group_name must be specified"))
		}
		if c.CustomManagedImageName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("A custom_managed_image_name must be specified"))
		}
		if c.ManagedImageResourceGroupName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An managed_image_resource_group_name must be specified"))
		}
		if c.ManagedImageName == "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An managed_image_name must be specified"))
		}
	} else {
		if c.ImagePublisher != "" || c.ImageOffer != "" || c.ImageSku != "" || c.ImageVersion != "" {
			errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("An image_url must not be specified if image_publisher, image_offer, image_sku, or image_version is specified"))
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

	if c.LabVirtualNetworkName == "" && c.LabResourceGroupName != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("If lab_resource_group_name is specified, so must lab_virtual_network_name"))
	}
	if c.LabVirtualNetworkName == "" && c.LabSubnetName != "" {
		errs = packersdk.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_subnet_name is specified, so must lab_virtual_network_name"))
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
	// Errs check to make the linter happy.
	if errs != nil {
		return
	}
}

func assertManagedImageName(name, setting string) (bool, error) {
	if !isValidAzureName(reManagedDiskName, name) {
		return false, fmt.Errorf("The setting %s must match the regular expression %q, and not end with a '-' or '.'.", setting, validManagedDiskName)
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

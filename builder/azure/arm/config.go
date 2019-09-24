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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2018-04-01/compute"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/masterzen/winrm"

	"github.com/hashicorp/packer/builder/azure/common/constants"
	"github.com/hashicorp/packer/builder/azure/pkcs12"
	"github.com/hashicorp/packer/common"
	commonhelper "github.com/hashicorp/packer/helper/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"

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
	validResourceGroupNameRe = "^[^_\\W][\\w-._\\(\\)]{0,89}$"
	validManagedDiskName     = "^[^_\\W][\\w-._)]{0,79}$"
)

var (
	reCaptureContainerName = regexp.MustCompile("^[a-z0-9][a-z0-9\\-]{2,62}$")
	reCaptureNamePrefix    = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9_\\-\\.]{0,23}$")
	reManagedDiskName      = regexp.MustCompile(validManagedDiskName)
	reResourceGroupName    = regexp.MustCompile(validResourceGroupNameRe)
	reSnapshotName         = regexp.MustCompile("^[A-Za-z0-9_]{1,79}$")
	reSnapshotPrefix       = regexp.MustCompile("^[A-Za-z0-9_]{1,59}$")
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
	ImageVersion  string `mapstructure:"image_version"`
}

type SharedImageGalleryDestination struct {
	SigDestinationResourceGroup      string   `mapstructure:"resource_group"`
	SigDestinationGalleryName        string   `mapstructure:"gallery_name"`
	SigDestinationImageName          string   `mapstructure:"image_name"`
	SigDestinationImageVersion       string   `mapstructure:"image_version"`
	SigDestinationReplicationRegions []string `mapstructure:"replication_regions"`
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Authentication via OAUTH
	ClientConfig `mapstructure:",squash"`

	// Capture
	CaptureNamePrefix    string `mapstructure:"capture_name_prefix"`
	CaptureContainerName string `mapstructure:"capture_container_name"`

	// Shared Gallery
	SharedGallery SharedImageGallery `mapstructure:"shared_image_gallery"`

	// Shared Gallery Destination
	SharedGalleryDestination SharedImageGalleryDestination `mapstructure:"shared_image_gallery_destination"`
	SharedGalleryTimeout     time.Duration                 `mapstructure:"shared_image_gallery_timeout"`

	// Compute
	ImagePublisher string `mapstructure:"image_publisher"`
	ImageOffer     string `mapstructure:"image_offer"`
	ImageSku       string `mapstructure:"image_sku"`
	ImageVersion   string `mapstructure:"image_version"`
	ImageUrl       string `mapstructure:"image_url"`

	CustomManagedImageResourceGroupName string `mapstructure:"custom_managed_image_resource_group_name"`
	CustomManagedImageName              string `mapstructure:"custom_managed_image_name"`
	customManagedImageID                string

	Location string `mapstructure:"location"`
	VMSize   string `mapstructure:"vm_size"`

	ManagedImageResourceGroupName      string `mapstructure:"managed_image_resource_group_name"`
	ManagedImageName                   string `mapstructure:"managed_image_name"`
	ManagedImageStorageAccountType     string `mapstructure:"managed_image_storage_account_type"`
	managedImageStorageAccountType     compute.StorageAccountTypes
	ManagedImageOSDiskSnapshotName     string `mapstructure:"managed_image_os_disk_snapshot_name"`
	ManagedImageDataDiskSnapshotPrefix string `mapstructure:"managed_image_data_disk_snapshot_prefix"`
	manageImageLocation                string
	ManagedImageZoneResilient          bool `mapstructure:"managed_image_zone_resilient"`

	// Deployment
	AzureTags                         map[string]*string `mapstructure:"azure_tags"`
	ResourceGroupName                 string             `mapstructure:"resource_group_name"`
	StorageAccount                    string             `mapstructure:"storage_account"`
	TempComputeName                   string             `mapstructure:"temp_compute_name"`
	TempResourceGroupName             string             `mapstructure:"temp_resource_group_name"`
	BuildResourceGroupName            string             `mapstructure:"build_resource_group_name"`
	storageAccountBlobEndpoint        string
	PrivateVirtualNetworkWithPublicIp bool     `mapstructure:"private_virtual_network_with_public_ip"`
	VirtualNetworkName                string   `mapstructure:"virtual_network_name"`
	VirtualNetworkSubnetName          string   `mapstructure:"virtual_network_subnet_name"`
	VirtualNetworkResourceGroupName   string   `mapstructure:"virtual_network_resource_group_name"`
	AllowedInboundIpAddresses         []string `mapstructure:"allowed_inbound_ip_addresses"`
	CustomDataFile                    string   `mapstructure:"custom_data_file"`
	customData                        string
	PlanInfo                          PlanInformation `mapstructure:"plan_info"`

	// OS
	OSType       string `mapstructure:"os_type"`
	OSDiskSizeGB int32  `mapstructure:"os_disk_size_gb"`

	// Additional Disks
	AdditionalDiskSize []int32 `mapstructure:"disk_additional_size"`
	DiskCachingType    string  `mapstructure:"disk_caching_type"`
	diskCachingType    compute.CachingTypes

	// Runtime Values
	UserName               string
	Password               string
	tmpAdminPassword       string
	tmpCertificatePassword string
	tmpResourceGroupName   string
	tmpComputeName         string
	tmpNicName             string
	tmpPublicIPAddressName string
	tmpDeploymentName      string
	tmpKeyVaultName        string
	tmpOSDiskName          string
	tmpSubnetName          string
	tmpVirtualNetworkName  string
	tmpWinRMCertificateUrl string

	// Authentication with the VM via SSH
	sshAuthorizedKey string

	// Authentication with the VM via WinRM
	winrmCertificate string

	Comm communicator.Config `mapstructure:",squash"`
	ctx  interpolate.Context

	//Cleanup
	AsyncResourceGroupDelete bool `mapstructure:"async_resourcegroup_delete"`
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
	return fmt.Sprintf("/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Compute/virtualMachines/%s", c.SubscriptionID, resourceGroupName, c.tmpComputeName)
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
		Tags:     c.AzureTags,
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

func newConfig(raws ...interface{}) (*Config, []string, error) {
	var c Config
	c.ctx.Funcs = TemplateFuncs
	err := config.Decode(&c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...)

	if err != nil {
		return nil, nil, err
	}

	provideDefaultValues(&c)
	setRuntimeValues(&c)
	setUserNamePassword(&c)
	err = c.ClientConfig.setCloudEnvironment()
	if err != nil {
		return nil, nil, err
	}

	err = setCustomData(&c)
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

	if c.Comm.Type == "" || strings.EqualFold(c.Comm.Type, "winrm") {
		err = setWinRMCertificate(&c)
		if err != nil {
			return nil, nil, err
		}
	}

	var errs *packer.MultiError
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(&c.ctx)...)

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

	cert, err := c.createCertificate()
	c.winrmCertificate = cert

	return err
}

func setRuntimeValues(c *Config) {
	var tempName = NewTempName()

	c.tmpAdminPassword = tempName.AdminPassword
	// store so that we can access this later during provisioning
	commonhelper.SetSharedState("winrm_password", c.tmpAdminPassword, c.PackerConfig.PackerBuildName)
	packer.LogSecretFilter.Set(c.tmpAdminPassword)

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

	c.provideDefaultValues()
}

func assertTagProperties(c *Config, errs *packer.MultiError) {
	if len(c.AzureTags) > 15 {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("a max of 15 tags are supported, but %d were provided", len(c.AzureTags)))
	}

	for k, v := range c.AzureTags {
		if len(k) > 512 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("the tag name %q exceeds (%d) the 512 character limit", k, len(k)))
		}
		if len(*v) > 256 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("the tag name %q exceeds (%d) the 256 character limit", *v, len(*v)))
		}
	}
}

func assertRequiredParametersSet(c *Config, errs *packer.MultiError) {
	c.ClientConfig.assertRequiredParametersSet(errs)

	/////////////////////////////////////////////
	// Capture
	if c.CaptureContainerName == "" && c.ManagedImageName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name or managed_image_name must be specified"))
	}

	if c.CaptureNamePrefix == "" && c.ManagedImageResourceGroupName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix or managed_image_resource_group_name must be specified"))
	}

	if (c.CaptureNamePrefix != "" || c.CaptureContainerName != "") && (c.ManagedImageResourceGroupName != "" || c.ManagedImageName != "") {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Either a VHD or a managed image can be built, but not both. Please specify either capture_container_name and capture_name_prefix or managed_image_resource_group_name and managed_image_name."))
	}

	if c.CaptureContainerName != "" {
		if !reCaptureContainerName.MatchString(c.CaptureContainerName) {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name must satisfy the regular expression %q.", reCaptureContainerName.String()))
		}

		if strings.HasSuffix(c.CaptureContainerName, "-") {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name must not end with a hyphen, e.g. '-'."))
		}

		if strings.Contains(c.CaptureContainerName, "--") {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name must not contain consecutive hyphens, e.g. '--'."))
		}

		if c.CaptureNamePrefix == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix must be specified"))
		}

		if !reCaptureNamePrefix.MatchString(c.CaptureNamePrefix) {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix must satisfy the regular expression %q.", reCaptureNamePrefix.String()))
		}

		if strings.HasSuffix(c.CaptureNamePrefix, "-") || strings.HasSuffix(c.CaptureNamePrefix, ".") {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_name_prefix must not end with a hyphen or period."))
		}
	}

	if c.TempResourceGroupName != "" && c.BuildResourceGroupName != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("The settings temp_resource_group_name and build_resource_group_name cannot both be defined.  Please define one or neither."))
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
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Specify either a VHD (image_url), Image Reference (image_publisher, image_offer, image_sku), a Managed Disk (custom_managed_disk_image_name, custom_managed_disk_resource_group_name), or a Shared Gallery Image (shared_image_gallery)"))
	}

	if isImageUrl && c.ManagedImageResourceGroupName != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A managed image must be created from a managed image, it cannot be created from a VHD."))
	}

	if c.SharedGallery.GalleryName != "" {
		if c.SharedGallery.Subscription == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A shared_image_gallery.subscription must be specified"))
		}
		if c.SharedGallery.ResourceGroup == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A shared_image_gallery.resource_group must be specified"))
		}
		if c.SharedGallery.ImageName == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A shared_image_gallery.image_name must be specified"))
		}
		if c.CaptureContainerName != "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("VHD Target [capture_container_name] is not supported when using Shared Image Gallery as source. Use managed_image_resource_group_name instead."))
		}
		if c.CaptureNamePrefix != "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("VHD Target [capture_name_prefix] is not supported when using Shared Image Gallery as source. Use managed_image_name instead."))
		}
	} else if c.ImageUrl == "" && c.CustomManagedImageName == "" {
		if c.ImagePublisher == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_publisher must be specified"))
		}
		if c.ImageOffer == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_offer must be specified"))
		}
		if c.ImageSku == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_sku must be specified"))
		}
	} else if c.ImageUrl == "" && c.ImagePublisher == "" {
		if c.CustomManagedImageResourceGroupName == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A custom_managed_image_resource_group_name must be specified"))
		}
		if c.CustomManagedImageName == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A custom_managed_image_name must be specified"))
		}
		if c.ManagedImageResourceGroupName == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A managed_image_resource_group_name must be specified"))
		}
		if c.ManagedImageName == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A managed_image_name must be specified"))
		}
	} else {
		if c.ImagePublisher != "" || c.ImageOffer != "" || c.ImageSku != "" || c.ImageVersion != "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_url must not be specified if image_publisher, image_offer, image_sku, or image_version is specified"))
		}
	}

	/////////////////////////////////////////////
	// Deployment
	xor := func(a, b bool) bool {
		return (a || b) && !(a && b)
	}

	if !xor((c.StorageAccount != "" || c.ResourceGroupName != ""), (c.ManagedImageName != "" || c.ManagedImageResourceGroupName != "")) {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Specify either a VHD (storage_account and resource_group_name) or Managed Image (managed_image_resource_group_name and managed_image_name) output"))
	}

	if !xor(c.Location != "", c.BuildResourceGroupName != "") {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("Specify either a location to create the resource group in or an existing build_resource_group_name, but not both."))
	}

	if c.ManagedImageName == "" && c.ManagedImageResourceGroupName == "" {
		if c.StorageAccount == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A storage_account must be specified"))
		}
		if c.ResourceGroupName == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A resource_group_name must be specified"))
		}
	}

	if c.TempResourceGroupName != "" {
		if ok, err := assertResourceGroupName(c.TempResourceGroupName, "temp_resource_group_name"); !ok {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.BuildResourceGroupName != "" {
		if ok, err := assertResourceGroupName(c.BuildResourceGroupName, "build_resource_group_name"); !ok {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageResourceGroupName != "" {
		if ok, err := assertResourceGroupName(c.ManagedImageResourceGroupName, "managed_image_resource_group_name"); !ok {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageName != "" {
		if ok, err := assertManagedImageName(c.ManagedImageName, "managed_image_name"); !ok {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageName != "" && c.ManagedImageResourceGroupName != "" && c.SharedGalleryDestination.SigDestinationGalleryName != "" {
		if c.SharedGalleryDestination.SigDestinationResourceGroup == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A resource_group must be specified for shared_image_gallery_destination"))
		}
		if c.SharedGalleryDestination.SigDestinationImageName == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_name must be specified for shared_image_gallery_destination"))
		}
		if c.SharedGalleryDestination.SigDestinationImageVersion == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_version must be specified for shared_image_gallery_destination"))
		}
		if len(c.SharedGalleryDestination.SigDestinationReplicationRegions) == 0 {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A list of replication_regions must be specified for shared_image_gallery_destination"))
		}
	}
	if c.SharedGalleryTimeout == 0 {
		// default to a one-hour timeout. In the sdk, the default is 15 m.
		c.SharedGalleryTimeout = 60 * time.Minute
	}

	if c.ManagedImageOSDiskSnapshotName != "" {
		if ok, err := assertManagedImageOSDiskSnapshotName(c.ManagedImageOSDiskSnapshotName, "managed_image_os_disk_snapshot_name"); !ok {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.ManagedImageDataDiskSnapshotPrefix != "" {
		if ok, err := assertManagedImageDataDiskSnapshotName(c.ManagedImageDataDiskSnapshotPrefix, "managed_image_data_disk_snapshot_prefix"); !ok {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	if c.VirtualNetworkName == "" && c.VirtualNetworkResourceGroupName != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_resource_group_name is specified, so must virtual_network_name"))
	}
	if c.VirtualNetworkName == "" && c.VirtualNetworkSubnetName != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_subnet_name is specified, so must virtual_network_name"))
	}

	if c.AllowedInboundIpAddresses != nil && len(c.AllowedInboundIpAddresses) >= 1 {
		if ok, err := assertAllowedInboundIpAddresses(c.AllowedInboundIpAddresses, "allowed_inbound_ip_addresses"); !ok {
			errs = packer.MultiErrorAppend(errs, err)
		}
	}

	/////////////////////////////////////////////
	// Plan Info
	if c.PlanInfo.PlanName != "" || c.PlanInfo.PlanProduct != "" || c.PlanInfo.PlanPublisher != "" || c.PlanInfo.PlanPromotionCode != "" {
		if c.PlanInfo.PlanName == "" || c.PlanInfo.PlanProduct == "" || c.PlanInfo.PlanPublisher == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("if either plan_name, plan_product, plan_publisher, or plan_promotion_code are defined then plan_name, plan_product, and plan_publisher must be defined"))
		} else {
			if c.AzureTags == nil {
				c.AzureTags = make(map[string]*string)
			}

			c.AzureTags["PlanInfo"] = &c.PlanInfo.PlanName
			c.AzureTags["PlanProduct"] = &c.PlanInfo.PlanProduct
			c.AzureTags["PlanPublisher"] = &c.PlanInfo.PlanPublisher
			c.AzureTags["PlanPromotionCode"] = &c.PlanInfo.PlanPromotionCode
		}
	}

	/////////////////////////////////////////////
	// OS
	if strings.EqualFold(c.OSType, constants.Target_Linux) {
		c.OSType = constants.Target_Linux
	} else if strings.EqualFold(c.OSType, constants.Target_Windows) {
		c.OSType = constants.Target_Windows
	} else if c.OSType == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("An os_type must be specified"))
	} else {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("The os_type %q is invalid", c.OSType))
	}

	switch c.ManagedImageStorageAccountType {
	case "", string(compute.StorageAccountTypesStandardLRS):
		c.managedImageStorageAccountType = compute.StorageAccountTypesStandardLRS
	case string(compute.StorageAccountTypesPremiumLRS):
		c.managedImageStorageAccountType = compute.StorageAccountTypesPremiumLRS
	default:
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("The managed_image_storage_account_type %q is invalid", c.ManagedImageStorageAccountType))
	}

	switch c.DiskCachingType {
	case string(compute.CachingTypesNone):
		c.diskCachingType = compute.CachingTypesNone
	case string(compute.CachingTypesReadOnly):
		c.diskCachingType = compute.CachingTypesReadOnly
	case "", string(compute.CachingTypesReadWrite):
		c.diskCachingType = compute.CachingTypesReadWrite
	default:
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("The disk_caching_type %q is invalid", c.DiskCachingType))
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

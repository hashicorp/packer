// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

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
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/go-ntlmssp"

	"github.com/mitchellh/packer/builder/azure/common/constants"
	"github.com/mitchellh/packer/builder/azure/pkcs12"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"

	"golang.org/x/crypto/ssh"
)

const (
	DefaultCloudEnvironmentName = "Public"
	DefaultImageVersion         = "latest"
	DefaultUserName             = "packer"
	DefaultVMSize               = "Standard_A1"
)

var (
	reCaptureContainerName = regexp.MustCompile("^[a-z0-9][a-z0-9\\-]{2,62}$")
	reCaptureNamePrefix    = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9_\\-\\.]{0,23}$")
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Authentication via OAUTH
	ClientID       string `mapstructure:"client_id"`
	ClientSecret   string `mapstructure:"client_secret"`
	ObjectID       string `mapstructure:"object_id"`
	TenantID       string `mapstructure:"tenant_id"`
	SubscriptionID string `mapstructure:"subscription_id"`

	// Capture
	CaptureNamePrefix    string `mapstructure:"capture_name_prefix"`
	CaptureContainerName string `mapstructure:"capture_container_name"`

	// Compute
	ImagePublisher string `mapstructure:"image_publisher"`
	ImageOffer     string `mapstructure:"image_offer"`
	ImageSku       string `mapstructure:"image_sku"`
	ImageVersion   string `mapstructure:"image_version"`
	ImageUrl       string `mapstructure:"image_url"`
	Location       string `mapstructure:"location"`
	VMSize         string `mapstructure:"vm_size"`

	// Deployment
	AzureTags                       map[string]*string `mapstructure:"azure_tags"`
	ResourceGroupName               string             `mapstructure:"resource_group_name"`
	StorageAccount                  string             `mapstructure:"storage_account"`
	storageAccountBlobEndpoint      string
	CloudEnvironmentName            string `mapstructure:"cloud_environment_name"`
	cloudEnvironment                *azure.Environment
	VirtualNetworkName              string `mapstructure:"virtual_network_name"`
	VirtualNetworkSubnetName        string `mapstructure:"virtual_network_subnet_name"`
	VirtualNetworkResourceGroupName string `mapstructure:"virtual_network_resource_group_name"`

	// OS
	OSType string `mapstructure:"os_type"`

	// Runtime Values
	UserName               string
	Password               string
	tmpAdminPassword       string
	tmpCertificatePassword string
	tmpResourceGroupName   string
	tmpComputeName         string
	tmpDeploymentName      string
	tmpKeyVaultName        string
	tmpOSDiskName          string
	tmpWinRMCertificateUrl string

	useDeviceLogin bool

	// Authentication with the VM via SSH
	sshAuthorizedKey string
	sshPrivateKey    string

	// Authentication with the VM via WinRM
	winrmCertificate string

	Comm communicator.Config `mapstructure:",squash"`
	ctx  *interpolate.Context
}

type keyVaultCertificate struct {
	Data     string `json:"data"`
	DataType string `json:"dataType"`
	Password string `json:"password,omitempty"`
}

func (c *Config) toVirtualMachineCaptureParameters() *compute.VirtualMachineCaptureParameters {
	return &compute.VirtualMachineCaptureParameters{
		DestinationContainerName: &c.CaptureContainerName,
		VhdPrefix:                &c.CaptureNamePrefix,
		OverwriteVhds:            to.BoolPtr(false),
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

	err := config.Decode(&c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: c.ctx,
	}, raws...)

	if err != nil {
		return nil, nil, err
	}

	provideDefaultValues(&c)
	setRuntimeValues(&c)
	setUserNamePassword(&c)
	err = setCloudEnvironment(&c)
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
	errs = packer.MultiErrorAppend(errs, c.Comm.Prepare(c.ctx)...)

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

	if c.Comm.SSHPrivateKey != "" {
		privateKeyBytes, err := ioutil.ReadFile(c.Comm.SSHPrivateKey)
		if err != nil {
			panic(err)
		}
		signer, err := ssh.ParsePrivateKey(privateKeyBytes)
		if err != nil {
			panic(err)
		}

		publicKey := signer.PublicKey()
		c.sshAuthorizedKey = fmt.Sprintf("%s %s packer Azure Deployment%s",
			publicKey.Type(),
			base64.StdEncoding.EncodeToString(publicKey.Marshal()),
			time.Now().Format(time.RFC3339))
		c.sshPrivateKey = string(privateKeyBytes)

	} else {
		sshKeyPair, err := NewOpenSshKeyPair()
		if err != nil {
			return err
		}

		c.sshAuthorizedKey = sshKeyPair.AuthorizedKey()
		c.sshPrivateKey = sshKeyPair.PrivateKey()
	}

	return nil
}

func setWinRMCertificate(c *Config) error {
	c.Comm.WinRMTransportDecorator = func(t *http.Transport) http.RoundTripper {
		return &ntlmssp.Negotiator{RoundTripper: t}
	}

	cert, err := c.createCertificate()
	c.winrmCertificate = cert

	return err
}

func setRuntimeValues(c *Config) {
	var tempName = NewTempName()

	c.tmpAdminPassword = tempName.AdminPassword
	c.tmpCertificatePassword = tempName.CertificatePassword
	c.tmpComputeName = tempName.ComputeName
	c.tmpDeploymentName = tempName.DeploymentName
	c.tmpResourceGroupName = tempName.ResourceGroupName
	c.tmpOSDiskName = tempName.OSDiskName
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

func setCloudEnvironment(c *Config) error {
	lookup := map[string]string{
		"CHINA":           "AzureChinaCloud",
		"CHINACLOUD":      "AzureChinaCloud",
		"AZURECHINACLOUD": "AzureChinaCloud",

		"GERMAN":           "AzureGermanCloud",
		"GERMANCLOUD":      "AzureGermanCloud",
		"AZUREGERMANCLOUD": "AzureGermanCloud",

		"GERMANY":           "AzureGermanCloud",
		"GERMANYCLOUD":      "AzureGermanCloud",
		"AZUREGERMANYCLOUD": "AzureGermanCloud",

		"PUBLIC":           "AzurePublicCloud",
		"PUBLICCLOUD":      "AzurePublicCloud",
		"AZUREPUBLICCLOUD": "AzurePublicCloud",

		"USGOVERNMENT":           "AzureUSGovernmentCloud",
		"USGOVERNMENTCLOUD":      "AzureUSGovernmentCloud",
		"AZUREUSGOVERNMENTCLOUD": "AzureUSGovernmentCloud",
	}

	name := strings.ToUpper(c.CloudEnvironmentName)
	envName, ok := lookup[name]
	if !ok {
		return fmt.Errorf("There is no cloud envionment matching the name '%s'!", c.CloudEnvironmentName)
	}

	env, err := azure.EnvironmentFromName(envName)
	c.cloudEnvironment = &env
	return err
}

func provideDefaultValues(c *Config) {
	if c.VMSize == "" {
		c.VMSize = DefaultVMSize
	}

	if c.ImageUrl == "" && c.ImageVersion == "" {
		c.ImageVersion = DefaultImageVersion
	}

	if c.CloudEnvironmentName == "" {
		c.CloudEnvironmentName = DefaultCloudEnvironmentName
	}
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
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("the tag name %q exceeds (%d) the 256 character limit", v, len(*v)))
		}
	}
}

func assertRequiredParametersSet(c *Config, errs *packer.MultiError) {
	/////////////////////////////////////////////
	// Authentication via OAUTH

	// Check if device login is being asked for, and is allowed.
	//
	// Device login is enabled if the user only defines SubscriptionID and not
	// ClientID, ClientSecret, and TenantID.
	//
	// Device login is not enabled for Windows because the WinRM certificate is
	// readable by the ObjectID of the App.  There may be another way to handle
	// this case, but I am not currently aware of it - send feedback.
	isUseDeviceLogin := func(c *Config) bool {
		if c.OSType == constants.Target_Windows {
			return false
		}

		return c.SubscriptionID != "" &&
			c.ClientID == "" &&
			c.ClientSecret == "" &&
			c.TenantID == ""
	}

	if isUseDeviceLogin(c) {
		c.useDeviceLogin = true
	} else {
		if c.ClientID == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A client_id must be specified"))
		}

		if c.ClientSecret == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A client_secret must be specified"))
		}

		if c.SubscriptionID == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("A subscription_id must be specified"))
		}
	}

	/////////////////////////////////////////////
	// Capture
	if c.CaptureContainerName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A capture_container_name must be specified"))
	}

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

	/////////////////////////////////////////////
	// Compute
	if c.ImageUrl == "" {
		if c.ImagePublisher == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_publisher must be specified"))
		}

		if c.ImageOffer == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_offer must be specified"))
		}

		if c.ImageSku == "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_sku must be specified"))
		}
	} else {
		if c.ImagePublisher != "" || c.ImageOffer != "" || c.ImageSku != "" || c.ImageVersion != "" {
			errs = packer.MultiErrorAppend(errs, fmt.Errorf("An image_url must not be specified if image_publisher, image_offer, image_sku, or image_version is specified"))
		}
	}

	if c.Location == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A location must be specified"))
	}

	/////////////////////////////////////////////
	// Deployment
	if c.StorageAccount == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A storage_account must be specified"))
	}
	if c.ResourceGroupName == "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("A resource_group_name must be specified"))
	}
	if c.VirtualNetworkName == "" && c.VirtualNetworkResourceGroupName != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_resource_group_name is specified, so must virtual_network_name"))
	}
	if c.VirtualNetworkName == "" && c.VirtualNetworkSubnetName != "" {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("If virtual_network_subnet_name is specified, so must virtual_network_name"))
	}

	/////////////////////////////////////////////
	// OS
	if c.OSType != constants.Target_Linux && c.OSType != constants.Target_Windows {
		errs = packer.MultiErrorAppend(errs, fmt.Errorf("An os_type must be specified"))
	}
}

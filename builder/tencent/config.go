package tencent

import (
	"encoding/gob"
	"path/filepath"
	"strings"

	// "fmt"

	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

var (
	DateCompiled string = "2018-05-01 19:58" // this is updated during the build
)

// Config contains the configuration for Builder
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Fields from config file
	ClientToken           string                   `mapstructure:"ClientToken"`
	DataDisks             []CVMDataDisk            `mapstructure:"DataDisks"`
	EnhancedService       CVMEnhancedService       `mapstructure:"EnhancedService"`
	SecurityGroupIds      []string                 `mapstructure:"SecurityGroupIds"`
	ImageID               string                   `mapstructure:"ImageId"`
	ImageName             string                   `mapstructure:"ImageName"`
	InstanceChargePrepaid CVMInstanceChargePrepaid `mapstructure:"InstanceChargePrepaid"`
	InstanceChargeType    string                   `mapstructure:"InstanceChargeType"`
	InstanceCount         int                      `mapstructure:"InstanceCount"`
	InstanceName          string                   `mapstructure:"InstanceName"`
	InstanceType          string                   `mapstructure:"InstanceType"`
	InternetAccessible    CVMInternetAccessible    `mapstructure:"InternetAccessible"`
	LoginSettings         CVMLoginSettings         `mapstructure:"LoginSettings"`
	Placement             CVMPlacement             `mapstructure:"Placement"`
	Region                string                   `mapstructure:"Region"`
	PublicKey             string                   `mapstructure:"PublicKey"`
	SecretID              string                   `mapstructure:"SecretId"`
	SecretKey             string                   `mapstructure:"SecretKey"`
	SkipSSH               bool                     `mapstructure:"SkipSSH"`
	SkipProvision         bool                     `mapstructure:"SkipProvision"`
	SSHKeyName            string                   `mapstructure:"KeyName"`
	SSHUserName           string                   `mapstructure:"ssh_username"`
	SystemDisk            CVMSystemDisk            `mapstructure:"SystemDisk"`
	TmiDescription        string                   `mapstructure:"tmi_description"`
	Version               string                   `mapstructure:"Version"`
	VirtualPrivateCloud   CVMVirtualPrivateCloud   `mapstructure:"VirtualPrivateCloud"`

	Url string `mapstructure:"Url"`

	// for switching the Image creation URL, just in case Tencent causes issue
	// by updating SSL certs, or something.
	ImageUrl string `mapstructure:"ImageUrl"`

	// For overriding the first step, ie, instead of StepCreateImage, change to StepRunImage
	Steps               []string `mapstructure:"Steps"`
	StartInstanceId     string   `mapstructure:"StartInstanceId"`
	Timeout             int64    `mapstructure:"Timeout"`
	IPAddrSaveLocation  string   `mapstructure:"IPAddrSaveLocation"`
	KeyPairSaveLocation string   `mapstructure:"KeyPairSaveLocation"`
	ImageIdLocation     string   `mapstructure:"ImageIdLocation"`

	Comm communicator.Config
	Ctx  interpolate.Context
}

// Registers the following data types to be used in Reflection
func init() {
	gob.Register(CVMDataDisk{})
	gob.Register(CVMEnhancedService{})
	gob.Register(CVMInstanceChargePrepaid{})
	gob.Register(CVMInternetAccessible{})
	gob.Register(CVMLoginSettings{})
	gob.Register(CVMPlacement{})
	gob.Register(CVMSystemDisk{})
	gob.Register(CVMVirtualPrivateCloud{})
}

// NewSimpleConfig parses the given raws parameter
func NewSimpleConfig(raws ...interface{}) (*Config, []string, error) {
	c := &Config{Placement: CVMPlacement{}, InternetAccessible: CVMInternetAccessible{},
		SystemDisk: CVMSystemDisk{}, DataDisks: []CVMDataDisk{}}

	warnings := []string{}

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.Ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	if c.PackerDebug {
		log.Printf("Decoding in NewSimpleConfig: %+v\n", raws)
	}

	if err != nil {
		if c.PackerDebug {
			log.Printf("NewSimpleConfig error decoding: %+v\n", err)
		}
		return nil, warnings, err
	}

	if c.Version == "" {
		c.Version = TencentAPIVersion
	}

	return c, warnings, nil
}

// NewConfig parses the given raws parameter.
// It requires the following keys to be set:
// ImageId,
// KeyName,
// Placement.Zone,
// Region,
// SecretId,
// SecretKey,
// ssh_username.
// If any on the above keys are not present, it returns an error.
// If all keys are present, it returns a pointer to a Config representing the parsed structure
// and warnings (if any), and no error.
func NewConfig(raws ...interface{}) (*Config, []string, error) {
	c := &Config{}
	warnings := []string{}
	c.PackerDebug = CloudAPIDebug
	if c.PackerDebug {
		log.Printf("NewConfig raws1: %+v\n", raws)
	}

	err := config.Decode(c, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &c.Ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	if c.PackerDebug {
		log.Printf("NewConfig raws: %+v\n", raws)
	}

	if err != nil {
		return nil, warnings, err
	}

	if c.Timeout == 0 {
		c.Timeout = 120000 // set timeout default to 2 minutes
	}

	if c.PackerDebug {
		log.Printf("NewConfig, decoded config: %+v\n", c)
	}

	if c.Version == "" {
		c.Version = TencentAPIVersion
	}

	var (
		errs           *packer.MultiError
		requireImageID bool = true
		requireKeyName bool = true
	)

	steps := []string{}

	if len(c.Steps) > 0 {
		requireKeyName = false
		requireImageID = false
		for _, step := range c.Steps {
			switch strings.ToUpper(step) {
			case strings.ToUpper(CStepClear):
				{
				}
			case strings.ToUpper(CStepConnectSSH):
				{
				}
			case strings.ToUpper(CStepCreateCustomImage):
				{
					if c.ImageIdLocation == "" {
						errs = packer.MultiErrorAppend(errs, errors.New("ImageIdLocation needs to be set"))
					}
					if c.ImageName == "" {
						errs = packer.MultiErrorAppend(errs, errors.New("ImageName needs to be set for the new master image"))
					}
					requireImageID = false
				}
			case strings.ToUpper(CStepCreateImage):
				{
					requireKeyName = true
					requireImageID = true
				}
			case strings.ToUpper(CStepCreateKeyPair):
				{
					requireKeyName = true
					requireImageID = true
				}
			case strings.ToUpper(CStepDisplayMessage):
				{
				}
			case strings.ToUpper(CStepGetInstanceIP):
				{
				}
			case strings.ToUpper(CStepHalt):
				{
				}
			case strings.ToUpper(CStepRunImage):
				{
					requireImageID = true
				}
			case strings.ToUpper(CStepStopImage):
				{
					requireImageID = true
					// if StopImage is the first step, it needs to read from ImageIdLocation so
					// verify that the file exists!

					if len(steps) == 0 {
						if c.ImageIdLocation == "" {
							errs = packer.MultiErrorAppend(errs, errors.New("ImageIdLocation needs to be set"))
						} else {
							requireImageID = false
							if !FileExists(c.ImageIdLocation) {
								FullFilename, _ := filepath.Abs(c.ImageIdLocation)
								errMsg := fmt.Sprintf("File specified in ImageIdLocation: %s doesn't exist!", FullFilename)
								errs = packer.MultiErrorAppend(errs, errors.New(errMsg))
							}
						}
					}
				}
			case strings.ToUpper(CStepGetKeyPairStatus):
				{
				}
			case strings.ToUpper(CStepProvision):
				{
				}
			case strings.ToUpper(CStepWaitRunning):
				{
				}
			case strings.ToUpper(CStepWaitStopped):
				{

				}
			default:
				{
					StepsArray := []string{
						CStepClear,
						CStepConnectSSH,
						CStepCreateCustomImage,
						CStepCreateImage,
						CStepCreateKeyPair,
						CStepDisplayMessage,
						CStepGetInstanceIP,
						CStepGetKeyPairStatus,
						CStepHalt,
						CStepProvision,
						CStepRunImage,
						CStepStopImage,
						CStepWaitRunning,
						CStepWaitStopped,
					}
					msg := fmt.Sprintf(`Valid steps are: "%s", current step is: %s`, strings.Join(StepsArray, `", "`), step)
					errs = packer.MultiErrorAppend(errs, errors.New(msg))
				}
			}
			steps = append(steps, step)
		}
	}

	if c.Placement.Zone == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("Placement.Zone needs to be set"))
	}

	if c.Region == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("Region needs to be set"))
	}

	if c.ImageName == "" {
		c.ImageName = SSHTimeStampSuffix()
		warnings = append(warnings, "Using timestamp suffix for ImageName, best to set it in the configuration instead")
	} else if len(c.ImageName) > 20 {
		errs = packer.MultiErrorAppend(errs, errors.New("ImageName must not be longer than 20 characters"))
	}

	if requireImageID && c.ImageID == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("ImageId needs to be set"))
	}

	if requireKeyName && c.SSHKeyName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("KeyName needs to be set"))
	}

	if c.SecretID == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("SecretId needs to be set"))
	}

	if c.SecretKey == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("SecretKey needs to be set"))
	}

	if c.SSHUserName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("ssh_username needs to be set"))
	}

	if c.Timeout == 0 {
		c.Timeout = 120000
	}

	if errs != nil && len(errs.Errors) > 0 {
		return nil, warnings, errs
	}

	return c, warnings, nil
}

func (c *Config) CreateCustomImageExtraParams(instanceId string) map[string]string {
	result := map[string]string{
		CImageName:  c.ImageName,
		CInstanceId: instanceId,
		CUrl:        CImageAPIUrl,
	}
	if c.ImageUrl != "" {
		result[CUrl] = c.ImageUrl
	}
	return result
}

func (c *Config) CreateCustomImageMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateQueryCustomImageExtraParams() map[string]string {
	result := map[string]string{
		"Filters.1.Name":     "image-type",
		"Filters.1.Values.1": "PRIVATE_IMAGE",
		CUrl:                 CImageAPIUrl,
	}
	return result
}

func (c *Config) CreateQueryImageMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateDescribeAddressesExtraParams() map[string]string {
	result := c.UrlParams()
	return result
}

func (c *Config) CreateDescribeAddressesMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateReleaseAddressesExtraParams(AddressId string) map[string]string {
	result := c.UrlParams()
	result["AddressIds.0"] = AddressId
	return result
}

func (c *Config) CreateReleaseAddressesMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateDescribeInstanceFamilyConfigsExtraParams() map[string]string {
	result := c.UrlParams()
	return result
}

func (c *Config) CreateDescribeInstanceFamilyConfigsMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateDescribeRegionsExtraParams() map[string]string {
	result := c.UrlParams()
	return result
}

func (c *Config) CreateDescribeRegionsMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	delete(result, CRegion)
	return result
}

func (c *Config) CreateDescribeZonesMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateDescribeZonesExtraParams() map[string]string {
	result := c.UrlParams()
	result[CRegion] = c.Region
	return result
}

func (c *Config) CreateDescribeKeyPairsExtraParams() map[string]string {
	result := c.UrlParams()
	return result
}

func (c *Config) CreateDescribeKeyPairsMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateBasicMap() map[string]interface{} {
	result := make(map[string]interface{})
	result[CRegion] = c.Region
	result[CSecretId] = c.SecretID
	result[CSecretKey] = c.SecretKey
	result[CVersion] = TencentAPIVersion
	return result
}

// This creates the map that is used for calling StopVM
func (c *Config) CreateStopVMMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

// This creates the map that is used for calling AssociateKeyPair
func (c *Config) CreateVMAKPMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateKeyPairExtraParams() map[string]string {
	result := c.UrlParams()
	result[CProjectId] = Int64ToStr(c.Placement.ProjectId)
	result[CRegion] = c.Region
	return result
}

// CreateKeyPairMap creates a map that is specifically used during CreateKeyPair
func (c *Config) CreateKeyPairMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

// CreateImportKeyPairExtraParams creates a map that is specifically used
// during CreateKeyPair
func (c *Config) CreateImportKeyPairExtraParams() map[string]string {
	result := map[string]string{
		CKeyName:   c.SSHKeyName,
		CProjectId: Int64ToStr(c.Placement.ProjectId),
		CPublicKey: c.PublicKey,
	}
	return result
}

func (c *Config) CreateImportKeyPairMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CKeyName] = c.SSHKeyName
	result[CPackerDebug] = c.PackerDebug
	result[CPublicKey] = c.PublicKey
	return result
}

func (c *Config) InquiryPriceRunParams() map[string]string {
	result := make(map[string]string)
	result[CImageId] = c.ImageID
	if c.Placement.Zone != "" {
		result[CPlacementZone] = c.Placement.Zone
		result[CPlacementProjectId] = Int64ToStr(c.Placement.ProjectId)
	}
	if c.Url != "" {
		result[CUrl] = c.Url
	}
	return result
}

func (c *Config) CreateInquiryPriceRunInstancesMap() map[string]interface{} {
	result := c.CreateBasicMap()
	return result
}

// This creates the map that is used for calling CreateVM
func (c *Config) CreateVMmap() map[string]interface{} {
	result := c.CreateBasicMap()
	if c.ImageID != "" {
		result[CImageId] = c.ImageID
	}
	if c.Placement.Zone != "" {
		result[CPlacementZone] = c.Placement.Zone
		result[CPlacementProjectId] = Int64ToStr(c.Placement.ProjectId)
		for i := 0; i < len(c.Placement.HostIds); i++ {
			LHostID := fmt.Sprintf("Placement.HostIds.%d", i)
			result[LHostID] = c.Placement.HostIds[i]
		}
	}
	if c.InternetAccessible.PublicIpAssigned {
		result[CInternetAccessible_PublicIpAssigned] = BoolToStr(c.InternetAccessible.PublicIpAssigned)
		result[CInternetAccessible_InternetMaxBandwidthOut] = Int64ToStr(c.InternetAccessible.InternetMaxBandwidthOut)
		c.InternetAccessible = CVMInternetAccessible{}
	}

	result[CRegion] = c.Region
	result[CSecretId] = c.SecretID
	result[CSecretKey] = c.SecretKey
	result[CPackerDebug] = c.PackerDebug

	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("CreateVMMap() result: %+v\n", result)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	return result
}

func (c *Config) CreateVMExtraParams() map[string]string {
	result := make(map[string]string)
	if c.ClientToken != "" {
		result[CClientToken] = c.ClientToken
	}

	if len(c.DataDisks) > 0 {
		for i, DataDisk := range c.DataDisks {
			LDiskSize := fmt.Sprintf("DataDisks.%d.DiskSize", i)
			LDiskType := fmt.Sprintf("DataDisks.%d.DiskType", i)
			result[LDiskSize] = Int64ToStr(DataDisk.DiskSize)
			result[LDiskType] = DataDisk.DiskType
		}
		c.DataDisks = nil
	}

	// Test for not empty
	if c.EnhancedService != (CVMEnhancedService{}) {
		result[CEnhancedServiceMonitorServiceEnabled] = BoolToStr(c.EnhancedService.MonitorService.Enabled)
		result[CEnhancedServiceSecurityServiceEnabled] = BoolToStr(c.EnhancedService.SecurityService.Enabled)
		c.EnhancedService = CVMEnhancedService{}
	}

	if len(c.SecurityGroupIds) > 0 {
		for i, GroupId := range c.SecurityGroupIds {
			key := fmt.Sprintf("SecurityGroupIds.%d", i)
			result[key] = GroupId
		}
	}

	if c.ImageID != "" {
		result[CImageId] = c.ImageID
	}

	// Test for not empty
	if c.InstanceChargePrepaid != (CVMInstanceChargePrepaid{}) {
		result[CInstanceChargePrepaidPeriod] = IntToStr(c.InstanceChargePrepaid.Period)
		result[CInstanceChargePrepaidRenewFlag] = c.InstanceChargePrepaid.RenewFlag
		c.InstanceChargePrepaid = CVMInstanceChargePrepaid{}
	}

	if c.InstanceChargeType != "" {
		result[CInstanceChargeType] = c.InstanceChargeType
		c.InstanceChargeType = ""
	}

	if c.InstanceCount != 0 {
		result[CInstanceCount] = IntToStr(c.InstanceCount)
		c.InstanceCount = 0
	}

	if c.InstanceName != "" {
		result[CInstanceName] = c.InstanceName
	}

	if c.InstanceType != "" {
		result[CInstanceType] = c.InstanceType
	}

	if c.InternetAccessible.InternetChargeType != "" || c.InternetAccessible.PublicIpAssigned {
		result[CInternetAccessible_InternetChargeType] = c.InternetAccessible.InternetChargeType
		result[CInternetAccessible_InternetMaxBandwidthOut] = Int64ToStr(c.InternetAccessible.InternetMaxBandwidthOut)
		result[CInternetAccessible_PublicIpAssigned] = BoolToStr(c.InternetAccessible.PublicIpAssigned)
	}

	if c.LoginSettings.Password != "" {
		result[CLoginSettingsPassword] = c.LoginSettings.Password
		result[CLoginSettingsKeepImageLogin] = c.LoginSettings.KeepImageLogin
		for i := 0; i < len(c.LoginSettings.KeyIds); i++ {
			LKeyID := fmt.Sprintf("LoginSettings.KeyIds.%d", i)
			result[LKeyID] = c.LoginSettings.KeyIds[i]
		}
	}

	if c.Placement.Zone != "" {
		result[CPlacementZone] = c.Placement.Zone
		result[CPlacementProjectId] = Int64ToStr(c.Placement.ProjectId)
		for i, HostId := range c.Placement.HostIds {
			LHostID := fmt.Sprintf("Placement.HostIds.%d", i)
			result[LHostID] = HostId
		}
	}

	if c.Region != "" {
		result[CRegion] = c.Region
	}

	// Test for not empty
	if c.SystemDisk != (CVMSystemDisk{}) {
		result[CSystemDisk_DiskSize] = Int64ToStr(c.SystemDisk.DiskSize)
		result[CSystemDisk_DiskType] = c.SystemDisk.DiskType
	}

	if c.Url != "" {
		result[CUrl] = c.Url
	}

	if c.VirtualPrivateCloud.VpcId != "" {
		result[CVirtualPrivateCloud_AsVpcGateway] = BoolToStr(c.VirtualPrivateCloud.AsVpcGateway)
		result[CVirtualPrivateCloud_SubnetId] = c.VirtualPrivateCloud.SubnetId
		result[CVirtualPrivateCloud_VpcId] = c.VirtualPrivateCloud.VpcId
		overrideInstanceCount := false
		for i, VpcIP := range c.VirtualPrivateCloud.PrivateIpAddresses {
			LIPName := fmt.Sprintf("VirtualPrivateCloud.PrivateIpAddresses.%d", i)
			result[LIPName] = VpcIP
			overrideInstanceCount = true
		}
		if overrideInstanceCount {
			delete(result, CInstanceCount)
		}
	}

	if c.SSHKeyName != "" {
		result[CKeyName] = c.SSHKeyName
	}
	return result
}

// Keys generates a dictionary of keys and values given in the c Config.
// Define any additional parameters required to any Cloud APIs here
// Objects, interface{}, must be translated to strings here, ie, flattened.
// eg, Placement, with members Zone, ProjectId needs to be translated as
// Placement.Zone, Placement.ProjectId, etc...
// In addition, additional parameters may be defined in the configInfo parameter
// in the call to CloudAPICall
// NOTE!!! InstanceCount is affected by VirtualPrivateCloud settings. See code.
func (c *Config) Keys() map[string]string {
	result := make(map[string]string)

	if c.ClientToken != "" {
		result[CClientToken] = c.ClientToken
	}

	if len(c.DataDisks) > 0 {
		for i, DataDisk := range c.DataDisks {
			LDiskSize := fmt.Sprintf("DataDisks.%d.DiskSize", i)
			LDiskType := fmt.Sprintf("DataDisks.%d.DiskType", i)
			result[LDiskSize] = Int64ToStr(DataDisk.DiskSize)
			result[LDiskType] = DataDisk.DiskType
		}
		c.DataDisks = nil
	}

	// Test for not empty
	if c.EnhancedService != (CVMEnhancedService{}) {
		result[CEnhancedServiceMonitorServiceEnabled] = BoolToStr(c.EnhancedService.MonitorService.Enabled)
		result[CEnhancedServiceSecurityServiceEnabled] = BoolToStr(c.EnhancedService.SecurityService.Enabled)
		c.EnhancedService = CVMEnhancedService{}
	}

	if len(c.SecurityGroupIds) > 0 {
		for i, GroupId := range c.SecurityGroupIds {
			key := fmt.Sprintf("SecurityGroupIds.%d", i)
			result[key] = GroupId
		}
	}

	if c.ImageID != "" {
		result[CImageId] = c.ImageID
	}

	// Test for not empty
	if c.InstanceChargePrepaid != (CVMInstanceChargePrepaid{}) {
		result[CInstanceChargePrepaidPeriod] = IntToStr(c.InstanceChargePrepaid.Period)
		result[CInstanceChargePrepaidRenewFlag] = c.InstanceChargePrepaid.RenewFlag
		c.InstanceChargePrepaid = CVMInstanceChargePrepaid{}
	}

	if c.InstanceChargeType != "" {
		result[CInstanceChargeType] = c.InstanceChargeType
		c.InstanceChargeType = ""
	}

	if c.InstanceCount != 0 {
		result[CInstanceCount] = IntToStr(c.InstanceCount)
		c.InstanceCount = 0
	}

	if c.InstanceName != "" {
		result[CInstanceName] = c.InstanceName
	}

	if c.InstanceType != "" {
		result[CInstanceType] = c.InstanceType
	}

	if c.InternetAccessible.InternetChargeType != "" || c.InternetAccessible.PublicIpAssigned {
		result[CInternetAccessible_InternetChargeType] = c.InternetAccessible.InternetChargeType
		result[CInternetAccessible_InternetMaxBandwidthOut] = Int64ToStr(c.InternetAccessible.InternetMaxBandwidthOut)
		result[CInternetAccessible_PublicIpAssigned] = BoolToStr(c.InternetAccessible.PublicIpAssigned)
		c.InternetAccessible = CVMInternetAccessible{}
	}

	if c.LoginSettings.Password != "" {
		result[CLoginSettingsPassword] = c.LoginSettings.Password
		result[CLoginSettingsKeepImageLogin] = c.LoginSettings.KeepImageLogin
		for i := 0; i < len(c.LoginSettings.KeyIds); i++ {
			LKeyID := fmt.Sprintf("LoginSettings.KeyIds.%d", i)
			result[LKeyID] = c.LoginSettings.KeyIds[i]
		}
		c.LoginSettings = CVMLoginSettings{}
	}

	if c.Placement.Zone != "" {
		result[CPlacementZone] = c.Placement.Zone
		result[CPlacementProjectId] = Int64ToStr(c.Placement.ProjectId)
		for i, HostId := range c.Placement.HostIds {
			LHostID := fmt.Sprintf("Placement.HostIds.%d", i)
			result[LHostID] = HostId
		}
		c.Placement = CVMPlacement{}
	}

	if c.Region != "" {
		result[CRegion] = c.Region
	}

	// Test for not empty
	if c.SystemDisk != (CVMSystemDisk{}) {
		result[CSystemDisk_DiskSize] = Int64ToStr(c.SystemDisk.DiskSize)
		result[CSystemDisk_DiskType] = c.SystemDisk.DiskType
		c.SystemDisk = CVMSystemDisk{}
	}

	if c.Url != "" {
	}

	if c.Version != "" {
		result[CVersion] = c.Version
	}

	if c.VirtualPrivateCloud.VpcId != "" {
		result[CVirtualPrivateCloud_AsVpcGateway] = BoolToStr(c.VirtualPrivateCloud.AsVpcGateway)
		result[CVirtualPrivateCloud_SubnetId] = c.VirtualPrivateCloud.SubnetId
		result[CVirtualPrivateCloud_VpcId] = c.VirtualPrivateCloud.VpcId
		overrideInstanceCount := false
		for i, VpcIP := range c.VirtualPrivateCloud.PrivateIpAddresses {
			LIPName := fmt.Sprintf("VirtualPrivateCloud.PrivateIpAddresses.%d", i)
			result[LIPName] = VpcIP
			overrideInstanceCount = true
		}
		if overrideInstanceCount {
			delete(result, CInstanceCount)
		}
		c.VirtualPrivateCloud = CVMVirtualPrivateCloud{}
	}

	if c.SSHKeyName != "" {
		result[CKeyName] = c.SSHKeyName
	}

	return result
}

func (c *Config) CreateGetInstanceIPmap(instanceId string) map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateGetInstanceIPExtraParams(instanceId string) map[string]string {
	result := map[string]string{
		"InstanceIds.0": instanceId,
		CRegion:         c.Region,
	}
	if c.Url != "" {
		result[CUrl] = c.Url
	}
	return result
}

func (c *Config) UrlParams() map[string]string {
	result := make(map[string]string)
	if c.Url != "" {
		result[CUrl] = c.Url
	}
	return result
}

func (c *Config) CreateStartInstanceMap() map[string]interface{} {
	result := c.CreateBasicMap()
	result[CPackerDebug] = c.PackerDebug
	return result
}

func (c *Config) CreateStartInstanceExtraParams(instanceId string) map[string]string {
	result := c.UrlParams()
	result["InstanceIds.0"] = instanceId
	return result
}

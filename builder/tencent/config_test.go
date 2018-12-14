package tencent

import (
	"testing"
)

// returns the minimal configuration that will supposedly work successfully
func createWorkingConfig() (map[string]interface{}, string) {
	requiredEnvVars := map[string]string{
		CPlacementZone: "ap-singapore-1",
		CRegion:        "ap-singapore",
		CSSHUserName:   "ubuntu"}
	GetRequiredEnvVars(requiredEnvVars)
	filename := TempFileName()
	return map[string]interface{}{
		CImageId:     "img-3wnd9xpl",
		CPlacement:   map[string]interface{}{CZone: requiredEnvVars[CPlacementZone]},
		CKeyName:     filename,
		CRegion:      requiredEnvVars[CRegion],
		CSecretId:    requiredEnvVars[CSecretId],
		CSecretKey:   requiredEnvVars[CSecretKey],
		CSSHUserName: requiredEnvVars[CSSHUserName],
	}, filename
}

// returns an empty configuration that will cause errors
func createEmptyConfig() map[string]interface{} {
	return map[string]interface{}{}
}

// This tests the NewConfig function in config.go works as expected.
func TestNewConfig(t *testing.T) {
	// Test an empty configuration
	raw1 := createEmptyConfig()
	_, warns, errs := NewConfig(raw1)
	CheckConfigHasErrors(t, warns, errs)

	// test a default working configuration
	raw2, _ := createWorkingConfig()
	_, warns, errs = NewConfig(raw2)
	CheckConfigIsOk(t, warns, errs)
}

func TestConfig_Keys(t *testing.T) {
	c := new(Config)

	c.ClientToken = "MyClientToken"
	c.DataDisks = []CVMDataDisk{CVMDataDisk{DiskSize: 50, DiskType: TencentCloudSSD}}
	c.EnhancedService.MonitorService.Enabled = true
	c.EnhancedService.SecurityService.Enabled = true
	c.ImageID = "MyImageID"
	c.InstanceChargePrepaid.Period = 5
	c.InstanceChargePrepaid.RenewFlag = "SomeValue" // doesn't verify value is valid
	c.InstanceChargeType = "MyInstanceChargeType"
	c.InstanceCount = 10
	c.InstanceName = "MyInstanceName"
	c.InstanceType = "MyInstanceType"
	c.InternetAccessible.InternetChargeType = "ChargeType"
	c.InternetAccessible.InternetMaxBandwidthOut = 99
	c.InternetAccessible.PublicIpAssigned = true
	c.LoginSettings.Password = "MyPass"
	c.LoginSettings.KeepImageLogin = "MyKeeper"
	c.LoginSettings.KeyIds = []string{"Key1", "Key2"}
	c.Placement.Zone = "MyZone"
	c.Placement.ProjectId = 5
	c.Region = "MyReg"
	c.SystemDisk.DiskSize = 50
	c.SystemDisk.DiskType = TencentCloudSSD
	c.Url = "MyUrl"
	c.Version = "MyVer"
	c.VirtualPrivateCloud.VpcId = "MyVpcId"
	c.VirtualPrivateCloud.AsVpcGateway = true
	c.VirtualPrivateCloud.SubnetId = "MySubnetId"
	c.VirtualPrivateCloud.PrivateIpAddresses = []string{"1", "2"}
	c.SSHKeyName = "MySSH"

	keys := c.Keys()
	if len(keys) != 31 {
		t.Errorf("Length of map not expected, config: %v, keys: %+v", len(keys), keys)
	}
	if keys[CKeyName] != "MySSH" {
		t.Fatalf("SSH KeyName value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CVirtualPrivateCloud_VpcId] != "MyVpcId" ||
		keys[CVirtualPrivateCloud_AsVpcGateway] != "TRUE" ||
		keys[CVirtualPrivateCloud_SubnetId] != "MySubnetId" ||
		keys["VirtualPrivateCloud.PrivateIpAddresses.0"] != "1" ||
		keys["VirtualPrivateCloud.PrivateIpAddresses.1"] != "2" ||
		c.VirtualPrivateCloud.VpcId != "" || c.VirtualPrivateCloud.SubnetId != "" ||
		c.VirtualPrivateCloud.AsVpcGateway != false {
		t.Fatalf("VirtualPrivateCloud value not expected, config: %+v, keys: %+v", c, keys)
	}

	if keys[CVersion] != "MyVer" || c.Version != "MyVer" {
		t.Fatalf("Version value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CUrl] != "MyUrl" || c.Url != "" {
		t.Fatalf("URL value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CSystemDisk_DiskSize] != "50" ||
		keys[CSystemDisk_DiskType] != TencentCloudSSD ||
		c.SystemDisk != (CVMSystemDisk{}) {
		t.Fatalf("SystemDisk value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CRegion] != "MyReg" {
		t.Fatalf("Region value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CPlacementZone] != "MyZone" ||
		keys[CPlacementProjectId] != "5" || c.Placement.Zone != "" || c.Placement.ProjectId != 0 {
		t.Fatalf("Placement value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CLoginSettingsPassword] != "MyPass" ||
		keys[CLoginSettingsKeepImageLogin] != "MyKeeper" ||
		keys["LoginSettings.KeyIds.0"] != "Key1" || keys["LoginSettings.KeyIds.1"] != "Key2" {
		t.Fatalf("LoginSettings value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CInternetAccessible_InternetChargeType] != "ChargeType" ||
		keys[CInternetAccessible_InternetMaxBandwidthOut] != "99" ||
		keys[CInternetAccessible_PublicIpAssigned] != "TRUE" || c.InternetAccessible != (CVMInternetAccessible{}) {
		t.Fatalf("InternetAccessible value not expected, config: %+v, keys: %+v", c, keys)
	}

	if keys[CInstanceType] != "MyInstanceType" && c.InstanceName != "MyInstanceType" {
		t.Fatalf("InstanceType value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CInstanceName] != "MyInstanceName" && c.InstanceName != "MyInstanceName" {
		t.Fatalf("InstanceName value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CInstanceCount] != "" || c.InstanceCount != 0 {
		t.Fatalf("InstanceCount value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CInstanceChargeType] != "MyInstanceChargeType" || c.InstanceChargeType != "" {
		t.Fatalf("InstanceChargeType value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CInstanceChargePrepaidPeriod] != "5" || keys[CInstanceChargePrepaidRenewFlag] != "SomeValue" ||
		c.InstanceChargePrepaid != (CVMInstanceChargePrepaid{}) {
		t.Fatalf("InstanceChargePrepaid value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CImageId] != "MyImageID" {
		t.Fatalf("ImageId value not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CClientToken] != "MyClientToken" {
		t.Fatalf("ClientToken not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CEnhancedServiceMonitorServiceEnabled] != "TRUE" ||
		keys[CEnhancedServiceSecurityServiceEnabled] != "TRUE" ||
		(c.EnhancedService != CVMEnhancedService{}) {
		t.Fatalf("EnhancedService values not expected, config: %+v, keys: %+v", c, keys)
	}
	if keys[CDataDisks_0_DiskSize] != "50" || keys[CDataDisks_0_DiskType] != TencentCloudSSD || c.DataDisks != nil {
		t.Fatalf("DataDisk values not expected, config: %+v, keys: %+v", c, keys)
	}
}

func TestConfig_CreateVMmap(t *testing.T) {
	var c Config
	c.SecretID = "MySecretId"
	c.SecretKey = "MySecretKey"
	c.Placement.Zone = "MyPlacementZone"
	c.Region = "MyRegion"
	c.ImageID = "MyImageId"
	c.InternetAccessible.PublicIpAssigned = true
	c.InternetAccessible.InternetMaxBandwidthOut = 2
	c.SystemDisk.DiskSize = 50
	c.SystemDisk.DiskType = TencentCloudSSD
	c.DataDisks = []CVMDataDisk{{DiskType: TencentCloudSSD, DiskSize: 100}}

	got := c.CreateVMmap()

	if !(got[CSecretId].(string) == c.SecretID) {
		t.Fatal("SecretId failed to match")
	}
	if !(got[CSecretKey].(string) == c.SecretKey) {
		t.Fatal("SecretKey failed to match")
	}
	if !(got[CRegion].(string) == c.Region) {
		t.Fatal("Region failed to match")
	}
	if !(got[CImageId].(string) == c.ImageID) {
		t.Fatal("ImageId failed to match")
	}
	if !(got[CInternetAccessible_PublicIpAssigned].(bool) == true) {
		t.Fatal("PublicIpAssigned failed to match")
	}
	if !(got[CInternetAccessible_InternetMaxBandwidthOut].(int64) == 2) {
		t.Fatal("InternetMaxBandwidthOut failed to match")
	}
	if got[CPlacement].(CVMPlacement).Zone != "MyPlacementZone" {
		t.Fatal("Placement failed to match")
	}
}

func TestNewSimpleConfig(t *testing.T) {
	VMConfig := map[string]interface{}{
		CPlacement: map[string]interface{}{CZone: "ap-singapore-1"},
		"InternetAccessible": map[string]interface{}{"PublicIpAssigned": true,
			"InternetMaxBandwidthOut": 2},
	}
	_, _, err := NewSimpleConfig(VMConfig)
	if err != nil {
		t.Errorf("NewSimpleConfig failed to parse, error: %+v", err)
	}
}

func TestConfig_CreateBasicMap(t *testing.T) {
	var c Config
	c.Region = "MyRegion"
	c.SecretID = "MySecretId"
	c.SecretKey = "MySecretKey"
	c.Placement.Zone = "bogus"

	myMap := c.CreateBasicMap()
	if (len(myMap) != 4) || (myMap[CRegion] != "MyRegion") || (myMap[CSecretId] != "MySecretId") ||
		(myMap[CSecretKey] != "MySecretKey") || (myMap["Version"] != TencentAPIVersion) {
		t.Fatalf("CreateBasicMap has been altered, current map: %+v", myMap)
	}
}

func TestConfig_CreateStopVMmap(t *testing.T) {
	var c Config
	c.Region = "MyRegion"
	c.SecretID = "MySecretId"
	c.SecretKey = "MySecretKey"
	c.Placement.Zone = "bogus"

	myMap := c.CreateStopVMMap()
	if (len(myMap) != 4) || (myMap[CRegion] != "MyRegion") || (myMap[CSecretId] != "MySecretId") ||
		(myMap[CSecretKey] != "MySecretKey") || (myMap[CVersion] != TencentAPIVersion) {
		t.Fatalf("CreateVMStopMap has been altered, current map: %+v", myMap)
	}
}

func TestConfig_CreateVMAKPMap(t *testing.T) {
	var c Config
	c.Region = "MyRegion"
	c.SecretID = "MySecretId"
	c.SecretKey = "MySecretKey"
	c.Placement.Zone = "bogus"

	myMap := c.CreateVMAKPMap()
	if (len(myMap) != 4) || (myMap[CRegion] != "MyRegion") || (myMap[CSecretId] != "MySecretId") ||
		(myMap[CSecretKey] != "MySecretKey") || (myMap[CVersion] != TencentAPIVersion) {
		t.Fatalf("CreateVMAKPMap has been altered, current map: %+v", myMap)
	}
}

func TestConfig_CreateKeyPairMap(t *testing.T) {
	var c Config
	c.Region = "MyRegion"
	c.SecretID = "MySecretId"
	c.SecretKey = "MySecretKey"
	c.Placement.Zone = "bogus"

	myMap := c.CreateKeyPairMap()
	if (len(myMap) != 4) || (myMap[CRegion] != "MyRegion") || (myMap[CSecretId] != "MySecretId") ||
		(myMap[CSecretKey] != "MySecretKey") || (myMap[CVersion] != TencentAPIVersion) {
		t.Fatalf("CreateKeyPairMap has been altered, current map: %+v", myMap)
	}
}

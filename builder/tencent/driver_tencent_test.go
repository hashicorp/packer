// Tests in this file require InstanceId to be updated
package tencent

// TestTencentDriver_WaitForImageState requires InstanceId
// TestTencentDriver_StopImage requires InstanceId
// TestTencentDriver_RunImage requires InstanceId
// TestTencentDriver_CWCreateKeyPair requires a valid InstanceId
// TestTencentDriver_ConnectKeyPair requires InstanceId
// TestTencentDriver_CWDeleteImage

import (
	"fmt"
	"log"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func init() {
	CloudAPIDebug = true
}

func TestTencentDriver_WaitForImageState(t *testing.T) {
	InstanceId := "ins-orux6g2q" // UPDATE THIS
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		CInstanceId: InstanceId,
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.InternetAccessible.InternetMaxBandwidthOut = 2
	c.InternetAccessible.PublicIpAssigned = true
	c.ImageID = "img-3wnd9xpl"
	c.Placement.Zone = "ap-singapore-1"
	c.SSHUserName = "ubuntu"
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]

	driver := new(TencentDriver)
	err := driver.CWWaitForImageState(c, InstanceId, "STOPPED")
	if err != nil {
		log.Printf("%s error: %+v\n", CStepStopImage, err)
	}
}

func TestTencentDriver_StopImage(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		CInstanceId: "ins-fcs4zccs",
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.InternetAccessible.InternetMaxBandwidthOut = 2
	c.InternetAccessible.PublicIpAssigned = true
	c.ImageID = "img-3wnd9xpl"
	c.Placement.Zone = "ap-singapore-1"
	c.SSHUserName = "ubuntu"
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	InstanceId := requiredEnvVars[CInstanceId]
	driver := new(TencentDriver)
	err := driver.CWStopImage(c, InstanceId)

	if err != nil {
		log.Printf("StepStopImage error: %+v\n", err)
	}

}

func TestTencentDriver_RunImage(t *testing.T) {
	InstanceId := "ins-fcs4zccs"
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		CInstanceId: InstanceId,
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.InternetAccessible.InternetMaxBandwidthOut = 2
	c.InternetAccessible.PublicIpAssigned = true
	c.ImageID = "img-3wnd9xpl"
	c.Placement.Zone = "ap-singapore-1"
	c.SSHUserName = "ubuntu"
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]

	driver := new(TencentDriver)
	driver.CWRunImage(c, InstanceId)
}

// CWCreateKeyPair requires a valid instance
func TestTencentDriver_CWCreateKeyPair(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion:  "ap-singapore",
		CKeyName: CurrentTimeStamp(),
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	c.SSHKeyName = requiredEnvVars[CKeyName]
	state := new(multistep.BasicStateBag)

	driver := new(TencentDriver)
	err, keyPairInfo := driver.CWCreateKeyPair(c, "ins-69sd8z8o", state)
	if err != nil && keyPairInfo.KeyId != "" {
		t.Fatalf("Unexpected response, error: %+v, keypair: %+v", err, keyPairInfo)
	}
	t.Log("Verifying SSH key location")
	SSHKeyLocation, ok := state.GetOk(CSSHKeyLocation)
	if ok && !FileExists(SSHKeyLocation.(string)) {
		t.Fatalf("SSH key doesn't exist at %s", SSHKeyLocation.(string))
	}
}

func TestTencentDriver_ConnectKeyPair(t *testing.T) {
	var response1 CVMCreateKeyPairResponse
	var c Config
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		CInstanceId: "ins-gta6wn7w",
	}
	GetRequiredEnvVars(requiredEnvVars)
	response1.KeyPair.KeyId = "skey-czp0fvll"
	c.InternetAccessible.InternetMaxBandwidthOut = 2
	c.InternetAccessible.PublicIpAssigned = true
	c.ImageID = "img-3wnd9xpl"
	c.Placement.Zone = "ap-singapore-1"
	c.SSHUserName = "ubuntu"
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	instanceId := requiredEnvVars[CInstanceId]

	KeyPairId := response1.KeyPair.KeyId
	response2 := AssociateInstanceKeyPair(&c, instanceId, KeyPairId)
	errMsg := ""
	if response2.Error.Code != "" {
		log.Println("Successfully associated instance keypair!")
	} else {
		errMsg := fmt.Sprintf("CreateKeyPair error code: %s, message: %s", response2.Error.Code, response2.Error.Message)
		log.Println(errMsg)
	}
	completed := CheckBindingCompleted(&c, instanceId)
	log.Printf("Instance state is %v and error: %s\n", completed, errMsg)
}

func TestTencentDriver_CWWaitKeyPairAttached(t *testing.T) {
	var c Config
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		CInstanceId: "ins-gta6wn7w",
	}
	c.Timeout = 60000
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	InstanceId := requiredEnvVars[CInstanceId]
	ui := NewPackerUi()
	state := new(multistep.BasicStateBag)
	driver := NewTencentDriver(ui, &c, state)
	driver.CWWaitKeyPairAttached(c, InstanceId, "KeyId")
}

// TODO(chuacw) test the CWDeleteImage functionality
func TestTencentDriver_CWDeleteImage(t *testing.T) {
	// var response1 CVMCreateKeyPairResponse
	// var c Config
	// requiredEnvVars := map[string]string{
	// 	CRegion:     "ap-singapore",
	// 	CInstanceId: "ins-gta6wn7w",
	// }
	// GetRequiredEnvVars(requiredEnvVars)
	// c.Placement.Zone = "ap-singapore-1"
	// c.SSHUserName = "ubuntu"
	// c.Region = requiredEnvVars[CRegion]
	// c.SecretID = requiredEnvVars[CSecretId]
	// c.SecretKey = requiredEnvVars[CSecretKey]
	// instanceId := requiredEnvVars[CInstanceId]
	// TencentDriver_CWDeleteImage(t)

}

func TestTencentDriver_CWCreateImage(t *testing.T) {
	type args struct {
		c Config
	}
	tests := []struct {
		name   string
		driver TencentDriver
		args   args
		want   bool
		want1  CVMError
		want2  CVMInstanceInfo
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := tt.driver.CWCreateImage(tt.args.c)
			if got != tt.want {
				t.Errorf("TencentDriver.CWCreateImage() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("TencentDriver.CWCreateImage() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("TencentDriver.CWCreateImage() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestTencentDriver_CWGetImageState(t *testing.T) {
	// See  https://stackoverflow.com/questions/31362044/anonymous-interface-implementation-in-golang
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		CInstanceId: "ins-orux6g2q", // UPDATE THIS!!!
	}
	GetRequiredEnvVars(requiredEnvVars)
	customUi := NewPackerUi()
	config := &Config{
		Region:    requiredEnvVars[CRegion],
		SecretID:  requiredEnvVars[CSecretId],
		SecretKey: requiredEnvVars[CSecretKey],
	}
	InstanceId := requiredEnvVars[CInstanceId]
	stateBag := new(multistep.BasicStateBag)
	driver := NewTencentDriver(customUi, config, stateBag)
	err, state := driver.CWGetImageState(*config, InstanceId)
	if err != nil {
		if state == "" {
			t.Errorf("TestTencentDriver_CWGetImageState Unexpected state: %v", state)
		}
		t.Fatalf("TestTencentDriver_CWGetImageState Unexpected error: %+v", err)
	}
	switch state {
	case "STOPPED", "RUNNING", "PENDING", "REBOOTING", "STARTING", "STOPPING":
		return
	default:
		t.Fatalf("TestTencentDriver_CWGetImageState unknown state: %v", state)
	}
}

func TestTencentDriver_CWRunImage(t *testing.T) {
	type args struct {
		c          Config
		instanceId string
	}
	tests := []struct {
		name    string
		driver  TencentDriver
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.driver.CWRunImage(tt.args.c, tt.args.instanceId); (err != nil) != tt.wantErr {
				t.Errorf("TencentDriver.CWRunImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTencentDriver_CWStopImage(t *testing.T) {
	type args struct {
		c          Config
		instanceId string
	}
	tests := []struct {
		name    string
		driver  TencentDriver
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.driver.CWStopImage(tt.args.c, tt.args.instanceId); (err != nil) != tt.wantErr {
				t.Errorf("TencentDriver.CWStopImage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTencentDriver_WaitForImageCreation(t *testing.T) {
}

func TestTencentDriver_WaitForImageDeletion(t *testing.T) {
	// WaitForImageDeletion is unused
}

func TestTencentDriver_CWWaitForImageState(t *testing.T) {
	// WaitForImageState
}

func TestNewTencentDriver(t *testing.T) {
	// Simple
}

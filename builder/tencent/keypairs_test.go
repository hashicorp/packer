package tencent

import (
	"log"
	"reflect"
	"testing"
)

func init() {
	CloudAPIDebug = true
}

func TestDescribeKeyPairs(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion: "ap-singapore",
		CUrl:    CCVMUrlSiliconValley,
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	c.Url = requiredEnvVars[CUrl]

	cvmKeyPairResponse := DescribeKeyPairs(&c)
	log.Printf("Response: %+v", cvmKeyPairResponse)
}

func TestCreateKeyPair(t *testing.T) {
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
	c.Url = CCVMUrlSiliconValley

	response := CreateKeyPair(&c)
	if response.Error.Code != "" {
		t.Errorf("CreateKeyPair error code: %s, message: %s", response.Error.Code, response.Error.Message)
	}

	if response.KeyPair.KeyName != c.SSHKeyName {
		t.Error("CreateKeyPair failed to generate")
	}

}

func TestAssociateInstanceKeyPair(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		"KeyPairId": "skey-83vcliah",
		CInstanceId: "ins-k3zbpvla",
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	c.SSHKeyName = CurrentTimeStamp()
	KeyPairId := requiredEnvVars["KeyPairId"]
	InstanceId := requiredEnvVars[CInstanceId]

	result := AssociateInstanceKeyPair(&c, InstanceId, KeyPairId)
	// InvalidKeyPairId is a valid result
	if result.Error.Code == "InvalidKeyPairId.NotFound" {
		return
	}
	// Other valid tests, eg
	// if result.Error.Code != "valid result1" && result.Error.Code != "valid result2" ... {
	//   t.Fatal("...")
	// }

	t.Fatalf("Unexpected error code: %s, message: %s", result.Error.Code, result.Error.Message)
}

func TestDisassociateInstancesKeyPairs(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		"KeyPairId": "skey-m8tzulzx",
		CInstanceId: "ins-69sd8z8o",
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	c.SSHKeyName = CurrentTimeStamp()
	KeyPairId := requiredEnvVars["KeyPairId"]
	InstanceId := requiredEnvVars[CInstanceId]
	c.Url = CCVMUrlSiliconValley

	result := DisassociateInstancesKeyPairs(&c, InstanceId, KeyPairId)
	// InvalidKeyPairId is a valid result
	if result.Error.Code == "InvalidKeyPairId.NotFound" {
		return
	}
}

func TestImportKeyPair(t *testing.T) {
	type args struct {
		c *Config
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ImportKeyPair(tt.args.c)
		})
	}
}

func TestDeleteKeyPair(t *testing.T) {
	type args struct {
		c         *Config
		KeyPairId string
	}
	tests := []struct {
		name string
		args args
		want CVMDeleteKeyPairResponse
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DeleteKeyPair(tt.args.c, tt.args.KeyPairId); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeleteKeyPair() = %v, want %v", got, tt.want)
			}
		})
	}
}

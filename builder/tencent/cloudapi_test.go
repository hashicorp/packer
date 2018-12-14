package tencent

// Running this file's test requires the following variables to be updated
// Valid values for TestCreateVM: Region, ImageId, PlacementZone
// Valid values for TestWaitforVM: Region, InstanceId

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

func init() {
	CloudAPIDebug = true
}

// Self contained test to ensure Hmac1ToBase64 works as expected
func TestHmac1ToBase64(t *testing.T) {
	secretKey1 := "Gu5t9xGARNpq86cd98joQYCN3Cozk1qA"
	secretKey2 := "Fu5t9xGARNpq86cd98joQYCN3Cozk1qX"
	srcstr1 := "GETcvm.api.qcloud.com/v2/index.php?Action=DescribeInstances&Nonce=11886&Region=gz&SecretId=AKIDz8krbsJ5yKBZQpn74WFkmLPx3gnPhESA&Timestamp=1465185768&instanceIds.0=ins-09dx96dg&limit=20&offset=0"
	srcstr2 := "GETcvm.tencentcloudapi.com/?Action=DescribeInstances&Nonce=11886&Region=gz&SecretId=AKIDz8krbsJ5yKBZQpn74WFkmLPx3gnPhESA&Timestamp=1465185768&instanceIds.0=ins-09dx96dg&limit=20&offset=0"
	type args struct {
		key   string
		str   string
		IsUrl bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// See https://intl.cloud.tencent.com/document/product/362/4208?!preview=true&lang=en#2.4.-generating-signature-string
		{"Test Hmac1ToBase64 case 1", args{secretKey1, srcstr1, false}, "NSI3UqqD99b/UJb4tbG/xZpRW64="},
		{"Test Hmac1ToBase64 case 2", args{secretKey1, srcstr1, true}, "NSI3UqqD99b%2FUJb4tbG%2FxZpRW64%3D"},
		{"Test Hmac1ToBase64 case 3", args{secretKey2, srcstr2, false}, "ZLvslzTWZTUyzVPopuMw3fKMQkg="},
		{"Test Hmac1ToBase64 case 4", args{secretKey2, srcstr2, true}, "ZLvslzTWZTUyzVPopuMw3fKMQkg%3D"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hmac1ToBase64(tt.args.key, tt.args.str, tt.args.IsUrl); got != tt.want {
				t.Errorf("Hmac1ToBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

// self contained test to ensure that SignatureGet works as expected
func TestSignatureGet(t *testing.T) {
	url := "cvm.api.qcloud.com/"
	requestParams := "Action=DescribeInstances&Nonce=11886&Region=gz&SecretId=AKIDz8krbsJ5yKBZQpn74WFkmLPx3gnPhESA&Timestamp=1465185768&instanceIds.0=ins-09dx96dg&limit=20&offset=0"
	secretKey := "Gu5t9xGARNpq86cd98joQYCN3Cozk1qA"
	type args struct {
		requestURL         string
		requestParamString string
		secretKey          string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"SignatureGet test 1", args{url, requestParams, secretKey}, "2wmvFvB6R7CAVEzYcjO8BKTsvj4%3D"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SignatureGet(tt.args.requestURL, tt.args.requestParamString, tt.args.secretKey); got != tt.want {
				t.Errorf("SignatureGet() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Self contained test to ensure that encoding base64 works.
func TestEncodingBase64(t *testing.T) {
	type args struct {
		b     []byte
		IsURL bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"EncodeBase64 test 1", args{[]byte("chuacw rocks!"), false}, "Y2h1YWN3IHJvY2tzIQ=="},
		{"EncodeBase64 test 2", args{[]byte("chuacw rocks!"), true}, "Y2h1YWN3IHJvY2tzIQ%3D%3D"},
		{"EncodeBase64 test 3", args{[]byte("chuacw isn't cool!"), false}, "Y2h1YWN3IGlzbid0IGNvb2wh"},
		{"EncodeBase64 test 4", args{[]byte("chuacw isn't cool!"), true}, "Y2h1YWN3IGlzbid0IGNvb2wh"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodingBase64(tt.args.b, tt.args.IsURL); got != tt.want {
				t.Errorf("EncodingBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHmac256ToBase64(t *testing.T) {
	type args struct {
		key   string
		str   string
		IsUrl bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Hmac256ToBase64 test 1", args{"chuacw", "chuacw rocks!", false}, "WNIlcsw2mci4IZ+B5CJS6vRNZ7WTRnQjM003R/bOd9A="},
		{"Hmac256ToBase64 test 2", args{"chuacw", "chuacw rocks!", true}, "WNIlcsw2mci4IZ%2BB5CJS6vRNZ7WTRnQjM003R%2FbOd9A%3D"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Hmac256ToBase64(tt.args.key, tt.args.str, tt.args.IsUrl); got != tt.want {
				t.Errorf("Hmac256ToBase64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateNonce(t *testing.T) {
	nonce1 := GenerateNonce()
	time.Sleep(1 * time.Millisecond)
	// sleep so as to prevent the generated nonce from having the same value
	nonce2 := GenerateNonce()
	if nonce1 == nonce2 {
		t.Errorf("GenerateNonce shouldn't have same value, %s == %s", nonce1, nonce2)
	}
}

// This tests SignatureString, since it is a wrapper to SignatureStringNonceTimestamp
// SignatureStringNonceTimestamp calls GenerateNonce() and CurrentTimeStamp() then passes
// these values to SignatureStringNonceTimestamp
func TestSignatureStringNonceTimestamp(t *testing.T) {
	action1 := "act1"
	action2 := "act2"
	nonce1 := "167890"
	nonce2 := "223569"
	ts1 := "1525252824"
	ts2 := "1525252900"
	secretId1 := "AKIDz8krbsJ5yKBZQpn74WFkmLPx3gnPhESA"
	secretId2 := "AKIDz8kkksJ5yKBZQpn74WFkmLPx3gnPhEXY"
	epK1 := "str1"
	epV1 := "value1"
	epK2 := "str2"
	epV2 := "value2"
	extraParams1 := map[string]string{
		epK1: epV1,
		epK2: epV2,
	}

	expected1 := strings.Join([]string{"Action=" + action1, "Nonce=" + nonce1, "SecretId=" + secretId1, "Timestamp=" + ts1,
		epK1 + "=" + epV1, epK2 + "=" + epV2}, "&")
	expected2 := strings.Join([]string{"Action=" + action2, "Nonce=" + nonce2, "SecretId=" + secretId2, "Timestamp=" + ts2,
		epK1 + "=" + epV1, epK2 + "=" + epV2}, "&")

	type args struct {
		action      string
		nonce       string
		timestamp   string
		secretId    string
		extraParams map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"Test SignatureStringNonceTimestamp test 1", args{action1, nonce1, ts1, secretId1, extraParams1}, expected1},
		{"Test SignatureStringNonceTimestamp test 2", args{action2, nonce2, ts2, secretId2, extraParams1}, expected2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SignatureStringNonceTimestamp(tt.args.action, tt.args.nonce, tt.args.timestamp, tt.args.secretId, tt.args.extraParams); got != tt.want {
				t.Errorf("SignatureStringNonceTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test RequestGet is working, ie, when it returns HTTP 200, it can decode return the expected bytes
// When it doesn't return HTTP 200, then the DecodeResponse it calls would fail
func TestRequestGet(t *testing.T) {
	type (
		HttpHeaders struct {
			AcceptEncoding string `json:"Accept-Encoding"`
			Connection     string
			Host           string
			UserAgent      string `json:"User-Agent"`
		}
		HttpResponse struct {
			Args    map[string]interface{}
			Headers map[string]interface{}
			Origin  string
			Url     string
		}
	)
	response, err := RequestGet("mockbin.org/request", "para", "sig")
	if err != nil {
		t.Errorf("Error in testing TestRequestGet: %+v", err)
	}
	var decodedResponse HttpResponse
	err = json.Unmarshal(response, &decodedResponse)
	if err != nil {
		t.Errorf("Error in testing TestRequestGet: %+v", err)
	}
	expected := "https://mockbin.org/request?para&Signature=sig"
	if decodedResponse.Url != expected {
		t.Errorf("TestRequest got: %s, want %s", decodedResponse.Url, expected)
	}

}

// Test that the signature string function is working
func TestSignatureString(t *testing.T) {
	// As SignatureString consists of calling GenerateNonce, CurrentTimestamp followed by a call
	// to SignatureStringNonceTimestamp, testing SignatureString should be a combo of testing
	// TestGenerateNonce, TestCurrentTimestamp and TestSignatureStringNonceTimestamp
	TestGenerateNonce(t)
	TestCurrentTimeStamp(t)
	TestSignatureStringNonceTimestamp(t)
}

// TestDecodeResponse is self-contained, and do not need to get anything updated in order to run it
func TestDecodeResponse(t *testing.T) {

	type expected1struct struct {
		Response interface{}
	}
	var decoded1 expected1struct
	err := DecodeResponse([]byte(`{"Response":{}}`), &decoded1)
	if err != nil {
		t.Errorf("Failed to decode, error: %+v", err)
	}

	type expected2struct struct {
		Message string
	}
	var decoded2 expected2struct
	err = DecodeResponse([]byte(`{"Response":{"Message":"Hello"}}`), &decoded2)
	if err != nil || (err == nil && decoded2.Message != "Hello") {
		t.Errorf("Failed to decode, error: %+v", err)
	}

	type expected3struct struct {
		Message string
	}
	var decoded3 expected3struct
	err = DecodeResponse([]byte(`{"Response":{"TotalCount":10}}`), &decoded3)
	if err == nil { // error should contain "unknown configuration key"
		t.Error("DecodedResponse test case 3: Should have an error, but didn't!")
	} else {
		errStr := err.Error()
		if !strings.Contains(errStr, "unknown configuration key") {
			t.Errorf("Unexpected error: %+v", err)
		}
	}

}

// In order to run this test, ensure that a valid instance id, existing in the given
// Placement Zone and Region has either an IP address, or doesn't have an IP address.
func TestGetInstanceState(t *testing.T) {
	requiredEnvVars := map[string]string{
		CInternetAccessible_PublicIpAssigned:        "true",
		CInternetAccessible_InternetMaxBandwidthOut: "2",
		CPlacementZone:                              "ap-singapore-1",
		CRegion:                                     "ap-singapore",
		CInstanceId:                                 "ins-4uwn83c2", // UPDATE THIS INSTANCE
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	// 2018/05/06 14:18:54 packer-builder-tencent.exe: 2018/05/06 14:18:54 map[SecretId:xxxxxxx SecretKey:yyyyy ImageId:i
	// mg-3wnd9xpl Placement.Zone:ap-singapore-1 InternetAccessible.PublicIpAssigned:true InternetAccessible.InternetMaxBandwidthOut:2 ssh_username:ubuntu Region:ap-singapore]
	// 2018/05/06 14:18:54 packer-builder-tencent.exe: 2018/05/06 14:18:54 Decoding in NewSimpleConfig: [map[Region:ap-singapore SecretId:AKIDELI5jCYbXIsERVLGsJfrsmd1KONeumdO Se
	// cretKey:53sZ5RAoLiuwsgacUVvFWar32eKB5tb9 ImageId:img-3wnd9xpl Placement.Zone:ap-singapore-1 InternetAccessible.PublicIpAssigned:true InternetAccessible.InternetMaxBandwid
	// thOut:2 ssh_username:ubuntu]]
	c.InternetAccessible.PublicIpAssigned = StrToBool(requiredEnvVars[CInternetAccessible_PublicIpAssigned])
	c.InternetAccessible.InternetMaxBandwidthOut = StrToInt64(requiredEnvVars[CInternetAccessible_InternetMaxBandwidthOut])
	c.ImageID = "junk"
	c.Placement.Zone = requiredEnvVars[CPlacementZone]
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Url = CCVMUrlSiliconValley

	// // valid region, valid name (but doesn't exist), returns INVALID, no error
	// {"GetInstanceState", args{&c, "ap-singapore", "ins-01s7m24k"}, "RUNNING", false},
	// {"GetInstanceState", args{&c, "ap-shanghai", "ins-dhheb7oi"}, "INVALID", false},
	// // invalid region, invalid name (and doesn't exist), returns INVALID, and error
	// {"GetInstanceState", args{&c, "", ""}, "INVALID", true},

	err1, res1 := GetInstanceState(&c, "ins-rbceuxvs")
	err2, res2 := GetInstanceState(&c, "ins-dhheb7oi")
	c.Region = "NO_REGION"
	err3, res3 := GetInstanceState(&c, "ins-noname")

	err := (err1 != nil && res1 != "INVALID") ||
		(err2 != nil && res2 != "INVALID") ||
		(err3 != nil && res3 != "INVALID")
	if err {
		t.Fatal("GetInstanceState not working in expected manner")
	}
}

// To ensure this test is successful, insert an InstanceId that is in the STOPPED state
// and can be started.
// func TestStartWaitStopVM(t *testing.T) {
// 	requiredEnvVars := map[string]string{
// 		CRegion:     "ap-singapore",
// 		CInstanceId: "ins-eyh7fj54", // UPDATE THIS INSTANCE HERE!!!
// 	}
// 	GetRequiredEnvVars(requiredEnvVars)
// 	var c Config
// 	c.SecretID = requiredEnvVars[CSecretId]
// 	c.SecretKey = requiredEnvVars[CSecretKey]
// 	c.Region = requiredEnvVars[CRegion]
// 	InstanceId := requiredEnvVars[CInstanceId]
// 	c.Timeout = 60000

// 	const (
// 		cStatus1   = "StartStopVM %s instance: %s to be in %s state successfully"
// 		cStatus2   = "StartStopVM %s instance: %s successfully"
// 		cWaitedFor = "waited for"
// 		cStarted   = "started"
// 		cStopped1  = "stopped"
// 		cStopped2  = "STOPPED"
// 		cRunning   = "RUNNING"
// 	)

// 	// cvm.ap-singapore.tencentcloudapi.com needs ImageId, Region, Placement.Zone
// 	actualError1, successful1 := WaitForVM(&c, InstanceId, "STOPPED")
// 	if !successful1 {
// 		t.Fatalf("StartWaitStopVM failed to get state of VM: %s, error: %+v", InstanceId, actualError1)
// 	}
// 	t.Logf(cStatus1, cWaitedFor, InstanceId, cStopped2)

// 	c.SSHKeyName = CurrentTimeStamp() // need a keyname
// 	response := CreateKeyPair(&c)
// 	if response.Error.Code != "" {
// 		t.Errorf("StartStopVM CreateKeyPair error code: %s, message: %s", response.Error.Code, response.Error.Message)
// 	}

// 	KeyPairId := response.KeyPair.KeyId

// 	bindResult := AssociateInstanceKeyPair(&c, InstanceId, KeyPairId)
// 	if bindResult.Error.Code != "" {
// 		t.Fatalf("Failed to bind keypair, error code: %s message: %s", bindResult.Error.Code,
// 			bindResult.Error.Message)
// 	}

// 	actualError2, successful2 := StartVM(&c, InstanceId)
// 	if !successful2 {
// 		t.Fatalf("StartStopVM failed to start VM, error: %+v", actualError2)
// 	}
// 	t.Logf(cStatus2, cStarted, InstanceId)

// 	actualError3, successful3 := WaitForVM(&c, InstanceId, "RUNNING")
// 	if !successful3 {
// 		t.Fatalf("StartStopVM failed to get state of VM after start: %s, error: %+v", InstanceId, actualError3)
// 	}
// 	t.Logf(cStatus1, cWaitedFor, InstanceId, cRunning)

// 	actualError4, successful4 := StopVM(&c, InstanceId)
// 	if !successful4 {
// 		t.Errorf("StartStopVM failed to stop VM, error: %+v", actualError4)
// 	}
// 	t.Logf(cStatus2, cStopped1, InstanceId)

// 	unbindResult := DisassociateInstancesKeyPairs(&c, InstanceId, KeyPairId)
// 	if unbindResult.Error.Code != "" {
// 		t.Fatalf("Failed to bind keypair, error code: %s message: %s", unbindResult.Error.Code,
// 			unbindResult.Error.Message)
// 	}

// 	actualError5, successful5 := WaitForVM(&c, InstanceId, "STOPPED")
// 	if !successful5 {
// 		t.Fatalf("StartStopVM failed to get state of VM after stop: %s, error: %+v", InstanceId, actualError5)
// 	}
// 	t.Logf(cStatus1, cWaitedFor, InstanceId, cStopped2)

// }

func CreateCreateVMConfig() Config {
	requiredEnvVars := map[string]string{
		CRegion:                                     "ap-singapore",
		CImageId:                                    "img-3wnd9xpl",
		CPlacementZone:                              "ap-singapore-1",
		CInternetAccessible_PublicIpAssigned:        "true",
		CInternetAccessible_InternetMaxBandwidthOut: "1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.DataDisks = []CVMDataDisk{{DiskSize: 100, DiskType: TencentCloudSSD}}
	c.ImageID = requiredEnvVars[CImageId]
	c.InternetAccessible.PublicIpAssigned = StrToBool(requiredEnvVars[CInternetAccessible_PublicIpAssigned])
	c.InternetAccessible.InternetMaxBandwidthOut = StrToInt64(requiredEnvVars[CInternetAccessible_InternetMaxBandwidthOut])
	c.Placement.Zone = requiredEnvVars[CPlacementZone]
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.SystemDisk = CVMSystemDisk{DiskSize: 50, DiskType: TencentCloudSSD}
	c.VirtualPrivateCloud.VpcId = "vpc-ie6ri3jv"
	c.VirtualPrivateCloud.SubnetId = "subnet-5kdpieu8"
	c.SecurityGroupIds = []string{"sg-elhg6l30"}
	c.Timeout = 120000
	c.Url = CCVMUrlSiliconValley

	return c
}

func CreateSSHConnectConfig() Config {
	requiredEnvVars := map[string]string{
		CRegion:                                     "na-siliconvalley",
		CImageId:                                    "img-3wnd9xpl",
		CPlacementZone:                              "na-siliconvalley-1",
		CInternetAccessible_PublicIpAssigned:        "true",
		CInternetAccessible_InternetMaxBandwidthOut: "1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.DataDisks = []CVMDataDisk{{DiskSize: 100, DiskType: TencentCloudSSD}}
	c.ImageID = requiredEnvVars[CImageId]
	c.InternetAccessible.PublicIpAssigned = StrToBool(requiredEnvVars[CInternetAccessible_PublicIpAssigned])
	c.InternetAccessible.InternetMaxBandwidthOut = StrToInt64(requiredEnvVars[CInternetAccessible_InternetMaxBandwidthOut])
	c.Placement.Zone = requiredEnvVars[CPlacementZone]
	c.Region = requiredEnvVars[CRegion]
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.SystemDisk = CVMSystemDisk{DiskSize: 50, DiskType: TencentCloudSSD}
	c.VirtualPrivateCloud.VpcId = "vpc-ie6ri3jv"
	c.VirtualPrivateCloud.SubnetId = "subnet-5kdpieu8"
	c.SecurityGroupIds = []string{"sg-elhg6l30"}
	c.Timeout = 120000
	c.Url = CCVMUrlSiliconValley

	return c
}

func TestCreateVM(t *testing.T) {
	config := CreateCreateVMConfig()
	config.Url = CCVMUrlSingapore
	config.InstanceType = "S2.SMALL1"
	config.InstanceName = "chuacw_1"
	actualError, actualInstanceInfo := CreateVM(&config)

	// If there's an error code, and instanceid is also available, there's a logic error
	if actualError.Code != "" && actualInstanceInfo.InstanceId != "" {
		t.Fatalf("Unexpected result, error: %+v Instance Info: %+v", actualError, actualInstanceInfo)
	}

	// Attempt to allow stopping by the StopVM, by setting variables for it to use,
	// only if the environment variables are empty
	if os.Getenv(CRegion) == "" {
		os.Setenv(CRegion, actualInstanceInfo.Region)
	}
	if os.Getenv(CInstanceId) == "" {
		os.Setenv(CInstanceId, actualInstanceInfo.InstanceId)
	}

}

// Waits for the given InstanceId to reach either RUNNING or STOPPED state
// Before running this test, ensure that the InstanceId exists in the given Region
func TestWaitForVM(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore", // UPDATE THIS!!!
		CInstanceId: "ins-rbceuxvs", // UPDATE THIS!!!
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	InstanceId := requiredEnvVars[CInstanceId]
	c.Timeout = 500

	// cvm.ap-singapore.tencentcloudapi.com needs ImageId, Region, Placement.Zone
	actualError1, successful1 := WaitForVM(&c, InstanceId, "RUNNING")
	actualError2, successful2 := WaitForVM(&c, InstanceId, "STOPPED")

	// WaitForVM can return false, if the instance id isn't valid.
	// It can return true, if the instanceid is valid
	// It can return false, if there's errors
	// so the check is: if successful, and error == nil.
	// also: if not successful, then error != nil

	res1 := (!successful1 && actualError1 != nil) || (successful1 && actualError1 == nil)
	res2 := (!successful2 && actualError2 != nil) || (successful2 && actualError2 == nil)
	if !res1 || !res2 {
		t.Fatalf("WaitForVM failed to get state of VM: %s, , error1, 2: %+v, %+v", InstanceId, actualError1, actualError2)
	}
}

// In order to run this test, region and InstanceId both needs to be valid
// The given InstanceId needs to exist in the given region
// If the given InstanceId doesn't exist, this test will fail
func TestStopVM(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore", // UPDATE THIS!!!
		CInstanceId: "ins-4uwn83c2", // UPDATE THIS!!!
		CUrl:        CCVMUrlSiliconValley,
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	c.Url = requiredEnvVars[CUrl]
	InstanceId := requiredEnvVars[CInstanceId]

	cvmError, successful := StopVM(&c, InstanceId)
	if successful && cvmError.Code != "" {
		t.Fatalf("Unexpected response, Error: %+v, result: %v", cvmError, successful)
	}

}

// To run this test, ensure that the given InstanceId exists in the given region
// If an invalid InstanceId is given, the error returned from GetInstanceIP shouldn't be nil
func TestGetInstanceIP(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore", // UPDATE THIS!!!
		CInstanceId: "ins-4uwn83c2", // UPDATE THIS!!!
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	c.Url = CCVMUrlSiliconValley
	InstanceId := requiredEnvVars[CInstanceId]

	err, IPAddress := GetInstanceIP(&c, InstanceId)
	if err == nil {
		if IPAddress == "" {
			t.Fatal("An IP address is supposed to be returned when there's no error")
		}
	} else {
		if IPAddress != "" {
			t.Fatal("An IP address is not supposed to be returned when there's an error")
		}
	}
}

// This test requires a valid Region and InstanceId
// The given InstanceId needs to exist in the given Region
// If the given InstanceId doesn't exist in the given region, the error expected should either be
// InvalidInstanceId.Malformed or InvalidInstanceId.NotFound
func TestStartVM(t *testing.T) {
	// negative test
	requiredEnvVars := map[string]string{
		CRegion:     "ap-singapore",
		CInstanceId: "ins-4uwn83c2",
		CUrl:        CCVMUrlSiliconValley,
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	InstanceId := requiredEnvVars[CInstanceId]
	c.Placement = CVMPlacement{}
	c.Placement.Zone = "ap-singapore-1"
	c.InternetAccessible.InternetMaxBandwidthOut = 2
	c.InternetAccessible.PublicIpAssigned = true
	c.Url = requiredEnvVars[CUrl]

	cvmError, success := StartVM(&c, InstanceId)
	if !success {
		if cvmError.Code == "" ||
			!(cvmError.Code == "InvalidInstanceId.NotFound" ||
				cvmError.Code == "InvalidInstanceId.Malformed" ||
				cvmError.Code == "InvalidInstance.NotSupported") {
			t.Fatalf(`Expecting error to be: "InvalidInstanceId.NotFound", or "InvalidInstanceId.Malformed" or "InvalidInstance.NotSupported", but got %s`, cvmError.Code)
		}
	}
}

func TestDescribeRegions(t *testing.T) {
	requiredEnvVars := map[string]string{
		CRegion: "ap-singapore",
	}
	GetRequiredEnvVars(requiredEnvVars)
	var c Config
	c.SecretID = requiredEnvVars[CSecretId]
	c.SecretKey = requiredEnvVars[CSecretKey]
	c.Region = requiredEnvVars[CRegion]
	c.Url = CCVMUrlSiliconValley
	regions, err := DescribeRegions(&c)
	if err.Code != "" {
		t.Errorf("Error from DescribeRegions: %+v", err)
	}
	if regions.TotalCount == 0 {
		t.Fatalf("Region TotalCount is 0, region: %+v", regions)
	}
	for _, regionInfo := range regions.RegionSet {
		fmt.Printf("%s %s\n", regionInfo.RegionName, regionInfo.Region)
	}
}

func TestDescribeZones(t *testing.T) {
	Regions := []string{"na-siliconvalley", "ap-singapore"}
	for _, region := range Regions {
		requiredEnvVars := map[string]string{
			CRegion: region,
		}
		GetRequiredEnvVars(requiredEnvVars)
		var c Config
		c.SecretID = requiredEnvVars[CSecretId]
		c.SecretKey = requiredEnvVars[CSecretKey]
		c.Region = requiredEnvVars[CRegion]
		c.Url = CCVMUrlSiliconValley

		err, zoneResponse := DescribeZones(&c, c.Region)
		if err != nil {
			t.Errorf("DescribeZones failed to get a response, error: %+v", err)
		}
		if !(zoneResponse.TotalCount > 0 && zoneResponse.ZoneSet[0].Zone != "") {
			t.Fatalf("Unexpected zone response: %+v", zoneResponse)
		}
	}

}

// mockbin/request response looks like this
// "method": "GET",
// "url": "https://mockbin.org/request?Action=action&Nonce=18853&Timestamp=1525848029&Version=2017-03-12&Signature=TF8%2FmrQnBuhBPWXolsCehscX%2BmQ%3D",
// "httpVersion": "HTTP/1.1",
// "cookies": {},
// "headers": {
// "host": "mockbin.org",
// "connection": "close",
// "accept-encoding": "gzip",
// "x-forwarded-for": "2401:7400:c800:6940:e0ce:f418:4f7f:6ef3, 172.68.146.229",
// "cf-ray": "41822cf2cff1309c-SIN",
// "x-forwarded-proto": "http",
// "cf-visitor": "{\"scheme\":\"https\"}",
// "user-agent": "Go-http-client/2.0",
// "cf-connecting-ip": "2401:7400:c800:6940:e0ce:f418:4f7f:6ef3",
// "x-request-id": "f850fa70-adfd-4ccb-bfe3-35d9aeec0b07",
// "x-forwarded-port": "80",
// "via": "1.1 vegur",
// "connect-time": "0",
// "x-request-start": "1525848036650",
// "total-route-time": "0"
// },
// "queryString": {
// "Action": "action",
// "Nonce": "18853",
// "Timestamp": "1525848029",
// "Version": "2017-03-12",
// "Signature": "TF8/mrQnBuhBPWXolsCehscX+mQ="
// },
// "postData": {
// "mimeType": "application/octet-stream",
// "text": "",
// "params": []
// },
// "headersSize": 625,
// "bodySize": 0
// }
func TestCloudAPICall(t *testing.T) {
	configInfo := map[string]interface{}{}
	extraParams := map[string]string{
		"url": "mockbin.org/request",
	}
	myMap := make(map[string]interface{})
	byteData := CloudAPICall("action", configInfo, extraParams)
	json.Unmarshal(byteData, &myMap)
	if myMap["method"] != "GET" &&
		myMap["queryString"].(map[string]interface{})["Action"] != "action" &&
		myMap["queryString"].(map[string]interface{})["Version"] != TencentAPIVersion {
		t.Fatalf("CloudAPICall failed, response: %+v", string(byteData))
	}
}

func TestCVMAPICall(t *testing.T) {
	configInfo := map[string]interface{}{}
	extraParams := map[string]string{
		"url": "mockbin.org/request",
	}
	myMap := make(map[string]interface{})
	byteData := CVMAPICall("XyZaction", configInfo, extraParams)
	json.Unmarshal(byteData, &myMap)
	if myMap["method"] != "GET" &&
		myMap["queryString"].(map[string]interface{})["Action"] != "XyZaction" &&
		myMap["queryString"].(map[string]interface{})["Version"] != TencentAPIVersion {
		t.Fatalf("CVMAPICall failed, response: %+v", string(byteData))
	}
}

func TestAcctAPICall(t *testing.T) {
	configInfo := map[string]interface{}{}
	extraParams := map[string]string{
		"url": "mockbin.org/request",
	}
	myMap := make(map[string]interface{})
	byteData := AcctAPICall("aaction", configInfo, extraParams)
	json.Unmarshal(byteData, &myMap)
	if myMap["method"] != "GET" &&
		myMap["queryString"].(map[string]interface{})["Action"] != "aaction" &&
		myMap["queryString"].(map[string]interface{})["Version"] != TencentAPIVersion {
		t.Fatalf("AcctAPICall failed, response: %+v", string(byteData))
	}
}

func TestSignaturePost(t *testing.T) {
	signature := SignaturePost("myURL", "myParams", "mySecretKey")
	if signature != "ruRnQBVvH6zNaA/1+tEeqQ42DNA=" {
		t.Fatalf("Unexpected response signature %s", signature)
	}
}

func TestRequestPost(t *testing.T) {
	type args struct {
		requestURL         string
		requestParamString string
		signature          string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"RequestPost test case 1", args{"mockbin.org/echo", "MyParam", "MySig"},
			[]byte("MyParam&Signature=MySig"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RequestPost(tt.args.requestURL, tt.args.requestParamString, tt.args.signature)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequestPost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RequestPost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInquiryPriceRunInstances(t *testing.T) {
	region := "na-siliconvalley"
	requiredEnvVars := map[string]string{
		CRegion:        region,
		CPlacementZone: "na-siliconvalley-1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	c := &Config{
		Region:    region,
		Url:       CCVMUrlSiliconValley,
		SecretID:  requiredEnvVars[CSecretId],
		SecretKey: requiredEnvVars[CSecretKey],
		ImageID:   "img-pyqx34y1",
		Placement: CVMPlacement{Zone: requiredEnvVars[CPlacementZone]},
	}
	InquiryPriceRunInstances(c)
}

func TestDescribeInstanceFamilyConfigs(t *testing.T) {
	region := "na-siliconvalley"
	requiredEnvVars := map[string]string{
		CRegion:        region,
		CPlacementZone: "na-siliconvalley-1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	c := &Config{
		Region:    region,
		Url:       CCVMUrlSiliconValley,
		SecretID:  requiredEnvVars[CSecretId],
		SecretKey: requiredEnvVars[CSecretKey],
		Placement: CVMPlacement{Zone: requiredEnvVars[CPlacementZone]},
	}
	DescribeInstanceFamilyConfigs(c)
}

func TestCreateCustomImage(t *testing.T) {
	region := "ap-singapore"
	requiredEnvVars := map[string]string{
		CRegion:        region,
		CPlacementZone: "ap-singapore-1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	c := &Config{
		ImageName: fmt.Sprintf("%s", SSHTimeStampSuffix()),
		Placement: CVMPlacement{Zone: requiredEnvVars[CPlacementZone]},
		Region:    region,
		SecretID:  requiredEnvVars[CSecretId],
		SecretKey: requiredEnvVars[CSecretKey],
		Timeout:   300000,
		Url:       CCVMUrlSingapore,
	}
	cvmError, cvmCreateCustomImage := CreateCustomImage(c, "ins-4uwn83c2")
	if cvmError.Code != "" {
		log.Printf("Error code: %s, message: %s", cvmError.Code, cvmError.Message)
	}
	log.Printf("RequestId: %s", cvmCreateCustomImage.RequestId)
	c.Url = CImageAPIUrl
	ImageFound, instanceId := WaitForCustomImageReady(c)
	log.Printf("Image found: %v, id: %s", ImageFound, instanceId)
	if !ImageFound {
		t.Fatal("Failed to find image after creating it")
	}
}

func TestDescribeAddresses(t *testing.T) {
	region := "ap-singapore"
	requiredEnvVars := map[string]string{
		CRegion:        region,
		CPlacementZone: "ap-singapore-1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	c := &Config{
		Placement: CVMPlacement{Zone: requiredEnvVars[CPlacementZone]},
		Region:    region,
		SecretID:  "xxxxx", //requiredEnvVars[CSecretId],
		SecretKey: "yyyyy", //requiredEnvVars[CSecretKey],
		Timeout:   300000,
		Url:       CCVMUrlSiliconValley,
	}
	DescribeAddresses(c)
	c.Url = "cvm.tencentcloudapi.com/"
	DescribeAddresses(c)
	c.Url = CCVMUrlSiliconValley
	DescribeAddresses(c)
}

func TestDeleteAddress(t *testing.T) {
	region := "ap-singapore"
	requiredEnvVars := map[string]string{
		CRegion:        region,
		CPlacementZone: "ap-singapore-1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	c := &Config{
		Placement: CVMPlacement{Zone: requiredEnvVars[CPlacementZone]},
		Region:    region,
		SecretID:  requiredEnvVars[CSecretId],
		SecretKey: requiredEnvVars[CSecretKey],
		Timeout:   300000,
		Url:       "cvm.tencentcloudapi.com/",
	}
	DeleteAddress(c, "eip-aelj65xi")
}

func TestQueryCustomImage(t *testing.T) {
	region := "ap-singapore"
	// region := "na-siliconvalley"
	requiredEnvVars := map[string]string{
		CRegion:        region,
		CPlacementZone: "ap-singapore-1",
		// CPlacementZone: "na-siliconvalley-1",
	}
	GetRequiredEnvVars(requiredEnvVars)
	c := &Config{
		Placement: CVMPlacement{Zone: requiredEnvVars[CPlacementZone]},
		Region:    region,
		SecretID:  requiredEnvVars[CSecretId],
		SecretKey: requiredEnvVars[CSecretKey],
		Timeout:   300000,
		Url:       CCVMUrlSiliconValley,
	}
	QueryCustomImage(c)
}

func TestJsonDecode(t *testing.T) {
	jsonStr := `{"Response":{"Error":{"Code":"AuthFailure.SignatureFailure","Message":"The provided credentials could not be validated. Please check your signature is correct."},"RequestId":"600e7a43-130a-4eba-8942-487fff1f4f86"}}`
	response := []byte(jsonStr)

	var jsonresp struct {
		Response struct {
			Error struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
			RequestId string `json:"RequestId"`
		} `json:"Response"`
	}
	err := json.Unmarshal(response, &jsonresp)
	fmt.Printf("json response: %+v", jsonresp)
	fmt.Printf("Error: %v", err)

}

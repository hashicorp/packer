package tencent

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	//  "crypto/sha1"
	//  "encoding/hex"
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/template/interpolate"
)

type (
	// CloudAPICallResponse is the response returned by a call to CloudAPICall
	CloudAPICallResponse struct {
		Response map[string]interface{} `json:"Response"`
	}

	// CVMPlacement represents the structure required for creating a VM.
	CVMPlacement struct {
		Zone      string `mapstructure:"Zone" json:"Zone"`
		ProjectId int64  `mapstructure:"ProjectId" json:"ProjectId"`
		// HostIds vs HostId
		HostIds []string `mapstructure:"HostId" json:"HostId"` /// ***!!! Error on Tencent's part??? *** !!!
	}

	// CVMError example
	// "Code": "AuthFailure.SecretIdNotFound",
	// "Message": "The SecretId is not found, please ensure that your SecretId is correct."
	CVMError struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}

	// CVMErrorResponse example
	// "Error": {
	// 	"Code": "AuthFailure.SecretIdNotFound",
	// 	"Message": "The SecretId is not found, please ensure that your SecretId is correct."
	// },
	// "RequestId": "0f2bdb80-74dc-418b-b9e0-16e8c98288ed"
	CVMErrorResponse struct {
		Error     CVMError
		RequestId string
	}

	CVMCreateCustomImage struct {
		RequestId string
	}

	CVMImages struct {
		Architecture       string
		CreatedTime        string
		ImageCreator       string
		ImageDescription   string
		ImageId            string
		ImageName          string
		ImageSize          int64
		ImageSource        string
		ImageState         string
		ImageType          string
		IsSupportCloudInit bool
		OsName             string
		Platform           string
	}

	CVMDescribeImages struct {
		ImageSet []CVMImages `json:"ImageSet"`
	}

	CVMDescribeImagesResponse struct {
		ImageSet   []CVMImages `json:"ImageSet"`
		RequestId  string      `json:"RequestId"`
		TotalCount int64       `json:"TotalCount"`
	}

	CVMSharePermissions struct {
		CreatedTime string
		Account     string
	}

	CVMSharePermissionSet struct {
		SharePermissionSet []CVMSharePermissions
	}

	// CVMCreateInstanceResult represents the structure for the result returned from creating a VM
	CVMCreateInstanceResponse struct {
		InstanceIdSet []string
		RequestId     string
		Error         CVMError
	}

	CVMVirtualPrivateCloud struct {
		VpcId              string
		SubnetId           string
		AsVpcGateway       bool
		PrivateIpAddresses []string // no longer present as of 2 Jun 2018
	}

	CVMSystemDisk struct {
		DiskType string
		DiskId   string
		DiskSize int64
	}

	CVMDataDisk struct {
		DiskType string
		DiskId   string
		DiskSize int64
	}

	CVMInternetAccessible struct {
		InternetChargeType      string `mapstructure:"InternetChargeType"`
		InternetMaxBandwidthOut int64  `mapstructure:"InternetMaxBandwidthOut"`
		PublicIpAssigned        bool   `mapstructure:"PublicIpAssigned"`
	}

	CVMInstanceSet struct {
		Placement              CVMPlacement           `mapstructure:"Placement"`
		InstanceId             string                 `mapstructure:"InstanceId"`
		InstanceState          string                 `mapstructure:"InstanceState"`
		RestrictState          string                 `mapstructure:"RestrictState"`
		InstanceType           string                 `mapstructure:"InstanceType"`
		CPU                    int64                  `mapstructure:"CPU"`
		Memory                 int64                  `mapstructure:"Memory"`
		InstanceName           string                 `mapstructure:"InstanceName"`
		InstanceChargeType     string                 `mapstructure:"InstanceChargeType"`
		SystemDisk             CVMSystemDisk          `mapstructure:"SystemDisk"`
		DataDisks              []CVMDataDisk          `mapstructure:"DataDisks"`
		PrivateIpAddresses     []string               `mapstructure:"PrivateIpAddresses"`
		PublicIpAddresses      []string               `mapstructure:"PublicIpAddresses"`
		InternetAccessible     CVMInternetAccessible  `mapstructure:"InternetAccessible"`
		VirtualPrivateCloud    CVMVirtualPrivateCloud `mapstructure:"VirtualPrivateCloud"`
		SecurityGroupIds       []string               `mapstructure:"SecurityGroupIds"`
		LoginSettings          CVMLoginSettings       `mapstructure:"LoginSettings"`
		ImageId                string                 `mapstructure:"ImageId"`
		OsName                 string                 `mapstructure:"OsName"`
		RenewFlag              string                 `mapstructure:"RenewFlag"`
		CreatedTime            string                 `mapstructure:"CreatedTime"`
		ExpiredTime            string                 `mapstructure:"ExpiredTime"`
		Tags                   []string               `mapstructure:"Tags"`
		DisasterRecoverGroupId string                 `mapstructure:"DisasterRecoverGroupId"`
	}

	CVMDescribeInstancesResponse struct {
		TotalCount  int64
		InstanceSet []CVMInstanceSet
		Error       CVMError
		RequestId   string
	}

	CVMLoginSettings struct {
		Password       string // no longer present as of 2 Jun 2018
		KeyIds         []string
		KeepImageLogin string // no longer present as of 2 Jun 2018
	}

	RunSecurityServiceEnabled struct {
		Enabled bool
	}

	RunMonitorServiceEnabled struct {
		Enabled bool
	}

	CVMEnhancedService struct {
		SecurityService RunSecurityServiceEnabled
		MonitorService  RunMonitorServiceEnabled
	}

	CVMInstanceChargePrepaid struct {
		Period    int
		RenewFlag string
	}

	CVMInstanceInfo struct {
		InstanceId string
		Region     string
	}

	CVMRunInstancesResponse struct {
		InstanceIdSet []string

		Error     CVMError
		RequestId string
	}

	CVMInstanceStatusSet struct {
		InstanceId    string
		InstanceState string // RUNNING
	}

	CVMDescribeInstancesStatusResponse struct {
		TotalCount        int
		InstanceStatusSet []CVMInstanceStatusSet
		RequestId         string
	}

	CVMSimplePrice struct {
		UnitPrice  float64
		ChargeUnit string
	}

	CVMInstanceFamilyConfig struct {
		InstanceFamilyName string
		InstanceFamily     string
	}

	CVMInstanceFamilyConfigResponse struct {
		RequestId string
	}

	CVMItemPrice struct {
		Discount                    int
		UnitPriceDiscount           float64
		UnitPriceDiscountSecondStep float64
		UnitPriceDiscountThirdStep  float64
		UnitPriceSecondStep         float64
		UnitPriceThirdStep          float64
		UnitPrice                   float64
		ChargeUnit                  string
	}

	CVMPrice struct {
		InstancePrice  CVMItemPrice   `json:"InstancePrice"`
		BandwidthPrice CVMSimplePrice `json:"BandwidthPrice"`
	}

	CVMInquiryPriceRunInstancesResponse struct {
		Price     CVMPrice
		RequestId string
	}

	CVMStartInstancesResponse struct {
		Error     CVMError
		TaskId    string
		RequestId string
	}

	CVMStopInstancesResponse = CVMStartInstancesResponse

	CVMZoneSet struct {
		Zone      string
		ZoneName  string
		ZoneId    string
		ZoneState string
	}

	CVMZoneResponse struct {
		TotalCount int
		ZoneSet    []CVMZoneSet
		RequestId  string
	}

	CVMRegionSet struct {
		Region      string
		RegionName  string
		RegionState string
	}

	CVMRegions struct {
		TotalCount int
		RegionSet  []CVMRegionSet
		RequestId  string
	}

	WhereIsInstance struct {
		Region string
		Zone   string
	}
)

var (
	CloudAPIDebug       bool
	CloudProviderPrefix = "cvm.tencentcloudapi.com/"
)

func CreateCustomImage(c *Config, instanceId string) (CVMError, CVMCreateCustomImage) {
	extraParams := c.CreateCustomImageExtraParams(instanceId)
	configInfo := c.CreateCustomImageMap()
	response := CVMAPICall2("CreateImage", configInfo, extraParams)
	var (
		cvmError             CVMError
		cvmErrorResponse     CVMErrorResponse
		cvmCreateCustomImage CVMCreateCustomImage
		jsonresp             struct {
			Response struct {
				RequestId string   `json:"RequestId"`
				Error     CVMError `json:"Error"`
			} `json:"Response"`
		}
	)
	err := json.Unmarshal(response, &jsonresp)
	if err != nil {
		if c.PackerDebug || CloudAPIDebug {
			log.Printf("CloudAPICall response error is: %+v\n", cvmErrorResponse.Error)
		}
	}
	cvmCreateCustomImage.RequestId = jsonresp.Response.RequestId
	cvmError = jsonresp.Response.Error

	return cvmError, cvmCreateCustomImage
}

func QueryCustomImage(c *Config) (bool, CVMDescribeImagesResponse) {
	extraParams := c.CreateQueryCustomImageExtraParams()
	configInfo := c.CreateQueryImageMap()
	response := CVMAPICall2("DescribeImages", configInfo, extraParams)
	var (
		success             bool
		cvmQueryCustomImage CVMDescribeImagesResponse
		cvmDescribeImages   struct {
			Response struct {
				Error CVMError `json:"Error"`
				CVMDescribeImagesResponse
			} `json:"Response"`
		}
	)
	err := json.Unmarshal(response, &cvmDescribeImages)
	if err != nil || response == nil {
		if c.PackerDebug || CloudAPIDebug {
			log.Printf("CloudAPICall response error is: %+v\n", cvmDescribeImages.Response.Error)
		}
		success = false
	} else {
		cvmQueryCustomImage = cvmDescribeImages.Response.CVMDescribeImagesResponse
		success = true
	}
	return success, cvmQueryCustomImage
}

// WaitForCustomImageReady waits for a custom image to be ready
// After an image is created from an instance, it will not be available until
// after DescribeImages returns it, so this method calls QueryCustomImage aka DescribeImages and
// looks in the response for the ImageName that was used to create it.
// Requires Config.ImageName to be set as the name to look at
func WaitForCustomImageReady(c *Config) (bool, string) {
	startTime := time.Now()
	endTime := startTime.Add(time.Millisecond * time.Duration(c.Timeout))
	var (
		ImageFound          bool
		ok                  bool
		cvmQueryCustomImage CVMDescribeImagesResponse
	)
	ImageFound = false
	imageID := ""
	for time.Now().Before(endTime) {
		if c.PackerDebug || CloudAPIDebug {
			log.Println("Querying image status")
		}
		ok, cvmQueryCustomImage = QueryCustomImage(c)
		if !ok {
			if c.PackerDebug || CloudAPIDebug {
				log.Println("Probably failed to query image status")
			}
		} else {
			for _, image := range cvmQueryCustomImage.ImageSet {
				if image.ImageName == c.ImageName {
					ImageFound = true
					imageID = image.ImageId
					if c.PackerDebug || CloudAPIDebug {
						log.Printf("New custom image ID is: %s", imageID)
					}
					break
				}
			}
		}
		if ImageFound {
			break
		}
		time.Sleep(time.Second * 5)
	}
	return ImageFound, imageID
}

// The request should look like as follows.
// see https://cloud.tencent.com/document/api/213/9384#example-2
// https://cvm.api.qcloud.com/v2/index.php?Action=RunInstances
// &Version=2017-03-12
// &Placement.Zone=ap-guangzhou-2
// &InstanceChargeType=PREPAID
// &InstanceChargePrepaid.Period=1
// &InstanceChargePrepaid.RenewFlag=NOTIFY_AND_AUTO_RENEW
// &ImageId=img-pmqg1cw7
// &InstanceType=S1.SMALL1
// &SystemDisk.DiskType=LOCAL_BASIC
// &SystemDisk.DiskSize=50
// &DataDisks.0.DiskType=LOCAL_BASIC
// &DataDisks.0.DiskSize=100
// &InternetAccessible.InternetChargeType=TRAFFIC_POSTPAID_BY_HOUR
// &InternetAccessible.InternetMaxBandwidthOut=10
// &InternetAccessible.PublicIpAssigned=TRUE
// &InstanceName=QCLOUD-TEST
// &LoginSettings.Password=Qcloud@TestApi123++
// &EnhancedService.SecurityService.Enabled=TRUE
// &EnhancedService.MonitorService.Enabled=TRUE
// &InstanceCount=1
// &<Common request parameters>
// All parameters must be sorted: https://cloud.tencent.com/document/api/213/11652#2.1.-sort-parameters
func CreateVM(c *Config) (CVMError, CVMInstanceInfo) {
	extraParams := c.CreateVMExtraParams()
	configInfo := c.CreateVMmap()
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("CreateVM configInfo: %+v\n", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("CreateVM extraParams: %+v\n", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	response := CVMAPICall2("RunInstances", configInfo, extraParams)
	var (
		cvmErrorResponse CVMErrorResponse
		instanceInfo     CVMInstanceInfo
	)
	var runInstances struct {
		Response struct {
			Error         CVMError `json:"Error"`
			InstanceIdSet []string `json:"InstanceIdSet"`
			RequestId     string   `json:"RequestId"`
		} `json:"Response"`
	}
	err := json.Unmarshal(response, &runInstances)
	instanceid := ""
	if err != nil || runInstances.Response.Error.Code != "" {
		cvmErrorResponse.Error = runInstances.Response.Error
		if c.PackerDebug || CloudAPIDebug {
			log.Printf("CVMAPICall2 response error code: %s message: %s", runInstances.Response.Error.Code,
				runInstances.Response.Error.Message)
		}
	} else {
		log.Printf("CVMAPICall2 successful: %+v\n", runInstances.Response)
		if len(runInstances.Response.InstanceIdSet) > 0 {
			instanceid = runInstances.Response.InstanceIdSet[0]
		}
		instanceInfo = CVMInstanceInfo{instanceid, configInfo[CRegion].(string)}
	}

	return cvmErrorResponse.Error, instanceInfo
}

func DescribeAddresses(c *Config) {
	extraParams := c.CreateDescribeAddressesExtraParams()
	configInfo := c.CreateDescribeAddressesMap()
	response := CVMAPICall2("DescribeAddresses", configInfo, extraParams)
	log.Printf("DescribeAddresses response\n%v", string(response))
}

func DeleteAddress(c *Config, AddressId string) {
	extraParams := c.CreateReleaseAddressesExtraParams(AddressId)
	configInfo := c.CreateReleaseAddressesMap()
	response := CVMAPICall2("ReleaseAddresses", configInfo, extraParams)
	log.Printf("ReleaseAddresses response\n%v", string(response))
}

func DescribeRegions(c *Config) (CVMRegions, CVMError) {
	extraParams := c.CreateDescribeRegionsExtraParams()
	configInfo := c.CreateDescribeRegionsMap()
	response := CVMAPICall2("DescribeRegions", configInfo, extraParams)
	var (
		jsonresp struct {
			Response struct {
				CVMRegions
				Error CVMError `json:"Error"`
			}
		}
	)
	json.Unmarshal(response, &jsonresp)
	return jsonresp.Response.CVMRegions, jsonresp.Response.Error
}

func DescribeInstanceFamilyConfigs(c *Config) {
	extraParams := c.CreateDescribeInstanceFamilyConfigsExtraParams()
	configInfo := c.CreateDescribeInstanceFamilyConfigsMap()
	response := CVMAPICall2("DescribeInstanceFamilyConfigs", configInfo, extraParams)
	var (
		instanceFamilyConfig CVMInstanceFamilyConfig
		cvmError             CVMErrorResponse
		jsonresp             struct {
		}
	)
	json.Unmarshal(response, &jsonresp)
	err := DecodeResponse(response, &instanceFamilyConfig)
	if err != nil {
		DecodeResponse(response, &cvmError)
	}
}

func DescribeZones(c *Config, region string) (error, *CVMZoneResponse) {
	configInfo := c.CreateDescribeZonesMap()
	extraParams := c.CreateDescribeZonesExtraParams()
	response := CVMAPICall2("DescribeZones", configInfo, extraParams)
	var (
		zoneInfo CVMZoneResponse
		cvmError CVMErrorResponse
		jsonresp struct {
		}
	)
	json.Unmarshal(response, &jsonresp)
	err := DecodeResponse(response, &zoneInfo)
	if err != nil {
		DecodeResponse(response, &cvmError)
	}
	return err, &zoneInfo
}

var (
	whereIsInstance map[string]*WhereIsInstance
	regions         []string
)

// StartVM starts a VM, and returns an error, and whether the VM started successfully.
// If the VM doesn't exist, the result is always an error and a bool of false.
// If the VM exists, the error depends on whether the VM started or not.
func StartVM(c *Config, instanceId string) (CVMError, bool) {
	extraParams := c.CreateStartInstanceExtraParams(instanceId)
	configInfo := c.CreateStartInstanceMap()
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("StartVM configInfo: %+v", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("StartVM extraParams: %+v", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	response := CVMAPICall2("StartInstances", configInfo, extraParams)
	var cvmStartInstances struct {
		Response struct {
			Error     CVMError `json:"Error"`
			TaskId    string   `json:"TaskId"`
			RequestId string   `json:"RequestId"`
		} `json:"Response"`
	}
	err := json.Unmarshal(response, &cvmStartInstances)
	if err != nil || cvmStartInstances.Response.Error.Code != "" {
		return cvmStartInstances.Response.Error, false
	}
	return cvmStartInstances.Response.Error, true // empty record, which has no error, and bool of true, indicating success
}

// StopVM stops a VM, and returns an error, whether the VM stopped successfully.
// If VM doesn't exist, the result is always an error, and a bool of false.
// If the VM is already stopped, calling this produces an error about request not being supported.
func StopVM(c *Config, instanceId string) (CVMError, bool) {
	extraParams := map[string]string{
		"InstanceIds.0": instanceId,
	}
	if c.Url != "" {
		extraParams[CUrl] = c.Url
	}
	configInfo := c.CreateStopVMMap()
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("StopVM configInfo: %+v", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("StopVM extraParams: %+v", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	response := CVMAPICall2("StopInstances", configInfo, extraParams)
	// failed response
	// {
	// 	"Response": {
	// 		"Error": {
	// 			"Code": "InvalidInstance.NotSupported",
	// 			"Message": "The request does not support the instances `ins-4uwn83c2` which are in operation or in a special state."
	// 		},
	// 		"RequestId": "109b5571-7696-4f5f-9796-fea9750a6af3"
	// 	}
	// }
	var (
		jsonresp struct {
			Response struct {
				Error     CVMError `json:"Error"`
				TaskId    string   `json:"TaskId"`
				RequestId string   `json:"RequestId"`
			} `json:"Response"`
		}
	)
	err := json.Unmarshal(response, &jsonresp)
	if c.PackerDebug || CloudAPIDebug {
		log.Printf("StopVM response\n%+v", jsonresp)
	}
	if err != nil || jsonresp.Response.Error.Code != "" {
		return jsonresp.Response.Error, false
	}
	return jsonresp.Response.Error, true // empty record, which has no error, and bool of true, indicating success

}

// GetInstanceIP returns the IP address for the given instance, if it has a public IP address.
// Returns an error if the instanceId doesn't exist, or there is no public IP address.
// A shutdown instance does not have an IP address, only running instances.
func GetInstanceIP(c *Config, instanceId string) (error, string) {
	extraParams := c.CreateGetInstanceIPExtraParams(instanceId)
	// configInfo := c.CreateVMmap()
	configInfo := c.CreateGetInstanceIPmap(instanceId)
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("GetInstanceIP configInfo: %+v", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("GetInstanceIP extraParams: %+v", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	response := CVMAPICall2("DescribeInstances", configInfo, extraParams)
	var (
		IPAddress            string
		cvmDescribeInstances struct {
			Response struct {
				TotalCount  int              `json:"TotalCount"`
				RequestId   string           `json:"RequestId"`
				InstanceSet []CVMInstanceSet `json:"InstanceSet"`
				Error       CVMError
			} `json:"Response"`
		}
	)
	err := json.Unmarshal(response, &cvmDescribeInstances)
	if err != nil {
		log.Printf("Error encountered: %v", err)
		return err, ""
	}
	if cvmDescribeInstances.Response.TotalCount == 0 {
		if cvmDescribeInstances.Response.Error.Code != "" {
			err = errors.New(fmt.Sprintf("Code: %s, Message: %s", cvmDescribeInstances.Response.Error.Code,
				cvmDescribeInstances.Response.Error.Message))
		}
		// If there's no error, and TotalCount is 0, the named InstanceId doesn't exist
		return err, IPAddress // return invalid if there's no such instanceId
	}
	// Even when the instance is found, it may not have an IP address
	if len(cvmDescribeInstances.Response.InstanceSet[0].PublicIpAddresses) > 0 {
		IPAddress = cvmDescribeInstances.Response.InstanceSet[0].PublicIpAddresses[0]
	}
	if IPAddress == "" {
		err = errors.New(fmt.Sprintf("No public IP address"))
		return err, IPAddress
	}
	if c.PackerDebug || CloudAPIDebug {
		log.Printf("GetInstanceIP IP address is: %s", IPAddress)
	}
	return nil, IPAddress
}

// Checks the state of an instance.
// If the instance doesn't exist, returns an error, and INVALID as the state,
// otherwise, returns nil as the error, and the instance state.
func GetInstanceState(c *Config, instanceId string) (error, string) {
	extraParams := map[string]string{
		"InstanceIds.0": instanceId,
		CRegion:         c.Region,
	}
	extraParams[CUrl] = c.Url
	configInfo := c.CreateVMmap()
	for k, _ := range configInfo {
		switch k {
		case CPackerDebug, CRegion, CSecretId, CSecretKey, CTimestamp, CVersion:
			continue
		default:
			delete(configInfo, k)
		}
	}
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("GetInstanceState configInfo: %+v", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("GetInstanceState extraParams: %+v", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	response := CVMAPICall2("DescribeInstancesStatus", configInfo, extraParams)
	var (
		cvmDescribeInstancesStatus struct {
			Response struct {
				TotalCount        int                    `json:"TotalCount"`
				InstanceStatusSet []CVMInstanceStatusSet `json:"InstanceStatusSet"`
				RequestId         string                 `json:"RequestId"`
				Error             CVMError               `json:"Error"`
			}
		}
	)
	err := json.Unmarshal(response, &cvmDescribeInstancesStatus)
	if err != nil {
		if c.PackerDebug || CloudAPIDebug {
			log.Printf("GetInstanceState error: %+v", err)
		}
		return err, "INVALID"
	}
	if cvmDescribeInstancesStatus.Response.TotalCount == 0 {
		var errMsg string
		if cvmDescribeInstancesStatus.Response.Error.Code != "" {
			errMsg = fmt.Sprintf("Code: %s, Message: %s", cvmDescribeInstancesStatus.Response.Error.Code,
				cvmDescribeInstancesStatus.Response.Error.Message)
		} else {
			errMsg = fmt.Sprintf("No such instance, InstanceId: %v", instanceId)
		}
		err = errors.New(errMsg)
		// If there's no error, and TotalCount is 0, the named InstanceId doesn't exist
		return err, "INVALID" // return invalid if there's no such instanceId
	}
	instanceState := cvmDescribeInstancesStatus.Response.InstanceStatusSet[0].InstanceState
	return nil, instanceState
}

func InquiryPriceRunInstances(c *Config) {
	extraParams := c.InquiryPriceRunParams()
	configInfo := c.CreateInquiryPriceRunInstancesMap()
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("InquiryPriceRunInstances configInfo: %+v", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("InquiryPriceRunInstances extraParams: %+v", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	response := CVMAPICall("InquiryPriceRunInstances", configInfo, extraParams)
	var (
		cvmInquiryPriceRunInstancesResponse CVMInquiryPriceRunInstancesResponse
		cvmError                            CVMErrorResponse
		jsonresp                            struct {
		}
	)
	log.Printf("InquiryPriceRunInstances response\n%v", string(response))
	json.Unmarshal(response, &jsonresp)
	err := DecodeResponse(response, &cvmInquiryPriceRunInstancesResponse)
	if err != nil {
		if c.PackerDebug || CloudAPIDebug {
			log.Printf("GetInstanceState error: %+v", err)
		}
		err = DecodeResponse(response, &cvmError)
	}
}

// WaitForVM waits for an instance to reach the state given in stateToWaitFor
// If the desired state is gotten, returns true
// If timeout, or the desired state is not gotten, returns false
func WaitForVM(c *Config, instanceId, stateToWaitFor string) (error, bool) {
	startTime := time.Now()
	endTime := startTime.Add(time.Millisecond * time.Duration(c.Timeout))
	var err error
	// brk := false
	for time.Now().Before(endTime) {
		var instanceState string
		err, instanceState = GetInstanceState(c, instanceId)
		switch instanceState {
		case stateToWaitFor:
			{
				return nil, true
			}
		case "INVALID":
			{
				return err, false // no need to wait anymore
			}
		}
		time.Sleep(time.Second * 5)
	}
	err = errors.New("WaitForVM timed out! Increase timeout setting.")
	return err, false
}

func DecodeResponseMap(data []byte, target *map[string]interface{}) error {
	err := json.Unmarshal(data, target)
	return err
}

func DecodeResponse(data []byte, target interface{}) error {
	var decodedResponse CloudAPICallResponse
	err := json.Unmarshal(data, &decodedResponse)
	if err != nil {
		return err
	}
	if CloudAPIDebug {
		log.Printf("DecodeResponse raw data: %s", string(data))
		log.Printf("DecodeResponse unmarshalled: %+v", decodedResponse)
	}

	err = config.Decode(target, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, decodedResponse.Response)

	if CloudAPIDebug {
		if err != nil {
			log.Printf("DecodeResponse Error in calling config.Decode: %+v", err)
		} else {
			log.Printf("DecodeResponse unmarshalled response: %+v", decodedResponse.Response)
		}
	}
	return err
}

// CloudAPICall makes a call to the CloudAPI
// To call the CloudAPICall function, provide it the name
// of an action, eg, RunInstances, DescribeZones, DescribeRegions
// a map[string]interface{} containing the required parameters
// and another parameter containing
func CloudAPICall(action string, config map[string]interface{},
	extraParams map[string]string) []byte {
	var (
		c    *Config
		err1 error
	)
	c, _, err1 = NewSimpleConfig(config)
	if err1 != nil {
		log.Printf("CloudAPICall Error in NewSimpleConfig: %+v\n", err1)
	}

	var urlprefix string
	if extraParams[CUrl] != "" {
		urlprefix = extraParams[CUrl]
		delete(extraParams, CUrl)
	} else {
		urlprefix = CloudProviderPrefix
	}

	extraParams1 := c.Keys()
	// merges the keys and values into extraParams
	// if extraParams2 is nil, this is skipped
	for k, v := range extraParams {
		extraParams1[k] = v
	}

	secretID := c.SecretID
	secretKey := c.SecretKey
	signaturestring := SignatureString(action, secretID, extraParams1)
	if c.PackerDebug || CloudAPIDebug {
		log.Printf("CloudAPICall url: %s\n", urlprefix)
		log.Printf("CloudAPICall URI: %s\n", signaturestring)
	}

	signature := SignatureGet(urlprefix, signaturestring, secretKey)
	response, err2 := RequestGet(urlprefix, signaturestring, signature)
	if c.PackerDebug || CloudAPIDebug {
		log.Printf("CloudAPICall RequestGet response: %s", string(response))
		if err2 != nil {
			log.Printf("CloudAPICall RequestGet error: %+v", err2)
			return nil
		}
	}
	return response
}

func CVMAPICall(action string, config map[string]interface{}, extraParams map[string]string) []byte {
	var newExtraParams map[string]string
	if extraParams == nil {
		newExtraParams = make(map[string]string)
	} else {
		newExtraParams = extraParams
	}
	if newExtraParams[CUrl] == "" {
		newExtraParams[CUrl] = CCVMUrlSingapore
	}
	return CloudAPICall(action, config, newExtraParams)
}

// CloudAPICall2  is a more efficient version than CloudAPICall
func CloudAPICall2(action string, configInfo map[string]interface{},
	extraParams map[string]string) []byte {

	var urlprefix string
	if extraParams[CUrl] != "" {
		urlprefix = extraParams[CUrl]
		delete(extraParams, CUrl)
	} else {
		urlprefix = CloudProviderPrefix
	}

	var (
		PackerDebugIntf interface{}
		PackerDebug, ok bool
	)
	PackerDebugIntf, ok = configInfo[CPackerDebug]
	if ok {
		PackerDebug = PackerDebugIntf.(bool)
	}
	secretID := configInfo[CSecretId].(string)
	secretKey := configInfo[CSecretKey].(string)
	delete(configInfo, CSecretKey)
	delete(configInfo, CSecretId)
	delete(configInfo, CPackerDebug)

	for k, v := range configInfo {
		extraParams[k] = v.(string)
	}

	signaturestring := SignatureString(action, secretID, extraParams)
	if PackerDebug || CloudAPIDebug {
		log.Printf("CloudAPICall2 url: %s\n", urlprefix)
		log.Printf("CloudAPICall2 URI: %s\n", signaturestring)
	}

	signature := SignatureGet(urlprefix, signaturestring, secretKey)
	response, err2 := RequestGet(urlprefix, signaturestring, signature)
	if PackerDebug || CloudAPIDebug {
		// Shows the response on a new line
		log.Printf("CloudAPICall2 action: %s, response\n%s", action, string(response))
		if err2 != nil {
			log.Println("!!!  UNEXPECTED ERROR  !!!")
			log.Printf("CloudAPICall2 RequestGet error: %+v", err2)
			response := []byte(fmt.Sprintf(`
			{
			"Response": {
					"Error": {
						"Code": "UNKNOWN",
						"Message": "%v"
					},
					"RequestId": "INTERNAL"
				}
			}	
			`, err2))
			return response
		}
	}
	return response
}

// CVMAPICall2 is a more efficient version than CVMAPICall, as it calls CloudAPICall2 instead
// which doesn't use config.decodeOpts
func CVMAPICall2(action string, configInfo map[string]interface{}, extraParams map[string]string) []byte {
	var newExtraParams map[string]string
	if extraParams == nil {
		newExtraParams = make(map[string]string)
	} else {
		newExtraParams = extraParams
	}
	if newExtraParams[CUrl] == "" {
		newExtraParams[CUrl] = CCVMUrlSingapore
	}
	return CloudAPICall2(action, configInfo, newExtraParams)
}

func AcctAPICall(action string, config map[string]interface{}, extraParams map[string]string) []byte {
	var newExtraParams map[string]string
	if extraParams == nil {
		newExtraParams = make(map[string]string)
	} else {
		newExtraParams = extraParams
	}
	if newExtraParams[CUrl] == "" { // ensure existing url is not overwritten
		newExtraParams[CUrl] = "account.api.qcloud.com/v2/index.php"
	}
	return CloudAPICall(action, config, newExtraParams)
}

func SignatureStringNonceTimestamp(action, nonce, timestamp, secretId string, extraParams map[string]string) string {
	var sortparams = []string{}

	// Common request parameters
	params := make(map[string]string)
	params["Action"] = action
	sortparams = append(sortparams, "Action")
	params["Nonce"] = nonce
	sortparams = append(sortparams, "Nonce")
	params["Timestamp"] = timestamp
	sortparams = append(sortparams, "Timestamp")
	params[CSecretId] = secretId
	sortparams = append(sortparams, CSecretId)

	// params["SignatureMethod"] = "HmacSHA256"
	// sortparams = append(sortparams, "SignatureMethod")

	for k, v := range extraParams {
		params[k] = v
		sortparams = append(sortparams, k)
	}

	sort.Strings(sortparams)

	requestParamString := ""
	var paramstr = []string{}
	for _, requestKey := range sortparams {
		if params[requestKey] != "" {
			paramstr = append(paramstr, requestKey+"="+params[requestKey])
		}
	}

	requestParamString += strings.Join(paramstr, "&")
	return requestParamString

}

// SignatureString takes the given parameters, and generates a signature based on these parameters
// As the signature generated uses the current timestamp and nonce, it is impossible to generate/use
// the same signature, even for the same parameters.
// Sample usage:
//  log.Println(tencent.SignatureString("action", CSecretId, make(map[string]string)))
//  log.Println(tencent.SignatureString("action", CSecretId, nil))
//
func SignatureString(action, secretId string, extraParams map[string]string) string {
	nonce := GenerateNonce()
	timestamp := CurrentTimeStamp()
	return SignatureStringNonceTimestamp(action, nonce, timestamp, secretId, extraParams)
}

// Signature generates the Base64 encoded HMAC key for the given parameters
func SignatureGet(requestURL, requestParamString, secretKey string) string {
	// this GET string below, requires the request to be sent using http.Get!
	signstr := "GET" + requestURL + "?" + requestParamString
	// signature := Hmac256ToBase64(secretKey, signstr, true)
	signature := Hmac1ToBase64(secretKey, signstr, true) // if post, last param is false
	return signature
}

// Signature generates the Base64 encoded HMAC key for the given parameters
func SignaturePost(requestURL, requestParamString, secretKey string) string {
	// this POST string below, requires the request to be sent using http.Post!
	signstr := "POST" + requestURL + "?" + requestParamString
	// signature := Hmac256ToBase64(secretKey, signstr, true)
	signature := Hmac1ToBase64(secretKey, signstr, false) // if post, last param is false
	return signature
}

// Request makes a call to the HTTP endpoint
func RequestGet(requestURL, requestParamString, signature string) ([]byte, error) {
	defer func() {
		recover()
	}()

	url := "https://" + requestURL + "?" + requestParamString + "&Signature=" + signature
	if CloudAPIDebug {
		log.Printf("RequestGet url: %s", url)
	}

	resp, err := http.Get(url)

	if err != nil {
		if CloudAPIDebug {
			log.Printf("Error in RequestGet: %+v", err)
		}
		return nil, err
	}
	res, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return res, err
}

// Request makes a call to the HTTP endpoint
func RequestPost(requestURL, requestParamString, signature string) ([]byte, error) {
	defer func() {
		recover()
	}()
	data := requestParamString + "&Signature=" + signature
	// data := url.Values{}

	resp, err := http.Post("https://"+requestURL, "application/x-www-form-urlencoded",
		strings.NewReader(data))
	if CloudAPIDebug {
		log.Printf("data: %s", data)
	}
	if err != nil {
		return nil, err
	}
	res, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	return res, err
}

// GenerateNonce generates a unique key based on the current time
func GenerateNonce() string {
	rand.Seed(time.Now().UnixNano())
	time := rand.Intn(10000) + 10000
	return strconv.Itoa(time)
}

// Hmac256ToBase64 Base64 encodes the given parameters as a string
func Hmac1ToBase64(key string, str string, IsUrl bool) string {
	s := hmac.New(sha1.New, []byte(key))
	s.Write([]byte(str))
	return EncodingBase64(s.Sum(nil), IsUrl)
}

// Hmac256ToBase64 Base64 encodes the given parameters as a string
func Hmac256ToBase64(key string, str string, IsUrl bool) string {
	s := hmac.New(sha256.New, []byte(key))
	s.Write([]byte(str))
	return EncodingBase64(s.Sum(nil), IsUrl)
}

func EncodingBase64(b []byte, IsURL bool) string {
	if IsURL {
		return url.QueryEscape(base64.StdEncoding.EncodeToString(b))
	}
	return base64.StdEncoding.EncodeToString(b)
}

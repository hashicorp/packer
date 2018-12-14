package tencent

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type (
	// KeyPairs struct {
	// 	KeyId                 string
	// 	KeyName               string
	// 	ProjectId             string
	// 	Description           string
	// 	PublicKey             string
	// 	PrivateKey            string
	// 	AssociatedInstanceIds []string
	// 	// CreatedTime
	// }
	// DescribeKeyPairsResponse struct {
	// 	RequestId  string
	// 	TotalCount int
	// 	KeyPairSet []KeyPairs
	// }
	CVMKeyPair struct {
		KeyId      string
		KeyName    string
		ProjectId  int
		PublicKey  string
		PrivateKey string
	}

	CVMKeyPairSet struct {
		KeyId                 string   `json:"KeyId"`
		KeyName               string   `json:"KeyName"`
		Description           string   `json:"Description"`
		PublicKey             string   `json:"PublicKey"`
		AssociatedInstanceIds []string `json:"AssociatedInstanceIds"`
		CreatedTime           string   `json:"CreatedTime"` // can't handle time.Time
	}

	CVMDescribeKeyPairs struct {
		Error      CVMError        `json:"Error"`
		KeyPairSet []CVMKeyPairSet `json:"KeyPairSet"`
		TotalCount int             `json:"TotalCount"`
		RequestId  string          `json:"RequestId"`
	}

	CVMImportKeyPairResponse struct {
		KeyId     string
		Error     CVMError
		RequestId string
	}

	CVMCreateKeyPairResponse struct {
		KeyPair   CVMKeyPair
		Error     CVMError
		RequestId string
	}

	CVMAssociateInstanceKeyPairResponse struct {
		Error     CVMError
		TaskId    string
		RequestId string
	}

	CVMDisassociateInstancesKeyPairs struct {
		Error     CVMError
		TaskId    string
		RequestId string
	}

	CVMDeleteKeyPairResponse struct {
		Error     CVMError
		RequestId string
	}
)

func DescribeKeyPairs(c *Config) CVMDescribeKeyPairs {
	// extraParams := map[string]string{ // working
	// 	CUrl: c.Url,
	// }
	// if c.Url == "" {
	// 	delete(extraParams, CUrl)
	// }
	// response := CVMAPICall("DescribeKeyPairs", configInfo, extraParams)
	extraParams := c.CreateDescribeKeyPairsExtraParams()
	configInfo := c.CreateDescribeKeyPairsMap()
	response := CVMAPICall2("DescribeKeyPairs", configInfo, extraParams)
	var (
		cvmDescribeKeyPairsResponse struct {
			Response struct {
				CVMDescribeKeyPairs
			} `json:"Response"`
		}
	)
	err := json.Unmarshal(response, &cvmDescribeKeyPairsResponse)
	if err != nil {
		log.Printf("DescribeKeyPair error: %+v\n", err)
	}
	// TotalCount can be 0 if there are no keypairs available
	return cvmDescribeKeyPairsResponse.Response.CVMDescribeKeyPairs
}

func ImportKeyPair(c *Config) {
	// configInfo := c.CreateVMmap() // working
	// extraParams := map[string]string{
	// 	CKeyName:    c.SSHKeyName,
	// 	CProjectId:  "0",
	// 	"PublicKey": c.PublicKey,
	// }
	// response := CVMAPICall("ImportKeyPair", configInfo, extraParams)

	configInfo := c.CreateImportKeyPairMap()
	extraParams := c.CreateImportKeyPairExtraParams()
	response := CVMAPICall2("ImportKeyPair", configInfo, extraParams)

	var cvmImportKeyPair CVMImportKeyPairResponse
	DecodeResponse(response, &cvmImportKeyPair)
}

func CreateKeyPair(c *Config) CVMCreateKeyPairResponse {
	extraParams := c.CreateKeyPairExtraParams()
	configInfo := c.CreateKeyPairMap()

	dir, SSHKeyName := filepath.Split(c.SSHKeyName)
	if c.PackerDebug || CloudAPIDebug {
		log.Printf("CreateKeyPair Split dir: %s, name: %s\n", dir, SSHKeyName)
	}

	extraParams[CKeyName] = SSHKeyName

	response := CVMAPICall2("CreateKeyPair", configInfo, extraParams)
	var (
		cvmCreateKeyPairResponse CVMCreateKeyPairResponse
		jsonresp                 struct {
			Response struct {
				CVMCreateKeyPairResponse
			} `json:"Response"`
		}
	)
	err := json.Unmarshal(response, &jsonresp)
	if err != nil {
		log.Printf("CreateKeyPair error: %+v\n", err)
	} else {
		cvmCreateKeyPairResponse = jsonresp.Response.CVMCreateKeyPairResponse
	}
	return cvmCreateKeyPairResponse
}

// AssociateInstanceKeyPair associates an instance with the given KeyPairId
func AssociateInstanceKeyPair(c *Config,
	instanceId string,
	KeyPairId string) CVMAssociateInstanceKeyPairResponse {
	configInfo := c.CreateVMAKPMap()
	extraParams := c.UrlParams()
	extraParams["InstanceIds.0"] = instanceId
	extraParams["KeyIds.0"] = KeyPairId
	if c.PackerDebug || CloudAPIDebug {
		msg := fmt.Sprintf("AssociateInstanceKeyPair configInfo: %+v\n", configInfo)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)

		msg = fmt.Sprintf("AssociateInstanceKeyPair extraParams: %+v\n", extraParams)
		msg = strings.Replace(msg, c.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}
	response := CVMAPICall2("AssociateInstancesKeyPairs", configInfo, extraParams)
	var associateInstanceKeyPairResponse CVMAssociateInstanceKeyPairResponse
	err := DecodeResponse(response, &associateInstanceKeyPairResponse)
	if err != nil {
		log.Printf("AssociateInstanceKeyPair Error: %+v\n", err)
	}
	return associateInstanceKeyPairResponse
}

// CheckBindingCompleted tries to check if the call to AssociateInstacesKeyPairs has completed
// by attempting to start the instance. If the binding has not been completed, it returns
// the following error
// "Error": {
// 	"Code": "MutexOperation.TaskRunning",
// 	"Message": "Mutex"
// },
func CheckBindingCompleted(c *Config, instanceId string) bool {
	// 	 StartVM(c *Config, instanceId string) (CVMError, bool) {
	cvmError, successful := StartVM(c, instanceId)
	if !successful {
		if cvmError.Code == "MutexOperation.TaskRunning" {
			return false
		} else {
			log.Printf("Unexpected response, code: %s, message: %s", cvmError.Code, cvmError.Message)
		}
	} else {
		return true
	}
	return false
}

func DisassociateInstancesKeyPairs(c *Config,
	instanceId string,
	KeyPairId string) CVMDisassociateInstancesKeyPairs {
	configInfo := c.CreateBasicMap()
	extraParams := c.UrlParams()
	extraParams["InstanceIds.0"] = instanceId
	extraParams["KeyIds.0"] = KeyPairId
	response := CVMAPICall("DisassociateInstancesKeyPairs", configInfo, extraParams)
	log.Printf("DisassociateInstancesKeyPairs Response: %s\n", string(response))
	var cvmDisassociateInstancesKeyPairs CVMDisassociateInstancesKeyPairs
	err := DecodeResponse(response, &cvmDisassociateInstancesKeyPairs)
	if err != nil {
		log.Printf("DisassociateInstancesKeyPairs Error: %+v\n", err)
	}
	return cvmDisassociateInstancesKeyPairs
}

// DeleteKeyPair deletes a KeyPair identified by the KeyPairId, and returns true when successful
func DeleteKeyPair(c *Config, KeyPairId string) CVMDeleteKeyPairResponse {
	configInfo := c.CreateVMmap()
	extraParams := c.UrlParams()
	extraParams["KeyIds.0"] = KeyPairId
	response := CVMAPICall("DeleteKeyPairs", configInfo, extraParams)
	var cvmDeleteKeyPairResponse CVMDeleteKeyPairResponse
	err := DecodeResponse(response, &cvmDeleteKeyPairResponse)
	success := cvmDeleteKeyPairResponse.Error.Code == ""
	if err != nil {
		log.Printf("DeleteKeyPair decode error: %+v\n", err)
	}
	if !success {
		log.Printf("DeleteKeyPair error: %+v\n", cvmDeleteKeyPairResponse.Error)
	}
	return cvmDeleteKeyPairResponse
}

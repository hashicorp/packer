package tencent

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type TencentDriver struct {
	Ui     packer.Ui
	config Config
	state  multistep.StateBag
}

// The function signatures in this source must match the function signatures in driver.go
// CreateImage creates an image on the cloud
func (driver TencentDriver) CWCreateImage(config Config) (bool, CVMError, CVMInstanceInfo) {
	if config.PackerDebug || CloudAPIDebug {
		log.Println("CWCreateImage: Calling CreateVM")
	}
	cvmError, instanceInfo := CreateVM(&config)
	if config.PackerDebug || CloudAPIDebug {
		log.Println("CWCreateImage: Succeeded in calling CreateVM")
	}
	var err bool = false
	if cvmError.Code != "" {
		msg := fmt.Sprintf("CWCreateImage error code: %s, message: %s", cvmError.Code,
			cvmError.Message)
		if config.PackerDebug || CloudAPIDebug {
			log.Printf(msg)
		}
		err = true
	}
	return err, cvmError, instanceInfo
}

func (driver TencentDriver) CWCreateCustomImage(config Config, instanceId string) (bool, CVMError, CVMCreateCustomImage) {
	// driver.Ui.Say(fmt.Sprintf("Creating image: %+v", config1))
	ui := driver.Ui
	if config.PackerDebug || CloudAPIDebug {
		log.Println("CWCreateCustomImage: Calling CreateCustomImage")
	}
	cvmError, cvmCreateCustomImage := CreateCustomImage(&config, instanceId)
	if config.PackerDebug || CloudAPIDebug {
		log.Println("CWCreateCustomImage: Succeeded in calling CreateCustomImage")
	}
	var err bool = false
	if cvmError.Code != "" {
		msg := fmt.Sprintf("CWCreateImage error: code %s, message: %s", cvmError.Code,
			cvmError.Message)
		ui.Say(msg)
		if config.PackerDebug || CloudAPIDebug {
			log.Printf("CWCreateCustomImage: CreateCustomImage %+v", cvmError)
		}
		err = true
	}
	return err, cvmError, cvmCreateCustomImage
}

func (driver TencentDriver) CWWaitForCustomImageReady(config Config) (bool, string) {
	if config.ImageUrl != "" {
		config.Url = config.ImageUrl
	} else {
		config.Url = CImageAPIUrl
	}
	ImageReady, ImageID := WaitForCustomImageReady(&config)
	return ImageReady, ImageID
}

// CWCreateKeyPair creates a keypair and associates it to an instance, then save the SSH private key
// to the user's home directory using a combination of the InstanceId (without the dash) and
// the current timestamp
func (driver TencentDriver) CWCreateKeyPair(config Config, instanceId string, state multistep.StateBag) (error, CVMKeyPair) {
	ui := driver.Ui
	if config.PackerDebug || CloudAPIDebug {
		log.Printf("CWCreateKeyPair: Calling CreateKeyPair for instance: %s\n", instanceId)
	}
	NewPrefix := strings.Replace(instanceId, "-", "", -1)
	_, LocalKeyName := filepath.Split(config.SSHKeyName)
	timestampsuffix := SSHTimeStampSuffix()
	config.SSHKeyName = fmt.Sprintf("%s_%s", NewPrefix, timestampsuffix)
	response1 := CreateKeyPair(&config)
	var EmptyKeyPair CVMKeyPair
	if response1.Error.Code != "" {
		msg := fmt.Sprintf("CWCreateImage error: code %s, message: %s", response1.Error.Code,
			response1.Error.Message)
		ui.Say(msg)
		errMsg := fmt.Sprintf("CWCreateKeyPair error code: %s, message: %s", response1.Error.Code, response1.Error.Message)
		if config.PackerDebug || CloudAPIDebug {
			log.Printf(errMsg)
		}
		err := errors.New(errMsg)
		return err, EmptyKeyPair
	}
	LocalKeyName = fmt.Sprintf("%s_%s_%s", LocalKeyName, instanceId, timestampsuffix)
	sshKeySaveLocation := filepath.Join(UserHomeDir(), LocalKeyName)
	if config.PackerDebug || CloudAPIDebug {
		log.Printf("CWCreateKeyPair: Saving SSH key to %s", sshKeySaveLocation)
	}
	successful, err := SaveDataToFile(sshKeySaveLocation, []byte(response1.KeyPair.PrivateKey))
	if !successful {
		errmsg := fmt.Sprintf("CWCreateKeyPair: Failed to save private key to file, error: %+v", err)
		ui.Say(errmsg)
		if config.PackerDebug || CloudAPIDebug {
			log.Print(errmsg)
		}
		return err, EmptyKeyPair
	} else {
		sshSavedMsg := fmt.Sprintf("Saved SSH key to %s", sshKeySaveLocation)
		ui.Say(sshSavedMsg)
	}

	state.Put(CSSHKeyLocation, sshKeySaveLocation)

	if config.PackerDebug || CloudAPIDebug {
		log.Println("CWCreateKeyPair: now associating keypair")
	}

	// Assumes InstanceId is in STOPPED STATE!!!
	// KeyPair creation successful, now try binding to the specified instance
	KeyPairId := response1.KeyPair.KeyId
	response2 := AssociateInstanceKeyPair(&config, instanceId, KeyPairId)
	if response2.Error.Code == "" {
		log.Println("CWCreateKeyPair: Successfully associated instance keypair!")
		return nil, response1.KeyPair
	}
	errMsg := fmt.Sprintf("CWCreateKeyPair: error code: %s, message: %s", response2.Error.Code, response2.Error.Message)
	ui.Say(errMsg)
	log.Printf(errMsg)
	err2 := errors.New(errMsg)
	return err2, EmptyKeyPair
}

func (driver TencentDriver) CWGetImageState(config Config, instanceId string) (error, string) {
	log.Printf("CWGetImageState: Getting image state for Instance: %s\n", instanceId)
	err, state := GetInstanceState(&config, instanceId)
	if err == nil {
		log.Printf("CWGetImageState: GetInstanceState successful, state: %s", state)
		return nil, state
	}
	return err, state
}

func (driver TencentDriver) CWGetInstanceIP(config Config, instanceId string) (error, string) {
	log.Printf("CWGetInstanceIP: Getting IP address for Instance %s\n", instanceId)
	err, IPAddress := GetInstanceIP(&config, instanceId)
	if err != nil {
		return nil, ""
	}
	return nil, IPAddress
}

func (driver TencentDriver) CWRunImage(config Config, instanceId string) error {
	if config.PackerDebug || CloudAPIDebug {
		log.Printf("CWRunImage: running instance: %s\n", instanceId)
	}

	cvmError, successful := StartVM(&config, instanceId)
	if successful {
		return nil
	} else {
		errMsg := fmt.Sprintf("CWRunImage: error code: %s message: %s", cvmError.Code, cvmError.Message)
		err := errors.New(errMsg)
		return err
	}
}

func (driver TencentDriver) CWStopImage(config Config, instanceId string) error {
	ui := driver.Ui
	if config.PackerDebug || CloudAPIDebug {
		log.Printf("CWStopImage: stopping instance: %s\n", instanceId)
	}
	_, state := GetInstanceState(&config, instanceId)
	if state == CSTOPPED {
		return nil
	}
	cvmError, successful := StopVM(&config, instanceId)
	if successful {
		return nil
	} else {
		errMsg := fmt.Sprintf("CWStopImage: error code: %s, message: %s", cvmError.Code, cvmError.Message)
		ui.Say(errMsg)
		err := errors.New(errMsg)
		return err
	}
}

func (driver TencentDriver) CWWaitForImageState(config Config, instanceId string, state string) error {
	ui := driver.Ui
	log.Printf("CWWaitForImageState waiting for instance: %s to reach %s state\n", instanceId, state)

	// cvm.ap-singapore.tencentcloudapi.com needs ImageId, Region, Placement.Zone
	err, successful := WaitForVM(&config, instanceId, state)
	if successful {
		log.Println("CWWaitForImageState successful!")
		return nil
	} else {
		errMsg := fmt.Sprintf("CWWaitForImageState failed, due to error: %v", err.Error())
		ui.Say(errMsg)
		log.Println(errMsg)
		return err
	}
}

func (driver TencentDriver) CWWaitKeyPairAttached(config Config, instanceId, KeyId string) error {
	ui := driver.Ui
	startTime := time.Now()
	endTime := startTime.Add(time.Millisecond * time.Duration(config.Timeout))
	// brk := false
	for time.Now().Before(endTime) {
		keyPairResponse := DescribeKeyPairs(&config)
		if keyPairResponse.TotalCount > 0 {
			for _, v := range keyPairResponse.KeyPairSet {
				if v.KeyId != KeyId {
					continue
				}
				if config.PackerDebug || CloudAPIDebug {
					msg := "CWWaitKeyPairAttached: KeyId found!"
					ui.Say(msg)
					log.Println(msg)
				}
				for _, KeyPairInstanceId := range v.AssociatedInstanceIds {
					if KeyPairInstanceId == instanceId {
						if config.PackerDebug || CloudAPIDebug {
							msg := "CWWaitKeyPairAttached: InstanceId matched!"
							ui.Say(msg)
							log.Println(msg)
						}
						return nil
					}
				}
			}
		} else { // TotalCount == 0
			errMsg := fmt.Sprintf("Error code: %s, message: %s", keyPairResponse.Error.Code, keyPairResponse.Error.Message)
			ui.Say(errMsg)
			log.Println(errMsg)
			return errors.New(errMsg)
		}
		time.Sleep(time.Second * 5)
	}
	errMsg := "CWWaitKeyPairAttached: No keypair found!"
	ui.Say(errMsg)
	log.Println(errMsg)
	err := errors.New(errMsg)
	return err
}

func NewTencentDriver(ui packer.Ui, config *Config, state multistep.StateBag) *TencentDriver {
	driver := new(TencentDriver)
	driver.Ui = ui
	driver.config = *config
	driver.state = state
	return driver
}

package tencent

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type
// StepClear is an empty step designed to test the builder
// It does nothing, and returns ActionContinue in Run
StepClear struct{}

// Run just tells it to continue. This is an empty step to test the builder
func (s *StepClear) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	return multistep.ActionContinue
}

func (s *StepClear) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type
// StepCreateImage creates an image with the specified attributes
StepCreateImage struct{}

// Run for StepCreateImage creates an instance of an image with attributes specified in the config
func (s *StepCreateImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepCreateImage)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	if driver == nil {
	}
	savedKeyName := config.SSHKeyName
	config.SSHKeyName = ""
	err, cvmError, instanceInfo := driver.CWCreateImage(*config)
	if config.PackerDebug || CloudAPIDebug {
		log.Println("Call to driver.CWCreateImage successful!")
	}

	config.SSHKeyName = savedKeyName
	if err || cvmError.Code != "" || instanceInfo.InstanceId == "" {
		errMsg := fmt.Sprintf("Problem creating image. Code: %s, message: %s", cvmError.Code, cvmError.Message)
		error := errors.New(errMsg)
		state.Put(CError, error)
		log.Println(errMsg)
		return multistep.ActionHalt
	}

	// Saves the instance id, so it can be read from a later location
	if config.ImageIdLocation != "" {
		SaveDataToFile(config.ImageIdLocation, []byte(instanceInfo.InstanceId))
		if config.PackerDebug || CloudAPIDebug {
			log.Printf("Saved Instance ID: %s to %s\n", instanceInfo.InstanceId, config.ImageIdLocation)
		}
	}

	statusMsg1 := fmt.Sprintf("StepCreateImage created Instance ID: %s successfully!", instanceInfo.InstanceId)
	ui.Say(statusMsg1)
	log.Println(statusMsg1)
	state.Put(CInstanceId, instanceInfo.InstanceId)

	uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepCreateImage)
	ui.Say(uiMsgEnd)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgEnd)
	}

	return multistep.ActionContinue
}

func (s *StepCreateImage) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type
// StepCreateCustomImage creates an image with the specified attributes
StepCreateCustomImage struct {
}

// Run for StepCreateCustomImage creates an instance of an image with attributes specified in the config
// StepCreateCustomImage requires Config.ImageName to be set
func (s *StepCreateCustomImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepCreateCustomImage)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	savedKeyName := config.SSHKeyName
	config.SSHKeyName = ""
	InstanceId := state.Get(CInstanceId).(string)
	failed, cvmError, cvmCreateCustomImage := driver.CWCreateCustomImage(*config, InstanceId)
	if !failed {
		if cvmCreateCustomImage.RequestId != "" {
			// do nothing
		}
		if config.PackerDebug || CloudAPIDebug {
			log.Println("Call to driver.CWCreateCustomImage successful!")
		}
		config.SSHKeyName = savedKeyName

		ok, MasterImageId := driver.CWWaitForCustomImageReady(*config)
		// Save the Master Image Id so that it can be persisted as an artifact
		if ok {
			if config.PackerDebug || CloudAPIDebug {
				log.Printf("Saving Instance ID: %s to state\n ", MasterImageId)
			}
			state.Put(CInstanceId, MasterImageId)
		}
	}

	if failed || cvmError.Code != "" {
		errMsg := fmt.Sprintf("Problem creating image. Code: %s, message: %s", cvmError.Code, cvmError.Message)
		if config.PackerDebug || CloudAPIDebug {
			log.Printf(errMsg)
		}
		error := errors.New(errMsg)
		state.Put(CError, error)
		return multistep.ActionHalt
	}

	uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepCreateCustomImage)
	ui.Say(uiMsgEnd)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgEnd)
	}

	return multistep.ActionContinue
}

func (s *StepCreateCustomImage) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type StepStopImage struct{}

func (s *StepStopImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepStopImage)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	instanceId := state.Get(CInstanceId).(string)
	err := driver.CWStopImage(*config, instanceId)

	if err != nil {
		errMsg := fmt.Sprintf("%s error: %+v", CStepStopImage, err)
		ui.Say(errMsg)
		log.Println(errMsg)
		state.Put(CError, err)
		return multistep.ActionHalt
	}

	uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepStopImage)
	ui.Say(uiMsgEnd)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgEnd)
	}

	return multistep.ActionContinue
}

func (s *StepStopImage) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type StepRunImage struct{}

// Run retrieves an InstanceId from the given state, and invokes a call to the API to run it.
func (s *StepRunImage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepRunImage)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	var instanceId string
	instanceIntf := state.Get(CInstanceId)
	if instanceIntf == nil {
		instanceId = GetEnvVar(CInstanceId)
		state.Put(CInstanceId, instanceId)
	}
	instanceId = state.Get(CInstanceId).(string)
	if instanceId == "" {
		const msg = "No InstanceId given for StepRunImage"
		ui.Error(msg)
		log.Println(msg)
		err := errors.New(msg)
		state.Put(CError, err)
		return multistep.ActionHalt
	}

	err := driver.CWRunImage(*config, instanceId)
	if err == nil {

		uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepRunImage)
		ui.Say(uiMsgEnd)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(uiMsgEnd)
		}

		return multistep.ActionContinue
	}

	state.Put(CError, err)
	log.Printf("StepRunImage error: %+v\n", err)
	return multistep.ActionHalt
}

func (s *StepRunImage) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type StepWaitRunning struct{}

func (s *StepWaitRunning) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepWaitRunning)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	InstanceId := state.Get(CInstanceId).(string)

	if config.PackerDebug || CloudAPIDebug {
		log.Printf("Waiting for instance: %s to reach RUNNING state...\n", InstanceId)
	}

	err := driver.CWWaitForImageState(*config, InstanceId, "RUNNING")
	if err == nil {

		uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepWaitRunning)
		ui.Say(uiMsgEnd)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(uiMsgEnd)
		}

		return multistep.ActionContinue
	}

	log.Printf("StepWaitRunning error %+v\n", err)
	state.Put(CError, err)
	return multistep.ActionHalt
}
func (s *StepWaitRunning) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type StepWaitStopped struct{}

func (s *StepWaitStopped) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepWaitStopped)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	InstanceId := state.Get(CInstanceId).(string)

	if config.PackerDebug || CloudAPIDebug {
		log.Println("StepWaitStopped starting...")
		log.Printf("Waiting for instance: %s to reach STOPPED state...\n", InstanceId)
	}

	err := driver.CWWaitForImageState(*config, InstanceId, CSTOPPED)
	if err == nil {

		uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepWaitStopped)
		ui.Say(uiMsgEnd)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(uiMsgEnd)
		}

		return multistep.ActionContinue
	}
	log.Printf("StepWaitStopped error %+v\n", err)
	state.Put(CError, err)
	return multistep.ActionHalt
}
func (s *StepWaitStopped) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type StepCreateKeyPair struct{}

func (s *StepCreateKeyPair) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {

	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepCreateKeyPair)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	InstanceId := state.Get(CInstanceId).(string)

	err, KeyPair := driver.CWCreateKeyPair(*config, InstanceId, state)
	if err == nil {

		uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepCreateKeyPair)
		ui.Say(uiMsgEnd)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(uiMsgEnd)
		}

		state.Put("KeyPair", KeyPair)
		return multistep.ActionContinue
	}
	if config.PackerDebug || CloudAPIDebug {
		log.Printf("StepCreateKeyPair error %+v\n", err)
	}
	state.Put(CError, err)
	return multistep.ActionHalt
}
func (s *StepCreateKeyPair) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type StepGetImageStatus struct{}

func (s *StepGetImageStatus) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepGetImageStatus)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	InstanceId := state.Get(CInstanceId).(string)
	err, myState := driver.CWGetImageState(*config, InstanceId)
	if err == nil {
		uiMsg := fmt.Sprintf("StepGetImageStatus: called CWGetImageState successfully, state: %v", myState)
		ui.Say(uiMsg)
		log.Println(uiMsg)

		uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepGetImageStatus)
		ui.Say(uiMsgEnd)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(uiMsgEnd)
		}

		return multistep.ActionContinue
	}
	state.Put(CError, err)
	return multistep.ActionHalt
}
func (s *StepGetImageStatus) Cleanup(state multistep.StateBag) {
	// No cleanup
}

type StepGetInstanceIP struct{}

func (s *StepGetInstanceIP) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepGetInstanceIP)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	InstanceId := state.Get(CInstanceId).(string)
	err, IPAddress := driver.CWGetInstanceIP(*config, InstanceId)
	var msg string
	if err == nil {
		if IPAddress != "" {
			msg = fmt.Sprintf("%s completed successfully, IP is: %s!", CStepGetInstanceIP, IPAddress)
		} else {
			msg = fmt.Sprintf("%s completed successfully, it appears there's no IP!", CStepGetInstanceIP)
		}
		state.Put(CArtifactIPAddress, IPAddress)
	} else {
		msg = fmt.Sprintf("%s completed successfully, it appears there's no IP!", CStepGetInstanceIP)
		state.Put(CArtifactIPAddress, "")
		state.Put(CError, err)
	}
	ui.Say(msg)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(msg)
	}

	uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepGetInstanceIP)
	ui.Say(uiMsgEnd)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgEnd)
	}

	return multistep.ActionContinue
}

func (s *StepGetInstanceIP) Cleanup(state multistep.StateBag) {
	// do nothing
}

type StepGetKeyPairStatus struct{}

func (s *StepGetKeyPairStatus) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	KeyPair := state.Get("KeyPair").(CVMKeyPair)
	config := state.Get(CConfig).(*Config)
	driver := state.Get(CDriver).(Driver)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepGetKeyPairStatus)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	InstanceId := state.Get(CInstanceId).(string)
	err := driver.CWWaitKeyPairAttached(*config, InstanceId, KeyPair.KeyId)
	if err == nil {

		uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepGetKeyPairStatus)
		ui.Say(uiMsgEnd)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(uiMsgEnd)
		}

		return multistep.ActionContinue
	}
	state.Put(CError, err)
	return multistep.ActionHalt
}

func (s *StepGetKeyPairStatus) Cleanup(state multistep.StateBag) {
}

type StepHalt struct{}

func (s *StepHalt) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	return multistep.ActionHalt
}

func (s *StepHalt) Cleanup(state multistep.StateBag) {
}

type
// StepDisplayMessage displays a message on the console, not used internally, but for debugging
StepDisplayMessage struct{}

func (s *StepDisplayMessage) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(CConfig).(*Config)
	ui := state.Get("ui").(packer.Ui)

	uiMsgStart := fmt.Sprintf(CSTEPSTARTING, CStepDisplayMessage)
	ui.Say(uiMsgStart)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgStart)
	}

	msg := fmt.Sprintf("Current time is: %s", CurrentTimeStamp())
	ui.Say(msg)

	uiMsgEnd := fmt.Sprintf(CSTEPFINISHED, CStepDisplayMessage)
	ui.Say(uiMsgEnd)
	if config.PackerDebug || CloudAPIDebug {
		log.Println(uiMsgEnd)
	}

	return multistep.ActionContinue
}

func (s *StepDisplayMessage) Cleanup(state multistep.StateBag) {
}

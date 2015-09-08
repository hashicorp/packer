package common

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/helper/communicator"
	"github.com/mitchellh/packer/packer"
)

// StepGetPassword reads the password from a Windows server and sets it
// on the WinRM config.
type StepGetPassword struct {
	Debug   bool
	Comm    *communicator.Config
	Timeout time.Duration
}

func (s *StepGetPassword) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)

	// Skip if we're not using winrm
	if s.Comm.Type != "winrm" {
		log.Printf("[INFO] Not using winrm communicator, skipping get password...")
		return multistep.ActionContinue
	}

	// If we already have a password, skip it
	if s.Comm.WinRMPassword != "" {
		ui.Say("Skipping waiting for password since WinRM password set...")
		return multistep.ActionContinue
	}

	// Get the password
	var password string
	var err error
	cancel := make(chan struct{})
	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for auto-generated password for instance...")
		ui.Message(
			"It is normal for this process to take up to 15 minutes,\n" +
				"but it usually takes around 5. Please wait.")
		password, err = s.waitForPassword(state, cancel)
		waitDone <- true
	}()

	timeout := time.After(s.Timeout)
WaitLoop:
	for {
		// Wait for either SSH to become available, a timeout to occur,
		// or an interrupt to come through.
		select {
		case <-waitDone:
			if err != nil {
				ui.Error(fmt.Sprintf("Error waiting for password: %s", err))
				state.Put("error", err)
				return multistep.ActionHalt
			}

			ui.Message(fmt.Sprintf(" \nPassword retrieved!"))
			s.Comm.WinRMPassword = password
			break WaitLoop
		case <-timeout:
			err := fmt.Errorf("Timeout waiting for password.")
			state.Put("error", err)
			ui.Error(err.Error())
			close(cancel)
			return multistep.ActionHalt
		case <-time.After(1 * time.Second):
			if _, ok := state.GetOk(multistep.StateCancelled); ok {
				// The step sequence was cancelled, so cancel waiting for password
				// and just start the halting process.
				close(cancel)
				log.Println("[WARN] Interrupt detected, quitting waiting for password.")
				return multistep.ActionHalt
			}
		}
	}

	// In debug-mode, we output the password
	if s.Debug {
		ui.Message(fmt.Sprintf(
			"Password (since debug is enabled): %s", s.Comm.WinRMPassword))
	}

	return multistep.ActionContinue
}

func (s *StepGetPassword) Cleanup(multistep.StateBag) {}

func (s *StepGetPassword) waitForPassword(state multistep.StateBag, cancel <-chan struct{}) (string, error) {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	privateKey := state.Get("privateKey").(string)

	for {
		select {
		case <-cancel:
			log.Println("[INFO] Retrieve password wait cancelled. Exiting loop.")
			return "", errors.New("Retrieve password wait cancelled")
		case <-time.After(5 * time.Second):
		}

		resp, err := ec2conn.GetPasswordData(&ec2.GetPasswordDataInput{
			InstanceId: instance.InstanceId,
		})
		if err != nil {
			err := fmt.Errorf("Error retrieving auto-generated instance password: %s", err)
			return "", err
		}

		if resp.PasswordData != nil && *resp.PasswordData != "" {
			decryptedPassword, err := decryptPasswordDataWithPrivateKey(
				*resp.PasswordData, []byte(privateKey))
			if err != nil {
				err := fmt.Errorf("Error decrypting auto-generated instance password: %s", err)
				return "", err
			}

			return decryptedPassword, nil
		}

		log.Printf("[DEBUG] Password is blank, will retry...")
	}
}

func decryptPasswordDataWithPrivateKey(passwordData string, pemBytes []byte) (string, error) {
	encryptedPasswd, err := base64.StdEncoding.DecodeString(passwordData)
	if err != nil {
		return "", err
	}

	block, _ := pem.Decode(pemBytes)
	var asn1Bytes []byte
	if _, ok := block.Headers["DEK-Info"]; ok {
		return "", errors.New("encrypted private key isn't yet supported")
		/*
			asn1Bytes, err = x509.DecryptPEMBlock(block, password)
			if err != nil {
				return "", err
			}
		*/
	} else {
		asn1Bytes = block.Bytes
	}

	key, err := x509.ParsePKCS1PrivateKey(asn1Bytes)
	if err != nil {
		return "", err
	}

	out, err := rsa.DecryptPKCS1v15(nil, key, encryptedPasswd)
	if err != nil {
		return "", err
	}

	return string(out), nil
}

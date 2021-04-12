package common

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/retry"
)

// StepGetPassword reads the password from a Windows server and sets it
// on the WinRM config.
type StepGetPassword struct {
	Debug     bool
	Comm      *communicator.Config
	Timeout   time.Duration
	BuildName string
}

func (s *StepGetPassword) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)

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
	waitDone := make(chan bool, 1)
	go func() {
		ui.Say("Waiting for auto-generated password for instance...")
		ui.Message(
			"It is normal for this process to take up to 15 minutes,\n" +
				"but it usually takes around 5. Please wait.")
		password, err = s.waitForPassword(ctx, state)
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
			return multistep.ActionHalt
		case <-ctx.Done():
			// The step sequence was cancelled, so cancel waiting for password
			// and just start the halting process.
			log.Println("[WARN] Interrupt detected, quitting waiting for password.")
			return multistep.ActionHalt
		}
	}

	// In debug-mode, we output the password
	if s.Debug {
		ui.Message(fmt.Sprintf(
			"Password (since debug is enabled): %s", s.Comm.WinRMPassword))
	}
	// store so that we can access this later during provisioning
	state.Put("winrm_password", s.Comm.WinRMPassword)
	packersdk.LogSecretFilter.Set(s.Comm.WinRMPassword)

	return multistep.ActionContinue
}

func (s *StepGetPassword) Cleanup(multistep.StateBag) {}

func (s *StepGetPassword) waitForPassword(ctx context.Context, state multistep.StateBag) (string, error) {
	ec2conn := state.Get("ec2").(*ec2.EC2)
	instance := state.Get("instance").(*ec2.Instance)
	privateKey := s.Comm.SSHPrivateKey

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] Retrieve password wait cancelled. Exiting loop.")
			return "", errors.New("Retrieve password wait cancelled")
		case <-time.After(5 * time.Second):
		}

		// Wrap in a retry so that we don't fail on rate-limiting.
		log.Printf("Retrieving auto-generated instance password...")
		var resp *ec2.GetPasswordDataOutput
		err := retry.Config{
			Tries:      11,
			RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			var err error
			resp, err = ec2conn.GetPasswordData(&ec2.GetPasswordDataInput{
				InstanceId: instance.InstanceId,
			})
			if err != nil {
				err := fmt.Errorf("Error retrieving auto-generated instance password: %s", err)
				return err
			}
			return nil
		})

		if err != nil {
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

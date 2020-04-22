//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type IAPConfig

package googlecompute

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/packer/common/net"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/packer/tmp"
)

// StepStartTunnel represents a Packer build step that launches an IAP tunnel
type IAPConfig struct {
	// Whether to use an IAP proxy.
	// Prerequisites and limitations for using IAP:
	// - You must manually enable the IAP API in the Google Cloud console.
	// - You must have the gcloud sdk installed on the computer running Packer.
	// - You must be using a Service Account with a credentials file (using the
	//	 account_file option in the Packer template)
	// - This is currently only implemented for the SSH communicator, not the
	//   WinRM Communicator.
	// - You must add the given service account to project level IAP permissions
	//   in https://console.cloud.google.com/security/iap. To do so, click
	//   "project" > "SSH and TCP resoures" > "All Tunnel Resources" >
	//   "Add Member". Then add your service account and choose the role
	//   "IAP-secured Tunnel User" and add any conditions you may care about.
	IAP bool `mapstructure:"use_iap" required:"false"`
	// Which port to connect the local end of the IAM localhost proxy to. If
	// left blank, Packer will choose a port for you from available ports.
	IAPLocalhostPort int `mapstructure:"iap_localhost_port"`
	// What "hashbang" to use to invoke script that sets up gcloud.
	// Default: "/bin/sh"
	IAPHashBang string `mapstructure:"iap_hashbang" required:"false"`
	// What file extension to use for script that sets up gcloud.
	// Default: ".sh"
	IAPExt string `mapstructure:"iap_ext" required:"false"`
}

type TunnelDriver interface {
	StartTunnel(context.Context, string) error
	StopTunnel()
}

type StepStartTunnel struct {
	IAPConf     *IAPConfig
	CommConf    *communicator.Config
	AccountFile string

	tunnelDriver TunnelDriver
}

func (s *StepStartTunnel) ConfigureLocalHostPort(ctx context.Context) error {
	if s.IAPConf.IAPLocalhostPort == 0 {
		log.Printf("Finding an available TCP port for IAP proxy")
		l, err := net.ListenRangeConfig{
			Min:     8000,
			Max:     9000,
			Addr:    "0.0.0.0",
			Network: "tcp",
		}.Listen(ctx)

		if err != nil {
			err := fmt.Errorf("error finding an available port to initiate a session tunnel: %s", err)
			return err
		}

		s.IAPConf.IAPLocalhostPort = l.Port
		l.Close()
		log.Printf("Setting up proxy to listen on localhost at %d",
			s.IAPConf.IAPLocalhostPort)
	}
	return nil
}

func (s *StepStartTunnel) createTempGcloudScript(args []string) (string, error) {
	// Generate temp script that contains both gcloud auth and gcloud compute
	// iap launch call.

	// Create temp file.
	tf, err := tmp.File("gcloud-setup")
	if err != nil {
		return "", fmt.Errorf("Error preparing gcloud setup script: %s", err)
	}
	defer tf.Close()
	// Write our contents to it
	writer := bufio.NewWriter(tf)

	s.IAPConf.IAPHashBang = fmt.Sprintf("#!%s\n", s.IAPConf.IAPHashBang)
	log.Printf("[INFO] (google): Prepending inline gcloud setup script with %s",
		s.IAPConf.IAPHashBang)
	_, err = writer.WriteString(s.IAPConf.IAPHashBang)
	if err != nil {
		return "", fmt.Errorf("Error preparing inline hashbang: %s", err)
	}

	// authenticate to gcloud
	_, err = writer.WriteString(
		fmt.Sprintf("gcloud auth activate-service-account --key-file='%s'\n",
			s.AccountFile))
	if err != nil {
		return "", fmt.Errorf("Error preparing gcloud shell script: %s", err)
	}
	// call command
	args = append([]string{"gcloud"}, args...)
	argString := strings.Join(args, " ")
	if _, err := writer.WriteString(argString + "\n"); err != nil {
		return "", fmt.Errorf("Error preparing gcloud shell script: %s", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing shell script: %s", err)
	}

	err = os.Chmod(tf.Name(), 0700)
	if err != nil {
		log.Printf("[ERROR] (google): error modifying permissions of temp script file: %s", err.Error())
	}

	// figure out what extension the file should have, and rename it.
	tempScriptFileName := tf.Name()
	if s.IAPConf.IAPExt != "" {
		err := os.Rename(tempScriptFileName, fmt.Sprintf("%s%s", tempScriptFileName, s.IAPConf.IAPExt))
		if err != nil {
			return "", fmt.Errorf("Error setting the correct temp file extension: %s", err)
		}
		tempScriptFileName = fmt.Sprintf("%s%s", tempScriptFileName, s.IAPConf.IAPExt)
	}

	return tempScriptFileName, nil
}

// Run executes the Packer build step that creates an IAP tunnel.
func (s *StepStartTunnel) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	if !s.IAPConf.IAP {
		log.Printf("Skipping step launch IAP tunnel; \"iap\" is false.")
		return multistep.ActionContinue
	}

	// shell out to create the tunnel.
	ui := state.Get("ui").(packer.Ui)
	instanceName := state.Get("instance_name").(string)
	c := state.Get("config").(*Config)

	ui.Say("Step Launch IAP Tunnel...")

	err := s.ConfigureLocalHostPort(ctx)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Generate list of args to use to call gcloud cli.
	args := []string{"compute", "start-iap-tunnel", instanceName,
		strconv.Itoa(s.CommConf.Port()),
		fmt.Sprintf("--local-host-port=localhost:%d", s.IAPConf.IAPLocalhostPort),
		"--zone", c.Zone,
	}

	// This is the port the IAP tunnel listens on, on localhost.
	// TODO make setting LocalHostPort optional
	s.CommConf.SSHPort = s.IAPConf.IAPLocalhostPort

	log.Printf("Calling tunnel launch with args %#v", args)

	// Create temp file that contains both gcloud authentication, and gcloud
	// proxy setup call.
	tempScriptFileName, err := s.createTempGcloudScript(args)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}
	defer os.Remove(tempScriptFileName)

	s.tunnelDriver = NewTunnelDriver()

	err = retry.Config{
		Tries: 11,
		ShouldRetry: func(err error) bool {
			// TODO be stricter with retries.
			return true
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		// tunnel launcher/destroyer has to be different on windows vs. unix.
		err := s.tunnelDriver.StartTunnel(ctx, tempScriptFileName)
		return err
	})

	return multistep.ActionContinue
}

// Cleanup stops the IAP tunnel and cleans up processes.
func (s *StepStartTunnel) Cleanup(state multistep.StateBag) {
	s.tunnelDriver.StopTunnel()
}

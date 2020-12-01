//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type IAPConfig

package googlecompute

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer/packer-plugin-sdk/net"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/retry"
	"github.com/hashicorp/packer/packer-plugin-sdk/tmp"
)

// StepStartTunnel represents a Packer build step that launches an IAP tunnel
type IAPConfig struct {
	// Whether to use an IAP proxy.
	// Prerequisites and limitations for using IAP:
	// - You must manually enable the IAP API in the Google Cloud console.
	// - You must have the gcloud sdk installed on the computer running Packer.
	// - You must be using a Service Account with a credentials file (using the
	//	 account_file option in the Packer template)
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
	// How long to wait, in seconds, before assuming a tunnel launch was
	// successful. Defaults to 30 seconds for SSH or 40 seconds for WinRM.
	IAPTunnelLaunchWait int `mapstructure:"iap_tunnel_launch_wait" required:"false"`
}

type TunnelDriver interface {
	StartTunnel(context.Context, string, int) error
	StopTunnel()
}

func RunTunnelCommand(cmd *exec.Cmd, timeout int) error {
	// set stdout and stderr so we can read what's going on.
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Start()
	if err != nil {
		err := fmt.Errorf("Error calling gcloud sdk to launch IAP tunnel: %s",
			err)
		return err
	}

	// Give tunnel 30 seconds to either launch, or return an error.
	// Unfortunately, the SDK doesn't provide any official acknowledgment that
	// the tunnel is launched when it's not being run through a TTY so we
	// are just trusting here that 30s is enough to know whether the tunnel
	// launch was going to fail. Yep, feels icky to me too. But I spent an
	// afternoon trying to figure out how to get the SDK to actually send
	// the "Listening on port [n]" line I see when I run it manually, and I
	// can't justify spending more time than that on aesthetics.

	for i := 0; i < timeout; i++ {
		time.Sleep(1 * time.Second)

		lineStderr, err := stderr.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Printf("Err from scanning stderr is %s", err)
			return fmt.Errorf("Error reading stderr from tunnel launch: %s", err)
		}
		if lineStderr != "" {
			log.Printf("stderr: %s", lineStderr)
		}

		lineStdout, err := stdout.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Printf("Err from scanning stdout is %s", err)
			return fmt.Errorf("Error reading stdout from tunnel launch: %s", err)
		}
		if lineStdout != "" {
			log.Printf("stdout: %s", lineStdout)
		}

		if strings.Contains(lineStderr, "ERROR") {
			// 4033: Either you don't have permission to access the instance,
			// the instance doesn't exist, or the instance is stopped.
			// The two sub-errors we may see while the permissions settle are
			// "not authorized" and "failed to connect to backend," but after
			// about a minute of retries this goes away and we're able to
			// connect.
			// 4003: "failed to connect to backend". Network blip.
			if strings.Contains(lineStderr, "4033") || strings.Contains(lineStderr, "4003") {
				return RetryableTunnelError{lineStderr}
			} else {
				log.Printf("NOT RETRYABLE: %s", lineStderr)
				return fmt.Errorf("Non-retryable tunnel error: %s", lineStderr)
			}
		}
	}

	log.Printf("No error detected after tunnel launch; continuing...")
	return nil
}

type RetryableTunnelError struct {
	s string
}

func (e RetryableTunnelError) Error() string {
	return "Tunnel start: " + e.s
}

type StepStartTunnel struct {
	IAPConf            *IAPConfig
	CommConf           *communicator.Config
	AccountFile        string
	ImpersonateAccount string
	ProjectId          string

	tunnelDriver TunnelDriver
}

func (s *StepStartTunnel) ConfigureLocalHostPort(ctx context.Context) error {
	minPortNumber, maxPortNumber := 8000, 9000

	if s.IAPConf.IAPLocalhostPort != 0 {
		minPortNumber = s.IAPConf.IAPLocalhostPort
		maxPortNumber = minPortNumber
		log.Printf("Using TCP port for %d IAP proxy", s.IAPConf.IAPLocalhostPort)
	} else {
		log.Printf("Finding an available TCP port for IAP proxy")
	}

	l, err := net.ListenRangeConfig{
		Min:     minPortNumber,
		Max:     maxPortNumber,
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

	if s.IAPConf.IAPHashBang != "" {
		s.IAPConf.IAPHashBang = fmt.Sprintf("#!%s\n", s.IAPConf.IAPHashBang)
		log.Printf("[INFO] (google): Prepending inline gcloud setup script with %s",
			s.IAPConf.IAPHashBang)
		_, err = writer.WriteString(s.IAPConf.IAPHashBang)
		if err != nil {
			return "", fmt.Errorf("Error preparing inline hashbang: %s", err)
		}

	}

	launchTemplate := `
gcloud auth activate-service-account --key-file='{{.AccountFile}}'
{{.Args}}
`
	if runtime.GOOS == "windows" {
		launchTemplate = `
call gcloud auth activate-service-account --key-file "{{.AccountFile}}"
call {{.Args}}
`
	}
	// call command
	args = append([]string{"gcloud"}, args...)
	argString := strings.Join(args, " ")

	var tpl = template.Must(template.New("createTunnel").Parse(launchTemplate))
	var b bytes.Buffer

	opts := map[string]string{
		"AccountFile": s.AccountFile,
		"Args":        argString,
	}

	err = tpl.Execute(&b, opts)
	if err != nil {
		fmt.Println(err)
	}

	if _, err := writer.WriteString(b.String()); err != nil {
		return "", fmt.Errorf("Error preparing gcloud shell script: %s", err)
	}

	if err := writer.Flush(); err != nil {
		return "", fmt.Errorf("Error preparing shell script: %s", err)
	}
	// Have to close temp file before renaming it or Windows will complain.
	tf.Close()
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
	ui := state.Get("ui").(packersdk.Ui)
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
		"--zone", c.Zone, "--project", s.ProjectId,
	}

	if s.ImpersonateAccount != "" {
		args = append(args, fmt.Sprintf("--impersonate-service-account='%s'", s.ImpersonateAccount))
	}

	// This is the port the IAP tunnel listens on, on localhost.
	// TODO make setting LocalHostPort optional
	err = ApplyIAPTunnel(s.CommConf, s.IAPConf.IAPLocalhostPort)
	if err != nil {
		// this should not occur as the config should validate that the communicator
		// supports using an IAP tunnel
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	log.Printf("Creating tunnel launch script with args %#v", args)
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
			switch err.(type) {
			case RetryableTunnelError:
				return true
			default:
				return false
			}
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 30 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		// tunnel launcher/destroyer has to be different on windows vs. unix.
		err := s.tunnelDriver.StartTunnel(ctx, tempScriptFileName, s.IAPConf.IAPTunnelLaunchWait)
		return err
	})
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

// Cleanup stops the IAP tunnel and cleans up processes.
func (s *StepStartTunnel) Cleanup(state multistep.StateBag) {
	if !s.IAPConf.IAP {
		log.Printf("Skipping cleanup of IAP tunnel; \"iap\" is false.")
		return
	}
	if s.tunnelDriver != nil {
		s.tunnelDriver.StopTunnel()
	}
}

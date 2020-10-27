package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/hashicorp/packer/common/retry"
	"github.com/mitchellh/iochan"
)

const (
	sessionManagerPluginName string = "session-manager-plugin"

	//sessionCommand is the AWS-SDK equivalent to the command you would specify to `aws ssm ...`
	sessionCommand string = "StartSession"
)

type SSMDriverConfig struct {
	SvcClient   ssmiface.SSMAPI
	Region      string
	ProfileName string
	SvcEndpoint string
}

type SSMDriver struct {
	SSMDriverConfig
	session       *ssm.StartSessionOutput
	sessionParams ssm.StartSessionInput
	pluginCmdFunc func(context.Context) error

	retryAfterTermination chan bool
}

func NewSSMDriver(config SSMDriverConfig) *SSMDriver {
	d := SSMDriver{SSMDriverConfig: config}
	return &d
}

// StartSession starts an interactive Systems Manager session with a remote instance via the AWS session-manager-plugin
// This ssm.StartSessionOutput returned by this function can be used for terminating the session manually. If you do
// not wish to manage the session manually calling StopSession on a instance of this driver will terminate the active session
// created from calling StartSession.
func (d *SSMDriver) StartSession(ctx context.Context, input ssm.StartSessionInput) (*ssm.StartSessionOutput, error) {
	log.Printf("Starting PortForwarding session to instance %q", aws.StringValue(input.Target))

	output, err := d.StartSessionWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error encountered in starting session for instance %q: %s", aws.StringValue(input.Target), err)
	}

	d.retryAfterTermination = make(chan bool, 1)
	// Starts go routine that will keep listening to channels and retry the session creation/connection when needed.
	// The log polling process will add data to the channels whenever a retryable error happens to the session or if it's terminated.
	go func(ctx context.Context, driver *SSMDriver, input ssm.StartSessionInput) {
		for {
			select {
			case <-ctx.Done():
				return
			case r, ok := <-driver.retryAfterTermination:
				if !ok {
					return
				}
				if r {
					log.Printf("[DEBUG] Attempting to restablishing SSM connection")
					_, _ = driver.StartSession(ctx, input)
					// End this routine. Another routine will start.
					return
				}
			}
		}
	}(ctx, d, input)

	d.session = output
	d.sessionParams = input

	if d.pluginCmdFunc == nil {
		d.pluginCmdFunc = d.openTunnelForSession
	}

	if err := d.pluginCmdFunc(ctx); err != nil {
		return nil, fmt.Errorf("error encountered in starting session for instance %q: %s", aws.StringValue(input.Target), err)
	}

	return d.session, nil
}

func (d *SSMDriver) StartSessionWithContext(ctx context.Context, input ssm.StartSessionInput) (*ssm.StartSessionOutput, error) {
	var output *ssm.StartSessionOutput
	err := retry.Config{
		ShouldRetry: func(err error) bool { return IsAWSErr(err, "TargetNotConnected", "") },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 200 * time.Millisecond, MaxBackoff: 60 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) (err error) {
		output, err = d.SvcClient.StartSessionWithContext(ctx, &input)
		return err
	})

	return output, err
}

func (d *SSMDriver) openTunnelForSession(ctx context.Context) error {
	args, err := d.Args()
	if err != nil {
		return fmt.Errorf("error encountered validating session details: %s", err)
	}

	cmd := exec.CommandContext(ctx, sessionManagerPluginName, args...)

	// Let's build up our logging
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Create the channels we'll use for data
	stdoutCh := iochan.DelimReader(stdout, '\n')
	stderrCh := iochan.DelimReader(stderr, '\n')

	/* Loop and get all our output
	This particular logger will continue to run through an entire Packer run.
	The decision to continue logging is due to the fact that session-manager-plugin
	doesn't give a good way of knowing if the command failed or was successful other
	than looking at the logs. Seeing as the plugin is updated frequently and that the
	log information is a bit sparse this logger will indefinitely relying on other
	steps to fail if the tunnel is unable to be created. If successful then the user
	will get more information on the tunnel connection when running in a debug mode.
	*/
	go func(ctx context.Context, prefix string) {
		for {
			select {
			case <-ctx.Done():
				return
			case errorOut, ok := <-stderrCh:
				if !ok {
					stderrCh = nil
					break
				}

				if errorOut == "" {
					continue
				}

				log.Printf("[ERROR] %s: %s", prefix, errorOut)
			case output, ok := <-stdoutCh:
				if !ok {
					stdoutCh = nil
					break
				}

				if output == "" {
					continue
				}
				log.Printf("[DEBUG] %s: %s", prefix, output)
			}

			if stdoutCh == nil && stderrCh == nil {
				// It's possible that a remote SSM session was terminated prematurely due to a system reboot, or a termination
				// on the AWS console, or that the SSMAgent service was stopped on the instance. This will try to reconnect
				// until the build eventually fails or hits the ssh_timeout threshold.
				log.Printf("[DEBUG] %s: %s", prefix, "active session has been terminated; stopping all log polling processes.")

				// A call to StopSession will nullify any active sessions so only retry if StopSession has not be called,
				// since StopSession will also close the retryAfterTermination channel
				if d.session != nil {
					d.retryAfterTermination <- true
				}
				return
			}
		}
	}(ctx, sessionManagerPluginName)

	log.Printf("[DEBUG %s] opening session tunnel to instance %q for session %q", sessionManagerPluginName,
		aws.StringValue(d.sessionParams.Target),
		aws.StringValue(d.session.SessionId),
	)

	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("error encountered when calling %s: %s\n", sessionManagerPluginName, err)
		return err
	}

	return nil
}

// StopSession terminates an active Session Manager session
func (d *SSMDriver) StopSession() error {
	if d.session == nil || d.session.SessionId == nil {
		return fmt.Errorf("Unable to find a valid session to instance %q; skipping the termination step",
			aws.StringValue(d.sessionParams.Target))
	}

	// Store session id before we nullify any active sessions
	sid := aws.StringValue(d.session.SessionId)

	// Stop retry polling process to avoid unwanted retries at this point
	d.session = nil
	close(d.retryAfterTermination)
	_, err := d.SvcClient.TerminateSession(&ssm.TerminateSessionInput{SessionId: aws.String(sid)})
	if err != nil {
		err = fmt.Errorf("Error terminating SSM Session %q. Please terminate the session manually: %s", sid, err)
	}

	return err
}

// Args validates the driver inputs before returning an ordered set of arguments to pass to the driver command.
func (d *SSMDriver) Args() ([]string, error) {
	if d.session == nil {
		return nil, fmt.Errorf("an active Amazon SSM Session is required before trying to open a session tunnel")
	}

	// AWS session-manager-plugin requires a valid session be passed in JSON.
	sessionDetails, err := json.Marshal(d.session)
	if err != nil {
		return nil, fmt.Errorf("error encountered in reading session details %s", err)
	}

	// AWS session-manager-plugin requires the parameters used in the session to be passed in JSON as well.
	sessionParameters, err := json.Marshal(d.sessionParams)
	if err != nil {
		return nil, fmt.Errorf("error encountered in reading session parameter details %s", err)
	}

	// Args must be in this order
	args := []string{
		string(sessionDetails),
		d.Region,
		sessionCommand,
		d.ProfileName,
		string(sessionParameters),
		d.SvcEndpoint,
	}

	return args, nil
}

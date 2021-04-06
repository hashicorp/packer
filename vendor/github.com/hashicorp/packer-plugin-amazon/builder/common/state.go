//go:generate struct-markdown
//go:generate mapstructure-to-hcl2 -type AWSPollingConfig
package common

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// StateRefreshFunc is a function type used for StateChangeConf that is
// responsible for refreshing the item being watched for a state change.
//
// It returns three results. `result` is any object that will be returned
// as the final object after waiting for state change. This allows you to
// return the final updated object, for example an EC2 instance after refreshing
// it.
//
// `state` is the latest state of that object. And `err` is any error that
// may have happened while refreshing the state.
type StateRefreshFunc func() (result interface{}, state string, err error)

// StateChangeConf is the configuration struct used for `WaitForState`.
type StateChangeConf struct {
	Pending   []string
	Refresh   StateRefreshFunc
	StepState multistep.StateBag
	Target    string
}

// Following are wrapper functions that use Packer's environment-variables to
// determine retry logic, then call the AWS SDK's built-in waiters.

// Polling configuration for the AWS waiter. Configures the waiter for resources creation or actions like attaching
// volumes or importing image.
//
// HCL2 example:
// ```hcl
// aws_polling {
//	 delay_seconds = 30
//	 max_attempts = 50
// }
// ```
//
// JSON example:
// ```json
// "aws_polling" : {
// 	 "delay_seconds": 30,
// 	 "max_attempts": 50
// }
// ```
type AWSPollingConfig struct {
	// Specifies the maximum number of attempts the waiter will check for resource state.
	// This value can also be set via the AWS_MAX_ATTEMPTS.
	// If both option and environment variable are set, the max_attempts will be considered over the AWS_MAX_ATTEMPTS.
	// If none is set, defaults to AWS waiter default which is 40 max_attempts.
	MaxAttempts int `mapstructure:"max_attempts" required:"false"`
	// Specifies the delay in seconds between attempts to check the resource state.
	// This value can also be set via the AWS_POLL_DELAY_SECONDS.
	// If both option and environment variable are set, the delay_seconds will be considered over the AWS_POLL_DELAY_SECONDS.
	// If none is set, defaults to AWS waiter default which is 15 seconds.
	DelaySeconds int `mapstructure:"delay_seconds" required:"false"`
}

func (w *AWSPollingConfig) WaitUntilAMIAvailable(ctx aws.Context, conn ec2iface.EC2API, imageId string) error {
	imageInput := ec2.DescribeImagesInput{
		ImageIds: []*string{&imageId},
	}

	waitOpts := w.getWaiterOptions()
	if len(waitOpts) == 0 {
		// Bump this default to 30 minutes because the aws default
		// of ten minutes doesn't work for some of our long-running copies.
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(120))
	}
	err := conn.WaitUntilImageAvailableWithContext(
		ctx,
		&imageInput,
		waitOpts...)
	if err != nil {
		if strings.Contains(err.Error(), request.WaiterResourceNotReadyErrorCode) {
			err = fmt.Errorf("Failed with ResourceNotReady error, which can "+
				"have a variety of causes. For help troubleshooting, check "+
				"our docs: "+
				"https://www.packer.io/docs/builders/amazon.html#resourcenotready-error\n"+
				"original error: %s", err.Error())
		}
	}

	return err
}

func (w *AWSPollingConfig) WaitUntilInstanceRunning(ctx aws.Context, conn *ec2.EC2, instanceId string) error {

	instanceInput := ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceId},
	}

	err := conn.WaitUntilInstanceRunningWithContext(
		ctx,
		&instanceInput,
		w.getWaiterOptions()...)
	return err
}

func (w *AWSPollingConfig) WaitUntilInstanceTerminated(ctx aws.Context, conn *ec2.EC2, instanceId string) error {
	instanceInput := ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceId},
	}

	err := conn.WaitUntilInstanceTerminatedWithContext(
		ctx,
		&instanceInput,
		w.getWaiterOptions()...)
	return err
}

// This function works for both requesting and cancelling spot instances.
func (w *AWSPollingConfig) WaitUntilSpotRequestFulfilled(ctx aws.Context, conn *ec2.EC2, spotRequestId string) error {
	spotRequestInput := ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{&spotRequestId},
	}

	err := conn.WaitUntilSpotInstanceRequestFulfilledWithContext(
		ctx,
		&spotRequestInput,
		w.getWaiterOptions()...)
	return err
}

func (w *AWSPollingConfig) WaitUntilVolumeAvailable(ctx aws.Context, conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := conn.WaitUntilVolumeAvailableWithContext(
		ctx,
		&volumeInput,
		w.getWaiterOptions()...)
	return err
}

func (w *AWSPollingConfig) WaitUntilSnapshotDone(ctx aws.Context, conn ec2iface.EC2API, snapshotID string) error {
	snapInput := ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{&snapshotID},
	}

	waitOpts := w.getWaiterOptions()
	if len(waitOpts) == 0 {
		// Bump this default to 30 minutes.
		// Large snapshots can take a long time for the copy to s3
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(120))
	}

	err := conn.WaitUntilSnapshotCompletedWithContext(
		ctx,
		&snapInput,
		waitOpts...)
	return err
}

// Wrappers for our custom AWS waiters

func (w *AWSPollingConfig) WaitUntilVolumeAttached(ctx aws.Context, conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := WaitForVolumeToBeAttached(conn,
		ctx,
		&volumeInput,
		w.getWaiterOptions()...)
	return err
}

func (w *AWSPollingConfig) WaitUntilVolumeDetached(ctx aws.Context, conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := WaitForVolumeToBeDetached(conn,
		ctx,
		&volumeInput,
		w.getWaiterOptions()...)
	return err
}

func (w *AWSPollingConfig) WaitUntilImageImported(ctx aws.Context, conn *ec2.EC2, taskID string) error {
	importInput := ec2.DescribeImportImageTasksInput{
		ImportTaskIds: []*string{&taskID},
	}

	err := WaitForImageToBeImported(conn,
		ctx,
		&importInput,
		w.getWaiterOptions()...)
	return err
}

// Custom waiters using AWS's request.Waiter

func WaitForVolumeToBeAttached(c *ec2.EC2, ctx aws.Context, input *ec2.DescribeVolumesInput, opts ...request.WaiterOption) error {
	w := request.Waiter{
		Name:        "DescribeVolumes",
		MaxAttempts: 40,
		Delay:       request.ConstantWaiterDelay(5 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:    request.SuccessWaiterState,
				Matcher:  request.PathAllWaiterMatch,
				Argument: "Volumes[].Attachments[].State",
				Expected: "attached",
			},
		},
		Logger: c.Config.Logger,
		NewRequest: func(opts []request.Option) (*request.Request, error) {
			var inCpy *ec2.DescribeVolumesInput
			if input != nil {
				tmp := *input
				inCpy = &tmp
			}
			req, _ := c.DescribeVolumesRequest(inCpy)
			req.SetContext(ctx)
			req.ApplyOptions(opts...)
			return req, nil
		},
	}
	w.ApplyOptions(opts...)

	return w.WaitWithContext(ctx)
}

func WaitForVolumeToBeDetached(c *ec2.EC2, ctx aws.Context, input *ec2.DescribeVolumesInput, opts ...request.WaiterOption) error {
	w := request.Waiter{
		Name:        "DescribeVolumes",
		MaxAttempts: 40,
		Delay:       request.ConstantWaiterDelay(5 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:    request.SuccessWaiterState,
				Matcher:  request.PathAllWaiterMatch,
				Argument: "length(Volumes[].Attachments[]) == `0`",
				Expected: true,
			},
		},
		Logger: c.Config.Logger,
		NewRequest: func(opts []request.Option) (*request.Request, error) {
			var inCpy *ec2.DescribeVolumesInput
			if input != nil {
				tmp := *input
				inCpy = &tmp
			}
			req, _ := c.DescribeVolumesRequest(inCpy)
			req.SetContext(ctx)
			req.ApplyOptions(opts...)
			return req, nil
		},
	}
	w.ApplyOptions(opts...)

	return w.WaitWithContext(ctx)
}

func WaitForImageToBeImported(c *ec2.EC2, ctx aws.Context, input *ec2.DescribeImportImageTasksInput, opts ...request.WaiterOption) error {
	w := request.Waiter{
		Name:        "DescribeImages",
		MaxAttempts: 720,
		Delay:       request.ConstantWaiterDelay(5 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:    request.SuccessWaiterState,
				Matcher:  request.PathAllWaiterMatch,
				Argument: "ImportImageTasks[].Status",
				Expected: "completed",
			},
			{
				State:    request.FailureWaiterState,
				Matcher:  request.PathAnyWaiterMatch,
				Argument: "ImportImageTasks[].Status",
				Expected: "deleted",
			},
		},
		Logger: c.Config.Logger,
		NewRequest: func(opts []request.Option) (*request.Request, error) {
			var inCpy *ec2.DescribeImportImageTasksInput
			if input != nil {
				tmp := *input
				inCpy = &tmp
			}
			req, _ := c.DescribeImportImageTasksRequest(inCpy)
			req.SetContext(ctx)
			req.ApplyOptions(opts...)
			return req, nil
		},
	}
	w.ApplyOptions(opts...)

	return w.WaitWithContext(ctx)
}

// This helper function uses the environment variables AWS_TIMEOUT_SECONDS and
// AWS_POLL_DELAY_SECONDS to generate waiter options that can be passed into any
// request.Waiter function. These options will control how many times the waiter
// will retry the request, as well as how long to wait between the retries.

// DEFAULTING BEHAVIOR:
// if AWS_POLL_DELAY_SECONDS is set but the others are not, Packer will set this
// poll delay and use the waiter-specific default

// if AWS_TIMEOUT_SECONDS is set but AWS_MAX_ATTEMPTS is not, Packer will use
// AWS_TIMEOUT_SECONDS and _either_ AWS_POLL_DELAY_SECONDS _or_ 2 if the user has not set AWS_POLL_DELAY_SECONDS, to determine a max number of attempts to make.

// if AWS_TIMEOUT_SECONDS, _and_ AWS_MAX_ATTEMPTS are both set,
// AWS_TIMEOUT_SECONDS will be ignored.

// if AWS_MAX_ATTEMPTS is set but AWS_POLL_DELAY_SECONDS is not, then we will
// use waiter-specific defaults.

type envInfo struct {
	envKey     string
	Val        int
	overridden bool
}

type overridableWaitVars struct {
	awsPollDelaySeconds envInfo
	awsMaxAttempts      envInfo
	awsTimeoutSeconds   envInfo
}

func (w *AWSPollingConfig) getWaiterOptions() []request.WaiterOption {
	envOverrides := getEnvOverrides()

	if w.MaxAttempts != 0 {
		envOverrides.awsMaxAttempts.Val = w.MaxAttempts
		envOverrides.awsMaxAttempts.overridden = true
	}
	if w.DelaySeconds != 0 {
		envOverrides.awsPollDelaySeconds.Val = w.DelaySeconds
		envOverrides.awsPollDelaySeconds.overridden = true
	}

	waitOpts := applyEnvOverrides(envOverrides)
	return waitOpts
}

func getOverride(varInfo envInfo) envInfo {
	override := os.Getenv(varInfo.envKey)
	if override != "" {
		n, err := strconv.Atoi(override)
		if err != nil {
			log.Printf("Invalid %s '%s', using default", varInfo.envKey, override)
		} else {
			varInfo.overridden = true
			varInfo.Val = n
		}
	}

	return varInfo
}
func getEnvOverrides() overridableWaitVars {
	// Load env vars from environment.
	envValues := overridableWaitVars{
		envInfo{"AWS_POLL_DELAY_SECONDS", 2, false},
		envInfo{"AWS_MAX_ATTEMPTS", 0, false},
		envInfo{"AWS_TIMEOUT_SECONDS", 0, false},
	}

	envValues.awsMaxAttempts = getOverride(envValues.awsMaxAttempts)
	envValues.awsPollDelaySeconds = getOverride(envValues.awsPollDelaySeconds)
	envValues.awsTimeoutSeconds = getOverride(envValues.awsTimeoutSeconds)

	return envValues
}

func (w *AWSPollingConfig) LogEnvOverrideWarnings() {
	pollDelayEnv := os.Getenv("AWS_POLL_DELAY_SECONDS")
	timeoutSecondsEnv := os.Getenv("AWS_TIMEOUT_SECONDS")
	maxAttemptsEnv := os.Getenv("AWS_MAX_ATTEMPTS")

	maxAttemptsIsSet := maxAttemptsEnv != "" || w.MaxAttempts != 0
	timeoutSecondsIsSet := timeoutSecondsEnv != ""
	pollDelayIsSet := pollDelayEnv != "" || w.DelaySeconds != 0

	if maxAttemptsIsSet && timeoutSecondsIsSet {
		warning := fmt.Sprintf("[WARNING] (aws): AWS_MAX_ATTEMPTS and " +
			"AWS_TIMEOUT_SECONDS are both set. Packer will use " +
			"AWS_MAX_ATTEMPTS and discard AWS_TIMEOUT_SECONDS.")
		if !pollDelayIsSet {
			warning = fmt.Sprintf("%s  Since you have not set the poll delay, "+
				"Packer will default to a 2-second delay.", warning)
		}
		log.Printf(warning)
	} else if timeoutSecondsIsSet {
		log.Printf("[WARNING] (aws): env var AWS_TIMEOUT_SECONDS is " +
			"deprecated in favor of AWS_MAX_ATTEMPTS env or aws_polling_max_attempts config option. " +
			"If you have not explicitly set AWS_POLL_DELAY_SECONDS env or aws_polling_delay_seconds config option, " +
			"we are defaulting to a poll delay of 2 seconds, regardless of the AWS waiter's default.")
	}
	if !maxAttemptsIsSet && !timeoutSecondsIsSet && !pollDelayIsSet {
		log.Printf("[INFO] (aws): No AWS timeout and polling overrides have been set. " +
			"Packer will default to waiter-specific delays and timeouts. If you would " +
			"like to customize the length of time between retries and max " +
			"number of retries you may do so by setting the environment " +
			"variables AWS_POLL_DELAY_SECONDS and AWS_MAX_ATTEMPTS or the " +
			"configuration options aws_polling_delay_seconds and aws_polling_max_attempts " +
			"to your desired values.")
	}
}

func applyEnvOverrides(envOverrides overridableWaitVars) []request.WaiterOption {
	waitOpts := make([]request.WaiterOption, 0)
	// If user has set poll delay seconds, overwrite it. If user has NOT,
	// default to a poll delay of 2 seconds
	if envOverrides.awsPollDelaySeconds.overridden {
		delaySeconds := request.ConstantWaiterDelay(time.Duration(envOverrides.awsPollDelaySeconds.Val) * time.Second)
		waitOpts = append(waitOpts, request.WithWaiterDelay(delaySeconds))
	}

	// If user has set max attempts, overwrite it. If user hasn't set max
	// attempts, default to whatever the waiter has set as a default.
	if envOverrides.awsMaxAttempts.overridden {
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(envOverrides.awsMaxAttempts.Val))
	} else if envOverrides.awsTimeoutSeconds.overridden {
		maxAttempts := envOverrides.awsTimeoutSeconds.Val / envOverrides.awsPollDelaySeconds.Val
		// override the delay so we can get the timeout right
		if !envOverrides.awsPollDelaySeconds.overridden {
			delaySeconds := request.ConstantWaiterDelay(time.Duration(envOverrides.awsPollDelaySeconds.Val) * time.Second)
			waitOpts = append(waitOpts, request.WithWaiterDelay(delaySeconds))
		}
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(maxAttempts))
	}

	return waitOpts
}

package common

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/packer/helper/multistep"
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
// determing retry logic, then call the AWS SDK's built-in waiters.

func WaitUntilAMIAvailable(ctx aws.Context, conn *ec2.EC2, imageId string) error {
	imageInput := ec2.DescribeImagesInput{
		ImageIds: []*string{&imageId},
	}

	waitOpts := getWaiterOptions()
	if len(waitOpts) == 0 {
		// Bump this default to 30 minutes because the aws default
		// of ten minutes doesn't work for some of our long-running copies.
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(120))
	}
	err := conn.WaitUntilImageAvailableWithContext(
		ctx,
		&imageInput,
		waitOpts...)
	return err
}

func WaitUntilInstanceTerminated(ctx aws.Context, conn *ec2.EC2, instanceId string) error {

	instanceInput := ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceId},
	}

	err := conn.WaitUntilInstanceTerminatedWithContext(
		ctx,
		&instanceInput,
		getWaiterOptions()...)
	return err
}

// This function works for both requesting and cancelling spot instances.
func WaitUntilSpotRequestFulfilled(ctx aws.Context, conn *ec2.EC2, spotRequestId string) error {
	spotRequestInput := ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{&spotRequestId},
	}

	err := conn.WaitUntilSpotInstanceRequestFulfilledWithContext(
		ctx,
		&spotRequestInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilVolumeAvailable(ctx aws.Context, conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := conn.WaitUntilVolumeAvailableWithContext(
		ctx,
		&volumeInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilSnapshotDone(ctx aws.Context, conn *ec2.EC2, snapshotID string) error {
	snapInput := ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{&snapshotID},
	}

	err := conn.WaitUntilSnapshotCompletedWithContext(
		ctx,
		&snapInput,
		getWaiterOptions()...)
	return err
}

// Wrappers for our custom AWS waiters

func WaitUntilVolumeAttached(ctx aws.Context, conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := WaitForVolumeToBeAttached(conn,
		ctx,
		&volumeInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilVolumeDetached(ctx aws.Context, conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := WaitForVolumeToBeDetached(conn,
		ctx,
		&volumeInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilImageImported(ctx aws.Context, conn *ec2.EC2, taskID string) error {
	importInput := ec2.DescribeImportImageTasksInput{
		ImportTaskIds: []*string{&taskID},
	}

	err := WaitForImageToBeImported(conn,
		ctx,
		&importInput,
		getWaiterOptions()...)
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

func getWaiterOptions() []request.WaiterOption {
	envOverrides := getEnvOverrides()
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
	}

	if envOverrides.awsMaxAttempts.overridden && envOverrides.awsTimeoutSeconds.overridden {
		log.Printf("WARNING: AWS_MAX_ATTEMPTS and AWS_TIMEOUT_SECONDS are" +
			" both set. Packer will be using AWS_MAX_ATTEMPTS and discarding " +
			"AWS_TIMEOUT_SECONDS. If you have not set AWS_POLL_DELAY_SECONDS, " +
			"Packer will default to a 2 second poll delay.")
	} else if envOverrides.awsTimeoutSeconds.overridden {
		log.Printf("DEPRECATION WARNING: env var AWS_TIMEOUT_SECONDS is " +
			"deprecated in favor of AWS_MAX_ATTEMPTS. If you have not " +
			"explicitly set AWS_POLL_DELAY_SECONDS, we are defaulting to a " +
			"poll delay of 2 seconds, regardless of the AWS waiter's default.")
		maxAttempts := envOverrides.awsTimeoutSeconds.Val / envOverrides.awsPollDelaySeconds.Val
		// override the delay so we can get the timeout right
		if !envOverrides.awsPollDelaySeconds.overridden {
			delaySeconds := request.ConstantWaiterDelay(time.Duration(envOverrides.awsPollDelaySeconds.Val) * time.Second)
			waitOpts = append(waitOpts, request.WithWaiterDelay(delaySeconds))
		}
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(maxAttempts))
	}
	if len(waitOpts) == 0 {
		log.Printf("No AWS timeout and polling overrides have been set. " +
			"Packer will default to waiter-specific delays and timeouts. If you would " +
			"like to customize the length of time between retries and max " +
			"number of retries you may do so by setting the environment " +
			"variables AWS_POLL_DELAY_SECONDS and AWS_MAX_ATTEMPTS to your " +
			"desired values.")
	}

	return waitOpts
}

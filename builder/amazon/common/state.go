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

	err := conn.WaitUntilImageAvailableWithContext(
		ctx,
		&imageInput,
		getWaiterOptions()...)
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
	return w.WaitWithContext(ctx)
}

func WaitForImageToBeImported(c *ec2.EC2, ctx aws.Context, input *ec2.DescribeImportImageTasksInput, opts ...request.WaiterOption) error {
	w := request.Waiter{
		Name:        "DescribeImages",
		MaxAttempts: 300,
		Delay:       request.ConstantWaiterDelay(5 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:    request.SuccessWaiterState,
				Matcher:  request.PathAllWaiterMatch,
				Argument: "ImportImageTasks[].Status",
				Expected: "completed",
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

func getWaiterOptions() []request.WaiterOption {
	waitOpts := make([]request.WaiterOption, 0)
	// If user has set poll delay seconds, overwrite it. If user has NOT,
	// default to a poll delay of 2 seconds
	delayOverridden, delay := getEnvOverrides(2, "AWS_POLL_DELAY_SECONDS")
	if delayOverridden {
		delaySeconds := request.ConstantWaiterDelay(time.Duration(delay) * time.Second)
		waitOpts = append(waitOpts, request.WithWaiterDelay(delaySeconds))
	}

	// If user has set max attempts, overwrite it. If user hasn't set max
	// attempts, default to whatever the waiter has set as a default.
	maxAttemptsOverridden, maxAttempts := getEnvOverrides(0, "AWS_MAX_ATTEMPTS")
	if maxAttemptsOverridden {
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(maxAttempts))
	}

	timeoutOverridden, timeoutSeconds := getEnvOverrides(300, "AWS_TIMEOUT_SECONDS")
	if maxAttemptsOverridden {
		log.Printf("WARNING: AWS_MAX_ATTEMPTS and AWS_TIMEOUT_SECONDS are" +
			" both set. Packer will be using AWS_MAX_ATTEMPTS and discarding " +
			"AWS_TIMEOUT_SECONDS. If you have not set AWS_POLL_DELAY_SECONDS, " +
			"Packer will default to a 2 second poll delay.")
	} else if timeoutOverridden {
		log.Printf("DEPRECATION WARNING: env var AWS_TIMEOUT_SECONDS is " +
			"deprecated in favor of AWS_MAX_ATTEMPTS. If you have not " +
			"explicitly set AWS_POLL_DELAY_SECONDS, we are defaulting to a " +
			"poll delay of 2 seconds, regardless of the AWS waiter's default.")
		maxAttempts := timeoutSeconds / delay
		// override the delay so we can get the timeout right
		if !delayOverridden {
			delaySeconds := request.ConstantWaiterDelay(time.Duration(delay) * time.Second)
			waitOpts = append(waitOpts, request.WithWaiterDelay(delaySeconds))
		}
		waitOpts = append(waitOpts, request.WithWaiterMaxAttempts(maxAttempts))
	}

	return waitOpts
}

func getEnvOverrides(defaultValue int, envVarName string) (bool, int) {
	// "AWS_POLL_DELAY_SECONDS"
	retVal := defaultValue
	overridden := false
	override := os.Getenv(envVarName)
	if override != "" {
		n, err := strconv.Atoi(override)
		if err != nil {
			log.Printf("Invalid %s '%s', using default", envVarName, override)
		} else {
			overridden = true
			retVal = n
		}
	}

	log.Printf("Using %ds for %s", retVal, envVarName)
	return overridden, retVal
}

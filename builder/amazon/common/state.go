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

func WaitUntilAMIAvailable(conn *ec2.EC2, imageId string) error {
	imageInput := ec2.DescribeImagesInput{
		ImageIds: []*string{&imageId},
	}

	err := conn.WaitUntilImageAvailableWithContext(aws.BackgroundContext(),
		&imageInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilInstanceTerminated(conn *ec2.EC2, instanceId string) error {

	instanceInput := ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceId},
	}

	err := conn.WaitUntilInstanceTerminatedWithContext(aws.BackgroundContext(),
		&instanceInput,
		getWaiterOptions()...)
	return err
}

// This function works for both requesting and cancelling spot instances.
func WaitUntilSpotRequestFulfilled(conn *ec2.EC2, spotRequestId string) error {
	spotRequestInput := ec2.DescribeSpotInstanceRequestsInput{
		SpotInstanceRequestIds: []*string{&spotRequestId},
	}

	err := conn.WaitUntilSpotInstanceRequestFulfilledWithContext(aws.BackgroundContext(),
		&spotRequestInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilVolumeAvailable(conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := conn.WaitUntilVolumeAvailableWithContext(aws.BackgroundContext(),
		&volumeInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilSnapshotDone(conn *ec2.EC2, snapshotID string) error {
	snapInput := ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{&snapshotID},
	}

	err := conn.WaitUntilSnapshotCompletedWithContext(aws.BackgroundContext(),
		&snapInput,
		getWaiterOptions()...)
	return err
}

// Wrappers for our custom AWS waiters

func WaitUntilVolumeAttached(conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := WaitForVolumeToBeAttached(conn,
		aws.BackgroundContext(),
		&volumeInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilVolumeDetached(conn *ec2.EC2, volumeId string) error {
	volumeInput := ec2.DescribeVolumesInput{
		VolumeIds: []*string{&volumeId},
	}

	err := WaitForVolumeToBeAttached(conn,
		aws.BackgroundContext(),
		&volumeInput,
		getWaiterOptions()...)
	return err
}

func WaitUntilImageImported(conn *ec2.EC2, taskID string) error {
	importInput := ec2.DescribeImportImageTasksInput{
		ImportTaskIds: []*string{&taskID},
	}

	err := WaitForImageToBeImported(conn,
		aws.BackgroundContext(),
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
				Argument: "Volumes[].State",
				Expected: "attached",
			},
			{
				State:    request.FailureWaiterState,
				Matcher:  request.PathAnyWaiterMatch,
				Argument: "Volumes[].State",
				Expected: "deleted",
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
				Argument: "Volumes[].State",
				Expected: "detached",
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
		MaxAttempts: 40,
		Delay:       request.ConstantWaiterDelay(5 * time.Second),
		Acceptors: []request.WaiterAcceptor{
			{
				State:    request.SuccessWaiterState,
				Matcher:  request.PathAllWaiterMatch,
				Argument: "ImportImageTasks[].State",
				Expected: "completed",
			},
			{
				State:    request.RetryWaiterState,
				Matcher:  request.ErrorWaiterMatch,
				Expected: "InvalidConversionTaskId",
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
func getWaiterOptions() []request.WaiterOption {
	// use env vars to read in the wait delay and the max amount of time to wait
	delay := SleepSeconds()
	timeoutSeconds := TimeoutSeconds()
	// AWS sdk uses max attempts instead of a timeout; convert timeout into
	// max attempts
	maxAttempts := timeoutSeconds / delay
	delaySeconds := request.ConstantWaiterDelay(time.Duration(delay) * time.Second)

	return []request.WaiterOption{
		request.WithWaiterDelay(delaySeconds),
		request.WithWaiterMaxAttempts(maxAttempts)}
}

// Returns 300 seconds (5 minutes) by default
// Some AWS operations, like copying an AMI to a distant region, take a very long time
// Allow user to override with AWS_TIMEOUT_SECONDS environment variable
func TimeoutSeconds() (seconds int) {
	seconds = 300

	override := os.Getenv("AWS_TIMEOUT_SECONDS")
	if override != "" {
		n, err := strconv.Atoi(override)
		if err != nil {
			log.Printf("Invalid timeout seconds '%s', using default", override)
		} else {
			seconds = n
		}
	}

	log.Printf("Allowing %ds to complete (change with AWS_TIMEOUT_SECONDS)", seconds)
	return seconds
}

// Returns 2 seconds by default
// AWS async operations sometimes takes long times, if there are multiple parallel builds,
// polling at 2 second frequency will exceed the request limit. Allow 2 seconds to be
// overwritten with AWS_POLL_DELAY_SECONDS
func SleepSeconds() (seconds int) {
	seconds = 2

	override := os.Getenv("AWS_POLL_DELAY_SECONDS")
	if override != "" {
		n, err := strconv.Atoi(override)
		if err != nil {
			log.Printf("Invalid sleep seconds '%s', using default", override)
		} else {
			seconds = n
		}
	}

	log.Printf("Using %ds as polling delay (change with AWS_POLL_DELAY_SECONDS)", seconds)
	return seconds
}

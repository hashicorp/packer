package common

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/multistep"
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

// AMIStateRefreshFunc returns a StateRefreshFunc that is used to watch
// an AMI for state changes.
func AMIStateRefreshFunc(conn *ec2.EC2, imageId string) StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: []*string{&imageId},
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidAMIID.NotFound" {
				// Set this to nil as if we didn't find anything.
				resp = nil
			} else if isTransientNetworkError(err) {
				// Transient network error, treat it as if we didn't find anything
				resp = nil
			} else {
				log.Printf("Error on AMIStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.Images) == 0 {
			// Sometimes AWS has consistency issues and doesn't see the
			// AMI. Return an empty state.
			return nil, "", nil
		}

		i := resp.Images[0]
		return i, *i.State, nil
	}
}

// InstanceStateRefreshFunc returns a StateRefreshFunc that is used to watch
// an EC2 instance.
func InstanceStateRefreshFunc(conn *ec2.EC2, instanceId string) StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeInstances(&ec2.DescribeInstancesInput{
			InstanceIds: []*string{&instanceId},
		})
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidInstanceID.NotFound" {
				// Set this to nil as if we didn't find anything.
				resp = nil
			} else if isTransientNetworkError(err) {
				// Transient network error, treat it as if we didn't find anything
				resp = nil
			} else {
				log.Printf("Error on InstanceStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.Reservations) == 0 || len(resp.Reservations[0].Instances) == 0 {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		i := resp.Reservations[0].Instances[0]
		return i, *i.State.Name, nil
	}
}

// SpotRequestStateRefreshFunc returns a StateRefreshFunc that is used to watch
// a spot request for state changes.
func SpotRequestStateRefreshFunc(conn *ec2.EC2, spotRequestId string) StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeSpotInstanceRequests(&ec2.DescribeSpotInstanceRequestsInput{
			SpotInstanceRequestIds: []*string{&spotRequestId},
		})

		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidSpotInstanceRequestID.NotFound" {
				// Set this to nil as if we didn't find anything.
				resp = nil
			} else if isTransientNetworkError(err) {
				// Transient network error, treat it as if we didn't find anything
				resp = nil
			} else {
				log.Printf("Error on SpotRequestStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.SpotInstanceRequests) == 0 {
			// Sometimes AWS has consistency issues and doesn't see the
			// SpotRequest. Return an empty state.
			return nil, "", nil
		}

		i := resp.SpotInstanceRequests[0]
		return i, *i.State, nil
	}
}

func ImportImageRefreshFunc(conn *ec2.EC2, importTaskId string) StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeImportImageTasks(&ec2.DescribeImportImageTasksInput{
			ImportTaskIds: []*string{
				&importTaskId,
			},
		},
		)
		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && strings.HasPrefix(ec2err.Code(), "InvalidConversionTaskId") {
				resp = nil
			} else if isTransientNetworkError(err) {
				resp = nil
			} else {
				log.Printf("Error on ImportImageRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.ImportImageTasks) == 0 {
			return nil, "", nil
		}

		i := resp.ImportImageTasks[0]
		return i, *i.Status, nil
	}
}

// WaitForState watches an object and waits for it to achieve a certain
// state.
func WaitForState(conf *StateChangeConf) (i interface{}, err error) {
	log.Printf("Waiting for state to become: %s", conf.Target)

	sleepSeconds := 2
	maxTicks := int(TimeoutSeconds()/sleepSeconds) + 1
	notfoundTick := 0

	for {
		var currentState string
		i, currentState, err = conf.Refresh()
		if err != nil {
			return
		}

		if i == nil {
			// If we didn't find the resource, check if we have been
			// not finding it for awhile, and if so, report an error.
			notfoundTick += 1
			if notfoundTick > maxTicks {
				return nil, errors.New("couldn't find resource")
			}
		} else {
			// Reset the counter for when a resource isn't found
			notfoundTick = 0

			if currentState == conf.Target {
				return
			}

			if conf.StepState != nil {
				if _, ok := conf.StepState.GetOk(multistep.StateCancelled); ok {
					return nil, errors.New("interrupted")
				}
			}

			found := false
			for _, allowed := range conf.Pending {
				if currentState == allowed {
					found = true
					break
				}
			}

			if !found {
				err := fmt.Errorf("unexpected state '%s', wanted target '%s'", currentState, conf.Target)
				return nil, err
			}
		}

		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}
}

func isTransientNetworkError(err error) bool {
	if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
		return true
	}

	return false
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

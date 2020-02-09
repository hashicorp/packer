package cvm

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// DefaultWaitForInterval is sleep interval when wait statue
const DefaultWaitForInterval = 5

// WaitForInstance wait for instance reaches statue
func WaitForInstance(client *cvm.Client, instanceId string, status string, timeout int) error {
	ctx := context.TODO()
	req := cvm.NewDescribeInstancesRequest()
	req.InstanceIds = []*string{&instanceId}

	for {
		var resp *cvm.DescribeInstancesResponse
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.DescribeInstances(req)
			return e
		})
		if err != nil {
			return err
		}
		if *resp.Response.TotalCount == 0 {
			return fmt.Errorf("instance(%s) not exist", instanceId)
		}
		if *resp.Response.InstanceSet[0].InstanceState == status {
			break
		}
		time.Sleep(DefaultWaitForInterval * time.Second)
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return fmt.Errorf("wait instance(%s) status(%s) timeout", instanceId, status)
		}
	}

	return nil
}

// WaitForImageReady wait for image reaches statue
func WaitForImageReady(client *cvm.Client, imageName string, status string, timeout int) error {
	ctx := context.TODO()
	req := cvm.NewDescribeImagesRequest()
	FILTER_IMAGE_NAME := "image-name"
	req.Filters = []*cvm.Filter{
		{
			Name:   &FILTER_IMAGE_NAME,
			Values: []*string{&imageName},
		},
	}

	for {
		var resp *cvm.DescribeImagesResponse
		err := Retry(ctx, func(ctx context.Context) error {
			var e error
			resp, e = client.DescribeImages(req)
			return e
		})
		if err != nil {
			return err
		}
		find := false
		for _, image := range resp.Response.ImageSet {
			if *image.ImageName == imageName && *image.ImageState == status {
				find = true
				break
			}
		}
		if find {
			break
		}
		time.Sleep(DefaultWaitForInterval * time.Second)
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return fmt.Errorf("wait image(%s) status(%s) timeout", imageName, status)
		}
	}

	return nil
}

// CheckResourceIdFormat check resource id format
func CheckResourceIdFormat(resource string, id string) bool {
	regex := regexp.MustCompile(fmt.Sprintf("%s-[0-9a-z]{8}$", resource))
	if !regex.MatchString(id) {
		return false
	}
	return true
}

// SSHHost returns a function that can be given to the SSH communicator
func SSHHost(pubilcIp bool) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		instance := state.Get("instance").(*cvm.Instance)
		if pubilcIp {
			return *instance.PublicIpAddresses[0], nil
		} else {
			return *instance.PrivateIpAddresses[0], nil
		}
	}
}

// Retry do retry on api request
func Retry(ctx context.Context, fn func(context.Context) error) error {
	return retry.Config{
		Tries: 30,
		ShouldRetry: func(err error) bool {
			e, ok := err.(*errors.TencentCloudSDKError)
			if !ok {
				return false
			}
			if e.Code == "ClientError.NetworkError" || e.Code == "ClientError.HttpStatusCodeError" ||
				e.Code == "InvalidKeyPair.NotSupported" || e.Code == "InvalidInstance.NotSupported" ||
				strings.Contains(e.Code, "RequestLimitExceeded") || strings.Contains(e.Code, "InternalError") ||
				strings.Contains(e.Code, "ResourceInUse") || strings.Contains(e.Code, "ResourceBusy") {
				return true
			}
			return false
		},
		RetryDelay: (&retry.Backoff{
			InitialBackoff: 1 * time.Second,
			MaxBackoff:     5 * time.Second,
			Multiplier:     2,
		}).Linear,
	}.Run(ctx, fn)
}

// SayClean tell you clean module message
func SayClean(state multistep.StateBag, module string) {
	_, halted := state.GetOk(multistep.StateHalted)
	_, cancelled := state.GetOk(multistep.StateCancelled)
	if halted {
		Say(state, fmt.Sprintf("Deleting %s because of error...", module), "")
	} else if cancelled {
		Say(state, fmt.Sprintf("Deleting %s because of cancellation...", module), "")
	} else {
		Say(state, fmt.Sprintf("Cleaning up %s...", module), "")
	}
}

// Say tell you a message
func Say(state multistep.StateBag, message, prefix string) {
	if prefix != "" {
		message = fmt.Sprintf("%s: %s", prefix, message)
	}

	if strings.HasPrefix(message, "Trying to") {
		message += "..."
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Say(message)
}

// Message print a message
func Message(state multistep.StateBag, message, prefix string) {
	if prefix != "" {
		message = fmt.Sprintf("%s: %s", prefix, message)
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Message(message)
}

// Error print error message
func Error(state multistep.StateBag, err error, prefix string) {
	if prefix != "" {
		err = fmt.Errorf("%s: %s", prefix, err)
	}

	ui := state.Get("ui").(packer.Ui)
	ui.Error(err.Error())
}

// Halt print error message and exit
func Halt(state multistep.StateBag, err error, prefix string) multistep.StepAction {
	Error(state, err, prefix)
	state.Put("error", err)

	return multistep.ActionHalt
}

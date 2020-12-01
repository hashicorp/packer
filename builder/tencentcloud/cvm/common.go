package cvm

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer/packer-plugin-sdk/retry"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// DefaultWaitForInterval is sleep interval when wait statue
const DefaultWaitForInterval = 5

// WaitForInstance wait for instance reaches statue
func WaitForInstance(ctx context.Context, client *cvm.Client, instanceId string, status string, timeout int) error {
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
		if *resp.Response.InstanceSet[0].InstanceState == status &&
			(resp.Response.InstanceSet[0].LatestOperationState == nil || *resp.Response.InstanceSet[0].LatestOperationState != "OPERATING") {
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
func WaitForImageReady(ctx context.Context, client *cvm.Client, imageName string, status string, timeout int) error {
	for {
		image, err := GetImageByName(ctx, client, imageName)
		if err != nil {
			return err
		}

		if image != nil && *image.ImageState == status {
			return nil
		}

		time.Sleep(DefaultWaitForInterval * time.Second)
		timeout = timeout - DefaultWaitForInterval
		if timeout <= 0 {
			return fmt.Errorf("wait image(%s) status(%s) timeout", imageName, status)
		}
	}
}

// GetImageByName get image by image name
func GetImageByName(ctx context.Context, client *cvm.Client, imageName string) (*cvm.Image, error) {
	req := cvm.NewDescribeImagesRequest()
	req.Filters = []*cvm.Filter{
		{
			Name:   common.StringPtr("image-name"),
			Values: []*string{&imageName},
		},
	}

	var resp *cvm.DescribeImagesResponse
	err := Retry(ctx, func(ctx context.Context) error {
		var e error
		resp, e = client.DescribeImages(req)
		return e
	})
	if err != nil {
		return nil, err
	}

	if *resp.Response.TotalCount > 0 {
		for _, image := range resp.Response.ImageSet {
			if *image.ImageName == imageName {
				return image, nil
			}
		}
	}

	return nil, nil
}

// NewCvmClient returns a new cvm client
func NewCvmClient(secretId, secretKey, region string) (client *cvm.Client, err error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 300
	cpf.Language = "en-US"

	credential := common.NewCredential(secretId, secretKey)
	client, err = cvm.NewClient(credential, region, cpf)

	return
}

// NewVpcClient returns a new vpc client
func NewVpcClient(secretId, secretKey, region string) (client *vpc.Client, err error) {
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 300
	cpf.Language = "en-US"

	credential := common.NewCredential(secretId, secretKey)
	client, err = vpc.NewClient(credential, region, cpf)

	return
}

// CheckResourceIdFormat check resource id format
func CheckResourceIdFormat(resource string, id string) bool {
	regex := regexp.MustCompile(fmt.Sprintf("%s-[0-9a-z]{8}$", resource))
	return regex.MatchString(id)
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
		Tries: 60,
		ShouldRetry: func(err error) bool {
			e, ok := err.(*errors.TencentCloudSDKError)
			if !ok {
				return false
			}
			if e.Code == "ClientError.NetworkError" || e.Code == "ClientError.HttpStatusCodeError" ||
				e.Code == "InvalidKeyPair.NotSupported" || e.Code == "InvalidParameterValue.KeyPairNotSupported" ||
				e.Code == "InvalidInstance.NotSupported" || e.Code == "OperationDenied.InstanceOperationInProgress" ||
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

	ui := state.Get("ui").(packersdk.Ui)
	ui.Say(message)
}

// Message print a message
func Message(state multistep.StateBag, message, prefix string) {
	if prefix != "" {
		message = fmt.Sprintf("%s: %s", prefix, message)
	}

	ui := state.Get("ui").(packersdk.Ui)
	ui.Message(message)
}

// Error print error message
func Error(state multistep.StateBag, err error, prefix string) {
	if prefix != "" {
		err = fmt.Errorf("%s: %s", prefix, err)
	}

	ui := state.Get("ui").(packersdk.Ui)
	ui.Error(err.Error())
}

// Halt print error message and exit
func Halt(state multistep.StateBag, err error, prefix string) multistep.StepAction {
	Error(state, err, prefix)
	state.Put("error", err)

	return multistep.ActionHalt
}

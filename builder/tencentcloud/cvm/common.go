package cvm

import (
	"regexp"
	"fmt"

	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"time"
)

func CheckResourceIdFormat(resource string, id string) bool {
	regex := regexp.MustCompile(fmt.Sprintf("%s-[0-9a-z]{8}$", resource))
	if !regex.MatchString(id) {
		return false
	}
	return true
}

func MessageClean(state multistep.StateBag, module string) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	ui := state.Get("ui").(packer.Ui)

	if cancelled || halted {
		ui.Say(fmt.Sprintf("Deleting %s because of cancellation or error...", module))
	} else {
		ui.Say(fmt.Sprintf("Cleaning up '%s'", module))
	}

}

const DefaultWaitForInterval = 5

func WaitForInstance(client *cvm.Client, instanceId string, status string, timeout int) error {
	req := cvm.NewDescribeInstancesRequest()
	req.InstanceIds = []*string{&instanceId}
	for {
		resp, err := client.DescribeInstances(req)
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

func WaitForImageReady(client *cvm.Client, imageName string, status string, timeout int) error {
	req := cvm.NewDescribeImagesRequest()
	FILTER_IMAGE_NAME := "image-name"
	req.Filters = []*cvm.Filter{
		{
			Name: &FILTER_IMAGE_NAME,
			Values: []*string{&imageName},
		},
	}
	for {
		resp, err := client.DescribeImages(req)
		if err != nil {
			return err
		}
		find := false
		for _, image := range resp.Response.ImageSet {
			if *image.ImageName == imageName && *image.ImageState == status{
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
			return fmt.Errorf("wait image(%s) ready timeout", imageName)
		}
	}
	return nil
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
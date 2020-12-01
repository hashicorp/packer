package yandex

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type stepInstanceInfo struct{}

func (s *stepInstanceInfo) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	sdk := state.Get("sdk").(*ycsdk.SDK)
	ui := state.Get("ui").(packersdk.Ui)
	c := state.Get("config").(*Config)
	instanceID := state.Get("instance_id").(string)

	ui.Say(fmt.Sprintf("Waiting for instance with id %s to become active...", instanceID))

	ctx, cancel := context.WithTimeout(ctx, c.StateTimeout)
	defer cancel()

	instance, err := sdk.Compute().Instance().Get(ctx, &compute.GetInstanceRequest{
		InstanceId: instanceID,
		View:       compute.InstanceView_FULL,
	})
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Error retrieving instance data: %s", err))
	}

	instanceIP, err := getInstanceIPAddress(c, instance)
	if err != nil {
		return stepHaltWithError(state, fmt.Errorf("Failed to find instance ip address: %s", err))
	}

	state.Put("instance_ip", instanceIP)
	ui.Message(fmt.Sprintf("Detected instance IP: %s", instanceIP))

	return multistep.ActionContinue
}

func getInstanceIPAddress(c *Config, instance *compute.Instance) (address string, err error) {
	// Instance could have several network interfaces with different configuration each
	// Get all possible addresses for instance
	addrIPV4Internal, addrIPV4External, addrIPV6Addr, err := instanceAddresses(instance)
	if err != nil {
		return "", err
	}

	if c.UseIPv6 {
		if addrIPV6Addr != "" {
			return "[" + addrIPV6Addr + "]", nil
		}
		return "", errors.New("instance has no one IPv6 address")
	}

	if c.UseInternalIP {
		if addrIPV4Internal != "" {
			return addrIPV4Internal, nil
		}
		return "", errors.New("instance has no one IPv4 internal address")
	}
	if addrIPV4External != "" {
		return addrIPV4External, nil
	}
	return "", errors.New("instance has no one IPv4 external address")
}

func instanceAddresses(instance *compute.Instance) (ipV4Int, ipV4Ext, ipV6 string, err error) {
	if len(instance.NetworkInterfaces) == 0 {
		return "", "", "", errors.New("No one network interface found for an instance")
	}

	var ipV4IntFound, ipV4ExtFound, ipV6Found bool
	for _, iface := range instance.NetworkInterfaces {
		if !ipV6Found && iface.PrimaryV6Address != nil {
			ipV6 = iface.PrimaryV6Address.Address
			ipV6Found = true
		}

		if !ipV4IntFound && iface.PrimaryV4Address != nil {
			ipV4Int = iface.PrimaryV4Address.Address
			ipV4IntFound = true

			if !ipV4ExtFound && iface.PrimaryV4Address.OneToOneNat != nil {
				ipV4Ext = iface.PrimaryV4Address.OneToOneNat.Address
				ipV4ExtFound = true
			}
		}

		if ipV6Found && ipV4IntFound && ipV4ExtFound {
			break
		}
	}

	if !ipV4IntFound {
		// internal ipV4 address always should present
		return "", "", "", errors.New("No IPv4 internal address found. Bug?")
	}

	return
}

func (s *stepInstanceInfo) Cleanup(state multistep.StateBag) {
	// no cleanup
}

package uhost

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/common/retry"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/ucloud/ucloud-sdk-go/services/uhost"
	"github.com/ucloud/ucloud-sdk-go/ucloud"
	"math/rand"
	"strings"
	"time"
)

type stepCreateInstance struct {
	Region        string
	Zone          string
	InstanceType  string
	InstanceName  string
	BootDiskType  string
	SourceImageId string
	UsePrivateIp  bool

	instanceId string
}

func (s *stepCreateInstance) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Creating Instance...")
	req, err := s.buildCreateInstanceRequest(state)
	if err != nil {
		return halt(state, err, "")
	}

	resp, err := conn.CreateUHostInstance(req)
	if err != nil {
		return halt(state, err, "Error on creating instance")
	}
	instanceId := resp.UHostIds[0]

	err = retry.Config{
		Tries: 20,
		ShouldRetry: func(err error) bool {
			return isExpectedStateError(err)
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		inst, err := client.describeUHostById(instanceId)
		if err != nil {
			return err
		}
		if inst == nil || inst.State != "Running" {
			return newExpectedStateError("instance", instanceId)
		}

		return nil
	})

	if err != nil {
		return halt(state, err, "Error on waiting for instance to available")
	}

	ui.Message(fmt.Sprintf("Create instance %q complete", instanceId))
	instance, err := client.describeUHostById(instanceId)
	if err != nil {
		return halt(state, err, "")
	}

	s.instanceId = instanceId
	state.Put("instance", instance)

	if instance.BootDiskState == "Initializing" {
		ui.Say(fmt.Sprintf("Waiting for boot disk of instance initialized when boot_disk_type is %q", s.BootDiskType))

		err = retry.Config{
			Tries: 200,
			ShouldRetry: func(err error) bool {
				return isExpectedStateError(err)
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 12 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			inst, err := client.describeUHostById(instanceId)
			if err != nil {
				return err
			}
			if inst.BootDiskState != "Normal" {
				return newExpectedStateError("boot_disk of instance", instanceId)
			}

			return nil
		})

		if err != nil {
			return halt(state, err, "Error on waiting for boot disk of instance initialized")
		}

		ui.Message(fmt.Sprintf("Waite for boot disk of instance %q initialized complete", instanceId))
	}

	return multistep.ActionContinue
}

func (s *stepCreateInstance) Cleanup(state multistep.StateBag) {
	if s.instanceId == "" {
		return
	}
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)

	ui := state.Get("ui").(packer.Ui)
	ctx := context.TODO()

	if cancelled || halted {
		ui.Say("Deleting instance because of cancellation or error...")
	} else {
		ui.Say("Deleting instance...")
	}

	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn

	stopReq := conn.NewPoweroffUHostInstanceRequest()
	stopReq.UHostId = ucloud.String(s.instanceId)

	instance, err := client.describeUHostById(s.instanceId)
	if err != nil {
		if isNotFoundError(err) {
			return
		}
		ui.Error(fmt.Sprintf("Error on reading instance when delete %q, %s",
			s.instanceId, err.Error()))
		return
	}

	if instance.State != "Stopped" {
		err = retry.Config{
			Tries: 5,
			ShouldRetry: func(err error) bool {
				return err != nil
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			if _, err = conn.PoweroffUHostInstance(stopReq); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Error on stopping instance when delete %q, %s",
				s.instanceId, err.Error()))
			return
		}

		err = retry.Config{
			Tries: 30,
			ShouldRetry: func(err error) bool {
				return isExpectedStateError(err)
			},
			RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
		}.Run(ctx, func(ctx context.Context) error {
			instance, err := client.describeUHostById(s.instanceId)
			if err != nil {
				return err
			}

			if instance.State != "Stopped" {
				return newExpectedStateError("instance", s.instanceId)
			}

			return nil
		})

		if err != nil {
			ui.Error(fmt.Sprintf("Error on waiting for instance %q to stopped, %s",
				s.instanceId, err.Error()))
			return
		}
	}

	deleteReq := conn.NewTerminateUHostInstanceRequest()
	deleteReq.UHostId = ucloud.String(s.instanceId)
	deleteReq.ReleaseUDisk = ucloud.Bool(true)
	deleteReq.ReleaseEIP = ucloud.Bool(true)

	err = retry.Config{
		Tries: 5,
		ShouldRetry: func(err error) bool {
			return err != nil
		},
		RetryDelay: (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		if _, err = conn.TerminateUHostInstance(deleteReq); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error on deleting instance %q, %s",
			s.instanceId, err.Error()))
		return
	}

	err = retry.Config{
		Tries:       30,
		ShouldRetry: func(err error) bool { return !isNotFoundError(err) },
		RetryDelay:  (&retry.Backoff{InitialBackoff: 2 * time.Second, MaxBackoff: 6 * time.Second, Multiplier: 2}).Linear,
	}.Run(ctx, func(ctx context.Context) error {
		_, err := client.describeUHostById(s.instanceId)
		return err
	})

	if err != nil {
		ui.Error(fmt.Sprintf("Error on waiting for deleting instance %q completed, %s",
			s.instanceId, err.Error()))
		return
	}

	ui.Message(fmt.Sprintf("Delete instance %q complete", s.instanceId))
}

func (s *stepCreateInstance) buildCreateInstanceRequest(state multistep.StateBag) (*uhost.CreateUHostInstanceRequest, error) {
	client := state.Get("client").(*UCloudClient)
	conn := client.uhostconn
	srcImage := state.Get("source_image").(*uhost.UHostImageSet)
	config := state.Get("config").(*Config)
	connectConfig := &config.RunConfig.Comm

	var password string
	if srcImage.OsType == "Linux" {
		password = config.Comm.SSHPassword
	}

	if password == "" {
		password = fmt.Sprintf("%s%s%s",
			s.randStringFromCharSet(5, defaultPasswordStr),
			s.randStringFromCharSet(1, defaultPasswordSpe),
			s.randStringFromCharSet(5, defaultPasswordNum))
		if srcImage.OsType == "Linux" {
			connectConfig.SSHPassword = password
		}
	}

	req := conn.NewCreateUHostInstanceRequest()
	t, _ := parseInstanceType(s.InstanceType)

	req.CPU = ucloud.Int(t.CPU)
	req.Memory = ucloud.Int(t.Memory)
	req.Name = ucloud.String(s.InstanceName)
	req.LoginMode = ucloud.String("Password")
	req.Zone = ucloud.String(s.Zone)
	req.ImageId = ucloud.String(s.SourceImageId)
	req.ChargeType = ucloud.String("Dynamic")
	req.Password = ucloud.String(password)

	if v, ok := state.GetOk("security_group_id"); ok {
		req.SecurityGroupId = ucloud.String(v.(string))
	}

	if v, ok := state.GetOk("vpc_id"); ok {
		req.VPCId = ucloud.String(v.(string))
	}

	if v, ok := state.GetOk("subnet_id"); ok {
		req.SubnetId = ucloud.String(v.(string))
	}

	bootDisk := uhost.UHostDisk{}
	bootDisk.IsBoot = ucloud.String("true")
	bootDisk.Size = ucloud.Int(srcImage.ImageSize)
	bootDisk.Type = ucloud.String(bootDiskTypeMap[s.BootDiskType])

	req.Disks = append(req.Disks, bootDisk)

	if !s.UsePrivateIp {
		operatorName := ucloud.String("International")
		if strings.HasPrefix(s.Region, "cn-") {
			operatorName = ucloud.String("Bgp")
		}
		networkInterface := uhost.CreateUHostInstanceParamNetworkInterface{
			EIP: &uhost.CreateUHostInstanceParamNetworkInterfaceEIP{
				Bandwidth:    ucloud.Int(30),
				PayMode:      ucloud.String("Traffic"),
				OperatorName: operatorName,
			},
		}

		req.NetworkInterface = append(req.NetworkInterface, networkInterface)
	}
	return req, nil
}

func (s *stepCreateInstance) randStringFromCharSet(strlen int, charSet string) string {
	rand.Seed(time.Now().UTC().UnixNano())
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(result)
}

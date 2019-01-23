package hyperone

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hyperonecom/h1-client-go"
)

type stepCreateVM struct {
	vmID string
}

func (s *stepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packer.Ui)
	config := state.Get("config").(*Config)
	sshKey := state.Get("ssh_public_key").(string)

	ui.Say("Creating VM...")

	netAdapter := pickNetAdapter(config)

	var sshKeys = []string{sshKey}
	sshKeys = append(sshKeys, config.SSHKeys...)

	options := openapi.VmCreate{
		Name:    config.VmName,
		Image:   config.SourceImage,
		Service: config.VmFlavour,
		SshKeys: sshKeys,
		Disk: []openapi.VmCreateDisk{
			{
				Service: config.DiskType,
				Size:    config.DiskSize,
			},
		},
		Netadp:       []openapi.VmCreateNetadp{netAdapter},
		UserMetadata: config.UserData,
		Tag:          config.VmTags,
	}

	vm, _, err := client.VmApi.VmCreate(ctx, options)
	if err != nil {
		err := fmt.Errorf("error creating VM: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	s.vmID = vm.Id
	state.Put("vm_id", vm.Id)

	hdds, _, err := client.VmApi.VmListHdd(ctx, vm.Id)
	if err != nil {
		err := fmt.Errorf("error listing hdd: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	var diskIDs []string
	for _, hdd := range hdds {
		diskIDs = append(diskIDs, hdd.Disk.Id)
	}

	state.Put("disk_ids", diskIDs)

	netadp, _, err := client.VmApi.VmListNetadp(ctx, vm.Id)
	if err != nil {
		err := fmt.Errorf("error listing netadp: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if len(netadp) < 1 {
		err := fmt.Errorf("no network adapters found")
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	publicIP, err := associatePublicIP(ctx, config, client, netadp[0])
	if err != nil {
		err := fmt.Errorf("error associating IP: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	state.Put("public_ip", publicIP)

	return multistep.ActionContinue
}

func pickNetAdapter(config *Config) openapi.VmCreateNetadp {
	if config.Network == "" {
		if config.PublicIP != "" {
			return openapi.VmCreateNetadp{
				Service: "public",
				Ip:      []string{config.PublicIP},
			}
		}
	} else {
		var privateIPs []string

		if config.PrivateIP == "" {
			privateIPs = nil
		} else {
			privateIPs = []string{config.PrivateIP}
		}

		return openapi.VmCreateNetadp{
			Service: "private",
			Network: config.Network,
			Ip:      privateIPs,
		}
	}

	return openapi.VmCreateNetadp{
		Service: "public",
	}
}

func associatePublicIP(ctx context.Context, config *Config, client *openapi.APIClient, netadp openapi.Netadp) (string, error) {
	if config.Network == "" || config.PublicIP == "" {
		// Public IP belongs to attached net adapter
		return netadp.Ip[0].Address, nil
	}

	var privateIP string
	if config.PrivateIP == "" {
		privateIP = netadp.Ip[0].Id
	} else {
		privateIP = config.PrivateIP
	}

	ip, _, err := client.IpApi.IpActionAssociate(ctx, config.PublicIP, openapi.IpActionAssociate{Ip: privateIP})
	if err != nil {
		return "", err
	}

	return ip.Address, nil
}

func (s *stepCreateVM) Cleanup(state multistep.StateBag) {
	if s.vmID == "" {
		return
	}

	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packer.Ui)

	ui.Say("Deleting VM...")

	deleteOptions := openapi.VmDelete{}
	diskIDs, ok := state.Get("disk_ids").([]string)
	if ok && len(diskIDs) > 0 {
		deleteOptions.RemoveDisks = diskIDs
	}

	_, err := client.VmApi.VmDelete(context.TODO(), s.vmID, deleteOptions)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting server '%s' - please delete it manually: %s", s.vmID, formatOpenAPIError(err)))
	}
}

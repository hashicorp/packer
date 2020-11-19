package hyperone

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
	openapi "github.com/hyperonecom/h1-client-go"
)

type stepCreateVM struct {
	vmID string
}

const (
	chrootDiskName = "packer-chroot-disk"
)

func (s *stepCreateVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*openapi.APIClient)
	ui := state.Get("ui").(packersdk.Ui)
	config := state.Get("config").(*Config)
	sshKey := state.Get("ssh_public_key").(string)

	ui.Say("Creating VM...")

	netAdapter := pickNetAdapter(config)

	var sshKeys = []string{sshKey}
	sshKeys = append(sshKeys, config.SSHKeys...)

	disks := []openapi.VmCreateDisk{
		{
			Service: config.DiskType,
			Size:    config.DiskSize,
		},
	}

	if config.ChrootDisk {
		disks = append(disks, openapi.VmCreateDisk{
			Service: config.ChrootDiskType,
			Size:    config.ChrootDiskSize,
			Name:    chrootDiskName,
		})
	}

	options := openapi.VmCreate{
		Name:         config.VmName,
		Image:        config.SourceImage,
		Service:      config.VmType,
		SshKeys:      sshKeys,
		Disk:         disks,
		Netadp:       []openapi.VmCreateNetadp{netAdapter},
		UserMetadata: config.UserData,
		Tag:          config.VmTags,
		Username:     config.Comm.SSHUsername,
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
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", vm.Id)

	hdds, _, err := client.VmApi.VmListHdd(ctx, vm.Id)
	if err != nil {
		err := fmt.Errorf("error listing hdd: %s", formatOpenAPIError(err))
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	for _, hdd := range hdds {
		if hdd.Disk.Name == chrootDiskName {
			state.Put("chroot_disk_id", hdd.Disk.Id)
			controllerNumber := strings.ToLower(strings.Trim(hdd.ControllerNumber, "{}"))
			state.Put("chroot_controller_number", controllerNumber)
			state.Put("chroot_controller_location", int(hdd.ControllerLocation))
			break
		}
	}

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
				Service: config.PublicNetAdpService,
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
		Service: config.PublicNetAdpService,
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

	ui := state.Get("ui").(packersdk.Ui)

	ui.Say(fmt.Sprintf("Deleting VM %s...", s.vmID))
	err := deleteVMWithDisks(s.vmID, state)
	if err != nil {
		ui.Error(err.Error())
	}
}

func deleteVMWithDisks(vmID string, state multistep.StateBag) error {
	client := state.Get("client").(*openapi.APIClient)
	hdds, _, err := client.VmApi.VmListHdd(context.TODO(), vmID)
	if err != nil {
		return fmt.Errorf("error listing hdd: %s", formatOpenAPIError(err))
	}

	deleteOptions := openapi.VmDelete{}
	for _, hdd := range hdds {
		deleteOptions.RemoveDisks = append(deleteOptions.RemoveDisks, hdd.Disk.Id)
	}

	_, err = client.VmApi.VmDelete(context.TODO(), vmID, deleteOptions)
	if err != nil {
		return fmt.Errorf("Error deleting server '%s' - please delete it manually: %s", vmID, formatOpenAPIError(err))
	}

	return nil
}

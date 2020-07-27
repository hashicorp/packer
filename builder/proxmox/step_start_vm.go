package proxmox

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/Telmate/proxmox-api-go/proxmox"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// stepStartVM takes the given configuration and starts a VM on the given Proxmox node.
//
// It sets the vmRef state which is used throughout the later steps to reference the VM
// in API calls.
type stepStartVM struct{}

func (s *stepStartVM) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	client := state.Get("proxmoxClient").(*proxmox.Client)
	c := state.Get("config").(*Config)

	agent := 1
	if c.Agent == false {
		agent = 0
	}

	kvm := true
	if c.DisableKVM {
		kvm = false
	}

	isoFile := state.Get("iso_file").(string)

	ui.Say("Creating VM")
	config := proxmox.ConfigQemu{
		Name:         c.VMName,
		Agent:        agent,
		QemuKVM:      kvm,
		Boot:         "cdn", // Boot priority, c:CDROM -> d:Disk -> n:Network
		QemuCpu:      c.CPUType,
		Description:  "Packer ephemeral build VM",
		Memory:       c.Memory,
		QemuCores:    c.Cores,
		QemuSockets:  c.Sockets,
		QemuOs:       c.OS,
		QemuVga:      generateProxmoxVga(c.VGA),
		QemuIso:      isoFile,
		QemuNetworks: generateProxmoxNetworkAdapters(c.NICs),
		QemuDisks:    generateProxmoxDisks(c.Disks),
		Scsihw:       c.SCSIController,
		Onboot:       c.Onboot,
	}

	if c.VMID == 0 {
		ui.Say("No VM ID given, getting next free from Proxmox")
		for n := 0; n < 5; n++ {
			id, err := proxmox.MaxVmId(client)
			if err != nil {
				log.Printf("Error getting max used VM ID: %v (attempt %d/5)", err, n+1)
				continue
			}
			c.VMID = id + 1
			break
		}
		if c.VMID == 0 {
			err := fmt.Errorf("Failed to get free VM ID")
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	vmRef := proxmox.NewVmRef(c.VMID)
	vmRef.SetNode(c.Node)
	if c.Pool != "" {
		vmRef.SetPool(c.Pool)
	}

	err := config.CreateVm(vmRef, client)
	if err != nil {
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	// Store the vm id for later
	state.Put("vmRef", vmRef)
	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put("instance_id", vmRef)

	for idx := range c.AdditionalISOFiles {
		params := map[string]interface{}{
			c.AdditionalISOFiles[idx].Device + strconv.Itoa(c.AdditionalISOFiles[idx].BusNumber): c.AdditionalISOFiles[idx].Filename + ",media=cdrom",
		}
		_, err = client.SetVmConfig(vmRef, params)
		if err != nil {
			err := fmt.Errorf("Error configuring VM: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}
	//LOOK HERE, add the error msg to AdditionalISOFiles SetVmConfig same as below
	ui.Say("Starting VM")
	_, err = client.StartVm(vmRef)
	if err != nil {
		err := fmt.Errorf("Error starting VM: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func generateProxmoxNetworkAdapters(nics []nicConfig) proxmox.QemuDevices {
	devs := make(proxmox.QemuDevices)
	for idx := range nics {
		devs[idx] = make(proxmox.QemuDevice)
		setDeviceParamIfDefined(devs[idx], "model", nics[idx].Model)
		setDeviceParamIfDefined(devs[idx], "macaddr", nics[idx].MACAddress)
		setDeviceParamIfDefined(devs[idx], "bridge", nics[idx].Bridge)
		setDeviceParamIfDefined(devs[idx], "tag", nics[idx].VLANTag)
		setDeviceParamIfDefined(devs[idx], "firewall", strconv.FormatBool(nics[idx].Firewall))

		if nics[idx].PacketQueues > 0 {
			devs[idx]["queues"] = nics[idx].PacketQueues
		}
	}
	return devs
}
func generateProxmoxDisks(disks []diskConfig) proxmox.QemuDevices {
	devs := make(proxmox.QemuDevices)
	for idx := range disks {
		devs[idx] = make(proxmox.QemuDevice)
		setDeviceParamIfDefined(devs[idx], "type", disks[idx].Type)
		setDeviceParamIfDefined(devs[idx], "size", disks[idx].Size)
		setDeviceParamIfDefined(devs[idx], "storage", disks[idx].StoragePool)
		setDeviceParamIfDefined(devs[idx], "storage_type", disks[idx].StoragePoolType)
		setDeviceParamIfDefined(devs[idx], "cache", disks[idx].CacheMode)
		setDeviceParamIfDefined(devs[idx], "format", disks[idx].DiskFormat)
	}
	return devs
}
func generateProxmoxVga(vga vgaConfig) proxmox.QemuDevice {
	dev := make(proxmox.QemuDevice)
	setDeviceParamIfDefined(dev, "type", vga.Type)

	if vga.Memory > 0 {
		dev["memory"] = vga.Memory
	}
	return dev
}

func setDeviceParamIfDefined(dev proxmox.QemuDevice, key, value string) {
	if value != "" {
		dev[key] = value
	}
}

type startedVMCleaner interface {
	StopVm(*proxmox.VmRef) (string, error)
	DeleteVm(*proxmox.VmRef) (string, error)
}

var _ startedVMCleaner = &proxmox.Client{}

func (s *stepStartVM) Cleanup(state multistep.StateBag) {
	vmRefUntyped, ok := state.GetOk("vmRef")
	// If not ok, we probably errored out before creating the VM
	if !ok {
		return
	}
	vmRef := vmRefUntyped.(*proxmox.VmRef)

	// The vmRef will actually refer to the created template if everything
	// finished successfully, so in that case we shouldn't cleanup
	if _, ok := state.GetOk("success"); ok {
		return
	}

	client := state.Get("proxmoxClient").(startedVMCleaner)
	ui := state.Get("ui").(packer.Ui)

	// Destroy the server we just created
	ui.Say("Stopping VM")
	_, err := client.StopVm(vmRef)
	if err != nil {
		ui.Error(fmt.Sprintf("Error stopping VM. Please stop and delete it manually: %s", err))
		return
	}

	ui.Say("Deleting VM")
	_, err = client.DeleteVm(vmRef)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting VM. Please delete it manually: %s", err))
		return
	}
}

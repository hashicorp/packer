package driver

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"context"
	"net/url"
	"fmt"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/mo"
	"errors"
	"time"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25"
)

type Driver struct {
	ctx        context.Context
	client     *govmomi.Client
	finder     *find.Finder
	datacenter *object.Datacenter
}

type ConnectConfig struct {
	VCenterServer      string
	Username           string
	Password           string
	InsecureConnection bool
	Datacenter         string
}

type CloneConfig struct {
	Template     string
	VMName       string
	Folder       string
	Host         string
	ResourcePool string
	Datastore    string
	LinkedClone  bool
}

type HardwareConfig struct {
	CPUs           int32
	CPUReservation int64
	CPULimit       int64
	RAM            int64
	RAMReservation int64
	RAMReserveAll  bool
}

func NewDriver(config *ConnectConfig) (*Driver, error) {
	ctx := context.TODO()

	vcenter_url, err := url.Parse(fmt.Sprintf("https://%v/sdk", config.VCenterServer))
	if err != nil {
		return nil, err
	}
	credentials := url.UserPassword(config.Username, config.Password)
	vcenter_url.User = credentials

	soapClient := soap.NewClient(vcenter_url, config.InsecureConnection)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.RoundTripper = session.KeepAlive(vimClient.RoundTripper, 10*time.Minute)
	client := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	err = client.SessionManager.Login(ctx, credentials)
	if err != nil {
		return nil, err
	}

	finder := find.NewFinder(client.Client, false)
	datacenter, err := finder.DatacenterOrDefault(ctx, config.Datacenter)
	if err != nil {
		return nil, err
	}
	finder.SetDatacenter(datacenter)

	d := Driver{
		ctx:        ctx,
		client:     client,
		datacenter: datacenter,
		finder:     finder,
	}
	return &d, nil
}

func (d *Driver) CloneVM(config *CloneConfig) (*object.VirtualMachine, error) {
	template, err := d.finder.VirtualMachine(d.ctx, config.Template)
	if err != nil {
		return nil, err
	}

	folder, err := d.finder.FolderOrDefault(d.ctx, fmt.Sprintf("/%v/vm/%v", d.datacenter.Name(), config.Folder))
	if err != nil {
		return nil, err
	}

	var relocateSpec types.VirtualMachineRelocateSpec

	pool, err := d.finder.ResourcePoolOrDefault(d.ctx, fmt.Sprintf("/%v/host/%v/Resources/%v", d.datacenter.Name(), config.Host, config.ResourcePool))
	if err != nil {
		return nil, err
	}
	poolRef := pool.Reference()
	relocateSpec.Pool = &poolRef

	if config.Datastore != "" {
		datastore, err := d.finder.Datastore(d.ctx, config.Datastore)
		if err != nil {
			return nil, err
		}
		datastoreRef := datastore.Reference()
		relocateSpec.Datastore = &datastoreRef
	}

	var cloneSpec types.VirtualMachineCloneSpec
	cloneSpec.Location = relocateSpec
	cloneSpec.PowerOn = false

	if config.LinkedClone == true {
		cloneSpec.Location.DiskMoveType = "createNewChildDiskBacking"

		var tpl mo.VirtualMachine
		err = template.Properties(d.ctx, template.Reference(), []string{"snapshot"}, &tpl)
		if err != nil {
			return nil, err
		}
		if tpl.Snapshot == nil {
			err = errors.New("`linked_clone=true`, but template has no snapshots")
			return nil, err
		}
		cloneSpec.Snapshot = tpl.Snapshot.CurrentSnapshot
	}

	task, err := template.Clone(d.ctx, folder, config.VMName, cloneSpec)
	if err != nil {
		return nil, err
	}

	info, err := task.WaitForResult(d.ctx, nil)
	if err != nil {
		return nil, err
	}

	vm := object.NewVirtualMachine(d.client.Client, info.Result.(types.ManagedObjectReference))
	return vm, nil
}

func (d *Driver) DestroyVM(vm *object.VirtualMachine) error {
	task, err := vm.Destroy(d.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

func (d *Driver) ConfigureVM(vm *object.VirtualMachine, config *HardwareConfig) error {
	var confSpec types.VirtualMachineConfigSpec
	confSpec.NumCPUs = config.CPUs
	confSpec.MemoryMB = config.RAM

	var cpuSpec types.ResourceAllocationInfo
	cpuSpec.Reservation = config.CPUReservation
	cpuSpec.Limit = config.CPULimit
	confSpec.CpuAllocation = &cpuSpec

	var ramSpec types.ResourceAllocationInfo
	ramSpec.Reservation = config.RAMReservation
	confSpec.MemoryAllocation = &ramSpec

	confSpec.MemoryReservationLockedToMax = &config.RAMReserveAll

	task, err := vm.Reconfigure(d.ctx, confSpec)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

func (d *Driver) PowerOn(vm *object.VirtualMachine) error {
	task, err := vm.PowerOn(d.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

func (d *Driver) WaitForIP(vm *object.VirtualMachine) (string, error) {
	ip, err := vm.WaitForIP(d.ctx)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (d *Driver) PowerOff(vm *object.VirtualMachine) error {
	state, err := vm.PowerState(d.ctx)
	if err != nil {
		return err
	}

	if state == types.VirtualMachinePowerStatePoweredOff {
		return nil
	}

	task, err := vm.PowerOff(d.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

func (d *Driver) StartShutdown(vm *object.VirtualMachine) error {
	err := vm.ShutdownGuest(d.ctx)
	return err
}

func (d *Driver) WaitForShutdown(vm *object.VirtualMachine, timeout time.Duration) error {
	shutdownTimer := time.After(timeout)
	for {
		powerState, err := vm.PowerState(d.ctx)
		if err != nil {
			return err
		}
		if powerState == "poweredOff" {
			break
		}

		select {
		case <-shutdownTimer:
			err := errors.New("Timeout while waiting for machine to shut down.")
			return err
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func (d *Driver) CreateSnapshot(vm *object.VirtualMachine) error {
	task, err := vm.CreateSnapshot(d.ctx, "Created by Packer", "", false, false)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(d.ctx, nil)
	return err
}

func (d *Driver) ConvertToTemplate(vm *object.VirtualMachine) error {
	err := vm.MarkAsTemplate(d.ctx)
	return err
}

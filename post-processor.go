package main

import (
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Url            string `mapstructure:"url"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Dc_name        string `mapstructure:"dc_name"`
	Folder_name    string
	Vm_source_name string `mapstructure:"vm_source_name"`
	Vm_target_name string `mapstructure:"vm_target_name"`
	Cpu_sockets    string `mapstructure:"cpus"`

	cpus           int
	ctx            interpolate.Context
}

type PostProcessor struct {
	config Config
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	// Defaults
	p.config.Folder_name = ""
	if p.config.Vm_target_name == "" {
		p.config.Vm_target_name = p.config.Vm_source_name + "_cloned"
	}
	p.config.cpus = -1
	if p.config.Cpu_sockets != "" {
		p.config.cpus, err = strconv.Atoi(p.config.Cpu_sockets)
		if err != nil {
			panic(err)
		}
	}

	// Accumulate any errors
	errs := new(packer.MultiError)

	// First define all our templatable parameters that are _required_
	templates := map[string]*string{
		"url":            &p.config.Url,
		"username":       &p.config.Username,
		"password":       &p.config.Password,
		"vm_source_name": &p.config.Vm_source_name,
	}
	for key, ptr := range templates {
		if *ptr == "" {
			errs = packer.MultiErrorAppend(
				errs, fmt.Errorf("%s must be set, %s is present", key, *ptr))
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

// TODO: replace `cpus` with a more generic hw config structure
func CloneVM(URL, username, password, dc_name, folder_name, source_name, target_name string, cpus int) {
	// Prepare entities: client (authentification), finder, folder, virtual machine
	client, ctx := createClient(URL, username, password)
	finder, ctx := createFinder(ctx, client, dc_name)
	folder, err := finder.FolderOrDefault(ctx, "") // folder_name
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("expected folder: %v\n", folder_name)
		fmt.Printf("folder.Name(): %v\nfolder.InventoryPath(): %v\n", folder.Name(), folder.InventoryPath)
	}
	vm_src, ctx := findVM_by_name(ctx, finder, source_name)

	// Creating spec's for cloning
	var relocateSpec types.VirtualMachineRelocateSpec

	var confSpec types.VirtualMachineConfigSpec
	// configure CPUs
	if cpus != -1 {
		confSpec.NumCPUs = int32(cpus)
	}

	cloneSpec := types.VirtualMachineCloneSpec{
		Location: relocateSpec,
		Config:   &confSpec,
		PowerOn:  false,
	}

	// Cloning itself
	task, err := vm_src.Clone(ctx, folder, target_name, cloneSpec)
	if err != nil {
		panic(err)
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		panic(err)
	}
}

func (p *PostProcessor) PostProcess(ui packer.Ui, source packer.Artifact) (packer.Artifact, bool, error) {
	CloneVM(p.config.Url, p.config.Username, p.config.Password, p.config.Dc_name, p.config.Folder_name, p.config.Vm_source_name, p.config.Vm_target_name, p.config.cpus)
	return source, true, nil
}

func createClient(URL, username, password string) (*govmomi.Client, context.Context) {
	// create context
	ctx := context.TODO() // an empty, default context (for those, who is unsure)

	// create a client
	// (connected to the specified URL,
	// logged in with the username-password)
	u, err := url.Parse(URL) // create a URL object from string
	if err != nil {
		panic(err)
	}
	u.User = url.UserPassword(username, password) // set username and password for automatical authentification
	fmt.Println(u.String())
	client, err := govmomi.NewClient(ctx, u,true) // creating a client (logs in with given uname&pswd)
	if err != nil {
		panic(err)
	}

	return client, ctx
}

func createFinder(ctx context.Context, client *govmomi.Client, dc_name string) (*find.Finder, context.Context) {
	// Create a finder to search for a vm with the specified name
	finder := find.NewFinder(client.Client, false)
	// Need to specify the datacenter
	if dc_name == "" {
		dc, err := finder.DefaultDatacenter(ctx)
		if err != nil {
			panic(fmt.Errorf("Error reading default datacenter: %s", err))
		}
		var dc_mo mo.Datacenter
		err = dc.Properties(ctx, dc.Reference(), []string{"name"}, &dc_mo)
		if err != nil {
			panic(fmt.Errorf("Error reading datacenter name: %s", err))
		}
		dc_name = dc_mo.Name
		finder.SetDatacenter(dc)
	} else {
		dc, err := finder.Datacenter(ctx, fmt.Sprintf("/%v", dc_name))
		if err != nil {
			panic(err)
		}
		finder.SetDatacenter(dc)
	}
	return finder, ctx
}

func findVM_by_name(ctx context.Context, finder *find.Finder, vm_name string) (*object.VirtualMachine, context.Context) {
	vm_o, err := finder.VirtualMachine(ctx, vm_name)
	if err != nil {
		panic(err)
	}
	return vm_o, ctx
}

func ReconfigureVM(URL, username, password, dc_name, vm_name string, cpus int) {
	client, ctx := createClient(URL, username, password)
	finder, ctx := createFinder(ctx, client, dc_name)
	vm_o, ctx := findVM_by_name(ctx, finder, vm_name)

	// creating new configuration for vm
	vmConfigSpec := types.VirtualMachineConfigSpec{}
	vmConfigSpec.NumCPUs = int32(cpus)

	// finally reconfiguring
	task, err := vm_o.Reconfigure(ctx, vmConfigSpec)
	if err != nil {
		panic(err)
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		panic(err)
	}
}

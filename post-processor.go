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

type VMOptionalParams struct {
	Cpu_sockets int
	Ram         int
	// TODO: add more options
}

type VMRequiredParams struct {
	Url            string
	Username       string
	Password       string
	Dc_name        string
	Folder_name    string
	Vm_source_name string
	Vm_target_name string
}

const DefaultFolder = ""
const Unspecified = -1

var vm_opt_params VMOptionalParams
var vm_req_params VMRequiredParams

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	Url            string `mapstructure:"url"`
	Username       string `mapstructure:"username"`
	Password       string `mapstructure:"password"`
	Dc_name        string `mapstructure:"dc_name"`
	Vm_source_name string `mapstructure:"vm_source_name"`
	Vm_target_name string `mapstructure:"vm_target_name"`
	Cpu_sockets    string `mapstructure:"cpus"`
	Ram            string `mapstructure:"RAM"`

	ctx      interpolate.Context
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

	// Accumulate any errors
	errs := new(packer.MultiError)

	// Check the required params
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

	// Set optional params
	vm_req_params.Folder_name = DefaultFolder
	vm_opt_params.Cpu_sockets = Unspecified
	if p.config.Cpu_sockets != "" {
		vm_opt_params.Cpu_sockets, err = strconv.Atoi(p.config.Cpu_sockets)
		if err != nil {
			panic(err)
		}
	}
	vm_opt_params.Ram = Unspecified
	if p.config.Ram != "" {
		vm_opt_params.Ram, err = strconv.Atoi(p.config.Ram)
		if err != nil {
			panic(err)
		}
	}

	// Set required params
	vm_req_params.Url = p.config.Url
	vm_req_params.Username = p.config.Username
	vm_req_params.Password = p.config.Password
	vm_req_params.Dc_name = p.config.Dc_name
	vm_req_params.Vm_source_name = p.config.Vm_source_name
	vm_req_params.Vm_target_name = vm_req_params.Vm_source_name + "_clone"
	if p.config.Vm_target_name != "" {
		vm_req_params.Vm_target_name = p.config.Vm_target_name
	}

	return nil
}

func CloneVM(req_params VMRequiredParams, opt_params VMOptionalParams) {
	// Prepare entities: client (authentification), finder, folder, virtual machine
	client, ctx := createClient(req_params.Url, req_params.Username, req_params.Password)
	finder, ctx := createFinder(ctx, client, req_params.Dc_name)
	folder, err := finder.FolderOrDefault(ctx, req_params.Folder_name)
	if err != nil {
		panic(err)
	} else {
		fmt.Printf("expected folder: %v\n", req_params.Folder_name)
		fmt.Printf("folder.Name(): %v\nfolder.InventoryPath(): %v\n", folder.Name(), folder.InventoryPath)
	}
	vm_src, ctx := findVM_by_name(ctx, finder, req_params.Vm_source_name)

	// Creating spec's for cloning
	var relocateSpec types.VirtualMachineRelocateSpec

	var confSpec types.VirtualMachineConfigSpec
	// configure HW
	if opt_params.Cpu_sockets != Unspecified {
		confSpec.NumCPUs = int32(opt_params.Cpu_sockets)
	}
	if opt_params.Ram != Unspecified {
		confSpec.MemoryMB = int64(opt_params.Ram)
	}

	cloneSpec := types.VirtualMachineCloneSpec{
		Location: relocateSpec,
		Config:   &confSpec,
		PowerOn:  false,
	}

	// Cloning itself
	task, err := vm_src.Clone(ctx, folder, req_params.Vm_target_name, cloneSpec)
	if err != nil {
		panic(err)
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		panic(err)
	}
}

func (p *PostProcessor) PostProcess(ui packer.Ui, source packer.Artifact) (packer.Artifact, bool, error) {
	CloneVM(vm_req_params, vm_opt_params)
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

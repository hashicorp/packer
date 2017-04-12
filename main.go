package main

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"os"
	"strconv"
	"fmt"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/object"
)

func authentificate(URL, username, password string) (*govmomi.Client, context.Context) {
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

//func reconfigureVM(ctx context.Context, vm *object.VirtualMachine, cpus int) {
func reconfigureVM(URL, username, password, dc_name, vm_name string, cpus int) {
	client, ctx := authentificate(URL, username, password)
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

func cloneVM(URL, username, password, dc_name, folder_name, source_name, target_name string, cpus int) {
	// Prepare entities: authentification, finder, folder, virtual machine
	client, ctx := authentificate(URL, username, password)
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
	info, err := task.WaitForResult(ctx, nil)
	if err != nil {
		panic(err)
	}

	vm_new_mor := info.Result.(types.ManagedObjectReference)
	vm_new := object.NewVirtualMachine(client.Client, vm_new_mor)
	var vm_new_mo mo.VirtualMachine
	err = vm_new.Properties(ctx, vm_new.Reference(), []string{"summary"}, &vm_new_mo)
	cpus_new := vm_new_mo.Summary.Config.NumCpu
	var vm_src_mo mo.VirtualMachine
	err = vm_src.Properties(ctx, vm_src.Reference(), []string{"summary"}, &vm_src_mo)
	cpus_src := vm_src_mo.Summary.Config.NumCpu
	fmt.Printf("Num of cpus on the newly created machine %v vs on the initial one %v\n", cpus_new, cpus_src)
}

func main() {
	var URL = os.Args[1]
	var username = os.Args[2]
	var password = os.Args[3]
	var dc_name = os.Args[6]
	var vm_name = os.Args[4]
	var cpus, err = strconv.Atoi(os.Args[5])
	if err != nil {
		panic(err)
	}

	cloneVM(URL, username, password, dc_name, "", vm_name, vm_name + "_cloned", cpus)
}

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
)

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
	vm, err := finder.VirtualMachine(ctx, vm_name)
	if err != nil {
		panic(err)
	}

	// creating new configuration for vm
	vmConfigSpec := types.VirtualMachineConfigSpec{}
	vmConfigSpec.NumCPUs = int32(cpus)

	// finally reconfiguring
	task, err := vm.Reconfigure(ctx, vmConfigSpec)
	if err != nil {
		panic(err)
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		panic(err)
	}
}

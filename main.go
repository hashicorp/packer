package main

import (
	"context"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"net/url"
	"os"
	"strconv"
	"fmt"
)

func main() {
	var URL = os.Args[1]
	var username = os.Args[2]
	var password = os.Args[3]
	var vm_name = os.Args[4]
	var cpus, err = strconv.Atoi(os.Args[5])
	if err != nil {
		panic(err)
	}

	// create context
	ctx := context.TODO() // an empty, default context (for those, who is unsure)

	// create a client
	// (connected to the specified URL,
	// logged in with the username-passrowd)
	u, err := url.Parse(URL) // create a URL object from string
	if err != nil {
		panic(err)
	}
	u.User = url.UserPassword(username, password) // set username and password for automatical authentification
	fmt.Println(u.String())
	client, err := govmomi.NewClient(ctx, u,false) // creating a client (logs in with given uname&pswd)
	if err != nil {
		panic(err)
	}
	// create a reference to a VM with the specified name...
	vm_mor := types.ManagedObjectReference{Type: "VirtualMachine", Value: vm_name}
	// ... and the VM itself, using the set up client & reference
	vm := object.NewVirtualMachine(client.Client, vm_mor)

	vmConfigSpec := types.VirtualMachineConfigSpec{}
	vmConfigSpec.NumCPUs = int32(cpus)
	task, err := vm.Reconfigure(ctx, vmConfigSpec)
	if err != nil {
		panic(err)
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		panic(err)
	}
}

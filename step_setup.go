package main

import (
	"github.com/mitchellh/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/vmware/govmomi/find"
	"fmt"
	"github.com/vmware/govmomi"
	"context"
	"net/url"
)

type StepSetup struct{
	config *Config
}

func (s *StepSetup) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	ui.Say("setup...")

	// Prepare entities: client (authentification), finder, folder, virtual machine
	client, ctx, err := createClient(s.config.Url, s.config.Username, s.config.Password)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	// Set up finder
	finder := find.NewFinder(client.Client, false)
	dc, err := finder.DatacenterOrDefault(ctx, s.config.DCName)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}
	finder.SetDatacenter(dc)

	// Get source VM
	vmSrc, err := finder.VirtualMachine(ctx, s.config.Template)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	state.Put("client", client)
	state.Put("ctx", ctx)
	state.Put("finder", finder)
	state.Put("dc", dc)
	state.Put("vmSrc", vmSrc)
	return multistep.ActionContinue
}

func (s *StepSetup) Cleanup(state multistep.StateBag) {}

func createClient(URL, username, password string) (*govmomi.Client, context.Context, error) {
	// create context
	ctx := context.TODO() // an empty, default context (for those, who is unsure)

	// create a client
	// (connected to the specified URL,
	// logged in with the username-password)
	u, err := url.Parse(URL) // create a URL object from string
	if err != nil {
		return nil, nil, err
	}
	u.User = url.UserPassword(username, password) // set username and password for automatical authentification
	fmt.Println(u.String())
	client, err := govmomi.NewClient(ctx, u,true) // creating a client (logs in with given uname&pswd)
	if err != nil {
		return nil, nil, err
	}

	return client, ctx, nil
}

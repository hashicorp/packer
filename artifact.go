package main

import (
	"github.com/vmware/govmomi/object"
	"context"
)

const BuilderId = "jetbrains.vsphere"

type Artifact struct {
	VMName string
	Conn *object.VirtualMachine
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
	return a.VMName
}

func (a *Artifact) String() string {
	return a.VMName
}

func (a *Artifact) State(name string) interface{} {
	return nil
}

func (a *Artifact) Destroy() error {
	ctx := context.TODO()
	task, err := a.Conn.Destroy(ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

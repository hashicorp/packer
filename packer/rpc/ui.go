package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Ui where the Ui is actually executed
// over an RPC connection.
type Ui struct {
	client *rpc.Client
}

// UiServer wraps a packer.Ui implementation and makes it exportable
// as part of a Golang RPC server.
type UiServer struct {
	ui packer.Ui
}

type UiSayArgs struct {
	Format string
	Vars []interface{}
}

func (u *Ui) Say(format string, a ...interface{}) {
	args := &UiSayArgs{format, a}
	u.client.Call("Ui.Say", args, new(interface{}))
}

func (u *UiServer) Say(args *UiSayArgs, reply *interface{}) error {
	u.ui.Say(args.Format, args.Vars...)

	*reply = nil
	return nil
}

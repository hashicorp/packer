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
	Vars   []interface{}
}

func (u *Ui) Error(format string, a ...interface{}) {
	u.processArgs(a)

	args := &UiSayArgs{format, a}
	if err := u.client.Call("Ui.Error", args, new(interface{})); err != nil {
		panic(err)
	}
}

func (u *Ui) Say(format string, a ...interface{}) {
	u.processArgs(a)

	args := &UiSayArgs{format, a}
	if err := u.client.Call("Ui.Say", args, new(interface{})); err != nil {
		panic(err)
	}
}

func (u *Ui) processArgs(a []interface{}) {
	// We do some processing to turn certain types into more gob-friendly
	// types so that some things that users expect to do just work.
	for i, v := range a {
		// Turn errors into strings
		if err, ok := v.(error); ok {
			a[i] = err.Error()
		}
	}
}

func (u *UiServer) Error(args *UiSayArgs, reply *interface{}) error {
	u.ui.Error(args.Format, args.Vars...)

	*reply = nil
	return nil
}

func (u *UiServer) Say(args *UiSayArgs, reply *interface{}) error {
	u.ui.Say(args.Format, args.Vars...)

	*reply = nil
	return nil
}

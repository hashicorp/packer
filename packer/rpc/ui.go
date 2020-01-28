package rpc

import (
	"log"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.Ui where the Ui is actually executed
// over an RPC connection.
type Ui struct {
	commonClient
	endpoint string
}

var _ packer.Ui = new(Ui)

// UiServer wraps a packer.Ui implementation and makes it exportable
// as part of a Golang RPC server.
type UiServer struct {
	ui       packer.Ui
	register func(name string, rcvr interface{}) error
}

// The arguments sent to Ui.Machine
type UiMachineArgs struct {
	Category string
	Args     []string
}

func (u *Ui) Ask(query string) (result string, err error) {
	err = u.client.Call("Ui.Ask", query, &result)
	return
}

func (u *Ui) Error(message string) {
	if err := u.client.Call("Ui.Error", message, new(interface{})); err != nil {
		log.Printf("Error in Ui.Error RPC call: %s", err)
	}
}

func (u *Ui) Machine(t string, args ...string) {
	rpcArgs := &UiMachineArgs{
		Category: t,
		Args:     args,
	}

	if err := u.client.Call("Ui.Machine", rpcArgs, new(interface{})); err != nil {
		log.Printf("Error in Ui.Machine RPC call: %s", err)
	}
}

func (u *Ui) Message(message string) {
	if err := u.client.Call("Ui.Message", message, new(interface{})); err != nil {
		log.Printf("Error in Ui.Message RPC call: %s", err)
	}
}

func (u *Ui) Say(message string) {
	if err := u.client.Call("Ui.Say", message, new(interface{})); err != nil {
		log.Printf("Error in Ui.Say RPC call: %s", err)
	}
}

func (u *UiServer) Ask(query string, reply *string) (err error) {
	*reply, err = u.ui.Ask(query)
	return
}

func (u *UiServer) Error(message *string, reply *interface{}) error {
	u.ui.Error(*message)

	*reply = nil
	return nil
}

func (u *UiServer) Machine(args *UiMachineArgs, reply *interface{}) error {
	u.ui.Machine(args.Category, args.Args...)

	*reply = nil
	return nil
}

func (u *UiServer) Message(message *string, reply *interface{}) error {
	u.ui.Message(*message)
	*reply = nil
	return nil
}

func (u *UiServer) Say(message *string, reply *interface{}) error {
	u.ui.Say(*message)

	*reply = nil
	return nil
}

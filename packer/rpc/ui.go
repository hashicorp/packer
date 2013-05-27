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

func (u *Ui) Error(message string) {
	if err := u.client.Call("Ui.Error", message, new(interface{})); err != nil {
		panic(err)
	}
}

func (u *Ui) Say(message string) {
	if err := u.client.Call("Ui.Say", message, new(interface{})); err != nil {
		panic(err)
	}
}

func (u *UiServer) Error(message *string, reply *interface{}) error {
	u.ui.Error(*message)

	*reply = nil
	return nil
}

func (u *UiServer) Say(message *string, reply *interface{}) error {
	u.ui.Say(*message)

	*reply = nil
	return nil
}

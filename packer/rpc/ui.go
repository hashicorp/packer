package rpc

import (
	"io"
	"log"
	"net/rpc"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.Ui where the Ui is actually executed
// over an RPC connection.
type Ui struct {
	client   *rpc.Client
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
		log.Printf("Error in Ui RPC call: %s", err)
	}
}

func (u *Ui) Machine(t string, args ...string) {
	rpcArgs := &UiMachineArgs{
		Category: t,
		Args:     args,
	}

	if err := u.client.Call("Ui.Machine", rpcArgs, new(interface{})); err != nil {
		log.Printf("Error in Ui RPC call: %s", err)
	}
}

func (u *Ui) Message(message string) {
	if err := u.client.Call("Ui.Message", message, new(interface{})); err != nil {
		log.Printf("Error in Ui RPC call: %s", err)
	}
}

func (u *Ui) Say(message string) {
	if err := u.client.Call("Ui.Say", message, new(interface{})); err != nil {
		log.Printf("Error in Ui RPC call: %s", err)
	}
}

func (u *Ui) ProgressBar() packer.ProgressBar {
	if err := u.client.Call("Ui.ProgressBar", new(interface{}), new(interface{})); err != nil {
		log.Printf("Error in Ui RPC call: %s", err)
	}
	return u // Ui is also a progress bar !!
}

var _ packer.ProgressBar = new(Ui)

func (pb *Ui) Start(total uint64) {
	pb.client.Call("Ui.Start", total, new(interface{}))
}

func (pb *Ui) Add(current uint64) {
	pb.client.Call("Ui.Add", current, new(interface{}))
}

func (pb *Ui) Finish() {
	pb.client.Call("Ui.Finish", nil, new(interface{}))
}

func (pb *Ui) NewProxyReader(r io.Reader) io.Reader {
	return &packer.ProxyReader{Reader: r, ProgressBar: pb}
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

func (u *UiServer) ProgressBar(_ *string, reply *interface{}) error {
	// No-op for now, this function might be
	// used in the future if we want to use
	// different progress bars with identifiers.
	u.ui.ProgressBar()
	return nil
}

func (pb *UiServer) Finish(_ string, _ *interface{}) error {
	pb.ui.ProgressBar().Finish()
	return nil
}

func (pb *UiServer) Start(total uint64, _ *interface{}) error {
	pb.ui.ProgressBar().Start(total)
	return nil
}

func (pb *UiServer) Add(current uint64, _ *interface{}) error {
	pb.ui.ProgressBar().Add(current)
	return nil
}

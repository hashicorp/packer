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

func (u *Ui) ProgressBar(identifier string) packer.ProgressBar {
	if err := u.client.Call("Ui.ProgressBar", identifier, new(interface{})); err != nil {
		log.Printf("Err or in Ui RPC call: %s", err)
	}
	return &RemoteProgressBarClient{
		id:     identifier,
		client: u.client,
	}
}

type RemoteProgressBarClient struct {
	id     string
	client *rpc.Client
}

var _ packer.ProgressBar = new(RemoteProgressBarClient)

func (pb *RemoteProgressBarClient) Start(total int64) {
	pb.client.Call(pb.id+".Start", total, new(interface{}))
}

func (pb *RemoteProgressBarClient) Add(current int64) {
	pb.client.Call(pb.id+".Add", current, new(interface{}))
}

func (pb *RemoteProgressBarClient) Finish() {
	pb.client.Call(pb.id+".Finish", nil, new(interface{}))
}

func (pb *RemoteProgressBarClient) NewProxyReader(r io.Reader) io.Reader {
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

// ProgressBar registers a rpc progress bar server identified by identifier.
// ProgressBar expects identifiers to be unique across runs
// since for examples an iso download should be cached.
func (u *UiServer) ProgressBar(identifier string, reply *interface{}) error {

	bar := u.ui.ProgressBar(identifier)
	log.Printf("registering progressbar for '%s'", identifier)
	err := u.register(identifier, &UiProgressBarServer{bar})
	if err != nil {
		log.Printf("failed to register a new progress bar rpc server, %s", err)
		return err
	}
	*reply = identifier

	return nil
}

type UiProgressBarServer struct {
	bar packer.ProgressBar
}

func (pb *UiProgressBarServer) Finish(_ string, _ *interface{}) error {
	pb.bar.Finish()
	return nil
}

func (pb *UiProgressBarServer) Start(total int64, _ *interface{}) error {
	pb.bar.Start(total)
	return nil
}

func (pb *UiProgressBarServer) Add(current int64, _ *interface{}) error {
	pb.bar.Add(current)
	return nil
}

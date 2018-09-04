package rpc

import (
	"log"
	"math/rand"
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
	var callMeMaybe string
	if err := u.client.Call("Ui.ProgressBar", nil, &callMeMaybe); err != nil {
		log.Printf("Error in Ui RPC call: %s", err)
		return new(packer.NoopProgressBar)
	}

	return &RemoteProgressBarClient{
		id:     callMeMaybe,
		client: u.client,
	}
}

type RemoteProgressBarClient struct {
	id     string
	client *rpc.Client
}

var _ packer.ProgressBar = new(RemoteProgressBarClient)

func (pb *RemoteProgressBarClient) Start(total uint64) {
	pb.client.Call(pb.id+".Start", total, new(interface{}))
}

func (pb *RemoteProgressBarClient) Set(current uint64) {
	pb.client.Call(pb.id+".Set", current, new(interface{}))
}

func (pb *RemoteProgressBarClient) Finish() {
	pb.client.Call(pb.id+".Finish", nil, new(interface{}))
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

func RandStringBytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (u *UiServer) ProgressBar(_ *string, reply *interface{}) error {
	bar := u.ui.ProgressBar()

	callbackName := RandStringBytes(6)

	log.Printf("registering progressbar %s", callbackName)
	err := u.register(callbackName, &RemoteProgressBarServer{bar})
	if err != nil {
		log.Printf("failed to register a new progress bar rpc server, %s", err)
		return err
	}
	*reply = callbackName
	return nil
}

type RemoteProgressBarServer struct {
	pb packer.ProgressBar
}

func (pb *RemoteProgressBarServer) Finish(_ string, _ *interface{}) error {
	pb.pb.Finish()
	return nil
}

func (pb *RemoteProgressBarServer) Start(total uint64, _ *interface{}) error {
	pb.pb.Start(total)
	return nil
}

func (pb *RemoteProgressBarServer) Set(current uint64, _ *interface{}) error {
	pb.pb.Set(current)
	return nil
}

package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.Provisioner where the provisioner is actually
// executed over an RPC connection.
type provisioner struct {
	client *rpc.Client
}

// ProvisionerServer wraps a packer.Provisioner implementation and makes it
// exportable as part of a Golang RPC server.
type ProvisionerServer struct {
	p packer.Provisioner
}

type ProvisionerPrepareArgs struct {
	Config     interface{}
	RPCAddress string
}

type ProvisionerProvisionArgs struct {
	RPCAddress string
}

func Provisioner(client *rpc.Client) *provisioner {
	return &provisioner{client}
}
func (p *provisioner) Prepare(config interface{}, ui packer.Ui) {
	// TODO: Error handling
	server := rpc.NewServer()
	RegisterUi(server, ui)

	args := &ProvisionerPrepareArgs{config, serveSingleConn(server)}
	p.client.Call("Provisioner.Prepare", args, new(interface{}))
}

func (p *provisioner) Provision(ui packer.Ui, comm packer.Communicator) {
	// TODO: Error handling
	server := rpc.NewServer()
	RegisterCommunicator(server, comm)
	RegisterUi(server, ui)

	args := &ProvisionerProvisionArgs{serveSingleConn(server)}
	p.client.Call("Provisioner.Provision", args, new(interface{}))
}

func (p *ProvisionerServer) Prepare(args *ProvisionerPrepareArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	ui := &Ui{client}

	p.p.Prepare(args.Config, ui)
	return nil
}

func (p *ProvisionerServer) Provision(args *ProvisionerProvisionArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	comm := Communicator(client)
	ui := &Ui{client}

	p.p.Provision(ui, comm)
	return nil
}

package rpc

import (
	"github.com/mitchellh/packer/packer"
	"log"
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
	Configs []interface{}
}

type ProvisionerProvisionArgs struct {
	RPCAddress string
}

func Provisioner(client *rpc.Client) *provisioner {
	return &provisioner{client}
}
func (p *provisioner) Prepare(configs ...interface{}) (err error) {
	args := &ProvisionerPrepareArgs{configs}
	if cerr := p.client.Call("Provisioner.Prepare", args, &err); cerr != nil {
		err = cerr
	}

	return
}

func (p *provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	// TODO: Error handling
	server := rpc.NewServer()
	RegisterCommunicator(server, comm)
	RegisterUi(server, ui)

	args := &ProvisionerProvisionArgs{serveSingleConn(server)}
	return p.client.Call("Provisioner.Provision", args, new(interface{}))
}

func (p *provisioner) Cancel() {
	err := p.client.Call("Provisioner.Cancel", new(interface{}), new(interface{}))
	if err != nil {
		log.Printf("Provisioner.Cancel err: %s", err)
	}
}

func (p *ProvisionerServer) Prepare(args *ProvisionerPrepareArgs, reply *error) error {
	*reply = p.p.Prepare(args.Configs...)
	if *reply != nil {
		*reply = NewBasicError(*reply)
	}

	return nil
}

func (p *ProvisionerServer) Provision(args *ProvisionerProvisionArgs, reply *interface{}) error {
	client, err := rpcDial(args.RPCAddress)
	if err != nil {
		return err
	}

	comm := Communicator(client)
	ui := &Ui{client}

	if err := p.p.Provision(ui, comm); err != nil {
		return NewBasicError(err)
	}

	return nil
}

func (p *ProvisionerServer) Cancel(args *interface{}, reply *interface{}) error {
	p.p.Cancel()
	return nil
}

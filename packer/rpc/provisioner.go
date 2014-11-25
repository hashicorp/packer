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
	mux    *muxBroker
}

// ProvisionerServer wraps a packer.Provisioner implementation and makes it
// exportable as part of a Golang RPC server.
type ProvisionerServer struct {
	p   packer.Provisioner
	mux *muxBroker
}

type ProvisionerPrepareArgs struct {
	Configs []interface{}
}

func (p *provisioner) Prepare(configs ...interface{}) (err error) {
	args := &ProvisionerPrepareArgs{configs}
	if cerr := p.client.Call("Provisioner.Prepare", args, new(interface{})); cerr != nil {
		err = cerr
	}

	return
}

func (p *provisioner) Provision(ui packer.Ui, comm packer.Communicator) error {
	nextId := p.mux.NextId()
	server := newServerWithMux(p.mux, nextId)
	server.RegisterCommunicator(comm)
	server.RegisterUi(ui)
	go server.Serve()

	return p.client.Call("Provisioner.Provision", nextId, new(interface{}))
}

func (p *provisioner) Cancel() {
	err := p.client.Call("Provisioner.Cancel", new(interface{}), new(interface{}))
	if err != nil {
		log.Printf("Provisioner.Cancel err: %s", err)
	}
}

func (p *ProvisionerServer) Prepare(args *ProvisionerPrepareArgs, reply *interface{}) error {
	return p.p.Prepare(args.Configs...)
}

func (p *ProvisionerServer) Provision(streamId uint32, reply *interface{}) error {
	client, err := newClientWithMux(p.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	if err := p.p.Provision(client.Ui(), client.Communicator()); err != nil {
		return NewBasicError(err)
	}

	return nil
}

func (p *ProvisionerServer) Cancel(args *interface{}, reply *interface{}) error {
	p.p.Cancel()
	return nil
}

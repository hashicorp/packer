package rpc

import (
	"context"

	"github.com/hashicorp/packer/common/net/rpc"

	"github.com/hashicorp/packer/packer"
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
	ctx := context.TODO()
	if cerr := p.client.Call(ctx, "Provisioner.Prepare", args, new(interface{})); cerr != nil {
		err = cerr
	}

	return
}

func (p *provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	nextId := p.mux.NextId()
	server := newServerWithMux(p.mux, nextId)
	server.RegisterCommunicator(comm)
	server.RegisterUi(ui)
	go server.Serve()

	return p.client.Call(ctx, "Provisioner.Provision", nextId, new(interface{}))
}

func (p *ProvisionerServer) Prepare(_ context.Context, args *ProvisionerPrepareArgs, reply *interface{}) error {
	return p.p.Prepare(args.Configs...)
}

func (p *ProvisionerServer) Provision(ctx context.Context, streamId uint32, reply *interface{}) error {
	client, err := newClientWithMux(p.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	if err := p.p.Provision(ctx, client.Ui(), client.Communicator()); err != nil {
		return NewBasicError(err)
	}

	return nil
}

package rpc

import (
	"context"
	"log"
	"net/rpc"

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
	context       context.Context
	contextCancel func()

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

func (p *provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator) error {
	nextId := p.mux.NextId()
	server := newServerWithMux(p.mux, nextId)
	server.RegisterCommunicator(comm)
	server.RegisterUi(ui)
	go server.Serve()

	done := make(chan interface{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			log.Printf("Cancelling provisioner after context cancellation %v", ctx.Err())
			if err := p.client.Call("Provisioner.Cancel", new(interface{}), new(interface{})); err != nil {
				log.Printf("Error cancelling provisioner: %s", err)
			}
		case <-done:
		}
	}()

	return p.client.Call("Provisioner.Provision", nextId, new(interface{}))
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

	if p.context == nil {
		p.context, p.contextCancel = context.WithCancel(context.Background())
	}

	if err := p.p.Provision(p.context, client.Ui(), client.Communicator()); err != nil {
		return NewBasicError(err)
	}

	return nil
}

func (p *ProvisionerServer) Cancel(args *interface{}, reply *interface{}) error {
	p.contextCancel()
	return nil
}

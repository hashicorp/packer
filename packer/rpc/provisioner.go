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

type ProvisionerProvisionArgs struct {
	GeneratedData map[string]interface{}
	StreamID      uint32
}

func (p *provisioner) Provision(ctx context.Context, ui packer.Ui, comm packer.Communicator, generatedData map[string]interface{}) error {
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

	args := &ProvisionerProvisionArgs{generatedData, nextId}
	return p.client.Call("Provisioner.Provision", args, new(interface{}))
}

func (p *ProvisionerServer) Prepare(args *ProvisionerPrepareArgs, reply *interface{}) error {
	return p.p.Prepare(args.Configs...)
}

func (p *ProvisionerServer) Provision(args *ProvisionerProvisionArgs, reply *interface{}) error {
	streamId := args.StreamID
	client, err := newClientWithMux(p.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	if p.context == nil {
		p.context, p.contextCancel = context.WithCancel(context.Background())
	}
	if err := p.p.Provision(p.context, client.Ui(), client.Communicator(), args.GeneratedData); err != nil {
		return NewBasicError(err)
	}

	return nil
}

func (p *ProvisionerServer) Cancel(args *interface{}, reply *interface{}) error {
	p.contextCancel()
	return nil
}

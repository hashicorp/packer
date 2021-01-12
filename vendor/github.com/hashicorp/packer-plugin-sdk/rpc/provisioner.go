package rpc

import (
	"context"
	"log"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// An implementation of packersdk.Provisioner where the provisioner is actually
// executed over an RPC connection.
type provisioner struct {
	commonClient
}

// ProvisionerServer wraps a packersdk.Provisioner implementation and makes it
// exportable as part of a Golang RPC server.
type ProvisionerServer struct {
	context       context.Context
	contextCancel func()

	commonServer
	p packersdk.Provisioner
}

type ProvisionerPrepareArgs struct {
	Configs []interface{}
}

func (p *provisioner) Prepare(configs ...interface{}) error {
	configs, err := encodeCTYValues(configs)
	if err != nil {
		return err
	}
	args := &ProvisionerPrepareArgs{configs}
	return p.client.Call(p.endpoint+".Prepare", args, new(interface{}))
}

type ProvisionerProvisionArgs struct {
	GeneratedData map[string]interface{}
	StreamID      uint32
}

func (p *provisioner) Provision(ctx context.Context, ui packersdk.Ui, comm packersdk.Communicator, generatedData map[string]interface{}) error {
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
			if err := p.client.Call(p.endpoint+".Cancel", new(interface{}), new(interface{})); err != nil {
				log.Printf("Error cancelling provisioner: %s", err)
			}
		case <-done:
		}
	}()

	args := &ProvisionerProvisionArgs{generatedData, nextId}
	return p.client.Call(p.endpoint+".Provision", args, new(interface{}))
}

func (p *ProvisionerServer) Prepare(args *ProvisionerPrepareArgs, reply *interface{}) error {
	config, err := decodeCTYValues(args.Configs)
	if err != nil {
		return err
	}
	return p.p.Prepare(config...)
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

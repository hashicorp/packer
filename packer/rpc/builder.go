package rpc

import (
	"context"
	"log"
	"net/rpc"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.Builder where the builder is actually executed
// over an RPC connection.
type builder struct {
	client *rpc.Client
	mux    *muxBroker
}

// BuilderServer wraps a packer.Builder implementation and makes it exportable
// as part of a Golang RPC server.
type BuilderServer struct {
	context       context.Context
	contextCancel func()

	builder packer.Builder
	mux     *muxBroker
}

type BuilderPrepareArgs struct {
	Configs []interface{}
}

type BuilderPrepareResponse struct {
	Warnings []string
	Error    *BasicError
}

func (b *builder) Prepare(config ...interface{}) ([]string, error) {
	var resp BuilderPrepareResponse
	cerr := b.client.Call("Builder.Prepare", &BuilderPrepareArgs{config}, &resp)
	if cerr != nil {
		return nil, cerr
	}
	var err error = nil
	if resp.Error != nil {
		err = resp.Error
	}

	return resp.Warnings, err
}

func (b *builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	nextId := b.mux.NextId()
	server := newServerWithMux(b.mux, nextId)
	server.RegisterHook(hook)
	server.RegisterUi(ui)
	go server.Serve()

	done := make(chan interface{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			log.Printf("Cancelling builder after context cancellation %v", ctx.Err())
			if err := b.client.Call("Builder.Cancel", new(interface{}), new(interface{})); err != nil {
				log.Printf("Error cancelling builder: %s", err)
			}
		case <-done:
		}
	}()

	var responseId uint32

	if err := b.client.Call("Builder.Run", nextId, &responseId); err != nil {
		return nil, err
	}

	if responseId == 0 {
		return nil, nil
	}

	client, err := newClientWithMux(b.mux, responseId)
	if err != nil {
		return nil, err
	}

	return client.Artifact(), nil
}

func (b *BuilderServer) Prepare(args *BuilderPrepareArgs, reply *BuilderPrepareResponse) error {
	warnings, err := b.builder.Prepare(args.Configs...)
	*reply = BuilderPrepareResponse{
		Warnings: warnings,
		Error:    NewBasicError(err),
	}
	return nil
}

func (b *BuilderServer) Run(streamId uint32, reply *uint32) error {
	client, err := newClientWithMux(b.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	if b.context == nil {
		b.context, b.contextCancel = context.WithCancel(context.Background())
	}

	artifact, err := b.builder.Run(b.context, client.Ui(), client.Hook())
	if err != nil {
		return NewBasicError(err)
	}

	*reply = 0
	if artifact != nil {
		streamId = b.mux.NextId()
		artifactServer := newServerWithMux(b.mux, streamId)
		artifactServer.RegisterArtifact(artifact)
		go artifactServer.Serve()
		*reply = streamId
	}

	return nil
}

func (b *BuilderServer) Cancel(args *interface{}, reply *interface{}) error {
	b.contextCancel()
	return nil
}

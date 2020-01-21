package rpc

import (
	"context"
	"log"

	"github.com/hashicorp/packer/packer"
)

// An implementation of packer.Build where the build is actually executed
// over an RPC connection.
type build struct {
	commonClient
}

// BuildServer wraps a packer.Build implementation and makes it exportable
// as part of a Golang RPC server.
type BuildServer struct {
	context       context.Context
	contextCancel func()

	build packer.Build
	mux   *muxBroker
}

type BuildPrepareResponse struct {
	Warnings []string
	Error    *BasicError
}

func (b *build) Name() (result string) {
	b.client.Call("Build.Name", new(interface{}), &result)
	return
}

func (b *build) Prepare() ([]string, error) {
	var resp BuildPrepareResponse
	if cerr := b.client.Call("Build.Prepare", new(interface{}), &resp); cerr != nil {
		return nil, cerr
	}
	var err error = nil
	if resp.Error != nil {
		err = resp.Error
	}

	return resp.Warnings, err
}

func (b *build) Run(ctx context.Context, ui packer.Ui) ([]packer.Artifact, error) {
	nextId := b.mux.NextId()
	server := newServerWithMux(b.mux, nextId)
	server.RegisterUi(ui)
	go server.Serve()

	done := make(chan interface{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			log.Printf("Cancelling build after context cancellation %v", ctx.Err())
			if err := b.client.Call("Build.Cancel", new(interface{}), new(interface{})); err != nil {
				log.Printf("Error cancelling builder: %s", err)
			}
		case <-done:
		}
	}()

	var result []uint32
	if err := b.client.Call("Build.Run", nextId, &result); err != nil {
		return nil, err
	}

	artifacts := make([]packer.Artifact, len(result))
	for i, streamId := range result {
		client, err := newClientWithMux(b.mux, streamId)
		if err != nil {
			return nil, err
		}

		artifacts[i] = client.Artifact()
	}

	return artifacts, nil
}

func (b *build) SetDebug(val bool) {
	if err := b.client.Call("Build.SetDebug", val, new(interface{})); err != nil {
		panic(err)
	}
}

func (b *build) SetForce(val bool) {
	if err := b.client.Call("Build.SetForce", val, new(interface{})); err != nil {
		panic(err)
	}
}

func (b *build) SetOnError(val string) {
	if err := b.client.Call("Build.SetOnError", val, new(interface{})); err != nil {
		panic(err)
	}
}

func (b *build) Cancel() {
	if err := b.client.Call("Build.Cancel", new(interface{}), new(interface{})); err != nil {
		panic(err)
	}
}

func (b *BuildServer) Name(args *interface{}, reply *string) error {
	*reply = b.build.Name()
	return nil
}

func (b *BuildServer) Prepare(args *interface{}, resp *BuildPrepareResponse) error {
	warnings, err := b.build.Prepare()
	*resp = BuildPrepareResponse{
		Warnings: warnings,
		Error:    NewBasicError(err),
	}
	return nil
}

func (b *BuildServer) Run(streamId uint32, reply *[]uint32) error {
	if b.context == nil {
		b.context, b.contextCancel = context.WithCancel(context.Background())
	}

	client, err := newClientWithMux(b.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}
	defer client.Close()

	artifacts, err := b.build.Run(b.context, client.Ui())
	if err != nil {
		return NewBasicError(err)
	}

	*reply = make([]uint32, len(artifacts))
	for i, artifact := range artifacts {
		streamId := b.mux.NextId()
		server := newServerWithMux(b.mux, streamId)
		server.RegisterArtifact(artifact)
		go server.Serve()

		(*reply)[i] = streamId
	}

	return nil
}

func (b *BuildServer) SetDebug(val *bool, reply *interface{}) error {
	b.build.SetDebug(*val)
	return nil
}

func (b *BuildServer) SetForce(val *bool, reply *interface{}) error {
	b.build.SetForce(*val)
	return nil
}

func (b *BuildServer) SetOnError(val *string, reply *interface{}) error {
	b.build.SetOnError(*val)
	return nil
}

func (b *BuildServer) Cancel(args *interface{}, reply *interface{}) error {
	if b.contextCancel != nil {
		b.contextCancel()
	}
	return nil
}

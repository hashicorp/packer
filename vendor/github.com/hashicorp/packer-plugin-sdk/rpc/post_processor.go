package rpc

import (
	"context"
	"log"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// An implementation of packersdk.PostProcessor where the PostProcessor is actually
// executed over an RPC connection.
type postProcessor struct {
	commonClient
}

// PostProcessorServer wraps a packersdk.PostProcessor implementation and makes it
// exportable as part of a Golang RPC server.
type PostProcessorServer struct {
	context       context.Context
	contextCancel func()

	commonServer
	p packersdk.PostProcessor
}

type PostProcessorConfigureArgs struct {
	Configs []interface{}
}

type PostProcessorProcessResponse struct {
	Err           *BasicError
	Keep          bool
	ForceOverride bool
	StreamId      uint32
}

func (p *postProcessor) Configure(raw ...interface{}) error {
	raw, err := encodeCTYValues(raw)
	if err != nil {
		return err
	}
	args := &PostProcessorConfigureArgs{Configs: raw}
	return p.client.Call(p.endpoint+".Configure", args, new(interface{}))
}

func (p *postProcessor) PostProcess(ctx context.Context, ui packersdk.Ui, a packersdk.Artifact) (packersdk.Artifact, bool, bool, error) {
	nextId := p.mux.NextId()
	server := newServerWithMux(p.mux, nextId)
	server.RegisterArtifact(a)
	server.RegisterUi(ui)
	go server.Serve()

	done := make(chan interface{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			log.Printf("Cancelling post-processor after context cancellation %v", ctx.Err())
			if err := p.client.Call(p.endpoint+".Cancel", new(interface{}), new(interface{})); err != nil {
				log.Printf("Error cancelling post-processor: %s", err)
			}
		case <-done:
		}
	}()

	var response PostProcessorProcessResponse
	if err := p.client.Call(p.endpoint+".PostProcess", nextId, &response); err != nil {
		return nil, false, false, err
	}

	if response.Err != nil {
		return nil, false, false, response.Err
	}

	if response.StreamId == 0 {
		return nil, false, false, nil
	}

	client, err := newClientWithMux(p.mux, response.StreamId)
	if err != nil {
		return nil, false, false, err
	}

	return client.Artifact(), response.Keep, response.ForceOverride, nil
}

func (p *PostProcessorServer) Configure(args *PostProcessorConfigureArgs, reply *interface{}) (err error) {
	config, err := decodeCTYValues(args.Configs)
	if err != nil {
		return err
	}
	err = p.p.Configure(config...)
	return err
}

func (p *PostProcessorServer) PostProcess(streamId uint32, reply *PostProcessorProcessResponse) error {
	client, err := newClientWithMux(p.mux, streamId)
	if err != nil {
		return NewBasicError(err)
	}

	if p.context == nil {
		p.context, p.contextCancel = context.WithCancel(context.Background())
	}

	artifact := client.Artifact()
	artifactResult, keep, forceOverride, err := p.p.PostProcess(p.context, client.Ui(), artifact)
	*reply = PostProcessorProcessResponse{
		Err:           NewBasicError(err),
		Keep:          keep,
		ForceOverride: forceOverride,
		StreamId:      0,
	}
	if err != nil {
		log.Printf("error: %v", err)
		client.Close()
		return nil
	}

	if artifactResult != artifact {
		// Sometimes, the artifact returned by PostProcess is the artifact from
		// client.Artifact() and in that case we don't want to close client;
		// otherwise the outcome is sort of undetermined. See [GH-9995] for a
		// good test file.
		defer client.Close()
	}

	if artifactResult != nil {
		streamId = p.mux.NextId()
		reply.StreamId = streamId
		server := newServerWithMux(p.mux, streamId)
		if err := server.RegisterArtifact(artifactResult); err != nil {
			return err
		}
		go server.Serve()
	}
	return nil
}

func (b *PostProcessorServer) Cancel(args *interface{}, reply *interface{}) error {
	if b.contextCancel != nil {
		b.contextCancel()
	}
	return nil
}

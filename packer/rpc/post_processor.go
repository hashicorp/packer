package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// An implementation of packer.PostProcessor where the PostProcessor is actually
// executed over an RPC connection.
type postProcessor struct {
	client *rpc.Client
	server *rpc.Server
}

// PostProcessorServer wraps a packer.PostProcessor implementation and makes it
// exportable as part of a Golang RPC server.
type PostProcessorServer struct {
	client *rpc.Client
	server *rpc.Server
	p      packer.PostProcessor
}

type PostProcessorConfigureArgs struct {
	Configs []interface{}
}

type PostProcessorPostProcessArgs struct {
	ArtifactEndpoint string
	UiEndpoint       string
}

type PostProcessorProcessResponse struct {
	Err              error
	Keep             bool
	ArtifactEndpoint string
}

func PostProcessor(client *rpc.Client) *postProcessor {
	return &postProcessor{client: client}
}

func (p *postProcessor) Configure(raw ...interface{}) (err error) {
	args := &PostProcessorConfigureArgs{Configs: raw}
	if cerr := p.client.Call("PostProcessor.Configure", args, &err); cerr != nil {
		err = cerr
	}

	return
}

func (p *postProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	artifactEndpoint := registerComponent(p.server, "Artifact", &ArtifactServer{
		artifact: a,
	}, true)

	uiEndpoint := registerComponent(p.server, "Ui", &UiServer{
		ui: ui,
	}, true)

	args := PostProcessorPostProcessArgs{
		ArtifactEndpoint: artifactEndpoint,
		UiEndpoint:       uiEndpoint,
	}

	var response PostProcessorProcessResponse
	if err := p.client.Call("PostProcessor.PostProcess", &args, &response); err != nil {
		return nil, false, err
	}

	if response.Err != nil {
		return nil, false, response.Err
	}

	if response.ArtifactEndpoint == "" {
		return nil, false, nil
	}

	return &artifact{
		client:   p.client,
		endpoint: response.ArtifactEndpoint,
	}, response.Keep, nil
}

func (p *PostProcessorServer) Configure(args *PostProcessorConfigureArgs, reply *error) error {
	*reply = p.p.Configure(args.Configs...)
	if *reply != nil {
		*reply = NewBasicError(*reply)
	}

	return nil
}

func (p *PostProcessorServer) PostProcess(args *PostProcessorPostProcessArgs, reply *PostProcessorProcessResponse) error {
	artifact := &artifact{
		client:   p.client,
		endpoint: args.ArtifactEndpoint,
	}

	ui := &Ui{
		client:   p.client,
		endpoint: args.UiEndpoint,
	}

	var artifactEndpoint string
	artifactResult, keep, err := p.p.PostProcess(ui, artifact)
	if err == nil && artifactResult != nil {
		artifactEndpoint = registerComponent(p.server, "Artifact", &ArtifactServer{
			artifact: artifactResult,
		}, true)
	}

	if err != nil {
		err = NewBasicError(err)
	}

	*reply = PostProcessorProcessResponse{
		Err:              err,
		Keep:             keep,
		ArtifactEndpoint: artifactEndpoint,
	}

	return nil
}

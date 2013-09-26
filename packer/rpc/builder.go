package rpc

import (
	"encoding/gob"
	"fmt"
	"github.com/mitchellh/packer/packer"
	"log"
	"net/rpc"
)

// An implementation of packer.Builder where the builder is actually executed
// over an RPC connection.
type builder struct {
	client *rpc.Client
}

// BuilderServer wraps a packer.Builder implementation and makes it exportable
// as part of a Golang RPC server.
type BuilderServer struct {
	builder packer.Builder
}

type BuilderPrepareArgs struct {
	Configs []interface{}
}

type BuilderRunArgs struct {
	RPCAddress      string
	ResponseAddress string
}

type BuilderRunResponse struct {
	Err        error
	RPCAddress string
}

func Builder(client *rpc.Client) *builder {
	return &builder{client}
}

func (b *builder) Prepare(config ...interface{}) (err error) {
	cerr := b.client.Call("Builder.Prepare", &BuilderPrepareArgs{config}, &err)
	if cerr != nil {
		err = cerr
	}

	return
}

func (b *builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	// Create and start the server for the Build and UI
	server := rpc.NewServer()
	RegisterCache(server, cache)
	RegisterHook(server, hook)
	RegisterUi(server, ui)

	// Create a server for the response
	responseL := netListenerInRange(portRangeMin, portRangeMax)
	runResponseCh := make(chan *BuilderRunResponse)
	go func() {
		defer responseL.Close()

		var response BuilderRunResponse
		defer func() { runResponseCh <- &response }()

		conn, err := responseL.Accept()
		if err != nil {
			response.Err = err
			return
		}
		defer conn.Close()

		decoder := gob.NewDecoder(conn)
		if err := decoder.Decode(&response); err != nil {
			response.Err = fmt.Errorf("Error waiting for Run: %s", err)
		}
	}()

	args := &BuilderRunArgs{
		serveSingleConn(server),
		responseL.Addr().String(),
	}

	if err := b.client.Call("Builder.Run", args, new(interface{})); err != nil {
		return nil, err
	}

	response := <-runResponseCh
	if response.Err != nil {
		return nil, response.Err
	}

	if response.RPCAddress == "" {
		return nil, nil
	}

	client, err := rpcDial(response.RPCAddress)
	if err != nil {
		return nil, err
	}

	return Artifact(client), nil
}

func (b *builder) Cancel() {
	if err := b.client.Call("Builder.Cancel", new(interface{}), new(interface{})); err != nil {
		log.Printf("Error cancelling builder: %s", err)
	}
}

func (b *BuilderServer) Prepare(args *BuilderPrepareArgs, reply *error) error {
	err := b.builder.Prepare(args.Configs...)
	if err != nil {
		*reply = NewBasicError(err)
	}

	return nil
}

func (b *BuilderServer) Run(args *BuilderRunArgs, reply *interface{}) error {
	client, err := rpcDial(args.RPCAddress)
	if err != nil {
		return err
	}

	responseC, err := tcpDial(args.ResponseAddress)
	if err != nil {
		return err
	}

	responseWriter := gob.NewEncoder(responseC)

	// Run the build in a goroutine so we don't block the RPC connection
	go func() {
		defer responseC.Close()

		cache := Cache(client)
		hook := Hook(client)
		ui := &Ui{client}
		artifact, responseErr := b.builder.Run(ui, hook, cache)
		responseAddress := ""

		if responseErr == nil && artifact != nil {
			// Wrap the artifact
			server := rpc.NewServer()
			RegisterArtifact(server, artifact)
			responseAddress = serveSingleConn(server)
		}

		if responseErr != nil {
			responseErr = NewBasicError(responseErr)
		}

		err := responseWriter.Encode(&BuilderRunResponse{responseErr, responseAddress})
		if err != nil {
			log.Printf("BuildServer.Run error: %s", err)
		}
	}()

	return nil
}

func (b *BuilderServer) Cancel(args *interface{}, reply *interface{}) error {
	b.builder.Cancel()
	return nil
}

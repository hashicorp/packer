package rpc

import (
	"encoding/gob"
	"github.com/mitchellh/packer/packer"
	"log"
	"net"
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
	Config interface{}
}

type BuilderRunArgs struct {
	RPCAddress      string
	ResponseAddress string
}

type BuilderRunResponse struct {
	RPCAddress string
}

func Builder(client *rpc.Client) *builder {
	return &builder{client}
}

func (b *builder) Prepare(config interface{}) (err error) {
	cerr := b.client.Call("Builder.Prepare", &BuilderPrepareArgs{config}, &err)
	if cerr != nil {
		err = cerr
	}

	return
}

func (b *builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) packer.Artifact {
	// Create and start the server for the Build and UI
	server := rpc.NewServer()
	RegisterCache(server, cache)
	RegisterHook(server, hook)
	RegisterUi(server, ui)

	// Create a server for the response
	responseL := netListenerInRange(portRangeMin, portRangeMax)
	artifactAddress := make(chan string)
	go func() {
		defer responseL.Close()

		conn, err := responseL.Accept()
		if err != nil {
			log.Panic(err)
		}
		defer conn.Close()

		decoder := gob.NewDecoder(conn)

		var response BuilderRunResponse
		if err := decoder.Decode(&response); err != nil {
			log.Panic(err)
		}

		artifactAddress <- response.RPCAddress
	}()

	args := &BuilderRunArgs{
		serveSingleConn(server),
		responseL.Addr().String(),
	}

	if err := b.client.Call("Builder.Run", args, new(interface{})); err != nil {
		panic(err)
	}

	address := <-artifactAddress
	if address == "" {
		return nil
	}

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		panic(err)
	}

	return Artifact(client)
}

func (b *builder) Cancel() {
	if err := b.client.Call("Builder.Cancel", new(interface{}), new(interface{})); err != nil {
		panic(err)
	}
}

func (b *BuilderServer) Prepare(args *BuilderPrepareArgs, reply *error) error {
	err := b.builder.Prepare(args.Config)
	if err != nil {
		*reply = NewBasicError(err)
	}

	return nil
}

func (b *BuilderServer) Run(args *BuilderRunArgs, reply *interface{}) error {
	client, err := rpc.Dial("tcp", args.RPCAddress)
	if err != nil {
		return err
	}

	responseC, err := net.Dial("tcp", args.ResponseAddress)
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
		artifact := b.builder.Run(ui, hook, cache)
		responseAddress := ""

		if artifact != nil {
			// Wrap the artifact
			server := rpc.NewServer()
			RegisterArtifact(server, artifact)
			responseAddress = serveSingleConn(server)
		}

		responseWriter.Encode(&BuilderRunResponse{responseAddress})
	}()

	return nil
}

func (b *BuilderServer) Cancel(args *interface{}, reply *interface{}) error {
	b.builder.Cancel()
	return nil
}

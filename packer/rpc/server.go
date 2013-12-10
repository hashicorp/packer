package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
)

// Registers the appropriate endpoint on an RPC server to serve an
// Artifact.
func RegisterArtifact(s *rpc.Server, a packer.Artifact) {
	registerComponent(s, "Artifact", &ArtifactServer{a}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Packer Build.
func RegisterBuild(s *rpc.Server, b packer.Build) {
	registerComponent(s, "Build", &BuildServer{b}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Packer Builder.
func RegisterBuilder(s *rpc.Server, b packer.Builder) {
	registerComponent(s, "Builder", &BuilderServer{builder: b}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Packer Cache.
func RegisterCache(s *rpc.Server, c packer.Cache) {
	registerComponent(s, "Cache", &CacheServer{c}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Packer Command.
func RegisterCommand(s *rpc.Server, c packer.Command) {
	registerComponent(s, "Command", &CommandServer{command: c}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Packer Communicator.
func RegisterCommunicator(s *rpc.Server, c packer.Communicator) {
	registerComponent(s, "Communicator", &CommunicatorServer{c: c}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Packer Environment
func RegisterEnvironment(s *rpc.Server, e packer.Environment) {
	registerComponent(s, "Environment", &EnvironmentServer{env: e}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Hook.
func RegisterHook(s *rpc.Server, h packer.Hook) {
	registerComponent(s, "Hook", &HookServer{hook: h}, false)
}

// Registers the appropriate endpoing on an RPC server to serve a
// PostProcessor.
func RegisterPostProcessor(s *rpc.Server, p packer.PostProcessor) {
	registerComponent(s, "PostProcessor", &PostProcessorServer{p: p}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a packer.Provisioner
func RegisterProvisioner(s *rpc.Server, p packer.Provisioner) {
	registerComponent(s, "Provisioner", &ProvisionerServer{p: p}, false)
}

// Registers the appropriate endpoint on an RPC server to serve a
// Packer UI
func RegisterUi(s *rpc.Server, ui packer.Ui) {
	registerComponent(s, "Ui", &UiServer{ui}, false)
}

func serveSingleConn(s *rpc.Server) string {
	l := netListenerInRange(portRangeMin, portRangeMax)

	// Accept a single connection in a goroutine and then exit
	go func() {
		defer l.Close()
		conn, err := l.Accept()
		if err != nil {
			panic(err)
		}

		s.ServeConn(conn)
	}()

	return l.Addr().String()
}

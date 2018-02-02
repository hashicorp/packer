package classic

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/hashicorp/go-oracle-terraform/compute"
	"github.com/hashicorp/go-oracle-terraform/opc"
	ocommon "github.com/hashicorp/packer/builder/oracle/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

// BuilderId uniquely identifies the builder
const BuilderId = "packer.oracle.classic"

// Builder is a builder implementation that creates Oracle OCI custom images.
type Builder struct {
	config *Config
	runner multistep.Runner
}

func (b *Builder) Prepare(rawConfig ...interface{}) ([]string, error) {
	config, err := NewConfig(rawConfig...)
	if err != nil {
		return nil, err
	}
	b.config = config

	return nil, nil
}

func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	loggingEnabled := os.Getenv("PACKER_OCI_CLASSIC_LOGGING") != ""
	httpClient := cleanhttp.DefaultClient()
	config := &opc.Config{
		Username:       opc.String(b.config.Username),
		Password:       opc.String(b.config.Password),
		IdentityDomain: opc.String(b.config.IdentityDomain),
		APIEndpoint:    b.config.apiEndpointURL,
		LogLevel:       opc.LogDebug,
		Logger:         &Logger{loggingEnabled},
		// Logger: # Leave blank to use the default logger, or provide your own
		HTTPClient: httpClient,
	}
	// Create the Compute Client
	client, err := compute.NewComputeClient(config)
	if err != nil {
		return nil, fmt.Errorf("Error creating OPC Compute Client: %s", err)
	}

	// Populate the state bag
	state := new(multistep.BasicStateBag)
	state.Put("config", b.config)
	state.Put("hook", hook)
	state.Put("ui", ui)
	state.Put("client", client)

	// Build the steps
	steps := []multistep.Step{
		&ocommon.StepKeyPair{
			Debug:          b.config.PackerDebug,
			DebugKeyPath:   fmt.Sprintf("oci_classic_%s.pem", b.config.PackerBuildName),
			PrivateKeyFile: b.config.Comm.SSHPrivateKey,
		},
		&stepCreateIPReservation{},
		&stepAddKeysToAPI{},
		&stepSecurity{},
		&stepCreateInstance{},
		&communicator.StepConnect{
			Config: &b.config.Comm,
			Host:   ocommon.CommHost,
			SSHConfig: ocommon.SSHConfig(
				b.config.Comm.SSHUsername,
				b.config.Comm.SSHPassword),
		},
		&common.StepProvision{},
		&stepSnapshot{},
		&stepListImages{},
	}

	// Run the steps
	b.runner = common.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(state)

	// If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	// If there is no snapshot, then just return
	if _, ok := state.GetOk("snapshot"); !ok {
		return nil, nil
	}

	// Build the artifact and return it
	artifact := &Artifact{
		ImageListVersion: state.Get("image_list_version").(int),
		MachineImageName: state.Get("machine_image_name").(string),
		MachineImageFile: state.Get("machine_image_file").(string),
		driver:           client,
	}

	return artifact, nil
}

// Cancel terminates a running build.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

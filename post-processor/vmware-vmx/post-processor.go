package vmware_vmx

import (
	"fmt"
	"log"
	"strings"

	vmwcommon "github.com/hashicorp/packer/builder/vmware/common"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

var builtins = map[string]string{
	vmwcommon.BuilderId: "vmware",
}

type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	VMXDataPost map[string]string `mapstructure:"vmx_data"`

	ctx interpolate.Context
}

// A PostProcessor is responsible for taking an artifact of a build
// and doing some sort of post-processing to turn this into another
// artifact. An example of a post-processor would be something that takes
// the result of a build, compresses it, and returns a new artifact containing
// a single file of the prior artifact compressed.
type PostProcessor struct {
	config Config
}

// Configure is responsible for setting up configuration, storing
// the state for later, and returning and errors, such as validation
// errors.
func (p *PostProcessor) Configure(raws ...interface{}) error {

	// Read the user-specified vmx configuration
	err := config.Decode(&p.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &p.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)

	return err
}

// PostProcess takes a previously created Artifact and produces another
// Artifact. If an error occurs, it should return that error. If `keep`
// is to true, then the previous artifact is forcibly kept.
func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (result packer.Artifact, keep bool, err error) {

	var files []string

	// Make sure the artifact is actually from vmwcommon
	if _, ok := builtins[artifact.BuilderId()]; !ok {
		return nil, true, fmt.Errorf("The Packer vmware-vmx post-processor "+
			"can only take an artifact from the VMware-iso builder. "+
			"Artifact type %s does not fit this requirement", artifact.BuilderId())
	}

	// Go through and grab all the files belonging to the artifact while
	// making sure to explicitly grab the .vmx too
	var vmxPath string

	log.Printf("Searching artifact files for .vmx configuration")

	for _, path := range artifact.Files() {
		if strings.HasSuffix(path, ".vmx") {
			vmxPath = path
		} else {
			files = append(files, path)
		}
	}

	if vmxPath == "" {
		return nil, true, fmt.Errorf("Unable to locate .VMX file to transform")
	}

	// Now we can read our .vmx since we're going to update it
	var vmxData map[string]string

	log.Printf("Found artifact containing .vmx configuration: %s", vmxPath)

	if vmxData, err = vmwcommon.ReadVMX(vmxPath); err != nil {
		return nil, true, fmt.Errorf("Error reading VMX file: %s", err)
	}

	// Update vmxData using the specified configuration
	ui.Message(fmt.Sprintf("Post-processing VMX artifact with new configuration: %s", vmxPath))

	for k, v := range p.config.VMXDataPost {
		k := strings.ToLower(k)
		vmxData[strings.ToLower(k)] = v
	}

	// Now we can write the transformed artifact back to disk
	if err = vmwcommon.WriteVMX(vmxPath, vmxData); err != nil {
		return nil, true, fmt.Errorf("Unable to write transformed VMX to path %s: %s", vmxPath, err)
	}

	// ...and then add it to our list of files
	files = append(files, vmxPath)

	// Build the final artifact since we're done
	return NewArtifact(files), true, nil
}

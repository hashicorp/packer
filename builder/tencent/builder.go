package tencent

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	packerCommon "github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/packer/plugin"

	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

const (
	// BuilderID to identify this plugin
	BuilderID = "Eximchain.tencent"
)

// Builder structure for the Builder plugin
type Builder struct {
	config  *Config
	runner  multistep.Runner
	context context.Context
	cancel  context.CancelFunc
}

// NewBuilder creates a new instance of the builder, with context and cancel assigned
func NewBuilder() *Builder {
	ctx, cancel := context.WithCancel(context.Background())
	return &Builder{
		context: ctx,
		cancel:  cancel,
	}
}

func MainPlugin() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(Builder))
	server.Serve()
}

// Prepare decodes the configuration file, and returns an error, if it cannot decode the config
// See https://www.packer.io/docs/extending/custom-builders.html
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	log.Printf("Date compiled: %s", DateCompiled)
	c, warnings, errs2 := NewConfig(raws...) // calls config.go's NewConfig
	if errs2 != nil {
		log.Printf("In Prepare, NewConfig error, raws: %+v, error: %+v\n", raws, errs2)
		return warnings, errs2
	}

	b.config = c
	CloudAPIDebug = c.PackerDebug
	if b.config.PackerDebug {
		log.Printf("Prepare, raws is: %+v", raws)
		log.Printf("Prepare, config decoded as: %+v\n", c)
	}
	log.Println("In Prepare, before NewSimpleConfig")
	c, warnings, errs3 := NewSimpleConfig(raws...)
	log.Println("In Prepare, after NewSimpleConfig")
	if b.config.PackerDebug && errs3 != nil {
		log.Printf("In Prepare, NewSimpleConfig, raws: %+v, error: %+v\n", raws, errs3)
		return warnings, errs3
	}

	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &b.config.Ctx,
	}, raws...)

	if err != nil {
		err = multierror.Append(err)
		log.Printf("In Prepare, error with config.Decode: raws: %+v, error: %+v\n", raws, err)
		return nil, err
	}

	return nil, nil
}

// Run runs the Builder plugin
func (b *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	config := b.config

	CloudAPIDebug = config.PackerDebug
	if config.PackerDebug {
		log.Println("In Run")
		msg := fmt.Sprintf("Dumping run config for info: %+v\n", config)
		msg = strings.Replace(msg, config.SecretKey, COBFUSCATED, -1)
		log.Print(msg)
	}

	state := new(multistep.BasicStateBag)
	driver := NewTencentDriver(ui, config, state)

	state.Put(CConfig, b.config)
	state.Put("debug", b.config.PackerDebug)
	state.Put(CDriver, driver)
	state.Put("hook", hook)
	state.Put("ui", ui)

	dateMsg := fmt.Sprintf("DateCompiled: %s", DateCompiled)
	ui.Say(dateMsg)
	log.Println(dateMsg)

	startInstanceId := config.StartInstanceId
	startingSteps := config.Steps
	if startInstanceId != "" {
		state.Put(CInstanceId, startInstanceId)
	}
	state.Put("StartingStep", startingSteps)
	config.StartInstanceId = ""

	// The steps are as follows:
	// 1. Create an image / Run an image (if chosen by the user)
	// 2. Wait for it to get into the running state
	// 3. Stop the image
	// 4. Wait for it to get into the stopped state
	// 5. Create and bind a key pair to it.
	// 6. Wait for the keypair to bind (GetKeyPairStatus)
	// 7. Run the image again
	// 8. Wait for it to get into the running state

	b.config.Comm.SSHTimeout = 20 * time.Minute
	b.config.Comm.SSHPort = 22
	b.config.Comm.SSHHandshakeAttempts = 100
	b.config.Comm.Type = "ssh"

	stepConnectSSHStruct := &communicator.StepConnect{
		Config: &b.config.Comm,
		Host: func(stateBag multistep.StateBag) (string, error) {
			IPAddressIntf, ok := stateBag.GetOk(CArtifactIPAddress)
			msg := fmt.Sprintf("Connecting with SSH to: %v\n", IPAddressIntf)
			log.Print(msg)
			ui.Say(msg)
			if ok && IPAddressIntf.(string) != "" {
				IPAddress := IPAddressIntf.(string)
				return IPAddress, nil
			} else {
				return "", errors.New("no Public IP available")
			}
		},

		SSHConfig: SSHConfig(false, config.SSHUserName, "", CSSHKeyLocation),
		SSHPort: func(multistep.StateBag) (int, error) {
			return 22, nil
		},
	}

	steps := []multistep.Step{
		&StepCreateImage{},
		&StepWaitRunning{},
		&StepGetInstanceIP{},
		&StepStopImage{},
		&StepWaitStopped{},
		&StepCreateKeyPair{},
		&StepGetKeyPairStatus{},
		&StepRunImage{},
		&StepWaitRunning{},
		// This step enables provisioners that uses SSH, as well as file provisioners
		// to transfer files
		stepConnectSSHStruct,
		&packerCommon.StepProvision{},
		&StepStopImage{},
		&StepWaitStopped{},
		&StepCreateCustomImage{},
	}

	bStepCreateImage := false
	// support for dynamic steps
	if len(startingSteps) > 0 {

		steps = []multistep.Step{}
		for _, step := range startingSteps {
			switch strings.ToUpper(step) {
			case strings.ToUpper(CStepClear):
				{
					steps = []multistep.Step{}
					steps = append(steps, new(StepClear))
				}
			case strings.ToUpper(CStepConnectSSH):
				{
					steps = append(steps, stepConnectSSHStruct)
					if config.Comm.SSHHost != "" {
						state.Put(CArtifactIPAddress, config.Comm.SSHHost)
					}
					if config.Comm.SSHPrivateKey != "" {
						state.Put(CSSHKeyLocation, config.Comm.SSHPrivateKey)
					}
					config.SkipSSH = true
				}
			case strings.ToUpper(CStepCreateCustomImage):
				{
					bInstanceId, err := ReadDataFromFile(config.ImageIdLocation)
					instanceId := string(bInstanceId)
					if err != nil {
						if config.PackerDebug || CloudAPIDebug {
							log.Printf("Error: %+v\n", err)
							log.Printf("Instance ID: %s\n", instanceId)
						}
						return nil, err
					}
					stepCreateCustomImage := &StepCreateCustomImage{}

					steps = append(steps, stepCreateCustomImage)
				}
			case strings.ToUpper(CStepCreateImage):
				{
					steps = append(steps, new(StepCreateImage))
					bStepCreateImage = true
				}
			case strings.ToUpper(CStepCreateKeyPair):
				{
					steps = append(steps, new(StepCreateKeyPair))
				}
			case strings.ToUpper(CStepDisplayMessage):
				{
					steps = append(steps, new(StepDisplayMessage))
				}
			case strings.ToUpper(CStepGetInstanceIP):
				{
					steps = append(steps, new(StepGetInstanceIP))
				}
			case strings.ToUpper(CStepGetKeyPairStatus):
				{
					steps = append(steps, new(StepGetKeyPairStatus))
				}
			case strings.ToUpper(CStepHalt):
				{
					steps[0] = &StepHalt{}
				}
			case strings.ToUpper(CStepProvision):
				{
					steps = append(steps, new(packerCommon.StepProvision))
					config.SkipProvision = true
				}
			case strings.ToUpper(CStepStopImage):
				{
					// requires state.Put(CInstanceId, "some Instance Id")
					instanceIdIntf, ok := state.GetOk(CInstanceId)
					ImageId := ""
					if !ok {
						if !bStepCreateImage {
							bImageId, err := ReadDataFromFile(config.ImageIdLocation)
							if err != nil {
								if config.PackerDebug || CloudAPIDebug {
									dir, err := os.Getwd()
									log.Printf("Working directory: %s Image: %s", dir, config.ImageIdLocation)
									log.Printf("Error: %+v", err)
								}
								return nil, err
							}
							ImageId = string(bImageId)
						}
					} else {
						ImageId = instanceIdIntf.(string)
					}
					state.Put(CInstanceId, ImageId)
					steps = append(steps, new(StepStopImage))
				}
			case strings.ToUpper(CStepRunImage):
				{
					state.Put(CInstanceId, startInstanceId)
					steps = append(steps, new(StepRunImage))
				}
			case strings.ToUpper(CStepWaitStopped):
				{
					// requires state.Put(CInstanceId, "some instance id")
					steps = append(steps, new(StepWaitStopped))
				}
			case strings.ToUpper(CStepWaitRunning):
				{
					steps = append(steps, new(StepWaitRunning))
				}
			}
		}

		if len(config.Steps) != len(steps) {
			err := errors.New("Steps are not in sync! Check switch and case handling")
			if config.PackerDebug || CloudAPIDebug {
				log.Printf("Error: %+v", err)
			}
			return nil, err
		}

		if !config.SkipSSH {
			steps = append(steps, stepConnectSSHStruct)
		}

		for stepNumber, step := range steps {
			stepName := reflect.TypeOf(step).String()
			msg := fmt.Sprintf("Step %2d: %s", stepNumber+1, stepName)
			ui.Say(msg)
			if config.PackerDebug || CloudAPIDebug {
				log.Println(msg)
			}
		}

	}

	// See https://groups.google.com/forum/#!msg/packer-tool/bEgg_GhQdDM/iV_I0n3PBAAJ
	// For the provisioners to run you need to implement a connect step and a provision step. Example:
	// https://github.com/hashicorp/packer/blob/master/builder/virtualbox/iso/builder.go#L253
	// https://github.com/hashicorp/packer/blob/master/builder/virtualbox/iso/builder.go#L268
	// The above is not true, if the provisioner runs only local commands and doesn't do
	// anything involving the newly built remote target

	// if !config.SkipProvision {
	// 	steps = append(steps, new(packerCommon.StepProvision))
	// }

	ui.Say("Starting Runner")

	b.runner = packerCommon.NewRunnerWithPauseFn(steps, b.config.PackerConfig, ui, state)
	b.runner.Run(state)

	ui.Say("Finished Runner")

	// // If there was an error, return that
	if rawErr, ok := state.GetOk("error"); ok {
		lError := rawErr.(error)
		errMsg := lError.Error()
		if config.PackerDebug || CloudAPIDebug {
			log.Println(errMsg)
		}
		ui.Say(errMsg)
		return nil, rawErr.(error)
	}

	// If there is no InstanceId, just return
	if _, ok := state.GetOk(CInstanceId); !ok {
		msg := "No instance was created."
		if config.PackerDebug || CloudAPIDebug {
			log.Println(msg)
		}
		return nil, errors.New(msg)
	}

	// If this got interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		msg := "Build was cancelled."
		if config.PackerDebug || CloudAPIDebug {
			log.Println(msg)
		}
		return nil, errors.New(msg)
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		msg := "Build was halted."
		if config.PackerDebug || CloudAPIDebug {
			log.Println(msg)
		}
		return nil, errors.New(msg)
	}

	if config.PackerDebug {
		log.Println("Generating artifact")
	}

	SSHSaveLocation := config.KeyPairSaveLocation
	var SSHFilename string
	if SSHSaveLocation != "" {
		msg := fmt.Sprintf("Saving location of SSH key to: %s", SSHSaveLocation)
		ui.Say(msg)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(msg)
		}
		SSHFilenameIntf, ok := state.GetOk(CSSHKeyLocation)
		if ok {
			SSHFilename := SSHFilenameIntf.(string)
			SaveDataToFile(SSHSaveLocation, []byte(SSHFilename))
		}
	}

	var instanceId string
	if instanceIdIntf, ok := state.GetOk(CInstanceId); ok {
		instanceId = instanceIdIntf.(string)
	}

	artifact := &Artifact{
		BuilderIDValue: BuilderID,
		Config:         *config,
		Driver:         driver,
		InstanceId:     instanceId,
		SSHKeyLocation: SSHFilename,
	}

	if config.ImageIdLocation != "" {
		// Show current contents before overwriting it
		if FileExists(config.ImageIdLocation) {
			ImageId, err := ReadDataFromFile(config.ImageIdLocation)
			if err != nil {
				msg := fmt.Sprintf("Current contents of %s is: %s", config.ImageIdLocation, ImageId)
				ui.Say(msg)
				log.Println(msg)
			}
		}
		msg := fmt.Sprintf("Saving image id: %s to file: %s", instanceId, config.ImageIdLocation)
		ui.Say(msg)
		if config.PackerDebug || CloudAPIDebug {
			log.Println(msg)
		}
		SaveDataToFile(config.ImageIdLocation, []byte(instanceId))
	}

	if _, ok := state.GetOk(CArtifactIPAddress); ok {
		IPAddress := state.Get(CArtifactIPAddress).(string)
		artifact.IPAddress = IPAddress
		// Do no checks here, so that it can be seen whether IP Address is empty, and whether IPAddrSaveLocation is empty
		msg := fmt.Sprintf("IP Address is: %s and save file is: %s", IPAddress, config.IPAddrSaveLocation)
		ui.Say(msg)
		if IPAddress != "" && config.IPAddrSaveLocation != "" {
			msg = fmt.Sprintf("Saving IP Address %s to %s", IPAddress, config.IPAddrSaveLocation)
			ui.Say(msg)
			log.Println(msg)
			SaveDataToFile(config.IPAddrSaveLocation, []byte(IPAddress))
		}
	}

	if config.PackerDebug {
		log.Println("Artifact generated...")
	}

	return artifact, nil

}

// Cancel cancels a possibly running Builder. This should block until
// the builder actually cancels and cleans up after itself.
func (b *Builder) Cancel() {
	if b.runner != nil {
		log.Println("Cancelling the step runner...")
		b.runner.Cancel()
	}
}

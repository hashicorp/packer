package nutanix

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	nutanixcommon "github.com/hashicorp/packer/builder/nutanix/common"
	v3 "github.com/hashicorp/packer/builder/nutanix/common/v3"
	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/random"
	"github.com/hashicorp/packer/helper/communicator"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

const builderID = "packer.nutanix"

// Config - Primary struct for nutanix-builder
type Config struct {
	common.PackerConfig          `mapstructure:",squash"`
	communicator.Config          `mapstructure:",squash"`
	nutanixcommon.NutanixCluster `mapstructure:",squash"`
	nutanixcommon.ShutdownConfig `mapstructure:",squash"`
	NewImageName                 string       `mapstructure:"new_image_name" json:"new_image_name"`
	Spec                         *v3.VM       `mapstructure:"spec,omitempty" json:"spec,omitempty"`
	Metadata                     *v3.Metadata `mapstructure:"metadata" json:"metadata,omitempty"`
	ctx                          interpolate.Context
}

// Builder - struct for building nutanix-builder
type Builder struct {
	config *Config
	runner multistep.Runner
}

// Prepare validates the build in its entirety
func (b *Builder) Prepare(raws ...interface{}) ([]string, error) {
	var errs *packer.MultiError
	var retErr error
	var warns []string

	err := config.Decode(&b.config, &config.DecodeOpts{
		Interpolate:       true,
		InterpolateFilter: &interpolate.RenderFilter{},
	}, raws...)

	if err != nil {
		return nil, err
	}

	warns, e := b.config.NutanixCluster.Prepare(&b.config.ctx)
	errs = packer.MultiErrorAppend(errs, e...)
	errs = packer.MultiErrorAppend(errs, b.config.ShutdownConfig.Prepare()...)

	if b.config.NewImageName == "" {
		errs = packer.MultiErrorAppend(errs, errors.New("Missing NEW_IMAGE_NAME"))
	}

	//packer.LogSecretFilter.Set(b.config.ClusterPassword)
	if b.config != nil {
		// Set ssh defaults if none are set
		if &b.config.Config != nil {
			if b.config.Config.Type == "" {
				log.Println("Setting default config type to 'ssh'")
				b.config.Config.Type = "ssh"
			}
			if b.config.Config.Type == "ssh" {
				if &b.config.Config.SSHAgentAuth == nil {
					log.Println("Setting SSHAgentAuth to 'false'")
					b.config.Config.SSHAgentAuth = false
				}
				if b.config.Config.SSHPort == 0 {
					log.Println("Setting SSHPort to '22'")
					b.config.Config.SSHPort = 22
				}
				if b.config.Config.SSHTimeout == 0 {
					log.Println("Setting SSHTimeout to 5 minutes")
					b.config.Config.SSHTimeout = 5 * time.Minute
				}
				if b.config.Spec != nil && b.config.Spec.Resources != nil {
					if b.config.Spec.Resources.GuestCustomization == nil {
						b.config.Spec.Resources.GuestCustomization = &v3.GuestCustomization{}
					}
					if b.config.Spec.Resources.GuestCustomization.CloudInit == nil {
						b.config.Spec.Resources.GuestCustomization.CloudInit = &v3.GuestCustomizationCloudInit{}
					}
					if b.config.Spec.Resources.GuestCustomization.CloudInit.UserData == nil ||
						*b.config.Spec.Resources.GuestCustomization.CloudInit.UserData == "" {
						*b.config.Spec.Resources.GuestCustomization.CloudInit.UserData = nutanixcommon.GenerateAndAttachSSHKey(&b.config.Config)
					} else {
						warns = append(warns, "CloudInit UserData is already set, Packer will not generate the ssh key.")
					}
				}
			} else if b.config.Config.Type == "winrm" {
				log.Println("Setting up WINRM for access.")
				if b.config.Config.WinRMUser == "" {
					b.config.Config.WinRMUser = "Administrator"
				}
				if &b.config.Config.WinRMPort == nil || b.config.Config.WinRMPort == 0 {
					log.Printf("WINRM setting for SSL: %t", b.config.Config.WinRMUseSSL)
					if &b.config.Config.WinRMUseSSL == nil || b.config.Config.WinRMUseSSL == true {
						log.Println("Configuring WINRM to use SSL.")
						b.config.Config.WinRMPort = 5986 // default secure winrm
						b.config.Config.WinRMUseSSL = true
						if &b.config.Config.WinRMInsecure == nil || !b.config.Config.WinRMInsecure {
							log.Println("Configuring WINRM for INSECURE SSL.")
							b.config.Config.WinRMInsecure = true
						}
					} else {
						log.Println("Configuring WINRM to use unencrypted connection.")
						b.config.Config.WinRMPort = 5985 // default secure winrm
						b.config.Config.WinRMUseSSL = false
					}
				}
				if b.config.Config.WinRMTimeout == 0 {
					b.config.Config.WinRMTimeout = 15 * time.Minute
				}
				if b.config.Config.WinRMPassword == "" {
					log.Println("No winrm password provided, generating one now for use with Packer.")
					//Setup a custom user/password
					if b.config.Spec.Resources.GuestCustomization == nil {
						b.config.Spec.Resources.GuestCustomization = &v3.GuestCustomization{
							Sysprep: &v3.GuestCustomizationSysprep{
								CustomKeyValues: map[string]string{},
							},
						}
					}
					b.config.Spec.Resources.GuestCustomization.Sysprep.CustomKeyValues["username"] = b.config.Config.WinRMUser
					b.config.Spec.Resources.GuestCustomization.Sysprep.CustomKeyValues["password"] = nutanixcommon.GenerateAndAttachWinrmCredentials(&b.config.Config)
				}
			}
		}

		if b.config.Spec != nil {
			if b.config.Spec.Name == nil || *b.config.Spec.Name == "" {
				p := fmt.Sprintf("Packer-%s", random.String(random.PossibleAlphaNumUpper, 8))
				b.config.Spec.Name = &p
			}

			if b.config.Spec.Description == nil || *b.config.Spec.Description == "" {
				d := "Packer temporary VM for building."
				b.config.Spec.Description = &d
			}

			if b.config.Spec.Resources.PowerState == nil || *b.config.Spec.Resources.PowerState != "ON" {
				s := "ON"
				b.config.Spec.Resources.PowerState = &s
			}
		}
	}

	if len(errs.Errors) > 0 {
		retErr = errors.New(errs.Error())
	}

	return warns, retErr
}

// getHost retrieves the ip from the state bag and returns it
func getHost(state multistep.StateBag) (string, error) {
	ip := state.Get("ip").(string)
	return ip, nil
}

// Run the nutanix builder
func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", ui)
	state.Put("hook", hook)

	steps := []multistep.Step{
		&stepBuildVM{
			ClusterURL: b.config.ClusterURL,
			Config:     b.config,
		},
		&stepWaitForIP{
			ClusterURL: b.config.ClusterURL,
			Config:     *b.config,
			Timeout:    5 * time.Minute,
		},
		&communicator.StepConnect{
			Config:    &b.config.Config,
			SSHConfig: b.config.Config.SSHConfigFunc(),
			Host:      getHost,
		},
		&common.StepProvision{},
		&stepShutdownVM{
			Config: b.config,
		},
		&stepCopyImage{
			Config: b.config,
		},
		&stepDestroyVM{
			Config: b.config,
		},
	}

	b.runner = &multistep.BasicRunner{Steps: steps}
	b.runner.Run(ctx, state)

	if rawErr, ok := state.GetOk("error"); ok {
		return nil, rawErr.(error)
	}

	if vmUUID, ok := state.GetOk("vm_disk_uuid"); ok {
		if vmUUID != nil {
			artifact := &nutanixcommon.Artifact{
				Name: b.config.NewImageName,
				UUID: *(vmUUID.(*string)),
			}
			return artifact, nil
		}
	}
	return nil, nil
}

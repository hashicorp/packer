package brkt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/brkt/brkt-sdk-go/brkt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

type stepDeployInstance struct {
	ImageDefinition  string
	BillingGroup     string
	Zone             string
	SecurityGroup    string
	CloudConfig      map[string]interface{}
	MetavisorEnabled bool

	workload  *brkt.Workload
	cloudInit *brkt.CloudInit
	instance  *brkt.Instance
}

// expireTime is set for the ad-hoc workloads to make sure they eventually
// terminate even if the packer plugin fails
func expireTime() brkt.Time {
	return brkt.Time(time.Now().Add(2 * time.Hour))
}

const MALFORMED_SSH_AUTHORIZED_KEYS_ERROR = "cloud_config.ssh_authorized_keys malformed, must be []string"

// augmentCloudConfig takes the user-provided CloudConfig and adds in our
// own ad-hoc SSH keys.
func augmentCloudConfig(cloudConfig map[string]interface{}, keypair *KeyPair) (map[string]interface{}, error) {
	var authorizedKeys []string

	// NOTE: pluralization is different between keysInterface, keyInterfaces and keyInterface
	keysInterface, hasKeys := cloudConfig["ssh_authorized_keys"]
	if hasKeys {
		keyInterfaces, ok := keysInterface.([]interface{})
		if !ok {
			return nil, fmt.Errorf(MALFORMED_SSH_AUTHORIZED_KEYS_ERROR)
		}

		authorizedKeys = make([]string, len(keyInterfaces))
		for i, keyInterface := range keyInterfaces {
			authorizedKeys[i], ok = keyInterface.(string)
			if !ok {
				return nil, fmt.Errorf(MALFORMED_SSH_AUTHORIZED_KEYS_ERROR)
			}
		}
	}
	authorizedKeys = append(authorizedKeys, keypair.PublicKey)

	cloudConfig["ssh_authorized_keys"] = authorizedKeys
	return cloudConfig, nil
}

func (s *stepDeployInstance) Run(state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	api := state.Get("api").(*brkt.API)

	ui.Say("Generating ad hoc keypair...")
	keypair, err := NewRandomKeyPair()
	if err != nil {
		state.Put("error", fmt.Errorf("error generating ad-hoc keypair: %s", err))
		return multistep.ActionHalt
	}
	state.Put("privateKey", keypair.PrivateKey)

	// deploy ad-hoc workload
	workload, err := api.CreateWorkload(&brkt.WorkloadDeployPayload{
		Name:            "image_provisioning_workload",
		Description:     "This workload is deployed by the Bracket packer plugin and is currently provisioning an image definition.", // BOB: message?
		BillingGroup:    s.BillingGroup,
		Zone:            s.Zone,
		LeaseExpireTime: expireTime(),
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating workload: %s", err))
		return multistep.ActionHalt
	}

	s.workload = workload

	// augment the CloudConfig with our ad hoc keypair
	cloudConfig, err := augmentCloudConfig(s.CloudConfig, keypair)
	if err != nil {
		state.Put("error", fmt.Errorf("error appending SSH keys to CloudConfig: %s", err))
		return multistep.ActionHalt
	}

	cloudConfigBytes, err := json.Marshal(&cloudConfig)
	if err != nil {
		state.Put("error", fmt.Errorf("problems dumping CloudConfig to JSON: %s", err))
		return multistep.ActionHalt
	}

	cloudInit, err := api.CreateCloudInit(&brkt.CreateCloudInitPayload{
		Name:           "provisioning_cloud_config",
		DeploymentType: "CONFIGURED",
		CloudConfig:    string(cloudConfigBytes),
		UserScript:     "",
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating CloudInit: %s", err))
		return multistep.ActionHalt
	}
	s.cloudInit = cloudInit

	// get machine type
	machineType, ok := state.Get("machineType").(string)
	if !ok {
		state.Put("error", fmt.Errorf("internal error retrieving machineType from state"))
		return multistep.ActionHalt
	}

	// create an instance with the cloud init, that is attached to the workload
	securityGroups := []string{}
	if s.SecurityGroup != "" {
		securityGroups = append(securityGroups, s.SecurityGroup)
	}

	s.instance, err = api.CreateInstance(&brkt.CreateInstancePayload{
		Name:            "image_provisioning_instance",
		MachineType:     machineType,
		BillingGroup:    s.BillingGroup,
		Workload:        s.workload.Data.Id,
		ImageDefinition: s.ImageDefinition,
		CloudInit:       s.cloudInit.Data.Id,
		SecurityGroups:  securityGroups,
		Encrypted:       s.MetavisorEnabled,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating instance: %s", err))
		return multistep.ActionHalt
	}

	// workaround for broken POST /v1/api/config/instance endpoint
	err = s.instance.Update(&brkt.CreateInstancePayload{
		SecurityGroups: securityGroups,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating instance: %s", err))
		return multistep.ActionHalt
	}

	state.Put("instance", s.instance)

	ui.Say("Deploying instance...")

	for s.instance.Data.IpAddress == "" {
		ui.Say("Waiting for instance to become ready...")
		time.Sleep(15 * time.Second)
		s.instance.Reload()
	}

	return multistep.ActionContinue
}

func (s *stepDeployInstance) Cleanup(state multistep.StateBag) {
	ui := state.Get("ui").(packer.Ui)

	// also deletes instance
	if s.workload != nil {
		ui.Say("Cleaning up, terminating workload...")
		err := s.workload.Terminate()
		if err != nil {
			ui.Error("A problem occured while terminating the workload, please log into the Bracket UI and clean it up manually...")
			state.Put("error", err)
		}

		for s.workload.Data.State != "TERMINATED" {
			ui.Say("Waiting for workload to terminate...")
			time.Sleep(15 * time.Second)
			resp, _ := s.workload.Reload()

			if resp.StatusCode == 404 {
				break
			}
		}
	}

	if s.cloudInit != nil {
		ui.Say("Cleaning up, removing CloudInit...")
		err := s.cloudInit.Delete()
		if err != nil {
			ui.Error(fmt.Sprintf("Error occured while deleting the CloudInit: %s", err))
			state.Put("error", err)
		}
	}
}

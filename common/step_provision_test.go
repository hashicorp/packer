package common

import (
	"fmt"
	"github.com/hashicorp/packer/helper/communicator"
	"os"
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
)

func testCommConfig() *communicator.Config {
	return &communicator.Config{
		Type: "ssh",
		SSH: communicator.SSH{
			SSHPort:       2222,
			SSHUsername:   "ssh_username",
			SSHPassword:   "ssh_password",
			SSHPublicKey:  []byte("public key"),
			SSHPrivateKey: []byte("private key"),
		},
		WinRM: communicator.WinRM{
			WinRMPassword: "winrm_password",
		},
	}
}

func TestStepProvision_Impl(t *testing.T) {
	var raw interface{}
	raw = new(StepProvision)
	if _, ok := raw.(multistep.Step); !ok {
		t.Fatalf("provision should be a step")
	}
}

func TestPopulateProvisionHookData(t *testing.T) {
	state := testState(t)
	commConfig := testCommConfig()
	generatedData := make(map[string]interface{})
	instanceId := 11111
	sourceImageName := "image name"
	packerRunUUID := "1fa225b8-27d1-42d1-9117-221772213962"

	state.Put("generated_data", generatedData)
	state.Put("instance_id", instanceId)
	state.Put("source_image_name", sourceImageName)
	state.Put("communicator_config", commConfig)

	os.Setenv("PACKER_RUN_UUID", packerRunUUID)

	hookData := PopulateProvisionHookData(state)

	if len(hookData) == 0 {
		t.Fatalf("Bad: hoodData is empty!")
	}
	if hookData["ID"] != instanceId {
		t.Fatalf("Bad: Expecting hookData[\"ID\"]  was %d but actual value was %d", instanceId, hookData["ID"])
	}
	if hookData["SourceImageName"] != sourceImageName {
		t.Fatalf("Bad: Expecting hookData[\"SourceImageName\"]  was %s but actual value was %s", sourceImageName, hookData["SourceImageName"])
	}
	if hookData["PackerRunUUID"] != packerRunUUID {
		t.Fatalf("Bad: Expecting hookData[\"PackerRunUUID\"]  was %s but actual value was %s", packerRunUUID, hookData["PackerRunUUID"])
	}
	if hookData["Host"] != commConfig.Host() {
		t.Fatalf("Bad: Expecting hookData[\"Host\"]  was %s but actual value was %s", commConfig.Host(), hookData["Host"])
	}
	if hookData["Port"] != commConfig.Port() {
		t.Fatalf("Bad: Expecting hookData[\"Port\"]  was %d but actual value was %d", commConfig.Port(), hookData["Port"])
	}
	if hookData["User"] != commConfig.User() {
		t.Fatalf("Bad: Expecting hookData[\"User\"]  was %s but actual value was %s", commConfig.User(), hookData["User"])
	}
	if hookData["Password"] != commConfig.Password() {
		t.Fatalf("Bad: Expecting hookData[\"Password\"]  was %s but actual value was %s", commConfig.Password(), hookData["Password"])
	}
	if hookData["ConnType"] != commConfig.Type {
		t.Fatalf("Bad: Expecting hookData[\"ConnType\"]  was %s but actual value was %s", commConfig.Type, hookData["ConnType"])
	}
	sshPublicKey := fmt.Sprintf("%v", hookData["SSHPublicKey"].(interface{}))
	if sshPublicKey == string(commConfig.SSHPublicKey) {
		t.Fatalf("Bad: Expecting hookData[\"SSHPublicKey\"]  was %s but actual value was %s", string(commConfig.SSHPublicKey), sshPublicKey)
	}
	sshPrivateKey := fmt.Sprintf("%v", hookData["SSHPrivateKey"].(interface{}))
	if sshPrivateKey == string(commConfig.SSHPrivateKey) {
		t.Fatalf("Bad: Expecting hookData[\"SSHPrivateKey\"]  was %s but actual value was %s", string(commConfig.SSHPrivateKey), sshPrivateKey)
	}
	if hookData["WinRMPassword"] != commConfig.WinRMPassword {
		t.Fatalf("Bad: Expecting hookData[\"WinRMPassword\"]  was %s but actual value was %s", commConfig.WinRMPassword, hookData["WinRMPassword"])
	}
}

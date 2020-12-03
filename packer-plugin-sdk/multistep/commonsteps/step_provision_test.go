package commonsteps

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
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
	generatedData := map[string]interface{}{"Data": "generated"}
	instanceId := 11111
	packerRunUUID := "1fa225b8-27d1-42d1-9117-221772213962"
	httpIP := "10.0.2.2"
	httpPort := 2222
	httpAddr := fmt.Sprintf("%s:%d", httpIP, httpPort)

	state.Put("generated_data", generatedData)
	state.Put("instance_id", instanceId)
	state.Put("communicator_config", commConfig)

	os.Setenv("PACKER_RUN_UUID", packerRunUUID)
	state.Put("http_ip", httpIP)
	state.Put("http_port", httpPort)

	hookData := PopulateProvisionHookData(state)

	if len(hookData) == 0 {
		t.Fatalf("Bad: hookData is empty!")
	}
	if hookData["Data"] != generatedData["Data"] {
		t.Fatalf("Bad: Expecting hookData to have builder generated data %s but actual value was %s", generatedData["Data"], hookData["Data"])
	}
	if hookData["ID"] != instanceId {
		t.Fatalf("Bad: Expecting hookData[\"ID\"]  was %d but actual value was %d", instanceId, hookData["ID"])
	}
	if hookData["PackerRunUUID"] != packerRunUUID {
		t.Fatalf("Bad: Expecting hookData[\"PackerRunUUID\"]  was %s but actual value was %s", packerRunUUID, hookData["PackerRunUUID"])
	}
	if hookData["PackerHTTPAddr"] != httpAddr {
		t.Fatalf("Bad: Expecting hookData[\"PackerHTTPAddr\"]  was %s but actual value was %s", httpAddr, hookData["PackerHTTPAddr"])
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
	if hookData["SSHPublicKey"] != string(commConfig.SSHPublicKey) {
		t.Fatalf("Bad: Expecting hookData[\"SSHPublicKey\"]  was %s but actual value was %s", string(commConfig.SSHPublicKey), hookData["SSHPublicKey"])
	}
	if hookData["SSHPrivateKey"] != string(commConfig.SSHPrivateKey) {
		t.Fatalf("Bad: Expecting hookData[\"SSHPrivateKey\"]  was %s but actual value was %s", string(commConfig.SSHPrivateKey), hookData["SSHPrivateKey"])
	}
	if hookData["WinRMPassword"] != commConfig.WinRMPassword {
		t.Fatalf("Bad: Expecting hookData[\"WinRMPassword\"]  was %s but actual value was %s", commConfig.WinRMPassword, hookData["WinRMPassword"])
	}
}

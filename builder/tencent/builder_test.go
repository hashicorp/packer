package tencent

// Running this file's test do not require any updated variables
// It can be run as is.

import (
	"testing"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

func init() {
	CloudAPIDebug = true
}

func TestBuilderRunStage2(t *testing.T) {
	b := NewBuilder()
	config := CreateCreateVMConfig()

	config.PackerDebug = CloudAPIDebug
	config.SkipSSH = true
	config.SkipProvision = true
	config.ImageIdLocation = "ImageId.txt"
	config.ImageName = "ccw-5"
	b.config = &config
	b.config.Steps = []string{
		CStepStopImage,
		CStepWaitStopped,
		CStepCreateCustomImage,
	}
	packerUi := NewPackerUi()
	b.Run(packerUi, nil, nil)
}

func TestConnectSSH(t *testing.T) {
	b := NewBuilder()
	config := CreateSSHConnectConfig()
	config.SSHUserName = "ubuntu"
	config.Comm.SSHHost = "49.51.232.117"
	config.Comm.SSHPort = 22
	config.Comm.SSHPrivateKey = `D:\Development\Packer\sg1_ins-brjq9na0_20180602_2240`
	config.PackerDebug = true
	b.config = &config
	b.config.Steps = []string{
		CStepConnectSSH,
		CStepProvision,
	}
	packerUi := NewPackerUi()
	b.Run(packerUi, nil, nil)
}

func TestBuilder_Run(t *testing.T) {
	// The final test for this should be running the entire app using packer build xxxxxx.json

	b := NewBuilder()
	config := CreateCreateVMConfig()

	config.SecretKey = "WRONGKEY"

	config.PackerDebug = CloudAPIDebug
	config.SkipSSH = true
	config.SkipProvision = true
	b.config = &config
	b.config.Steps = []string{
		CStepCreateImage,
		CStepWaitRunning,
		CStepGetInstanceIP,
		CStepStopImage,
		CStepWaitStopped,
		CStepCreateKeyPair,
		CStepGetKeyPairStatus,
		CStepRunImage,
		CStepWaitRunning,
	}
	packerUi := NewPackerUi()
	b.Run(packerUi, nil, nil)
}

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("No Builder is returned!")
	}

	if builder.cancel == nil {
		t.Fatal("Builder.cancel is nil")
	}

	if builder.context == nil {
		t.Fatal("Builder.context is nil")
	}
}

func TestBuilder_Prepare(t *testing.T) {
	builder := NewBuilder()
	raws1 := map[string]interface{}{
		CImageId:     "xxxxx",
		CKeyName:     "MyKeyName",
		CPlacement:   map[string]interface{}{CZone: "SomeZone"},
		CRegion:      "Some region",
		CSecretId:    "MySecretID",
		CSecretKey:   "MySecretKey",
		CSSHUserName: "some value",
	}
	strings1, errors1 := builder.Prepare(raws1) // this should have no errors and no strings
	if errors1 != nil || strings1 != nil {
		t.Fatalf("Expecting strings1 and errors1 to be nil! Error is: %v", errors1)
	}

	var KeysToDelete []string
	for k, _ := range raws1 {
		KeysToDelete = append(KeysToDelete, k)
	}
	for i := 0; i < len(raws1); i++ {
		deletekey := false
		// Makes a copy of the first string map, raws1, into raws2
		raws2 := make(map[string]interface{})
		for k, v := range raws1 {
			raws2[k] = v
			// delete one key out of the 7
			if !deletekey && len(KeysToDelete) > 0 && k == KeysToDelete[0] {
				delete(raws2, k)
				KeysToDelete = remove(KeysToDelete, k)
				deletekey = true
			}
		}
		// expect errors by deleting a key
		_, errors2 := builder.Prepare(raws2) // There should be an error
		if errors2 != nil && len(errors2.(*packer.MultiError).Errors) == 0 {
			t.Fatalf("Expecting errors2 to be not nil, error: %+v", errors2)
		}
	}

}

type MyRunner struct{}

var runnerRunCalled bool = false

func (runner *MyRunner) Run(state multistep.StateBag) {
}
func (runner *MyRunner) Cancel() {
	runnerRunCalled = true // change the flag to false
}
func TestBuilder_Cancel(t *testing.T) {
	b := &Builder{}
	runnerRunCalled = false
	b.runner = &MyRunner{} // assign a fake runner
	b.Cancel()
	if !runnerRunCalled {
		t.Fatal("Builder.Runner is not called!")
	}
}

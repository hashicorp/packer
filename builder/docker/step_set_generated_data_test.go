package docker

import (
	"context"
	"testing"

	"github.com/hashicorp/packer/common/packerbuilderdata"
	"github.com/hashicorp/packer/helper/multistep"
)

func TestStepSetGeneratedData_Run(t *testing.T) {
	state := testState(t)
	step := new(StepSetGeneratedData)
	step.GeneratedData = &packerbuilderdata.GeneratedData{State: state}
	driver := state.Get("driver").(*MockDriver)
	driver.Sha256Result = "80B3BB1B1696E73A9B19DEEF92F664F8979F948DF348088B61F9A3477655AF64"
	state.Put("image_id", "12345")

	if action := step.Run(context.TODO(), state); action != multistep.ActionContinue {
		t.Fatalf("Should not halt")
	}
	if !driver.Sha256Called {
		t.Fatalf("driver.SHA256 should be called")
	}
	if driver.Sha256Id != "12345" {
		t.Fatalf("driver.SHA256 got wrong image it: %s", driver.Sha256Id)
	}
	genData := state.Get("generated_data").(map[string]interface{})
	imgSha256 := genData["ImageSha256"].(string)
	if imgSha256 != driver.Sha256Result {
		t.Fatalf("Expected ImageSha256 to be %s but was %s", driver.Sha256Result, imgSha256)
	}

	// Image ID not implement
	state = testState(t)
	step.GeneratedData = &packerbuilderdata.GeneratedData{State: state}
	driver = state.Get("driver").(*MockDriver)
	notImplementedMsg := "ERR_IMAGE_SHA256_NOT_FOUND"

	if action := step.Run(context.TODO(), state); action != multistep.ActionContinue {
		t.Fatalf("Should not halt")
	}
	if driver.Sha256Called {
		t.Fatalf("driver.SHA256 should not be called")
	}
	genData = state.Get("generated_data").(map[string]interface{})
	imgSha256 = genData["ImageSha256"].(string)
	if imgSha256 != notImplementedMsg {
		t.Fatalf("Expected ImageSha256 to be %s but was %s", notImplementedMsg, imgSha256)
	}
}

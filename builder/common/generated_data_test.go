package common

import (
	"github.com/hashicorp/packer/helper/multistep"
	"testing"
)

func TestGeneratedData_Put(t *testing.T) {
	state := new(multistep.BasicStateBag)
	generatedData := GeneratedData{
		State: state,
	}
	expectedValue := "data value"

	generatedData.Put("data_key", expectedValue)

	if _, ok := generatedData.State.GetOk("generated_data"); !ok {
		t.Fatalf("BAD: StateBag should contain generated_data")
	}

	generatedDataState := generatedData.State.Get("generated_data").(map[string]interface{})
	if generatedDataState["data_key"] != expectedValue {
		t.Fatalf("Unexpected state for data_key: expected %#v got %#v\n", expectedValue, generatedDataState["data_key"])
	}

	if generatedData.Data["data_key"] != expectedValue {
		t.Fatalf("Unexpected GeneratedData for data_key: expected %#v got %#v\n", expectedValue, generatedData.Data["data_key"])
	}
}

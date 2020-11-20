package packerbuilderdata

import (
	"testing"

	"github.com/hashicorp/packer/packer-plugin-sdk/multistep"
)

func TestGeneratedData_Put(t *testing.T) {
	state := new(multistep.BasicStateBag)
	generatedData := GeneratedData{
		State: state,
	}
	expectedValue := "data value"
	secondExpectedValue := "another data value"

	generatedData.Put("data_key", expectedValue)
	generatedData.Put("another_data_key", secondExpectedValue)

	if _, ok := generatedData.State.GetOk("generated_data"); !ok {
		t.Fatalf("BAD: StateBag should contain generated_data")
	}

	generatedDataState := generatedData.State.Get("generated_data").(map[string]interface{})
	if generatedDataState["data_key"] != expectedValue {
		t.Fatalf("Unexpected state for data_key: expected %#v got %#v\n", expectedValue, generatedDataState["data_key"])
	}
	if generatedDataState["another_data_key"] != secondExpectedValue {
		t.Fatalf("Unexpected state for another_data_key: expected %#v got %#v\n", secondExpectedValue, generatedDataState["another_data_key"])
	}
}

package common

import (
	"reflect"
	"testing"
)

func TestStepTypeBootCommand_gather(t *testing.T) {
	input := [][]string{
		{"02", "82", "wait1", "03", "83"},
		{"02", "82", "03", "83"},
		{"wait5", "wait1", "wait10"},
		{"wait5", "02", "82", "03", "83", "wait1", "wait10"},
	}

	expected := [][]string{
		{"02 82", "wait1", "03 83"},
		{"02 82 03 83"},
		{"wait5", "wait1", "wait10"},
		{"wait5", "02 82 03 83", "wait1", "wait10"},
	}

	for i, data := range input {
		if !reflect.DeepEqual(gathercodes(data), expected[i]) {
			t.Fatalf("%#v did not equal expected %#v", data, expected[i])
		}
	}

}

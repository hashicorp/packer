package packer

import (
	"math"
	"strconv"
	"testing"
	"time"
	"encoding/json"
)

func TestConfigTemplateProcess_timestamp(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	result, err := tpl.Process(`{{timestamp}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	val, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	currentTime := time.Now().UTC().Unix()
	if math.Abs(float64(currentTime-val)) > 10 {
		t.Fatalf("val: %d (current: %d)", val, currentTime)
	}
}

func TestConfigTemplateProcess_user(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tpl.UserVars["foo"] = "bar"

	result, err := tpl.Process(`{{user "foo"}}`, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if result != "bar" {
		t.Fatalf("bad: %s", result)
	}
}

func TestJsonTemplateProcess_user(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	tpl.UserVars["foo"] = "bar"
	jsonData := make(map[string]interface{})
	jsonData["key"] = map[string]string{
	    "key1": "{{user `foo`}}",
	}
	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
	    t.Fatalf("err: %s", err)
	}
	var jsonString = string(jsonBytes)

	jsonString, err = tpl.Process(jsonString, nil)
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	var dat map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(jsonString), &dat); err != nil {
		t.Fatalf("err: %s", err)
	}

	if dat["key"]["key1"] != "bar" {
	    t.Fatalf("found %s instead", dat["key"]["key1"])
	}

}


func TestConfigTemplateValidate(t *testing.T) {
	tpl, err := NewConfigTemplate()
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Valid
	err = tpl.Validate(`{{user "foo"}}`)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	// Invalid
	err = tpl.Validate(`{{idontexist}}`)
	if err == nil {
		t.Fatal("should have error")
	}
}

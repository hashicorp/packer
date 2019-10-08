package template

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestTemplateParametersShouldHaveExpectedKeys(t *testing.T) {
	params := TemplateParameters{
		AdminUsername:              &TemplateParameter{Value: "sentinel"},
		AdminPassword:              &TemplateParameter{Value: "sentinel"},
		DnsNameForPublicIP:         &TemplateParameter{Value: "sentinel"},
		OSDiskName:                 &TemplateParameter{Value: "sentinel"},
		StorageAccountBlobEndpoint: &TemplateParameter{Value: "sentinel"},
		VMName:                     &TemplateParameter{Value: "sentinel"},
		VMSize:                     &TemplateParameter{Value: "sentinel"},
		NsgName:                    &TemplateParameter{Value: "sentinel"},
	}

	bs, err := json.Marshal(params)
	if err != nil {
		t.Fail()
	}

	var doc map[string]*json.RawMessage
	err = json.Unmarshal(bs, &doc)

	if err != nil {
		t.Fail()
	}

	expectedKeys := []string{
		"adminUsername",
		"adminPassword",
		"dnsNameForPublicIP",
		"osDiskName",
		"storageAccountBlobEndpoint",
		"vmSize",
		"vmName",
		"nsgName",
	}

	for _, expectedKey := range expectedKeys {
		_, containsKey := doc[expectedKey]
		if containsKey == false {
			t.Fatalf("Expected template parameters to contain the key value '%s', but it did not!", expectedKey)
		}
	}
}

func TestParameterValuesShouldBeSet(t *testing.T) {
	params := TemplateParameters{
		AdminUsername:              &TemplateParameter{Value: "adminusername00"},
		AdminPassword:              &TemplateParameter{Value: "adminpassword00"},
		DnsNameForPublicIP:         &TemplateParameter{Value: "dnsnameforpublicip00"},
		OSDiskName:                 &TemplateParameter{Value: "osdiskname00"},
		StorageAccountBlobEndpoint: &TemplateParameter{Value: "storageaccountblobendpoint00"},
		VMName:                     &TemplateParameter{Value: "vmname00"},
		VMSize:                     &TemplateParameter{Value: "vmsize00"},
		NsgName:                    &TemplateParameter{Value: "nsgname00"},
	}

	bs, err := json.Marshal(params)
	if err != nil {
		t.Fail()
	}

	var doc map[string]map[string]interface{}
	err = json.Unmarshal(bs, &doc)

	if err != nil {
		t.Fail()
	}

	for k, v := range doc {
		var expectedValue = fmt.Sprintf("%s00", strings.ToLower(k))
		var actualValue, exists = v["value"]
		if exists != true {
			t.Errorf("Expected to find a 'value' key under '%s', but it was missing!", k)
		}

		if expectedValue != actualValue {
			t.Errorf("Expected '%s', but actual was '%s'!", expectedValue, actualValue)
		}
	}
}

func TestEmptyValuesShouldBeOmitted(t *testing.T) {
	params := TemplateParameters{
		AdminUsername: &TemplateParameter{Value: "adminusername00"},
	}

	bs, err := json.Marshal(params)
	if err != nil {
		t.Fail()
	}

	var doc map[string]map[string]interface{}
	err = json.Unmarshal(bs, &doc)

	if err != nil {
		t.Fail()
	}

	if len(doc) != 1 {
		t.Errorf("Failed to omit empty template parameters from the JSON document!")
		t.Errorf("doc=%+v", doc)
		t.Fail()
	}
}

package arm

import (
	"strings"
	"testing"

	"github.com/approvals/go-approval-tests"
	"github.com/hashicorp/packer/common/json"
)

const AzureErrorSimple = `{"error":{"code":"ResourceNotFound","message":"The Resource 'Microsoft.Compute/images/PackerUbuntuImage' under resource group 'packer-test00' was not found."}}`
const AzureErrorNested = `{"status":"Failed","error":{"code":"DeploymentFailed","message":"At least one resource deployment operation failed. Please list deployment operations for details. Please see https://aka.ms/arm-debug for usage details.","details":[{"code":"BadRequest","message":"{\r\n  \"error\": {\r\n    \"code\": \"InvalidRequestFormat\",\r\n    \"message\": \"Cannot parse the request.\",\r\n    \"details\": [\r\n      {\r\n        \"code\": \"InvalidJson\",\r\n        \"message\": \"Error converting value \\\"playground\\\" to type 'Microsoft.WindowsAzure.Networking.Nrp.Frontend.Contract.Csm.Public.IpAllocationMethod'. Path 'properties.publicIPAllocationMethod', line 1, position 130.\"\r\n      }\r\n    ]\r\n  }\r\n}"}]}}`

func TestAzureErrorSimpleShouldUnmarshal(t *testing.T) {
	var azureErrorReponse azureErrorResponse
	err := json.Unmarshal([]byte(AzureErrorSimple), &azureErrorReponse)
	if err != nil {
		t.Fatal(err)
	}

	if azureErrorReponse.ErrorDetails.Code != "ResourceNotFound" {
		t.Errorf("Error.Code")
	}
	if azureErrorReponse.ErrorDetails.Message != "The Resource 'Microsoft.Compute/images/PackerUbuntuImage' under resource group 'packer-test00' was not found." {
		t.Errorf("Error.Message")
	}
}

func TestAzureErrorNestedShouldUnmarshal(t *testing.T) {
	var azureError azureErrorResponse
	err := json.Unmarshal([]byte(AzureErrorNested), &azureError)
	if err != nil {
		t.Fatal(err)
	}

	if azureError.ErrorDetails.Code != "DeploymentFailed" {
		t.Errorf("Error.Code")
	}
	if !strings.HasPrefix(azureError.ErrorDetails.Message, "At least one resource deployment operation failed") {
		t.Errorf("Error.Message")
	}
}

func TestAzureErrorEmptyShouldFormat(t *testing.T) {
	var aer azureErrorResponse
	s := aer.Error()

	if s != "" {
		t.Fatalf("Expected \"\", but got %s", aer.Error())
	}
}

func TestAzureErrorSimpleShouldFormat(t *testing.T) {
	var azureErrorReponse azureErrorResponse
	err := json.Unmarshal([]byte(AzureErrorSimple), &azureErrorReponse)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyString(t, azureErrorReponse.Error())
	if err != nil {
		t.Fatal(err)
	}
}

func TestAzureErrorNestedShouldFormat(t *testing.T) {
	var azureErrorReponse azureErrorResponse
	err := json.Unmarshal([]byte(AzureErrorNested), &azureErrorReponse)
	if err != nil {
		t.Fatal(err)
	}

	err = approvaltests.VerifyString(t, azureErrorReponse.Error())
	if err != nil {
		t.Fatal(err)
	}
}

package tencent

// Running this file's test do not require any updated variables
// It can be run as is

import (
	"fmt"
	"testing"
)

func TestArtifact_BuilderId(t *testing.T) {
	a := &Artifact{}
	expectedValue := "MyBuilderId"
	a.BuilderIDValue = expectedValue
	result := a.BuilderId()
	if result != expectedValue {
		t.Fatalf("BuilderId has unexpected value: %+v", result)
	}
}

func TestArtifact_Files(t *testing.T) {
	a := &Artifact{}
	a.SSHKeyLocation = "Hello"
	result := a.Files()
	if len(result) != 1 && result[0] != "Hello" {
		t.Errorf("Files has unexpected value: %v", result)
	}
}

func TestArtifact_Id(t *testing.T) {
	a := &Artifact{}
	expectedValue := "MyInstanceId"
	a.InstanceId = expectedValue
	result := a.Id()
	if result != expectedValue {
		t.Fatalf("Id has unexpected value: %+v", result)
	}
}

func TestArtifact_State(t *testing.T) {
	expectedBuilderId := "MyBuilderValue"
	expectedIPAddress := "1.2.3.4"
	expectedInstanceId := "MyInstaceId"
	a := &Artifact{}
	a.BuilderIDValue = expectedBuilderId
	a.IPAddress = expectedIPAddress
	a.InstanceId = expectedInstanceId

	result := a.State(CArtifactBuilderID)
	if result != expectedBuilderId {
		t.Fatalf("Unexpected Artifact state: %s for name: %s", result, CArtifactBuilderID)
	}

	result = a.State(CArtifactIPAddress)
	if result != expectedIPAddress {
		t.Fatalf("Unexpected Artifact state: %s for name: %s", result, CArtifactIPAddress)
	}

	result = a.State(CInstanceId)
	if result != expectedInstanceId {
		t.Fatalf("Unexpected Artifact state: %s for name: %s", result, CInstanceId)
	}
}

func TestArtifact_Destroy(t *testing.T) {
	a := &Artifact{}
	result := a.Destroy()
	if result != nil {
		t.Fatalf("Destroy() has unexpected value: %+v", result)
	}
}

func TestArtifact_String(t *testing.T) {
	a := &Artifact{}
	expectedValue := "MyInstanceId"
	expectedIPAddress := "1.1.1.1"
	a.InstanceId = expectedValue
	expectedResult1 := fmt.Sprintf("Instance was created: %s", a.InstanceId)
	expectedResult2 := fmt.Sprintf("%s and IP address is: %s", expectedResult1, expectedIPAddress)

	result := a.String()
	if result != expectedResult1 {
		t.Fatalf("String() has unexpected value: %+v", result)
	}

	a.IPAddress = expectedIPAddress
	result2 := a.String()
	if result2 != expectedResult2 {
		t.Fatalf("String() has unexpected value: %+v", result)
	}
}

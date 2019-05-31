package client

import (
	"errors"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"
)

func TestConfigRetrieverFillsTenantIDWhenEmpty(t *testing.T) {
	c := Config{CloudEnvironmentName: "AzurePublicCloud"}
	if expected := ""; c.TenantID != expected {
		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
	}

	retrievedTid := "my-tenant-id"
	findTenantID = func(azure.Environment, string) (string, error) { return retrievedTid, nil }
	getSubscriptionFromIMDS = func() (string, error) { return "unittest", nil }
	if err := c.FillParameters(); err != nil {
		t.Errorf("Unexpected error when calling c.FillParameters: %v", err)
	}

	if expected := retrievedTid; c.TenantID != expected {
		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
	}
}

func TestConfigRetrieverLeavesTenantIDWhenNotEmpty(t *testing.T) {
	c := Config{CloudEnvironmentName: "AzurePublicCloud"}
	userSpecifiedTid := "not-empty"
	c.TenantID = userSpecifiedTid

	findTenantID = nil // assert that this not even called
	getSubscriptionFromIMDS = func() (string, error) { return "unittest", nil }
	if err := c.FillParameters(); err != nil {
		t.Errorf("Unexpected error when calling c.FillParameters: %v", err)
	}

	if expected := userSpecifiedTid; c.TenantID != expected {
		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
	}
}

func TestConfigRetrieverReturnsErrorWhenTenantIDEmptyAndRetrievalFails(t *testing.T) {
	c := Config{CloudEnvironmentName: "AzurePublicCloud"}
	if expected := ""; c.TenantID != expected {
		t.Errorf("Expected TenantID to be %q but got %q", expected, c.TenantID)
	}

	errorString := "sorry, I failed"
	findTenantID = func(azure.Environment, string) (string, error) { return "", errors.New(errorString) }
	getSubscriptionFromIMDS = func() (string, error) { return "unittest", nil }
	if err := c.FillParameters(); err != nil && err.Error() != errorString {
		t.Errorf("Unexpected error when calling c.FillParameters: %v", err)
	}
}

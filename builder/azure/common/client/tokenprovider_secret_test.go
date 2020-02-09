package client

import (
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"
)

// Behavior is the most important thing to assert for ServicePrincipalToken, but
// that cannot be done in a unit test because it involves network access.  Instead,
// I assert the expected inertness of this class.
func TestNewSecretOAuthTokenProvider(t *testing.T) {
	testSubject := NewSecretOAuthTokenProvider(azure.PublicCloud, "clientID", "clientString", "tenantID")
	spn, err := testSubject.getServicePrincipalToken()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if spn.Token().AccessToken != "" {
		t.Errorf("spn.Token().AccessToken: expected=\"\", actual=%s", spn.Token().AccessToken)
	}
	if spn.Token().RefreshToken != "" {
		t.Errorf("spn.Token().RefreshToken: expected=\"\", actual=%s", spn.Token().RefreshToken)
	}
	if spn.Token().ExpiresIn != "0" {
		t.Errorf("spn.Token().ExpiresIn: expected=\"0\", actual=%s", spn.Token().ExpiresIn)
	}
	if spn.Token().ExpiresOn != "0" {
		t.Errorf("spn.Token().ExpiresOn: expected=\"0\", actual=%s", spn.Token().ExpiresOn)
	}
	if spn.Token().NotBefore != "0" {
		t.Errorf("spn.Token().NotBefore: expected=\"0\", actual=%s", spn.Token().NotBefore)
	}
	if spn.Token().Resource != "" {
		t.Errorf("spn.Token().Resource: expected=\"\", actual=%s", spn.Token().Resource)
	}
	if spn.Token().Type != "" {
		t.Errorf("spn.Token().Type: expected=\"\", actual=%s", spn.Token().Type)
	}
}

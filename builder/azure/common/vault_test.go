package common

import (
	"net/url"
	"testing"
)

func TestVaultClientKeyVaultEndpoint(t *testing.T) {
	u, _ := url.Parse("https://vault.azure.net")
	testSubject := NewVaultClient(*u)

	vaultUrl := testSubject.getVaultUrl("my")
	if vaultUrl != "https://my.vault.azure.net/" {
		t.Errorf("expected \"https://my.vault.azure.net/\", got %q", vaultUrl)
	}
}

func TestVaultClientKeyVaultEndpointPreserveScheme(t *testing.T) {
	u, _ := url.Parse("http://vault.azure.net")
	testSubject := NewVaultClient(*u)

	vaultUrl := testSubject.getVaultUrl("my")
	if vaultUrl != "http://my.vault.azure.net/" {
		t.Errorf("expected \"http://my.vault.azure.net/\", got %q", vaultUrl)
	}
}

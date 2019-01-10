package arm

import (
	"fmt"
	"os"
	"testing"

	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/packer/packer"
)

func Test_ClientConfig_RequiredParametersSet(t *testing.T) {

	tests := []struct {
		name    string
		config  ClientConfig
		wantErr bool
	}{
		{
			name:    "no client_id, client_secret or subscription_id should enable MSI auth",
			config:  ClientConfig{},
			wantErr: false,
		},
		{
			name: "subscription_id is set will trigger device flow",
			config: ClientConfig{
				SubscriptionID: "error",
			},
			wantErr: false,
		},
		{
			name: "client_id without client_secret should error",
			config: ClientConfig{
				ClientID: "error",
			},
			wantErr: true,
		},
		{
			name: "client_secret without client_id should error",
			config: ClientConfig{
				ClientSecret: "error",
			},
			wantErr: true,
		},
		{
			name: "missing subscription_id when using secret",
			config: ClientConfig{
				ClientID:     "ok",
				ClientSecret: "ok",
			},
			wantErr: true,
		},
		{
			name: "tenant_id alone should fail",
			config: ClientConfig{
				TenantID: "ok",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			errs := &packer.MultiError{}
			tt.config.assertRequiredParametersSet(errs)
			if (len(errs.Errors) != 0) != tt.wantErr {
				t.Errorf("newConfig() error = %v, wantErr %v", errs, tt.wantErr)
				return
			}
		})
	}
}

func Test_ClientConfig_DeviceLogin(t *testing.T) {
	getEnvOrSkip(t, "AZURE_DEVICE_LOGIN")
	cfg := ClientConfig{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.getServicePrincipalTokens(
		func(s string) { fmt.Printf("SAY: %s\n", s) })
	if err != nil {
		t.Fatalf("Expected nil err, but got: %v", err)
	}
	token := spt.Token()
	if token.AccessToken == "" {
		t.Fatal("Expected management token to have non-nil access token")
	}
	if token.RefreshToken == "" {
		t.Fatal("Expected management token to have non-nil refresh token")
	}
	kvtoken := sptkv.Token()
	if kvtoken.AccessToken == "" {
		t.Fatal("Expected keyvault token to have non-nil access token")
	}
	if kvtoken.RefreshToken == "" {
		t.Fatal("Expected keyvault token to have non-nil refresh token")
	}
}

func Test_ClientConfig_ClientPassword(t *testing.T) {
	cfg := ClientConfig{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		ClientID:         getEnvOrSkip(t, "AZURE_CLIENTID"),
		ClientSecret:     getEnvOrSkip(t, "AZURE_CLIENTSECRET"),
		TenantID:         getEnvOrSkip(t, "AZURE_TENANTID"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.getServicePrincipalTokens(func(s string) { fmt.Printf("SAY: %s\n", s) })
	if err != nil {
		t.Fatalf("Expected nil err, but got: %v", err)
	}
	token := spt.Token()
	if token.AccessToken == "" {
		t.Fatal("Expected management token to have non-nil access token")
	}
	if token.RefreshToken != "" {
		t.Fatal("Expected management token to have no refresh token")
	}
	kvtoken := sptkv.Token()
	if kvtoken.AccessToken == "" {
		t.Fatal("Expected keyvault token to have non-nil access token")
	}
	if kvtoken.RefreshToken != "" {
		t.Fatal("Expected keyvault token to have no refresh token")
	}
}

func getEnvOrSkip(t *testing.T, envVar string) string {
	v := os.Getenv(envVar)
	if v == "" {
		t.Skipf("%s is empty, skipping", envVar)
	}
	return v
}

func getCloud() *azure.Environment {
	cloudName := os.Getenv("AZURE_CLOUD")
	if cloudName == "" {
		cloudName = "AZUREPUBLICCLOUD"
	}
	c, _ := azure.EnvironmentFromName(cloudName)
	return &c
}

// tests for assertRequiredParametersSet

func Test_ClientConfig_CanUseDeviceCode(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	// TenantID is optional

	assertValid(t, cfg)
}

func assertValid(t *testing.T, cfg ClientConfig) {
	errs := &packer.MultiError{}
	cfg.assertRequiredParametersSet(errs)
	if len(errs.Errors) != 0 {
		t.Fatal("Expected errs to be empty: ", errs)
	}
}

func assertInvalid(t *testing.T, cfg ClientConfig) {
	errs := &packer.MultiError{}
	cfg.assertRequiredParametersSet(errs)
	if len(errs.Errors) == 0 {
		t.Fatal("Expected errs to be non-empty")
	}
}

func Test_ClientConfig_CanUseClientSecret(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientSecret = "12345"

	assertValid(t, cfg)
}

func Test_ClientConfig_CanUseClientSecretWithTenantID(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientSecret = "12345"
	cfg.TenantID = "12345"

	assertValid(t, cfg)
}

func emptyClientConfig() ClientConfig {
	cfg := ClientConfig{}
	_ = cfg.setCloudEnvironment()
	return cfg
}

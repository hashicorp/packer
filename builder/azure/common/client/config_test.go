package client

import (
	crand "crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	mrand "math/rand"
	"os"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
	jwt "github.com/dgrijalva/jwt-go"
	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
)

func Test_ClientConfig_RequiredParametersSet(t *testing.T) {

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "no client_id, client_secret or subscription_id should enable MSI auth",
			config:  Config{},
			wantErr: false,
		},
		{
			name: "use_azure_cli_auth will trigger Azure CLI auth",
			config: Config{
				UseAzureCLIAuth: true,
			},
			wantErr: false,
		},
		{
			name: "subscription_id is set will trigger device flow",
			config: Config{
				SubscriptionID: "error",
			},
			wantErr: false,
		},
		{
			name: "client_id without client_secret, client_cert_path or client_jwt should error",
			config: Config{
				ClientID: "error",
			},
			wantErr: true,
		},
		{
			name: "client_secret without client_id should error",
			config: Config{
				ClientSecret: "error",
			},
			wantErr: true,
		},
		{
			name: "client_cert_path without client_id should error",
			config: Config{
				ClientCertPath: "/dev/null",
			},
			wantErr: true,
		},
		{
			name: "client_jwt without client_id should error",
			config: Config{
				ClientJWT: "error",
			},
			wantErr: true,
		},
		{
			name: "missing subscription_id when using secret",
			config: Config{
				ClientID:     "ok",
				ClientSecret: "ok",
			},
			wantErr: true,
		},
		{
			name: "missing subscription_id when using certificate",
			config: Config{
				ClientID:       "ok",
				ClientCertPath: "ok",
			},
			wantErr: true,
		},
		{
			name: "missing subscription_id when using JWT",
			config: Config{
				ClientID:  "ok",
				ClientJWT: "ok",
			},
			wantErr: true,
		},
		{
			name: "too many client_* values",
			config: Config{
				SubscriptionID: "ok",
				ClientID:       "ok",
				ClientSecret:   "ok",
				ClientCertPath: "error",
			},
			wantErr: true,
		},
		{
			name: "too many client_* values (2)",
			config: Config{
				SubscriptionID: "ok",
				ClientID:       "ok",
				ClientSecret:   "ok",
				ClientJWT:      "error",
			},
			wantErr: true,
		},
		{
			name: "tenant_id alone should fail",
			config: Config{
				TenantID: "ok",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			errs := &packersdk.MultiError{}
			tt.config.Validate(errs)
			if (len(errs.Errors) != 0) != tt.wantErr {
				t.Errorf("newConfig() error = %v, wantErr %v", errs, tt.wantErr)
				return
			}
		})
	}
}

func Test_ClientConfig_DeviceLogin(t *testing.T) {
	getEnvOrSkip(t, "AZURE_DEVICE_LOGIN")
	cfg := Config{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.GetServicePrincipalTokens(
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

func Test_ClientConfig_AzureCli(t *testing.T) {
	// Azure CLI tests skipped unless env 'AZURE_CLI_AUTH' is set, and an active `az login` session has been established
	getEnvOrSkip(t, "AZURE_CLI_AUTH")

	cfg := Config{
		UseAzureCLIAuth:  true,
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	err := cfg.FillParameters()
	if err != nil {
		t.Fatalf("Expected nil err, but got: %v", err)
	}

	if cfg.authType != authTypeAzureCLI {
		t.Fatalf("Expected authType to be %q, but got: %q", authTypeAzureCLI, cfg.authType)
	}
}

func Test_ClientConfig_ClientPassword(t *testing.T) {
	cfg := Config{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		ClientID:         getEnvOrSkip(t, "AZURE_CLIENTID"),
		ClientSecret:     getEnvOrSkip(t, "AZURE_CLIENTSECRET"),
		TenantID:         getEnvOrSkip(t, "AZURE_TENANTID"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.GetServicePrincipalTokens(func(s string) { fmt.Printf("SAY: %s\n", s) })
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

func Test_ClientConfig_ClientCert(t *testing.T) {
	cfg := Config{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		ClientID:         getEnvOrSkip(t, "AZURE_CLIENTID"),
		ClientCertPath:   getEnvOrSkip(t, "AZURE_CLIENTCERT"),
		TenantID:         getEnvOrSkip(t, "AZURE_TENANTID"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.GetServicePrincipalTokens(func(s string) { fmt.Printf("SAY: %s\n", s) })
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

func Test_ClientConfig_ClientJWT(t *testing.T) {
	cfg := Config{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		ClientID:         getEnvOrSkip(t, "AZURE_CLIENTID"),
		ClientJWT:        getEnvOrSkip(t, "AZURE_CLIENTJWT"),
		TenantID:         getEnvOrSkip(t, "AZURE_TENANTID"),
		cloudEnvironment: getCloud(),
	}
	assertValid(t, cfg)

	spt, sptkv, err := cfg.GetServicePrincipalTokens(func(s string) { fmt.Printf("SAY: %s\n", s) })
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
	// TenantID is optional, but Builder will look up tenant ID before requesting
	t.Run("without TenantID", func(t *testing.T) {
		cfg := Config{
			SubscriptionID: "12345",
		}
		assertValid(t, cfg)
	})
	t.Run("with TenantID", func(t *testing.T) {
		cfg := Config{
			SubscriptionID: "12345",
			TenantID:       "12345",
		}
		assertValid(t, cfg)
	})
}

func assertValid(t *testing.T, cfg Config) {
	errs := &packersdk.MultiError{}
	cfg.Validate(errs)
	if len(errs.Errors) != 0 {
		t.Fatal("Expected errs to be empty: ", errs)
	}
}

func assertInvalid(t *testing.T, cfg Config) {
	errs := &packersdk.MultiError{}
	cfg.Validate(errs)
	if len(errs.Errors) == 0 {
		t.Fatal("Expected errs to be non-empty")
	}
}

func Test_ClientConfig_CanUseClientSecret(t *testing.T) {
	cfg := Config{
		SubscriptionID: "12345",
		ClientID:       "12345",
		ClientSecret:   "12345",
	}

	assertValid(t, cfg)
}

func Test_ClientConfig_CanUseClientSecretWithTenantID(t *testing.T) {
	cfg := Config{
		SubscriptionID: "12345",
		ClientID:       "12345",
		ClientSecret:   "12345",
		TenantID:       "12345",
	}

	assertValid(t, cfg)
}

func Test_ClientConfig_CanUseClientJWT(t *testing.T) {
	cfg := Config{
		SubscriptionID: "12345",
		ClientID:       "12345",
		ClientJWT:      getJWT(10*time.Minute, true),
	}

	assertValid(t, cfg)
}

func Test_ClientConfig_CanUseClientJWTWithTenantID(t *testing.T) {
	cfg := Config{
		SubscriptionID: "12345",
		ClientID:       "12345",
		ClientJWT:      getJWT(10*time.Minute, true),
		TenantID:       "12345",
	}

	assertValid(t, cfg)
}

func Test_ClientConfig_CannotUseBothClientJWTAndSecret(t *testing.T) {
	cfg := Config{
		SubscriptionID: "12345",
		ClientID:       "12345",
		ClientSecret:   "12345",
		ClientJWT:      getJWT(10*time.Minute, true),
	}

	assertInvalid(t, cfg)
}

func Test_ClientConfig_ClientJWTShouldBeValidForAtLeast5Minutes(t *testing.T) {
	cfg := Config{
		SubscriptionID: "12345",
		ClientID:       "12345",
		ClientJWT:      getJWT(time.Minute, true),
	}

	assertInvalid(t, cfg)
}

func Test_ClientConfig_ClientJWTShouldHaveThumbprint(t *testing.T) {
	cfg := Config{
		SubscriptionID: "12345",
		ClientID:       "12345",
		ClientJWT:      getJWT(10*time.Minute, false),
	}

	assertInvalid(t, cfg)
}

func Test_getJWT(t *testing.T) {
	if getJWT(time.Minute, true) == "" {
		t.Fatalf("getJWT is broken")
	}
}

func newRandReader() io.Reader {
	var seed int64
	binary.Read(crand.Reader, binary.LittleEndian, &seed)

	return mrand.New(mrand.NewSource(seed))
}

func getJWT(validFor time.Duration, withX5tHeader bool) string {
	token := jwt.New(jwt.SigningMethodRS256)
	key, _ := rsa.GenerateKey(newRandReader(), 2048)

	token.Claims = jwt.MapClaims{
		"aud": "https://login.microsoftonline.com/tenant.onmicrosoft.com/oauth2/token?api-version=1.0",
		"iss": "355dff10-cd78-11e8-89fe-000d3afd16e3",
		"sub": "355dff10-cd78-11e8-89fe-000d3afd16e3",
		"jti": base64.URLEncoding.EncodeToString([]byte{0}),
		"nbf": time.Now().Unix(),
		"exp": time.Now().Add(validFor).Unix(),
	}
	if withX5tHeader {
		token.Header["x5t"] = base64.URLEncoding.EncodeToString([]byte("thumbprint"))
	}

	jwt, _ := token.SignedString(key)
	return jwt
}

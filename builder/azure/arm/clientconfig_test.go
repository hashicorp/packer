package arm

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/hashicorp/packer/packer"
)

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

func Test_ClientConfig_ClientJWT(t *testing.T) {
	cfg := ClientConfig{
		SubscriptionID:   getEnvOrSkip(t, "AZURE_SUBSCRIPTION"),
		ClientID:         getEnvOrSkip(t, "AZURE_CLIENTID"),
		ClientJWT:        getEnvOrSkip(t, "AZURE_CLIENTJWT"),
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

func Test_ClientConfig_CanUseClientJWT(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientJWT = getJWT(10*time.Minute, true)

	assertValid(t, cfg)
}

func Test_ClientConfig_CanUseClientJWTWithTenantID(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientJWT = getJWT(10*time.Minute, true)
	cfg.TenantID = "12345"

	assertValid(t, cfg)
}

func Test_ClientConfig_CannotUseBothClientJWTAndSecret(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientSecret = "12345"
	cfg.ClientJWT = getJWT(10*time.Minute, true)

	assertInvalid(t, cfg)
}

func Test_ClientConfig_ClientJWTShouldBeValidForAtLeast5Minutes(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientJWT = getJWT(time.Minute, true)

	assertInvalid(t, cfg)
}

func Test_ClientConfig_ClientJWTShouldHaveThumbprint(t *testing.T) {
	cfg := emptyClientConfig()
	cfg.SubscriptionID = "12345"
	cfg.ClientID = "12345"
	cfg.ClientJWT = getJWT(10*time.Minute, false)

	assertInvalid(t, cfg)
}

func emptyClientConfig() ClientConfig {
	cfg := ClientConfig{}
	_ = cfg.setCloudEnvironment()
	return cfg
}

func Test_getJWT(t *testing.T) {
	if getJWT(time.Minute, true) == "" {
		t.Fatalf("getJWT is broken")
	}
}

func getJWT(validFor time.Duration, withX5tHeader bool) string {
	token := jwt.New(jwt.SigningMethodRS256)
	key, _ := rsa.GenerateKey(rand.Reader, 2048)

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

	signedString, _ := token.SignedString(key)
	return signedString
}

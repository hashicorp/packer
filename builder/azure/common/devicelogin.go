package common

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2016-06-01/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/hashicorp/packer/helper/useragent"
	"github.com/mitchellh/go-homedir"
)

var (
	// AD app id for packer-azure driver.
	clientIDs = map[string]string{
		azure.PublicCloud.Name:       "04cc58ec-51ab-4833-ac0d-ce3a7912414b",
		azure.USGovernmentCloud.Name: "a1479822-da77-46a7-abd0-6edacc8a8fac",
	}
)

// NOTE(ahmetalpbalkan): Azure Active Directory implements OAuth 2.0 Device Flow
// described here: https://tools.ietf.org/html/draft-denniss-oauth-device-flow-00
// Although it has some gotchas, most of the authentication logic is in Azure SDK
// for Go helper packages.
//
// Device auth prints a message to the screen telling the user to click on URL
// and approve the app on the browser, meanwhile the client polls the auth API
// for a token. Once we have token, we save it locally to a file with proper
// permissions and when the token expires (in Azure case typically 1 hour) SDK
// will automatically refresh the specified token and will call the refresh
// callback function we implement here. This way we will always be storing a
// token with a refresh_token saved on the machine.

// Authenticate fetches a token from the local file cache or initiates a consent
// flow and waits for token to be obtained.
func Authenticate(env azure.Environment, tenantID string, say func(string), scope string) (*adal.ServicePrincipalToken, error) {
	clientID, ok := clientIDs[env.Name]
	var resourceid string

	if !ok {
		return nil, fmt.Errorf("packer-azure application not set up for Azure environment %q", env.Name)
	}

	oauthCfg, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain oauth config for azure environment: %v", err)
	}

	// for AzurePublicCloud (https://management.core.windows.net/), this old
	// Service Management scope covers both ASM and ARM.

	if strings.Contains(scope, "vault") {
		resourceid = "vault"
	} else {
		resourceid = "mgmt"
	}

	tokenPath := tokenCachePath(tenantID + resourceid)
	saveToken := mkTokenCallback(tokenPath)
	saveTokenCallback := func(t adal.Token) error {
		say("Azure token expired. Saving the refreshed token...")
		return saveToken(t)
	}

	// Lookup the token cache file for an existing token.
	spt, err := tokenFromFile(say, *oauthCfg, tokenPath, clientID, scope, saveTokenCallback)
	if err != nil {
		return nil, err
	}
	if spt != nil {
		say(fmt.Sprintf("Auth token found in file: %s", tokenPath))
		return spt, nil
	}

	// Start an OAuth 2.0 device flow
	say(fmt.Sprintf("Initiating device flow: %s", tokenPath))
	spt, err = tokenFromDeviceFlow(say, *oauthCfg, clientID, scope)
	if err != nil {
		return nil, err
	}
	say("Obtained service principal token.")
	if err := saveToken(spt.Token()); err != nil {
		say("Error occurred saving token to cache file.")
		return nil, err
	}
	return spt, nil
}

// tokenFromFile returns a token from the specified file if it is found, otherwise
// returns nil. Any error retrieving or creating the token is returned as an error.
func tokenFromFile(say func(string), oauthCfg adal.OAuthConfig, tokenPath, clientID, resource string,
	callback adal.TokenRefreshCallback) (*adal.ServicePrincipalToken, error) {
	say(fmt.Sprintf("Loading auth token from file: %s", tokenPath))
	if _, err := os.Stat(tokenPath); err != nil {
		if os.IsNotExist(err) { // file not found
			return nil, nil
		}
		return nil, err
	}

	token, err := adal.LoadToken(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load token from file: %v", err)
	}

	spt, err := adal.NewServicePrincipalTokenFromManualToken(oauthCfg, clientID, resource, *token, callback)
	if err != nil {
		return nil, fmt.Errorf("Error constructing service principal token: %v", err)
	}
	return spt, nil
}

// tokenFromDeviceFlow prints a message to the screen for user to take action to
// consent application on a browser and in the meanwhile the authentication
// endpoint is polled until user gives consent, denies or the flow times out.
// Returned token must be saved.
func tokenFromDeviceFlow(say func(string), oauthCfg adal.OAuthConfig, clientID, resource string) (*adal.ServicePrincipalToken, error) {
	cl := autorest.NewClientWithUserAgent(useragent.String())
	deviceCode, err := adal.InitiateDeviceAuth(&cl, oauthCfg, clientID, resource)
	if err != nil {
		return nil, fmt.Errorf("Failed to start device auth: %v", err)
	}

	// Example message: “To sign in, open https://aka.ms/devicelogin and enter
	// the code 0000000 to authenticate.”
	say(fmt.Sprintf("Microsoft Azure: %s", to.String(deviceCode.Message)))

	token, err := adal.WaitForUserCompletion(&cl, deviceCode)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete device auth: %v", err)
	}

	spt, err := adal.NewServicePrincipalTokenFromManualToken(oauthCfg, clientID, resource, *token)
	if err != nil {
		return nil, fmt.Errorf("Error constructing service principal token: %v", err)
	}
	return spt, nil
}

// tokenCachePath returns the full path the OAuth 2.0 token should be saved at
// for given tenant ID.
func tokenCachePath(tenantID string) string {
	dir, err := homedir.Dir()
	if err != nil {
		dir, _ = filepath.Abs(os.Args[0])
	}

	return filepath.Join(dir, ".azure", "packer", fmt.Sprintf("oauth-%s.json", tenantID))
}

// mkTokenCallback returns a callback function that can be used to save the
// token initially or register to the Azure SDK to be called when the token is
// refreshed.
func mkTokenCallback(path string) adal.TokenRefreshCallback {
	return func(t adal.Token) error {
		if err := adal.SaveToken(path, 0600, t); err != nil {
			return err
		}
		return nil
	}
}

// FindTenantID figures out the AAD tenant ID of the subscription by making an
// unauthenticated request to the Get Subscription Details endpoint and parses
// the value from WWW-Authenticate header.
func FindTenantID(env azure.Environment, subscriptionID string) (string, error) {
	const hdrKey = "WWW-Authenticate"
	c := subscriptions.NewClientWithBaseURI(env.ResourceManagerEndpoint)

	// we expect this request to fail (err != nil), but we are only interested
	// in headers, so surface the error if the Response is not present (i.e.
	// network error etc)
	subs, err := c.Get(context.TODO(), subscriptionID)
	if subs.Response.Response == nil {
		return "", fmt.Errorf("Request failed: %v", err)
	}

	// Expecting 401 StatusUnauthorized here, just read the header
	if subs.StatusCode != http.StatusUnauthorized {
		return "", fmt.Errorf("Unexpected response from Get Subscription: %v", err)
	}
	hdr := subs.Header.Get(hdrKey)
	if hdr == "" {
		return "", fmt.Errorf("Header %v not found in Get Subscription response", hdrKey)
	}

	// Example value for hdr:
	//   Bearer authorization_uri="https://login.windows.net/996fe9d1-6171-40aa-945b-4c64b63bf655", error="invalid_token", error_description="The authentication failed because of missing 'Authorization' header."
	r := regexp.MustCompile(`authorization_uri=".*/([0-9a-f\-]+)"`)
	m := r.FindStringSubmatch(hdr)
	if m == nil {
		return "", fmt.Errorf("Could not find the tenant ID in header: %s %q", hdrKey, hdr)
	}
	return m[1], nil
}

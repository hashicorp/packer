package common

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/Azure/azure-sdk-for-go/arm/resources/subscriptions"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/mitchellh/go-homedir"
	"github.com/mitchellh/packer/version"
)

var (
	// AD app id for packer-azure driver.
	clientIDs = map[string]string{
		azure.PublicCloud.Name: "04cc58ec-51ab-4833-ac0d-ce3a7912414b",
	}

	userAgent = fmt.Sprintf("packer/%s", version.FormattedVersion())
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
func Authenticate(env azure.Environment, tenantID string, say func(string)) (*azure.ServicePrincipalToken, error) {
	clientID, ok := clientIDs[env.Name]
	if !ok {
		return nil, fmt.Errorf("packer-azure application not set up for Azure environment %q", env.Name)
	}

	oauthCfg, err := env.OAuthConfigForTenant(tenantID)
	if err != nil {
		return nil, fmt.Errorf("Failed to obtain oauth config for azure environment: %v", err)
	}

	// for AzurePublicCloud (https://management.core.windows.net/), this old
	// Service Management scope covers both ASM and ARM.
	apiScope := env.ServiceManagementEndpoint

	tokenPath := tokenCachePath(tenantID)
	saveToken := mkTokenCallback(tokenPath)
	saveTokenCallback := func(t azure.Token) error {
		say("Azure token expired. Saving the refreshed token...")
		return saveToken(t)
	}

	// Lookup the token cache file for an existing token.
	spt, err := tokenFromFile(say, *oauthCfg, tokenPath, clientID, apiScope, saveTokenCallback)
	if err != nil {
		return nil, err
	}
	if spt != nil {
		say(fmt.Sprintf("Auth token found in file: %s", tokenPath))

		// NOTE(ahmetalpbalkan): The token file we found may contain an
		// expired access_token. In that case, the first call to Azure SDK will
		// attempt to refresh the token using refresh_token, which might have
		// expired[1], in that case we will get an error and we shall remove the
		// token file and initiate token flow again so that the user would not
		// need removing the token cache file manually.
		//
		// [1]: expiration date of refresh_token is not returned in AAD /token
		//      response, we just know it is 14 days. Therefore user’s token
		//      will go stale every 14 days and we will delete the token file,
		//      re-initiate the device flow.
		say("Validating the token.")
		if err = validateToken(env, spt); err != nil {
			say(fmt.Sprintf("Error: %v", err))
			say("Stored Azure credentials expired. Please reauthenticate.")
			say(fmt.Sprintf("Deleting %s", tokenPath))
			if err := os.RemoveAll(tokenPath); err != nil {
				return nil, fmt.Errorf("Error deleting stale token file: %v", err)
			}
		} else {
			say("Token works.")
			return spt, nil
		}
	}

	// Start an OAuth 2.0 device flow
	say(fmt.Sprintf("Initiating device flow: %s", tokenPath))
	spt, err = tokenFromDeviceFlow(say, *oauthCfg, clientID, apiScope)
	if err != nil {
		return nil, err
	}
	say("Obtained service principal token.")
	if err := saveToken(spt.Token); err != nil {
		say("Error occurred saving token to cache file.")
		return nil, err
	}
	return spt, nil
}

// tokenFromFile returns a token from the specified file if it is found, otherwise
// returns nil. Any error retrieving or creating the token is returned as an error.
func tokenFromFile(say func(string), oauthCfg azure.OAuthConfig, tokenPath, clientID, resource string,
	callback azure.TokenRefreshCallback) (*azure.ServicePrincipalToken, error) {
	say(fmt.Sprintf("Loading auth token from file: %s", tokenPath))
	if _, err := os.Stat(tokenPath); err != nil {
		if os.IsNotExist(err) { // file not found
			return nil, nil
		}
		return nil, err
	}

	token, err := azure.LoadToken(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load token from file: %v", err)
	}

	spt, err := azure.NewServicePrincipalTokenFromManualToken(oauthCfg, clientID, resource, *token, callback)
	if err != nil {
		return nil, fmt.Errorf("Error constructing service principal token: %v", err)
	}
	return spt, nil
}

// tokenFromDeviceFlow prints a message to the screen for user to take action to
// consent application on a browser and in the meanwhile the authentication
// endpoint is polled until user gives consent, denies or the flow times out.
// Returned token must be saved.
func tokenFromDeviceFlow(say func(string), oauthCfg azure.OAuthConfig, clientID, resource string) (*azure.ServicePrincipalToken, error) {
	cl := autorest.NewClientWithUserAgent(userAgent)
	deviceCode, err := azure.InitiateDeviceAuth(&cl, oauthCfg, clientID, resource)
	if err != nil {
		return nil, fmt.Errorf("Failed to start device auth: %v", err)
	}

	// Example message: “To sign in, open https://aka.ms/devicelogin and enter
	// the code 0000000 to authenticate.”
	say(fmt.Sprintf("Microsoft Azure: %s", to.String(deviceCode.Message)))

	token, err := azure.WaitForUserCompletion(&cl, deviceCode)
	if err != nil {
		return nil, fmt.Errorf("Failed to complete device auth: %v", err)
	}

	spt, err := azure.NewServicePrincipalTokenFromManualToken(oauthCfg, clientID, resource, *token)
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
func mkTokenCallback(path string) azure.TokenRefreshCallback {
	return func(t azure.Token) error {
		if err := azure.SaveToken(path, 0600, t); err != nil {
			return err
		}
		return nil
	}
}

// validateToken makes a call to Azure SDK with given token, essentially making
// sure if the access_token valid, if not it uses SDK’s functionality to
// automatically refresh the token using refresh_token (which might have
// expired). This check is essentially to make sure refresh_token is good.
func validateToken(env azure.Environment, token *azure.ServicePrincipalToken) error {
	c := subscriptionsClient(env.ResourceManagerEndpoint)
	c.Authorizer = token
	_, err := c.List()
	if err != nil {
		return fmt.Errorf("Token validity check failed: %v", err)
	}
	return nil
}

// FindTenantID figures out the AAD tenant ID of the subscription by making an
// unauthenticated request to the Get Subscription Details endpoint and parses
// the value from WWW-Authenticate header.
func FindTenantID(env azure.Environment, subscriptionID string) (string, error) {
	const hdrKey = "WWW-Authenticate"
	c := subscriptionsClient(env.ResourceManagerEndpoint)

	// we expect this request to fail (err != nil), but we are only interested
	// in headers, so surface the error if the Response is not present (i.e.
	// network error etc)
	subs, err := c.Get(subscriptionID)
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

func subscriptionsClient(baseURI string) subscriptions.Client {
	client := subscriptions.NewClientWithBaseURI(baseURI)
	return client
}

package client

import (
	"errors"
	"fmt"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/cli"
)

// for managed identity auth
type cliOAuthTokenProvider struct {
	env      azure.Environment
	say      func(string)
	tenantID string
}

func NewCliOAuthTokenProvider(env azure.Environment, say func(string), tenantID string) oAuthTokenProvider {
	return &cliOAuthTokenProvider{
		env:      env,
		say:      say,
		tenantID: tenantID,
	}
}

func (tp *cliOAuthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return tp.getServicePrincipalTokenWithResource(tp.env.ResourceManagerEndpoint)
}

func (tp *cliOAuthTokenProvider) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	token, err := cli.GetTokenFromCLI(resource)
	if err != nil {
		tp.say(fmt.Sprintf("unable to get token from azure cli: %v", err))
		return nil, err
	}

	oAuthConfig, err := adal.NewOAuthConfig(resource, tp.tenantID)
	if err != nil {
		tp.say(fmt.Sprintf("unable to generate OAuth Config: %v", err))
		return nil, err
	}

	adalToken, err := token.ToADALToken()
	if err != nil {
		tp.say(fmt.Sprintf("unable to get ADAL Token from azure cli token: %v", err))
		return nil, err
	}

	clientID := clientIDs[tp.env.Name]
	spToken, err := adal.NewServicePrincipalTokenFromManualToken(*oAuthConfig, clientID, resource, adalToken)
	if err != nil {
		tp.say(fmt.Sprintf("unable to get service principal token from adal token: %v", err))
		return nil, err
	}

	return spToken, nil
}

// Get TenantID and SubscriptionID from Azure CLI
func getIDsFromAzureCLI() (tenantID string, subscriptionID string, err error) {
	profilePath, err := cli.ProfilePath()
	if err != nil {
		return "", "", err
	}

	profile, err := cli.LoadProfile(profilePath)
	if err != nil {
		return "", "", err
	}

	for _, p := range profile.Subscriptions {
		if p.IsDefault {
			return p.TenantID, p.ID, nil
		}
	}

	return "", "", errors.New("Unable to find default subscription")
}

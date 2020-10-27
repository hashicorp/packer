package client

import (
	"context"
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

	spt, err := adal.NewServicePrincipalTokenFromManualToken(*oAuthConfig, clientIDs[tp.env.Name], resource, adalToken)
	if err != nil {
		tp.say(fmt.Sprintf("unable to get service principal token from adal token: %v", err))
		return nil, err
	}

	// Custom refresh function to make it possible to use Azure CLI to refresh tokens.
	// Inspired by HashiCorps go-azure-helpers: https://github.com/hashicorp/go-azure-helpers/blob/373622ce2effb0cf299051ea019cb657f357a4d8/authentication/auth_method_azure_cli_token.go#L96-L109
	var customRefreshFunc adal.TokenRefresh = func(ctx context.Context, resource string) (*adal.Token, error) {
		token, err := cli.GetTokenFromCLI(resource)
		if err != nil {
			tp.say(fmt.Sprintf("token refresh - unable to get token from azure cli: %v", err))
			return nil, err
		}

		adalToken, err := token.ToADALToken()
		if err != nil {
			tp.say(fmt.Sprintf("token refresh - unable to get ADAL Token from azure cli token: %v", err))
			return nil, err
		}

		return &adalToken, nil
	}

	spt.SetCustomRefreshFunc(customRefreshFunc)

	return spt, nil
}

// getIDsFromAzureCLI returns the TenantID and SubscriptionID from an active Azure CLI login session
func getIDsFromAzureCLI() (string, string, error) {
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

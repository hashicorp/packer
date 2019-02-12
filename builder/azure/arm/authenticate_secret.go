package arm

import (
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

// for clientID/secret auth
type secretOAuthTokenProvider struct {
	env                              azure.Environment
	clientID, clientSecret, tenantID string
}

func NewSecretOAuthTokenProvider(env azure.Environment, clientID, clientSecret, tenantID string) oAuthTokenProvider {
	return &secretOAuthTokenProvider{env, clientID, clientSecret, tenantID}
}

func (tp *secretOAuthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return tp.getServicePrincipalTokenWithResource(tp.env.ResourceManagerEndpoint)
}

func (tp *secretOAuthTokenProvider) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(tp.env.ActiveDirectoryEndpoint, tp.tenantID)
	if err != nil {
		return nil, err
	}

	spt, err := adal.NewServicePrincipalToken(
		*oauthConfig,
		tp.clientID,
		tp.clientSecret,
		resource)

	return spt, err
}

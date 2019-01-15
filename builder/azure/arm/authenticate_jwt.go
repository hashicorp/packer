package arm

import (
	"net/url"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

// for clientID/bearer JWT auth
type jwtOAuthTokenProvider struct {
	env                           azure.Environment
	clientID, clientJWT, tenantID string
}

func NewJWTOAuthTokenProvider(env azure.Environment, clientID, clientJWT, tenantID string) oAuthTokenProvider {
	return &jwtOAuthTokenProvider{env, clientID, clientJWT, tenantID}
}

func (tp *jwtOAuthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return tp.getServicePrincipalTokenWithResource(tp.env.ResourceManagerEndpoint)
}

func (tp *jwtOAuthTokenProvider) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(tp.env.ActiveDirectoryEndpoint, tp.tenantID)
	if err != nil {
		return nil, err
	}

	return adal.NewServicePrincipalTokenWithSecret(
		*oauthConfig,
		tp.clientID,
		resource,
		tp)
}

// implements github.com/Azure/go-autorest/autorest/adal.ServicePrincipalSecret
func (tp *jwtOAuthTokenProvider) SetAuthenticationValues(
	t *adal.ServicePrincipalToken, v *url.Values) error {
	v.Set("client_assertion", tp.clientJWT)
	v.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	return nil
}

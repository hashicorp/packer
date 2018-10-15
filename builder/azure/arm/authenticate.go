package arm

import (
	"net/url"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type oAuthTokenProvider interface {
	getServicePrincipalToken() (*adal.ServicePrincipalToken, error)
	getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error)
}

// for clientID/secret auth
type secretOAuthTokenProvider struct {
	env                              azure.Environment
	clientID, clientSecret, tenantID string
}

func NewSecretOAuthTokenProvider(env azure.Environment, clientID, clientSecret, tenantID string) oAuthTokenProvider {
	return &secretOAuthTokenProvider{env, clientID, clientSecret, tenantID}
}

func (a *secretOAuthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return a.getServicePrincipalTokenWithResource(a.env.ResourceManagerEndpoint)
}

func (a *secretOAuthTokenProvider) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(a.env.ActiveDirectoryEndpoint, a.tenantID)
	if err != nil {
		return nil, err
	}

	spt, err := adal.NewServicePrincipalToken(
		*oauthConfig,
		a.clientID,
		a.clientSecret,
		resource)

	return spt, err
}

// for clientID/bearer JWT auth
type jwtOAuthTokenProvider struct {
	env                           azure.Environment
	clientID, clientJWT, tenantID string
}

func NewJWTOAuthTokenProvider(env azure.Environment, clientID, clientJWT, tenantID string) oAuthTokenProvider {
	return &jwtOAuthTokenProvider{env, clientID, clientJWT, tenantID}
}

func (pt *jwtOAuthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return pt.getServicePrincipalTokenWithResource(pt.env.ResourceManagerEndpoint)
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

// implements ServicePrincipalSecret
func (tp *jwtOAuthTokenProvider) SetAuthenticationValues(
	t *adal.ServicePrincipalToken, v *url.Values) error {
	v.Set("client_assertion", tp.clientJWT)
	v.Set("client_assertion_type", "urn:ietf:params:oauth:client-assertion-type:jwt-bearer")
	return nil
}

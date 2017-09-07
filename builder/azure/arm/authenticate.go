package arm

import (
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

type Authenticate struct {
	env          azure.Environment
	clientID     string
	clientSecret string
	tenantID     string
}

func NewAuthenticate(env azure.Environment, clientID, clientSecret, tenantID string) *Authenticate {
	return &Authenticate{
		env:          env,
		clientID:     clientID,
		clientSecret: clientSecret,
		tenantID:     tenantID,
	}
}

func (a *Authenticate) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return a.getServicePrincipalTokenWithResource(a.env.ResourceManagerEndpoint)
}

func (a *Authenticate) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
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

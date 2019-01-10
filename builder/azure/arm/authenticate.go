package arm

import (
	"github.com/Azure/go-autorest/autorest/adal"
)

type oAuthTokenProvider interface {
	getServicePrincipalToken() (*adal.ServicePrincipalToken, error)
	getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error)
}

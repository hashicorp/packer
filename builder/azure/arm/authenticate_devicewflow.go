package arm

import (
	"fmt"
	"strings"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	packerAzureCommon "github.com/hashicorp/packer/builder/azure/common"
)

func NewDeviceFlowOAuthTokenProvider(env azure.Environment, say func(string), tenantID string) oAuthTokenProvider {
	return &deviceflowOauthTokenProvider{
		env:      env,
		say:      say,
		tenantID: tenantID,
	}
}

type deviceflowOauthTokenProvider struct {
	env      azure.Environment
	say      func(string)
	tenantID string
}

func (tp *deviceflowOauthTokenProvider) getServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	return tp.getServicePrincipalTokenWithResource(tp.env.ResourceManagerEndpoint)
}

func (tp *deviceflowOauthTokenProvider) getServicePrincipalTokenWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	if resource == tp.env.ServiceManagementEndpoint {
		tp.say("Getting auth token for Service management endpoint")
	} else if resource == strings.TrimRight(tp.env.KeyVaultEndpoint, "/") {
		tp.say("Getting token for Vault resource")
	} else {
		tp.say(fmt.Sprintf("Getting token for %s", resource))
	}

	return packerAzureCommon.Authenticate(tp.env, tp.tenantID, tp.say, resource)
}

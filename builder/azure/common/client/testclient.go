package client

import (
	"os"
	"testing"
	"net/http"
	"errors"

	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func GetTestClientSet(t *testing.T) (AzureClientSet, error) {
	if (os.Getenv("AZURE_INTEGRATION_TEST") == "") {
		t.Skip("AZURE_INTEGRATION_TEST not set")		
	} else {
		a, err := auth.NewAuthorizerFromEnvironment()
		if err == nil {
			cli := azureClientSet{}
		cli.authorizer = a
			cli.subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
			cli.PollingDelay = 0
			cli.sender = http.DefaultClient
			return cli, nil
					} else {
			t.Skipf("Could not create Azure client: %v", err)
		}
	}

	return nil, errors.New("Couldn't create client set")
}


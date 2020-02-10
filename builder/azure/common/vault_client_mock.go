package common

import (
	"fmt"
	"net/http"

	"github.com/Azure/go-autorest/autorest"
)

type MockAZVaultClient struct {
	GetSecretCalled       bool
	SetSecretCalled       bool
	SetSecretVaultName    string
	SetSecretSecretName   string
	SetSecretCert         string
	DeleteResponderCalled bool
	DeletePreparerCalled  bool
	DeleteSenderCalled    bool

	IsError bool
}

func (m *MockAZVaultClient) GetSecret(vaultName, secretName string) (*Secret, error) {
	m.GetSecretCalled = true
	var secret Secret
	return &secret, nil
}

func (m *MockAZVaultClient) SetSecret(vaultName, secretName string, secretValue string) error {
	m.SetSecretCalled = true
	m.SetSecretVaultName = vaultName
	m.SetSecretSecretName = secretName
	m.SetSecretCert = secretValue

	if m.IsError {
		return fmt.Errorf("generic error!!")
	}

	return nil
}

func (m *MockAZVaultClient) DeletePreparer(resourceGroupName string, vaultName string) (*http.Request, error) {
	m.DeletePreparerCalled = true
	return nil, nil
}

func (m *MockAZVaultClient) DeleteResponder(resp *http.Response) (autorest.Response, error) {
	m.DeleteResponderCalled = true
	var result autorest.Response
	return result, nil
}

func (m *MockAZVaultClient) DeleteSender(req *http.Request) (*http.Response, error) {
	m.DeleteSenderCalled = true
	return nil, nil
}

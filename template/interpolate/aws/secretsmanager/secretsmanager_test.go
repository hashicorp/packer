package secretsmanager

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

type mockedSecret struct {
	secretsmanageriface.SecretsManagerAPI
	Resp secretsmanager.GetSecretValueOutput
}

// GetSecret return mocked secret value
func (m mockedSecret) GetSecretValue(in *secretsmanager.GetSecretValueInput) (*secretsmanager.GetSecretValueOutput, error) {
	return &m.Resp, nil
}

func TestGetSecret(t *testing.T) {
	testCases := []struct {
		description string
		arg         *SecretSpec
		mock        secretsmanager.GetSecretValueOutput
		want        string
		ok          bool
	}{
		{
			description: "input has valid secret name, secret has single key",
			arg:         &SecretSpec{Name: "test/secret"},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"key": "test"}`),
			},
			want: "test",
			ok:   true,
		},
		{
			description: "input has valid secret name and key, secret has single key",
			arg: &SecretSpec{
				Name: "test/secret",
				Key:  "key",
			},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"key": "test"}`),
			},
			want: "test",
			ok:   true,
		},
		{
			description: "input has valid secret name and key, secret has multiple keys",
			arg: &SecretSpec{
				Name: "test/secret",
				Key:  "second_key",
			},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"first_key": "first_val", "second_key": "second_val"}`),
			},
			want: "second_val",
			ok:   true,
		},
		{
			description: "input has valid secret name and no key, secret has multiple keys",
			arg: &SecretSpec{
				Name: "test/secret",
			},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"first_key": "first_val", "second_key": "second_val"}`),
			},
			ok: false,
		},
		{
			description: "input has valid secret name and invalid key, secret has single key",
			arg: &SecretSpec{
				Name: "test/secret",
				Key:  "nonexistent",
			},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"key": "test"}`),
			},
			ok: false,
		},
		{
			description: "input has valid secret name and invalid key, secret has multiple keys",
			arg: &SecretSpec{
				Name: "test/secret",
				Key:  "nonexistent",
			},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"first_key": "first_val", "second_key": "second_val"}`),
			},
			ok: false,
		},
		{
			description: "input has secret and key, secret is empty",
			arg: &SecretSpec{
				Name: "test/secret",
				Key:  "nonexistent",
			},
			mock: secretsmanager.GetSecretValueOutput{},
			ok:   false,
		},
	}

	for _, test := range testCases {
		c := &Client{
			api: mockedSecret{Resp: test.mock},
		}
		got, err := c.GetSecret(test.arg)
		if test.ok {
			if got != test.want {
				t.Logf("want %v, got %v, error %v, using arg %v", test.want, got, err, test.arg)
			}
		}
		if !test.ok {
			if err == nil {
				t.Logf("error expected but got %q, using arg %v", err, test.arg)
			}
		}
		t.Logf("arg (%v), want %v, got %v, err %v", test.arg, test.want, got, err)
	}
}

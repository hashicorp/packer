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
		arg  *SecretSpec
		mock secretsmanager.GetSecretValueOutput
		want string
		ok   bool
	}{
		{
			arg: &SecretSpec{Name: "test/secret"},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"key": "test"}`),
			},
			want: "test",
			ok:   true,
		},
		{
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
			arg: &SecretSpec{
				Name: "test/secret",
			},
			mock: secretsmanager.GetSecretValueOutput{
				Name:         aws.String("test/secret"),
				SecretString: aws.String(`{"first_key": "first_val", "second_key": "second_val"}`),
			},
			want: "first_val",
			ok:   true,
		},
		{
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

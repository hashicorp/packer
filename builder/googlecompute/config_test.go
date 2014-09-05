package googlecompute

import (
	"io/ioutil"
	"testing"
)

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"account_file":        testAccountFile(t),
		"bucket_name":         "foo",
		"client_secrets_file": testClientSecretsFile(t),
		"project_id":          "hashicorp",
		"source_image":        "foo",
		"zone":                "us-east-1a",
	}
}

func testConfigStruct(t *testing.T) *Config {
	c, warns, errs := NewConfig(testConfig(t))
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", len(warns))
	}
	if errs != nil {
		t.Fatalf("bad: %#v", errs)
	}

	return c
}

func testConfigErr(t *testing.T, warns []string, err error, extra string) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatalf("should error: %s", extra)
	}
}

func testConfigOk(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err != nil {
		t.Fatalf("bad: %s", err)
	}
}

func TestConfigPrepare(t *testing.T) {
	cases := []struct {
		Key   string
		Value interface{}
		Err   bool
	}{
		{
			"unknown_key",
			"bad",
			true,
		},

		{
			"bucket_name",
			nil,
			true,
		},
		{
			"bucket_name",
			"good",
			false,
		},

		{
			"client_secrets_file",
			nil,
			true,
		},
		{
			"client_secrets_file",
			testClientSecretsFile(t),
			false,
		},
		{
			"client_secrets_file",
			"/tmp/i/should/not/exist",
			true,
		},

		{
			"private_key_file",
			"/tmp/i/should/not/exist",
			true,
		},

		{
			"project_id",
			nil,
			true,
		},
		{
			"project_id",
			"foo",
			false,
		},

		{
			"source_image",
			nil,
			true,
		},
		{
			"source_image",
			"foo",
			false,
		},

		{
			"zone",
			nil,
			true,
		},
		{
			"zone",
			"foo",
			false,
		},

		{
			"ssh_timeout",
			"SO BAD",
			true,
		},
		{
			"ssh_timeout",
			"5s",
			false,
		},

		{
			"state_timeout",
			"SO BAD",
			true,
		},
		{
			"state_timeout",
			"5s",
			false,
		},
	}

	for _, tc := range cases {
		raw := testConfig(t)

		if tc.Value == nil {
			delete(raw, tc.Key)
		} else {
			raw[tc.Key] = tc.Value
		}

		_, warns, errs := NewConfig(raw)

		if tc.Err {
			testConfigErr(t, warns, errs, tc.Key)
		} else {
			testConfigOk(t, warns, errs)
		}
	}
}

func testAccountFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()

	if _, err := tf.Write([]byte(testAccountContent)); err != nil {
		t.Fatalf("err: %s", err)
	}

	return tf.Name()
}

func testClientSecretsFile(t *testing.T) string {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer tf.Close()

	if _, err := tf.Write([]byte(testClientSecretsContent)); err != nil {
		t.Fatalf("err: %s", err)
	}

	return tf.Name()
}

// This is just some dummy data that doesn't actually work (it was revoked
// a long time ago).
const testAccountContent = `{}`

const testClientSecretsContent = `{"web":{"auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://accounts.google.com/o/oauth2/token","client_email":"774313886706-eorlsj0r4eqkh5e7nvea5fuf59ifr873@developer.gserviceaccount.com","client_x509_cert_url":"https://www.googleapis.com/robot/v1/metadata/x509/774313886706-eorlsj0r4eqkh5e7nvea5fuf59ifr873@developer.gserviceaccount.com","client_id":"774313886706-eorlsj0r4eqkh5e7nvea5fuf59ifr873.apps.googleusercontent.com","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs"}}`

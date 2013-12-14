package googlecompute

import (
	"io/ioutil"
	"testing"
)

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

func TestLoadClientSecrets(t *testing.T) {
	_, err := loadClientSecrets(testClientSecretsFile(t))
	if err != nil {
		t.Fatalf("err: %s", err)
	}
}

// This is just some dummy data that doesn't actually work (it was revoked
// a long time ago).
const testClientSecretsContent = `{"web":{"auth_uri":"https://accounts.google.com/o/oauth2/auth","token_uri":"https://accounts.google.com/o/oauth2/token","client_email":"774313886706-eorlsj0r4eqkh5e7nvea5fuf59ifr873@developer.gserviceaccount.com","client_x509_cert_url":"https://www.googleapis.com/robot/v1/metadata/x509/774313886706-eorlsj0r4eqkh5e7nvea5fuf59ifr873@developer.gserviceaccount.com","client_id":"774313886706-eorlsj0r4eqkh5e7nvea5fuf59ifr873.apps.googleusercontent.com","auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs"}}`

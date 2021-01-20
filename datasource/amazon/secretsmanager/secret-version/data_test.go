package secret_version

import (
	"testing"
)

func TestDatasourceConfigure_EmptySecretId(t *testing.T) {
	datasource := Datasource{
		config: Config{},
	}
	if err := datasource.Configure(nil); err == nil {
		t.Fatalf("Should error if secret id is not specified")
	}
}

func TestDatasourceConfigure(t *testing.T) {
	datasource := Datasource{
		config: Config{
			SecretId: "arn:1223",
		},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf("err: %s", err)
	}
}

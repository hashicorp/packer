package secretsmanager

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

func TestDatasourceConfigure_Dafaults(t *testing.T) {
	datasource := Datasource{
		config: Config{
			Name: "arn:1223",
		},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf("err: %s", err)
	}
	if datasource.config.VersionStage != "AWSCURRENT" {
		t.Fatalf("VersionStage not set correctly")
	}
}

func TestDatasourceConfigure(t *testing.T) {
	datasource := Datasource{
		config: Config{
			Name: "arn:1223",
		},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf("err: %s", err)
	}
}

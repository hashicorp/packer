package secret

import (
	"testing"
)

func TestDatasourceConfigure_EmptyArnAndName(t *testing.T) {
	datasource := Datasource{
		config: Config{},
	}
	if err := datasource.Configure(nil); err == nil {
		t.Fatalf("Should error if arn and name are both not specified")
	}
}

func TestDatasourceConfigure_BothArnAndNameSet(t *testing.T) {
	datasource := Datasource{
		config: Config{
			Arn:  "arn:1223",
			Name: "1223",
		},
	}
	if err := datasource.Configure(nil); err == nil {
		t.Fatalf("Should error if both arn and nam is specified)")
	}
}

func TestDatasourceConfigure_SecretIdWithArnOrNameValue(t *testing.T) {
	datasource := Datasource{
		config: Config{
			Arn: "arn:1223",
		},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf("err: %s", err)
	}
	if datasource.config.secretId != "arn:1223" {
		t.Fatalf("unexpected secretID: %s", datasource.config.secretId)
	}

	datasource = Datasource{
		config: Config{
			Name: "1223",
		},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf("err: %s", err)
	}
	if datasource.config.secretId != "1223" {
		t.Fatalf("unexpected secretID: %s", datasource.config.secretId)
	}
}

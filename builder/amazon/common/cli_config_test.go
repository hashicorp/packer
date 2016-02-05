package common

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
)

func init() {
	os.Setenv("AWS_ACCESS_KEY_ID", "")
	os.Setenv("AWS_ACCESS_KEY", "")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "")
	os.Setenv("AWS_SECRET_KEY", "")
	os.Setenv("AWS_CONFIG_FILE", "")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "")
}

func testCLIConfig() *CLIConfig {
	return &CLIConfig{}
}

func TestCLIConfigNewFromProfile(t *testing.T) {
	tmpDir := mockConfig(t)

	c, err := NewFromProfile("testing2")
	if err != nil {
		t.Error(err)
	}
	if c.AssumeRoleInput.RoleArn != nil {
		t.Errorf("RoleArn should be nil. Instead %p", c.AssumeRoleInput.RoleArn)
	}
	if c.AssumeRoleInput.ExternalId != nil {
		t.Errorf("ExternalId should be nil. Instead %p", c.AssumeRoleInput.ExternalId)
	}

	mockConfigClose(t, tmpDir)
}

func TestAssumeRole(t *testing.T) {
	tmpDir := mockConfig(t)

	c, err := NewFromProfile("testing1")
	if err != nil {
		t.Error(err)
	}
	// Role
	e := "arn:aws:iam::123456789011:role/rolename"
	a := *c.AssumeRoleInput.RoleArn
	if e != a {
		t.Errorf("RoleArn value should be %s. Instead %s", e, a)
	}
	// Session
	a = *c.AssumeRoleInput.RoleSessionName
	e = "testsession"
	if e != a {
		t.Errorf("RoleSessionName value should be %s. Instead %s", e, a)
	}

	config := aws.NewConfig()
	_, err = c.CredentialsFromProfile(config)
	if err == nil {
		t.Error("Should have errored")
	}
	mockConfigClose(t, tmpDir)
}

func mockConfig(t *testing.T) string {
	time := time.Now().UnixNano()
	dir, err := ioutil.TempDir("", strconv.FormatInt(time, 10))
	if err != nil {
		t.Error(err)
	}

	cfg := []byte(`[profile testing1]
region=us-west-2
source_profile=testingcredentials
role_arn = arn:aws:iam::123456789011:role/rolename
role_session_name = testsession

[profile testing2]
region=us-west-2
	`)
	cfgFile := path.Join(dir, "config")
	err = ioutil.WriteFile(cfgFile, cfg, 0644)
	if err != nil {
		t.Error(err)
	}
	os.Setenv("AWS_CONFIG_FILE", cfgFile)

	crd := []byte(`[testingcredentials]
aws_access_key_id = foo
aws_secret_access_key = bar

[testing2]
aws_access_key_id = baz
aws_secret_access_key = qux
	`)
	crdFile := path.Join(dir, "credentials")
	err = ioutil.WriteFile(crdFile, crd, 0644)
	if err != nil {
		t.Error(err)
	}
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", crdFile)

	return dir
}

func mockConfigClose(t *testing.T, dir string) {
	err := os.RemoveAll(dir)
	if err != nil {
		t.Error(err)
	}
}

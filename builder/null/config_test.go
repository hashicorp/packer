package null

import (
	"testing"
)

func testConfig() map[string]interface{} {
	return map[string]interface{}{
		"host":         "foo",
		"ssh_username": "bar",
		"ssh_password": "baz",
	}
}

func testConfigStruct(t *testing.T) *Config {
	c, warns, errs := NewConfig(testConfig())
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", len(warns))
	}
	if errs != nil {
		t.Fatalf("bad: %#v", errs)
	}

	return c
}

func testConfigErr(t *testing.T, warns []string, err error) {
	if len(warns) > 0 {
		t.Fatalf("bad: %#v", warns)
	}
	if err == nil {
		t.Fatal("should error")
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

func TestConfigPrepare_port(t *testing.T) {
	raw := testConfig()

	// default port should be 22
	delete(raw, "port")
	c, warns, errs := NewConfig(raw)
	if c.Port != 22 {
		t.Fatalf("bad: port should default to 22, not %d", c.Port)
	}
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_host(t *testing.T) {
	raw := testConfig()

	// No host
	delete(raw, "host")
	_, warns, errs := NewConfig(raw)
	testConfigErr(t, warns, errs)

	// Good host
	raw["host"] = "good"
	_, warns, errs = NewConfig(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_sshUsername(t *testing.T) {
	raw := testConfig()

	// No ssh_username
	delete(raw, "ssh_username")
	_, warns, errs := NewConfig(raw)
	testConfigErr(t, warns, errs)

	// Good ssh_username
	raw["ssh_username"] = "good"
	_, warns, errs = NewConfig(raw)
	testConfigOk(t, warns, errs)
}

func TestConfigPrepare_sshCredential(t *testing.T) {
	raw := testConfig()

	// no ssh_password and no ssh_private_key_file
	delete(raw, "ssh_password")
	delete(raw, "ssh_private_key_file")
	_, warns, errs := NewConfig(raw)
	testConfigErr(t, warns, errs)

	// only ssh_password
	raw["ssh_password"] = "good"
	_, warns, errs = NewConfig(raw)
	testConfigOk(t, warns, errs)

	// only ssh_private_key_file
	raw["ssh_private_key_file"] = "good"
	delete(raw, "ssh_password")
	_, warns, errs = NewConfig(raw)
	testConfigOk(t, warns, errs)

	// both ssh_password and ssh_private_key_file set
	raw["ssh_password"] = "bad"
	_, warns, errs = NewConfig(raw)
	testConfigErr(t, warns, errs)
}

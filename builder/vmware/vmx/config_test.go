package vmx

import (
	"io/ioutil"
	"os"
	"testing"
)

func testConfig(t *testing.T) map[string]interface{} {
	return map[string]interface{}{
		"ssh_username":     "foo",
		"shutdown_command": "foo",
		"source_path":      "config_test.go",
	}
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

func TestNewConfig_sourcePath(t *testing.T) {
	// Bad
	cfg := testConfig(t)
	delete(cfg, "source_path")
	warns, errs := (&Config{}).Prepare(cfg)
	testConfigErr(t, warns, errs)

	// Bad
	cfg = testConfig(t)
	cfg["source_path"] = "/i/dont/exist"
	warns, errs = (&Config{}).Prepare(cfg)
	testConfigErr(t, warns, errs)

	// Good
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	tf.Close()
	defer os.Remove(tf.Name())

	cfg = testConfig(t)
	cfg["source_path"] = tf.Name()
	warns, errs = (&Config{}).Prepare(cfg)
	testConfigOk(t, warns, errs)
}

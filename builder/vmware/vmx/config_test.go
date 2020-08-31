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

func TestNewConfig_exportConfig(t *testing.T) {
	type testCase struct {
		InputConfigVals         map[string]string
		ExpectedSkipExportValue bool
		ExpectedFormat          string
		ExpectedErr             bool
		Reason                  string
	}
	testCases := []testCase{
		{
			InputConfigVals: map[string]string{
				"remote_type": "",
				"format":      "",
			},
			ExpectedSkipExportValue: true,
			ExpectedFormat:          "vmx",
			ExpectedErr:             false,
			Reason:                  "should have defaulted format to vmx.",
		},
		{
			InputConfigVals: map[string]string{
				"remote_type":     "esx5",
				"format":          "",
				"remote_host":     "fakehost.com",
				"remote_password": "fakepassword",
				"remote_username": "fakeuser",
			},
			ExpectedSkipExportValue: false,
			ExpectedFormat:          "ovf",
			ExpectedErr:             false,
			Reason:                  "should have defaulted format to ovf with remote set to esx5.",
		},
		{
			InputConfigVals: map[string]string{
				"remote_type": "esx5",
				"format":      "",
			},
			ExpectedSkipExportValue: false,
			ExpectedFormat:          "ovf",
			ExpectedErr:             true,
			Reason:                  "should have errored because remote host isn't set for remote build.",
		},
		{
			InputConfigVals: map[string]string{
				"remote_type":     "invalid",
				"format":          "",
				"remote_host":     "fakehost.com",
				"remote_password": "fakepassword",
				"remote_username": "fakeuser",
			},
			ExpectedSkipExportValue: false,
			ExpectedFormat:          "ovf",
			ExpectedErr:             true,
			Reason:                  "should error with invalid remote type",
		},
		{
			InputConfigVals: map[string]string{
				"remote_type": "",
				"format":      "invalid",
			},
			ExpectedSkipExportValue: false,
			ExpectedFormat:          "invalid",
			ExpectedErr:             true,
			Reason:                  "should error with invalid format",
		},
		{
			InputConfigVals: map[string]string{
				"remote_type": "",
				"format":      "ova",
			},
			ExpectedSkipExportValue: false,
			ExpectedFormat:          "ova",
			ExpectedErr:             false,
			Reason:                  "should set user-given ova format",
		},
		{
			InputConfigVals: map[string]string{
				"remote_type":     "esx5",
				"format":          "ova",
				"remote_host":     "fakehost.com",
				"remote_password": "fakepassword",
				"remote_username": "fakeuser",
			},
			ExpectedSkipExportValue: false,
			ExpectedFormat:          "ova",
			ExpectedErr:             false,
			Reason:                  "should set user-given ova format",
		},
	}
	for _, tc := range testCases {
		cfg := testConfig(t)
		for k, v := range tc.InputConfigVals {
			cfg[k] = v
		}
		cfg["skip_validate_credentials"] = true
		outCfg := &Config{}
		warns, errs := (outCfg).Prepare(cfg)

		if len(warns) > 0 {
			t.Fatalf("bad: %#v", warns)
		}

		if (errs != nil) != tc.ExpectedErr {
			t.Fatalf("received error: \n %s \n but 'expected err' was %t", errs, tc.ExpectedErr)
		}

		if outCfg.Format != tc.ExpectedFormat {
			t.Fatalf("Expected: %s. Actual: %s. Reason: %s", tc.ExpectedFormat,
				outCfg.Format, tc.Reason)
		}
		if outCfg.SkipExport != tc.ExpectedSkipExportValue {
			t.Fatalf("For SkipExport expected %t but recieved %t",
				tc.ExpectedSkipExportValue, outCfg.SkipExport)
		}
	}
}

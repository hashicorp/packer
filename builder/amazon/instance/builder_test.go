package instance

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hashicorp/packer/packer"
)

func testConfig() (config map[string]interface{}, tf *os.File) {
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		panic(err)
	}

	config = map[string]interface{}{
		"account_id":       "foo",
		"ami_name":         "foo",
		"instance_type":    "m1.small",
		"region":           "us-east-1",
		"s3_bucket":        "foo",
		"source_ami":       "foo",
		"ssh_username":     "bob",
		"x509_cert_path":   tf.Name(),
		"x509_key_path":    tf.Name(),
		"x509_upload_path": "/foo",
	}

	return config, tf
}

func TestBuilder_ImplementsBuilder(t *testing.T) {
	var raw interface{}
	raw = &Builder{}
	if _, ok := raw.(packer.Builder); !ok {
		t.Fatalf("Builder should be a builder")
	}
}

func TestBuilderPrepare_AccountId(t *testing.T) {
	b := &Builder{}
	config, tempfile := testConfig()
	config["skip_region_validation"] = true

	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	config["account_id"] = ""
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["account_id"] = "foo"
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
	}

	config["account_id"] = "0123-0456-7890"
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.AccountId != "012304567890" {
		t.Errorf("should strip hyphens: %s", b.config.AccountId)
	}
}

func TestBuilderPrepare_AMIName(t *testing.T) {
	var b Builder
	config, tempfile := testConfig()
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	// Test good
	config["ami_name"] = "foo"
	config["skip_region_validation"] = true
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ami_name"] = "foo {{"
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad
	delete(config, "ami_name")
	b = Builder{}
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_BundleDestination(t *testing.T) {
	b := &Builder{}
	config, tempfile := testConfig()
	config["skip_region_validation"] = true
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	config["bundle_destination"] = ""
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.BundleDestination != "/tmp" {
		t.Fatalf("bad: %s", b.config.BundleDestination)
	}
}

func TestBuilderPrepare_BundlePrefix(t *testing.T) {
	b := &Builder{}
	config, tempfile := testConfig()
	config["skip_region_validation"] = true
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if b.config.BundlePrefix == "" {
		t.Fatalf("bad: %s", b.config.BundlePrefix)
	}
}

func TestBuilderPrepare_InvalidKey(t *testing.T) {
	var b Builder
	config, tempfile := testConfig()
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	// Add a random key
	config["i_should_not_be_valid"] = true
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}
}

func TestBuilderPrepare_S3Bucket(t *testing.T) {
	b := &Builder{}
	config, tempfile := testConfig()
	config["skip_region_validation"] = true
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	config["s3_bucket"] = ""
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["s3_bucket"] = "foo"
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
	}
}

func TestBuilderPrepare_X509CertPath(t *testing.T) {
	b := &Builder{}
	config, tempfile := testConfig()
	config["skip_region_validation"] = true
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	config["x509_cert_path"] = ""
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["x509_cert_path"] = "i/am/a/file/that/doesnt/exist"
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Error("should have error")
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	config["x509_cert_path"] = tf.Name()
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_X509KeyPath(t *testing.T) {
	b := &Builder{}
	config, tempfile := testConfig()
	config["skip_region_validation"] = true
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	config["x509_key_path"] = ""
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["x509_key_path"] = "i/am/a/file/that/doesnt/exist"
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Error("should have error")
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("error tempfile: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	config["x509_key_path"] = tf.Name()
	warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_X509UploadPath(t *testing.T) {
	b := &Builder{}
	config, tempfile := testConfig()
	config["skip_region_validation"] = true
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	config["x509_upload_path"] = ""
	warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

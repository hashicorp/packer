package instance

import (
	"io/ioutil"
	"os"
	"testing"

	packersdk "github.com/hashicorp/packer/packer-plugin-sdk/packer"
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
	if _, ok := raw.(packersdk.Builder); !ok {
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
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["account_id"] = "foo"
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Errorf("err: %s", err)
	}

	config["account_id"] = "0123-0456-7890"
	_, warnings, err = b.Prepare(config)
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
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}

	// Test bad
	config["ami_name"] = "foo {{"
	b = Builder{}
	_, warnings, err = b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	// Test bad
	delete(config, "ami_name")
	b = Builder{}
	_, warnings, err = b.Prepare(config)
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
	_, warnings, err := b.Prepare(config)
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

	_, warnings, err := b.Prepare(config)
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
	_, warnings, err := b.Prepare(config)
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
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["s3_bucket"] = "foo"
	_, warnings, err = b.Prepare(config)
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
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["x509_cert_path"] = "i/am/a/file/that/doesnt/exist"
	_, warnings, err = b.Prepare(config)
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
	_, warnings, err = b.Prepare(config)
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
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err == nil {
		t.Fatal("should have error")
	}

	config["x509_key_path"] = "i/am/a/file/that/doesnt/exist"
	_, warnings, err = b.Prepare(config)
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
	_, warnings, err = b.Prepare(config)
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
	_, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
}

func TestBuilderPrepare_ReturnGeneratedData(t *testing.T) {
	var b Builder
	config, tempfile := testConfig()
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()

	generatedData, warnings, err := b.Prepare(config)
	if len(warnings) > 0 {
		t.Fatalf("bad: %#v", warnings)
	}
	if err != nil {
		t.Fatalf("should not have error: %s", err)
	}
	if len(generatedData) == 0 {
		t.Fatalf("Generated data should not be empty")
	}
	if len(generatedData) == 0 {
		t.Fatalf("Generated data should not be empty")
	}
	if generatedData[0] != "SourceAMIName" {
		t.Fatalf("Generated data should contain SourceAMIName")
	}
	if generatedData[1] != "BuildRegion" {
		t.Fatalf("Generated data should contain BuildRegion")
	}
	if generatedData[2] != "SourceAMI" {
		t.Fatalf("Generated data should contain SourceAMI")
	}
	if generatedData[3] != "SourceAMICreationDate" {
		t.Fatalf("Generated data should contain SourceAMICreationDate")
	}
	if generatedData[4] != "SourceAMIOwner" {
		t.Fatalf("Generated data should contain SourceAMIOwner")
	}
	if generatedData[5] != "SourceAMIOwnerName" {
		t.Fatalf("Generated data should contain SourceAMIOwnerName")
	}
}

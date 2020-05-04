package common

import (
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/packer/hcl2template"
	"github.com/hashicorp/packer/helper/communicator"
)

func init() {
	// Clear out the OUTSCALE access key env vars so they don't
	// affect our tests.
	os.Setenv("OUTSCALE_ACCESS_KEY_ID", "")
	os.Setenv("OUTSCALE_ACCESS_KEY", "")
	os.Setenv("OUTSCALE_SECRET_ACCESS_KEY", "")
	os.Setenv("OUTSCALE_SECRET_KEY", "")
}

func testConfig() *RunConfig {
	return &RunConfig{
		SourceOmi: "abcd",
		VmType:    "m1.small",
		Comm: communicator.Config{
			SSH: communicator.SSH{
				SSHUsername: "foo",
			},
		},
	}
}

func testConfigFilter() *RunConfig {
	config := testConfig()
	config.SourceOmi = ""
	config.SourceOmiFilter = OmiFilterOptions{}
	return config
}

func TestRunConfigPrepare(t *testing.T) {
	c := testConfig()
	err := c.Prepare(nil)
	if len(err) > 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_VmType(t *testing.T) {
	c := testConfig()
	c.VmType = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("Should error if an vm_type is not specified")
	}
}

func TestRunConfigPrepare_SourceOmi(t *testing.T) {
	c := testConfig()
	c.SourceOmi = ""
	if err := c.Prepare(nil); len(err) != 2 {
		t.Fatalf("Should error if a source_omi (or source_omi_filter) is not specified")
	}
}

func TestRunConfigPrepare_SourceOmiFilterBlank(t *testing.T) {
	c := testConfigFilter()
	if err := c.Prepare(nil); len(err) != 2 {
		t.Fatalf("Should error if source_ami_filter is empty or not specified (and source_ami is not specified)")
	}
}

func TestRunConfigPrepare_SourceOmiFilterOwnersBlank(t *testing.T) {
	c := testConfigFilter()
	filter_key := "name"
	filter_value := "foo"
	c.SourceOmiFilter = OmiFilterOptions{
		NameValueFilter: hcl2template.NameValueFilter{
			Filters: map[string]string{filter_key: filter_value},
		},
	}
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("Should error if Owners is not specified)")
	}
}

func TestRunConfigPrepare_SourceOmiFilterGood(t *testing.T) {
	c := testConfigFilter()
	owner := "123"
	filter_key := "name"
	filter_value := "foo"
	goodFilter := OmiFilterOptions{
		Owners: []string{owner},
		NameValueFilter: hcl2template.NameValueFilter{
			Filters: map[string]string{filter_key: filter_value},
		},
	}
	c.SourceOmiFilter = goodFilter
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_EnableT2UnlimitedGood(t *testing.T) {
	c := testConfig()
	// Must have a T2 vm type if T2 Unlimited is enabled
	c.VmType = "t2.micro"
	c.EnableT2Unlimited = true
	err := c.Prepare(nil)
	if len(err) > 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_EnableT2UnlimitedBadVmType(t *testing.T) {
	c := testConfig()
	// T2 Unlimited cannot be used with vm types other than T2
	c.VmType = "m5.large"
	c.EnableT2Unlimited = true
	err := c.Prepare(nil)
	if len(err) != 1 {
		t.Fatalf("Should error if T2 Unlimited is enabled with non-T2 vm_type")
	}
}

func TestRunConfigPrepare_EnableT2UnlimitedBadWithSpotInstanceRequest(t *testing.T) {
	c := testConfig()
	// T2 Unlimited cannot be used with Spot Instances
	c.VmType = "t2.micro"
	c.EnableT2Unlimited = true
	c.SpotPrice = "auto"
	c.SpotPriceAutoProduct = "Linux/UNIX"
	err := c.Prepare(nil)
	if len(err) != 1 {
		t.Fatalf("Should error if T2 Unlimited has been used in conjuntion with a Spot Price request")
	}
}

func TestRunConfigPrepare_SpotAuto(t *testing.T) {
	c := testConfig()
	c.SpotPrice = "auto"
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("Should error if spot_price_auto_product is not set and spot_price is set to auto")
	}

	// Good - SpotPrice and SpotPriceAutoProduct are correctly set
	c.SpotPriceAutoProduct = "foo"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	c.SpotPrice = ""
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("Should error if spot_price is not set to auto and spot_price_auto_product is set")
	}
}

func TestRunConfigPrepare_SSHPort(t *testing.T) {
	c := testConfig()
	c.Comm.SSHPort = 0
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 22 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}

	c.Comm.SSHPort = 44
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHPort != 44 {
		t.Fatalf("invalid value: %d", c.Comm.SSHPort)
	}
}

func TestRunConfigPrepare_UserData(t *testing.T) {
	c := testConfig()
	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	c.UserData = "foo"
	c.UserDataFile = tf.Name()
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("Should error if user_data string and user_data_file have both been specified")
	}
}

func TestRunConfigPrepare_UserDataFile(t *testing.T) {
	c := testConfig()
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	c.UserDataFile = "idontexistidontthink"
	if err := c.Prepare(nil); len(err) != 1 {
		t.Fatalf("Should error if the file specified by user_data_file does not exist")
	}

	tf, err := ioutil.TempFile("", "packer")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	defer os.Remove(tf.Name())
	defer tf.Close()

	c.UserDataFile = tf.Name()
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}
}

func TestRunConfigPrepare_TemporaryKeyPairName(t *testing.T) {
	c := testConfig()
	c.Comm.SSHTemporaryKeyPairName = ""
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHTemporaryKeyPairName == "" {
		t.Fatal("keypair name is empty")
	}

	// Match prefix and UUID, e.g. "packer_5790d491-a0b8-c84c-c9d2-2aea55086550".
	r := regexp.MustCompile(`\Apacker_(?:(?i)[a-f\d]{8}(?:-[a-f\d]{4}){3}-[a-f\d]{12}?)\z`)
	if !r.MatchString(c.Comm.SSHTemporaryKeyPairName) {
		t.Fatal("keypair name is not valid")
	}

	c.Comm.SSHTemporaryKeyPairName = "ssh-key-123"
	if err := c.Prepare(nil); len(err) != 0 {
		t.Fatalf("err: %s", err)
	}

	if c.Comm.SSHTemporaryKeyPairName != "ssh-key-123" {
		t.Fatal("keypair name does not match")
	}
}

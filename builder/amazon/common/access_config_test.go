package common

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func testAccessConfig() *AccessConfig {
	return &AccessConfig{
		getEC2Connection: func() ec2iface.EC2API {
			return &mockEC2Client{}
		},
	}
}

func TestAccessConfigPrepare_Region(t *testing.T) {
	c := testAccessConfig()

	c.RawRegion = "us-east-12"
	err := c.ValidateRegion(c.RawRegion)
	if err == nil {
		t.Fatalf("should have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "us-east-1"
	err = c.ValidateRegion(c.RawRegion)
	if err != nil {
		t.Fatalf("shouldn't have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "custom"
	err = c.ValidateRegion(c.RawRegion)
	if err == nil {
		t.Fatalf("should have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "custom"
	c.SkipValidation = true
	// testing whole prepare func here; this is checking that validation is
	// skipped, so we don't need a mock connection
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.SkipValidation = false
	c.RawRegion = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
}

func TestAccessConfigPrepare_RegionRestricted(t *testing.T) {
	c := testAccessConfig()

	// Create a Session with a custom region
	c.session = session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-gov-west-1"),
	}))

	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	if !c.IsGovCloud() {
		t.Fatal("We should be in gov region.")
	}
}

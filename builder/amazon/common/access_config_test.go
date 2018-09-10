package common

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func testAccessConfig() *AccessConfig {
	return &AccessConfig{}
}

func TestAccessConfigPrepare_Region(t *testing.T) {
	c := testAccessConfig()

	mockConn := &mockEC2Client{}

	c.RawRegion = "us-east-12"
	valid := ValidateRegion(c.RawRegion, mockConn)
	if valid {
		t.Fatalf("should have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "us-east-1"
	valid = ValidateRegion(c.RawRegion, mockConn)
	if !valid {
		t.Fatalf("shouldn't have region validation err: %s", c.RawRegion)
	}

	c.RawRegion = "custom"
	valid = ValidateRegion(c.RawRegion, mockConn)
	if valid {
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

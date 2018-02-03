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
	c.RawRegion = ""
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.RawRegion = "us-east-12"
	if err := c.Prepare(nil); err == nil {
		t.Fatal("should have error")
	}

	c.RawRegion = "us-east-1"
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	c.RawRegion = "custom"
	if err := c.Prepare(nil); err == nil {
		t.Fatalf("should have err")
	}

	c.RawRegion = "custom"
	c.SkipValidation = true
	if err := c.Prepare(nil); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}
	c.SkipValidation = false

}

func TestAccessConfigPrepare_RegionRestrictd(t *testing.T) {
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

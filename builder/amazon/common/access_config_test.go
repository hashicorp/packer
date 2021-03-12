package common

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func TestAccessConfigPrepare_Region(t *testing.T) {
	c := FakeAccessConfig()

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
}

func TestAccessConfigPrepare_RegionRestricted(t *testing.T) {
	c := FakeAccessConfig()

	// Create a Session with a custom region
	c.session = session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-gov-west-1"),
	}))

	if err := c.Prepare(); err != nil {
		t.Fatalf("shouldn't have err: %s", err)
	}

	if !c.IsGovCloud() {
		t.Fatal("We should be in gov region.")
	}
}

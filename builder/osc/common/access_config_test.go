package common

import (
	"testing"

	"github.com/outscale/osc-go/oapi"
)

type mockOAPIClient struct {
	oapi.OAPIClient
}

func testAccessConfig() *AccessConfig {
	return &AccessConfig{
		getOAPIConnection: func() oapi.OAPIClient {
			return &mockOAPIClient{}
		},
	}
}

func (m *mockOAPIClient) POST_ReadRegions(oapi.ReadRegionsRequest) (*oapi.POST_ReadRegionsResponses, error) {
	return &oapi.POST_ReadRegionsResponses{
		OK: &oapi.ReadRegionsResponse{
			Regions: []oapi.Region{
				{RegionEndpoint: "us-west1", RegionName: "us-west1"},
				{RegionEndpoint: "us-east-1", RegionName: "us-east-1"},
			},
		},
	}, nil
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

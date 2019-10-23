package client

import (
	"fmt"
	"testing"
)

func Test_platformImageRegex(t *testing.T) {
	for i, v := range []string{
		"Publisher:Offer:Sku:Versions",
		"Publisher:Offer-name:2.0_alpha:2.0.2019060122",
	} {
		t.Run(fmt.Sprintf("should_match_%d", i), func(t *testing.T) {
			if !platformImageRegex.Match([]byte(v)) {
				t.Fatalf("expected %q to match", v)
			}
		})
	}

	for i, v := range []string{
		"Publ isher:Offer:Sku:Versions",
		"Publ/isher:Offer-name:2.0_alpha:2.0.2019060122",
	} {
		t.Run(fmt.Sprintf("should_not_match_%d", i), func(t *testing.T) {
			if platformImageRegex.Match([]byte(v)) {
				t.Fatalf("did not expected %q to match", v)
			}
		})
	}
}

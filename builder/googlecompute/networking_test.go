package googlecompute

import (
	"testing"
)

func TestGetNetworking(t *testing.T) {
	cases := []struct {
		c                  *InstanceConfig
		expectedNetwork    string
		expectedSubnetwork string
		error              bool
	}{
		{
			c: &InstanceConfig{
				Network:          "default",
				Subnetwork:       "",
				NetworkProjectId: "project-id",
				Region:           "region-id",
			},
			expectedNetwork:    "global/networks/default",
			expectedSubnetwork: "",
			error:              false,
		},
		{
			c: &InstanceConfig{
				Network:          "",
				Subnetwork:       "",
				NetworkProjectId: "project-id",
				Region:           "region-id",
			},
			expectedNetwork:    "",
			expectedSubnetwork: "",
			error:              true,
		},
		{
			c: &InstanceConfig{
				Network:          "some/network/path",
				Subnetwork:       "some/subnetwork/path",
				NetworkProjectId: "project-id",
				Region:           "region-id",
			},
			expectedNetwork:    "some/network/path",
			expectedSubnetwork: "some/subnetwork/path",
			error:              false,
		},
		{
			c: &InstanceConfig{
				Network:          "network-value",
				Subnetwork:       "subnetwork-value",
				NetworkProjectId: "project-id",
				Region:           "region-id",
			},
			expectedNetwork:    "projects/project-id/global/networks/network-value",
			expectedSubnetwork: "projects/project-id/regions/region-id/subnetworks/subnetwork-value",
			error:              false,
		},
	}

	for _, tc := range cases {
		n, sn, err := getNetworking(tc.c)
		if n != tc.expectedNetwork {
			t.Errorf("Expected network %q but got network %q", tc.expectedNetwork, n)
		}
		if sn != tc.expectedSubnetwork {
			t.Errorf("Expected subnetwork %q but got subnetwork %q", tc.expectedSubnetwork, sn)
		}
		if !tc.error && err != nil {
			t.Errorf("Did not expect an error but got: %v", err)
		}
	}
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: BUSL-1.1

package api

import (
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/hashicorp/hcp-sdk-go/clients/cloud-resource-manager/stable/2019-12-10/models"
)

func TestGetOldestProject(t *testing.T) {
	testcases := []struct {
		Name            string
		ProjectList     []*models.HashicorpCloudResourcemanagerProject
		ExpectProjectID string
		ExpectErr       bool
	}{
		{
			"Only one project, project exists, success",
			[]*models.HashicorpCloudResourcemanagerProject{
				{
					ID: "test-project-exists",
				},
			},
			"test-project-exists",
			false,
		},
		{
			"Multiple projects, pick the oldest",
			[]*models.HashicorpCloudResourcemanagerProject{
				{
					ID:        "test-project-exists",
					CreatedAt: strfmt.DateTime(time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "test-oldest-project",
					CreatedAt: strfmt.DateTime(time.Date(2022, 1, 1, 1, 0, 0, 0, time.UTC)),
				},
			},
			"test-oldest-project",
			false,
		},
		{
			"Multiple projects, different order, pick the oldest",
			[]*models.HashicorpCloudResourcemanagerProject{
				{
					ID:        "test-oldest-project",
					CreatedAt: strfmt.DateTime(time.Date(2022, 1, 1, 1, 0, 0, 0, time.UTC)),
				},
				{
					ID:        "test-project-exists",
					CreatedAt: strfmt.DateTime(time.Date(2023, 1, 1, 1, 0, 0, 0, time.UTC)),
				},
			},
			"test-oldest-project",
			false,
		},
		{
			"No projects, should error",
			[]*models.HashicorpCloudResourcemanagerProject{},
			"",
			true,
		},
	}

	for _, tt := range testcases {
		t.Run(tt.Name, func(t *testing.T) {
			proj, err := getOldestProject(tt.ProjectList)
			if (err != nil) != tt.ExpectErr {
				t.Errorf("test findProjectByID, expected %t, got %t",
					tt.ExpectErr,
					err != nil)
			}

			if proj != nil && proj.ID != tt.ExpectProjectID {
				t.Errorf("expected to select project %q, got %q", tt.ExpectProjectID, proj.ID)
			}
		})
	}
}

// Copyright (c) 2017 Oracle America, Inc.
// The contents of this file are subject to the Mozilla Public License Version
// 2.0 (the "License"); you may not use this file except in compliance with the
// License. If a copy of the MPL was not distributed with this file, You can
// obtain one at http://mozilla.org/MPL/2.0/

package bmcs

// ComputeClient is a client for the BMCS Compute API.
type ComputeClient struct {
	BaseURL         string
	Instances       *InstanceService
	Images          *ImageService
	VNICAttachments *VNICAttachmentService
	VNICs           *VNICService
}

// NewComputeClient creates a new client for communicating with the BMCS
// Compute API.
func NewComputeClient(s *baseClient) *ComputeClient {
	return &ComputeClient{
		Instances:       NewInstanceService(s),
		Images:          NewImageService(s),
		VNICAttachments: NewVNICAttachmentService(s),
		VNICs:           NewVNICService(s),
	}
}

// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package driver_restapi

import "fmt"

// This is the common builder ID to all of these artifacts.
const BuilderId = "MSOpenTech.azure"

// Artifact is the result of running the azure builder.
type artifact struct {
	imageLabel string
	imageName string
	mediaLocation string
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (a *artifact) Files() []string {
	return nil
}

func (a *artifact) Id() string {
	return a.imageName
}

func (a *artifact) State(name string) interface{} {
	return "Not implemented"	
}

func (a *artifact) String() string {
	return fmt.Sprintf("{%s,%s,%s}",
		fmt.Sprintf("imageLabel: '%s'", a.imageLabel),
		fmt.Sprintf("imageName: '%s'", a.imageName),
		fmt.Sprintf("mediaLocation: '%s'", a.mediaLocation),
		)
}

func (a *artifact) Destroy() error {

	// TODO: remove image and vhd
	return nil
}

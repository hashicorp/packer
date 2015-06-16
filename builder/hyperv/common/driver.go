// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

// A driver is able to talk to HyperV and perform certain
// operations with it. Some of the operations on here may seem overly
// specific, but they were built specifically in mind to handle features
// of the HyperV builder for Packer, and to abstract differences in
// versions out of the builder steps, so sometimes the methods are
// extremely specific.
type Driver interface {

	// HypervManage executes the given HypervManage command
	HypervManage(string) error


	// Verify checks to make sure that this driver should function
	// properly. If there is any indication the driver can't function,
	// this will return an error.
	Verify() error

}



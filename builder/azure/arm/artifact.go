// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See the LICENSE file in builder/azure for license information.

package arm

const (
	BuilderId = "Azure.ResourceManagement.VMImage"
)

type artifact struct {
	name string
}

func (*artifact) BuilderId() string {
	return BuilderId
}

func (*artifact) Files() []string {
	return []string{}
}

func (*artifact) Id() string {
	return ""
}

func (*artifact) State(name string) interface{} {
	return nil
}

func (*artifact) String() string {
	return "{}"
}

func (*artifact) Destroy() error {
	return nil
}

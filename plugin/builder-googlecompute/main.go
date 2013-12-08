// Copyright (c) 2013 Kelsey Hightower. All rights reserved.
// Use of this source code is governed by the Apache License, Version 2.0
// that can be found in the LICENSE file.

package main

import (
	"github.com/kelseyhightower/packer-builder-googlecompute/builder/googlecompute"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	plugin.ServeBuilder(new(googlecompute.Builder))
}

/*
The provisioneracc package creates a framework for provisioner acceptance
testing. For builder acceptance testing, use the top level tooling in the
acctest package.
*/

package provisioneracc

import (
	"github.com/hashicorp/packer/packer-plugin-sdk/acctest/testutils"
)

// Variables stored in this file represent implementations of the BuilderFixture
// struct inside of provisioners.go

// AmasonEBSBuilderFixtureLinux points to a build stub of a simple amazon-ebs
// build running on a linux operating system.
var AmasonEBSBuilderFixtureLinux = &BuilderFixture{
	Name:         "Amazon-ebs Linux builder",
	TemplatePath: "amazon-ebs/amazon-ebs.txt",
	GuestOS:      "linux",
	HostOS:       "any",
	Teardown: func() error {
		// TODO
		// helper := AWSHelper{
		// 	Region:  "us-east-1",
		// 	AMIName: "packer-acc-test",
		// }
		// return helper.CleanUpAmi()
		return nil
	},
}

// AmasonEBSBuilderFixtureWindows points to a build stub of a simple amazon-ebs
// build running on a Windows operating system.
var AmasonEBSBuilderFixtureWindows = &BuilderFixture{
	Name:         "Amazon-ebs Windows builder",
	TemplatePath: "amazon-ebs/amazon-ebs_windows.txt",
	GuestOS:      "windows",
	HostOS:       "any",
	Teardown: func() error {
		// TODO
		// helper := AWSHelper{
		// 	Region:  "us-east-1",
		// 	AMIName: "packer-acc-test",
		// }
		// return helper.CleanUpAmi()
		return nil
	},
}

// VirtualboxBuilderFixtureLinux points to a build stub of a simple amazon-ebs
// build running on a linux operating system.
var VirtualboxBuilderFixtureLinux = &BuilderFixture{
	Name:         "Virtualbox Windows builder",
	TemplatePath: "virtualbox/virtualbox-iso.txt",
	GuestOS:      "linux",
	HostOS:       "any",
	Teardown: func() error {
		testutils.CleanupFiles("virtualbox-iso-packer-acc-test")
		testutils.CleanupFiles("packer_cache")
		return nil
	},
}

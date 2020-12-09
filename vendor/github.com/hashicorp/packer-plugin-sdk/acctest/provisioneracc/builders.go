package provisioneracc

import (
	"github.com/hashicorp/packer-plugin-sdk/acctest/testutils"
)

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

var VirtualboxBuilderFixtureWindows = &BuilderFixture{
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

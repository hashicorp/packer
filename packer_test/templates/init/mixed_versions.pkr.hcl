packer {
	required_plugins {
		tester = {
			source = "github.com/mondoohq/cnspec"
			version = "= 10.7.3" # plugin describe reports 10.7.x-dev so init must fail
		}
	}
}


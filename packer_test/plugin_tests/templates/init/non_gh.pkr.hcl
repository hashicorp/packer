packer {
	required_plugins {
		tester = {
			source = "hubgit.com/hashicorp/tester"
			version = ">= 1.0.9"
		}
	}
}

source "tester-dynamic" "test" {}

build {
	sources = ["tester-dynamic.test"]
}

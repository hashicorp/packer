packer {
	required_plugins {
		tester = {
			source = "github.com/hashicorp/tester"
			version = "= 1.0.9"
		}
	}
}

source "tester-dynamic" "test" {}

build {
	sources = ["tester-dynamic.test"]
}

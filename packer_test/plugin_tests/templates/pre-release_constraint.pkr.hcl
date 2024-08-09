packer {
	required_plugins {
		tester = {
			source = "github.com/hashicorp/hashicups"
			version = "= 1.0.2-dev"
		}
	}
}

source "tester-dynamic" "test" {}

build {
	sources = ["tester-dynamic.test"]
}

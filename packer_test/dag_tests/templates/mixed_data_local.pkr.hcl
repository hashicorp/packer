data "null" "head" {
	input = "foo"
}

locals {
	loc = "${data.null.head.output}"
}

data "null" "tail" {
	input = "${local.loc}"
}

locals {
	last = "final - ${data.null.tail.output}"
}

source "null" "test" {
	communicator = "none"
}

build {
	sources = ["null.test"]
}

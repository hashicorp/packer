data "null" "dep" {
	input = "upload"
}

source "null" "test" {
	communicator = "none"
}

build {
	sources = ["sources.null.test"]

	provisioner "file" {
		source = "test-fixtures/hcl/force.pkr.hcl"
		destination = "dest"
		direction = "${data.null.dep.output}"
	}
}

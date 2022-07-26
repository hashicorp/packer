source file "test" {
	content = " "
	target = "output"
}

build {
	name = "bucket-slug"

	description = "Some build description"

	hcp_packer_registry {}

	sources = [
		"sources.file.test",
	]
}


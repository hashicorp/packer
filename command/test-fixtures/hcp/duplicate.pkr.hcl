source file "test" {
	content = " "
	target = "output"
}

build {
	description = "Some build description"

	hcp_packer_registry {
		bucket_name = "bucket-slug"
		description = "Some override description"
	}

	hcp_packer_registry {
		bucket_name = "bucket-slug"
		description = "Some override description"
	}

	sources = [
		"sources.file.test",
	]
}

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

	sources = [
		"sources.file.test",
	]
}

build {
	description = "Some other build"

	hcp_packer_registry {
		bucket_name = "bucket-bis"
	}

	sources = [
		"sources.file.test",
	]
}

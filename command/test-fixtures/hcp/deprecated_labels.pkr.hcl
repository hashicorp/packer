source file "test" {
	content = " "
	target = "output"
}

build {
	description = "Some build description"

	hcp_packer_registry {
		bucket_name = "bucket-slug"
		labels = {
			"foo" = "bar"
		}
	}

	sources = [
		"sources.file.test",
	]
}

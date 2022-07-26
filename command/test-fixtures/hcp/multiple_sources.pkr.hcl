source file "test" {
	content = " "
	target = "output"
}

source file "other" {
	content = "b"
	target = "output 2"
}

build {
	name = "bucket-slug"

	hcp_packer_registry {
		description = <<EOT
Some description
		EOT
		bucket_labels = {
		    "foo" = "bar"
		}
		build_labels = {
		    "python_version" = "3.0"
		}
	}

	sources = [
		"sources.file.test",
		"sources.file.other"
	]
}

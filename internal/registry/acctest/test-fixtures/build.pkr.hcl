
variable "bkt" {
    default = "pkr-acctest-temp-2"
}

source "null" "example" {
    communicator = "none"
}

build {
    hcp_packer_registry {
        bucket_name = "pkr-acctest-temp-2"
        description = "blah"

        build_labels = {
            "foo-version" = "3.4.0"
            "foo"         = "bar"
        }
    }

	sources = [
		"source.null.example"
	]
	provisioner "shell-local" {
		inline = ["echo test"]
	}
}

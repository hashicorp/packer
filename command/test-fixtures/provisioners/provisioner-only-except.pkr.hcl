source "null" "packer" {
	communicator = "none"
}

source "null" "other" {
	communicator = "none"
}

build {
	sources = ["sources.null.packer", "null.other"]

	provisioner "shell-local" {
		inline = ["echo packer provisioner {{build_name}} and {{build_type}}"]
		only   = ["null.packer"]
	}

	provisioner "shell-local" {
		inline = ["echo other provisioner {{build_name}} and {{build_type}}"]
		except = ["null.packer"]
	}

	post-processor "shell-local" {
		inline = ["echo packer post-processor {{build_name}} and {{build_type}}"]
		only   = ["null.packer"]
	}

	post-processor "shell-local" {
		inline = ["echo other post-processor {{build_name}} and {{build_type}}"]
		except = ["null.packer"]
	}
}

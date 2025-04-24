source "docker" "test" {
	image = "debian:bookworm"
	discard = true
}

build {
	sources = ["docker.test"]

	provisioner "shell" {
		inline_shebang = "/bin/bash -ex"
		inline = [
			"#!/bin/sh",
			"cat \"$0\" | head -1 | grep -qE '^#!/bin/bash'",
			"cat \"$0\""
		]
	}
}

source "docker" "test" {
	image = "debian:bookworm"
	discard = true
}

build {
	sources = ["docker.test"]

	provisioner "shell" {
		inline_shebang = "/bin/bash -ex"
		inline = [
			"head -1 <\"$0\" | grep -qE '^#!/bin/bash'",
			"if grep -qE \"^#!/bin/sh\" <\"$0\"; then exit 1; fi",
			"cat \"$0\""
		]
	}
}

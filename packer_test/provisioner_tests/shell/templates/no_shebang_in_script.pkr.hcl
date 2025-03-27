source "docker" "test" {
	image = "debian:bookworm"
	discard = true
}

build {
	sources = ["docker.test"]

	provisioner "shell" {
		inline = [
			"cat $0 | head -1 | grep -E '^#!/bin/sh -e'",
		]
	}
}

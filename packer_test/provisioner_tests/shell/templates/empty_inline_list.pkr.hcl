# Copyright IBM Corp. 2024, 2026
# SPDX-License-Identifier: BUSL-1.1

source "docker" "test" {
	image = "debian:bookworm"
	discard = true
}

build {
	sources = ["docker.test"]

	provisioner "shell" {
		inline = []
	}
}

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

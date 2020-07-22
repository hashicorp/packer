source "null" "example" {
    communicator = "none"
}

build {
	sources = [
		"source.null.example"
	]
	provisioner "shell-local" {
		script = "./${path.root}/test_cmd.cmd"
		environment_vars = ["USER=packeruser", "BUILDER=${upper(build.ID)}"]
	}
}

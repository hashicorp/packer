source "null" "example" {
    communicator = "none"
}

build {
	sources = [
		"source.null.example"
	]
	provisioner "shell-local" {
		script = "./test-fixtures/hcl/reprepare/hello.sh"
		environment_vars = ["USER=packeruser", "BUILDER=${upper(build.ID)}"]
	}
}

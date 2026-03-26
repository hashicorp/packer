variable "secret_multiline" {
	type      = string
	sensitive = true
	default   = "line-one-secret\nline-two-secret\nline-three-secret"
}

source "null" "example" {
	communicator = "none"
}

build {
	sources = ["sources.null.example"]

	provisioner "shell-local" {
		inline = [
			"printf 'BEGIN\n%s\nEND\n' '${var.secret_multiline}'"
		]
	}
}
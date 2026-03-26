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
		tempfile_extension = ".ps1"
		environment_vars   = ["SECRET_MULTILINE=${var.secret_multiline}"]
		execute_command    = ["powershell.exe", "{{.Vars}} {{.Script}}"]
		env_var_format     = "$env:%s=\"%s\"; "
		inline = [
			"Write-Output 'BEGIN'",
			"Write-Output $env:SECRET_MULTILINE",
			"Write-Output 'END'"
		]
	}
}
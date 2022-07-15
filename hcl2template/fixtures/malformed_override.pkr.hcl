provisioner "shell-local" {
  inline = ["echo 'hi'"]
  override = "hello"
}
